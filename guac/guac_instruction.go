package guac

import (
	"fmt"
	"strconv"
)

type Instruction struct {
	Opcode string
	Args   []string
	cache  string
}

func NewInstruction(opcode string, args ...string) *Instruction {
	return &Instruction{
		Opcode: opcode,
		Args:   args,
	}
}

func (i *Instruction) String() string {
	if len(i.cache) > 0 {
		return i.cache
	}

	i.cache = fmt.Sprintf("%d.%s", len(i.Opcode), i.Opcode)
	for _, value := range i.Args {
		i.cache += fmt.Sprintf(",%d.%s", len(value), value)
	}
	i.cache += ";"

	return i.cache
}

func (i *Instruction) Byte() []byte {
	return []byte(i.String())
}

func Parse(data []byte) (*Instruction, error) {
	elementStart := 0

	elements := make([]string, 0, 1)
	for elementStart < len(data) {
		lengthEnd := -1
		for i := elementStart; i < len(data); i++ {
			if data[i] == '.' {
				lengthEnd = i
				break
			}
		}
		if lengthEnd == -1 {
			return nil, ErrServer.NewError("ReadSome returned incomplete instruction.")
		}

		length, e := strconv.Atoi(string(data[elementStart:lengthEnd]))
		if e != nil {
			return nil, ErrServer.NewError("ReadSome returned wrong pattern instruction.", e.Error())
		}

		elementStart = lengthEnd + 1
		element := string(data[elementStart : elementStart+length])

		elements = append(elements, element)

		elementStart += length
		terminator := data[elementStart]

		elementStart++

		if terminator == ';' {
			break
		}

	}

	return NewInstruction(elements[0], elements[1:]...), nil
}

func ReadOne(stream *Stream) (instruction *Instruction, err error) {
	var instructionBuffer []byte
	instructionBuffer, err = stream.ReadSome()
	if err != nil {
		return
	}

	return Parse(instructionBuffer)
}
