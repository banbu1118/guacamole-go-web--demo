package guac

import (
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	SocketTimeout  = 15 * time.Second
	MaxGuacMessage = 8192
)

type Stream struct {
	conn         net.Conn
	ConnectionID string
	parseStart   int
	buffer       []byte
	reset        []byte
	timeout      time.Duration
}

func NewStream(conn net.Conn, timeout time.Duration) *Stream {
	buffer := make([]byte, 0, MaxGuacMessage*3)
	return &Stream{
		conn:    conn,
		timeout: timeout,
		buffer:  buffer,
		reset:   buffer[:cap(buffer)],
	}
}

func (s *Stream) Write(data []byte) (n int, err error) {
	if err = s.conn.SetWriteDeadline(time.Now().Add(s.timeout)); err != nil {
		logrus.Error(err)
		return
	}
	return s.conn.Write(data)
}

func (s *Stream) Available() bool {
	return len(s.buffer) > 0
}

func (s *Stream) Flush() {
	copy(s.reset, s.buffer)
	s.buffer = s.reset[:len(s.buffer)]
}

func (s *Stream) ReadSome() (instruction []byte, err error) {
	if err = s.conn.SetReadDeadline(time.Now().Add(s.timeout)); err != nil {
		logrus.Error(err)
		return
	}

	var n int
	for {
		var elementLength int

		i := s.parseStart

	parseLoop:
		for i < len(s.buffer) {
			readChar := s.buffer[i]
			i++

			switch readChar {
			case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
				elementLength = elementLength*10 + int(readChar-'0')

			case '.':
				if i+elementLength >= len(s.buffer) {
					break parseLoop
				}
				terminator := s.buffer[i+elementLength]
				i += elementLength + 1

				elementLength = 0

				s.parseStart = i

				switch terminator {
				case ';':
					instruction = s.buffer[0:i]
					s.parseStart = 0
					s.buffer = s.buffer[i:]
					return
				case ',':
				default:
					err = ErrServer.NewError("Element terminator of instruction was not ';' nor ','")
					return
				}
			default:
				err = ErrServer.NewError("Non-numeric character in element length:", string(readChar))
				return
			}
		}

		if cap(s.buffer) < MaxGuacMessage {
			s.Flush()
		}

		n, err = s.conn.Read(s.buffer[len(s.buffer):cap(s.buffer)])
		if err != nil && n == 0 {
			switch e := err.(type) {
			case net.Error:
				if e.Timeout() {
					err = ErrUpstreamTimeout.NewError("Connection to guacd timed out.", err.Error())
				} else {
					err = ErrConnectionClosed.NewError("Connection to guacd is closed.", err.Error())
				}
			default:
				err = ErrServer.NewError(err.Error())
			}
			return
		}
		s.buffer = s.buffer[:len(s.buffer)+n]
	}
}

func (s *Stream) Close() error {
	return s.conn.Close()
}

func (s *Stream) Handshake(config *Config) error {
	selectArg := config.ConnectionID
	if len(selectArg) == 0 {
		selectArg = config.Protocol
	}

	_, err := s.Write(NewInstruction("select", selectArg).Byte())
	if err != nil {
		return err
	}

	args, err := s.AssertOpcode("args")
	if err != nil {
		return err
	}

	argNameS := args.Args
	argValueS := make([]string, 0, len(argNameS))
	for _, argName := range argNameS {
		value := config.Parameters[argName]
		if len(value) == 0 {
			value = ""
		}
		argValueS = append(argValueS, value)
	}

	_, err = s.Write(NewInstruction("size",
		fmt.Sprintf("%v", config.OptimalScreenWidth),
		fmt.Sprintf("%v", config.OptimalScreenHeight),
		fmt.Sprintf("%v", config.OptimalResolution)).Byte(),
	)

	if err != nil {
		return err
	}

	_, err = s.Write(NewInstruction("audio", config.AudioMimetypes...).Byte())
	if err != nil {
		return err
	}

	_, err = s.Write(NewInstruction("video", config.VideoMimetypes...).Byte())
	if err != nil {
		return err
	}

	_, err = s.Write(NewInstruction("image", config.ImageMimetypes...).Byte())
	if err != nil {
		return err
	}

	_, err = s.Write(NewInstruction("connect", argValueS...).Byte())
	if err != nil {
		return err
	}

	ready, err := s.AssertOpcode("ready")
	if err != nil {
		return err
	}

	readyArgs := ready.Args
	if len(readyArgs) == 0 {
		err = ErrServer.NewError("No connection ID received")
		return err
	}

	s.Flush()
	s.ConnectionID = readyArgs[0]

	return nil
}

func (s *Stream) AssertOpcode(opcode string) (instruction *Instruction, err error) {
	instruction, err = ReadOne(s)
	if err != nil {
		return
	}

	if len(instruction.Opcode) == 0 {
		err = ErrServer.NewError("End of stream while waiting for \"" + opcode + "\".")
		return
	}

	if instruction.Opcode != opcode {
		err = ErrServer.NewError("Expected \"" + opcode + "\" instruction but instead received \"" + instruction.Opcode + "\".")
		return
	}
	return
}
