package guac

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"log"
)

const InternalDataOpcode = ""

var InternalOpcodeIns = []byte(fmt.Sprint(len(InternalDataOpcode), ".", InternalDataOpcode))

type InstructionReader interface {
	ReadSome() ([]byte, error)
	Available() bool
	Flush()
}

type TunnelPipe interface {
	AcquireReader() InstructionReader
	ReleaseReader()
	HasQueuedReaderThreads() bool
	AcquireWriter() io.Writer
	ReleaseWriter()
	HasQueuedWriterThreads() bool
	GetUUID() string
	ConnectionID() string
	Close() error
}

type SimpleTunnel struct {
	stream     *Stream
	readerLock CountedLock
	writerLock CountedLock
}

func NewSimpleTunnel(stream *Stream) *SimpleTunnel {
	return &SimpleTunnel{
		stream: stream,
	}
}

func (t *SimpleTunnel) AcquireReader() InstructionReader {
	t.readerLock.Lock()
	return t.stream
}

func (t *SimpleTunnel) ReleaseReader() {
	t.readerLock.Unlock()
}

func (t *SimpleTunnel) HasQueuedReaderThreads() bool {
	return t.readerLock.HasQueued()
}

func (t *SimpleTunnel) AcquireWriter() io.Writer {
	t.writerLock.Lock()
	return t.stream
}

func (t *SimpleTunnel) ReleaseWriter() {
	t.writerLock.Unlock()
}

func (t *SimpleTunnel) ConnectionID() string {
	return t.stream.ConnectionID
}

func (t *SimpleTunnel) HasQueuedWriterThreads() bool {
	return t.writerLock.HasQueued()
}

func (t *SimpleTunnel) Close() (err error) {
	return t.stream.Close()
}

func (t *SimpleTunnel) GetUUID() string {
	data := make([]byte, 32)
	_, err := io.ReadFull(rand.Reader, data)
	if err != nil {
		log.Println(err)
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(data)
}
