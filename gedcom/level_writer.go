package gedcom

import (
	"strconv"
	"strings"
)

type Block struct {
	Buffer strings.Builder
	Level  int
}

func NewBlock() *Block {
	return &Block{}
}

func (b *Block) SetLevel(level int) *Block {
	b.Level = level
	return b
}

func (b *Block) AddRaw(text string) *Block {
	b.Buffer.WriteString(text)
	return b
}

func (b *Block) AddRawNewLine(text string) *Block {
	b.Buffer.WriteString(text + "\n")
	return b
}

func (b *Block) Add(text string) *Block {
	b.Buffer.WriteString(strconv.Itoa(b.Level) + " ")
	b.AddRawNewLine(text)
	return b
}

func (b *Block) Down() *Block {
	b.Level++
	return b
}

func (b *Block) Up() *Block {
	b.Level--
	return b
}

func (b *Block) Reset() *Block {
	b.Level = 0
	return b
}

func (b *Block) AddDown(text string) *Block {
	b.Down()
	b.Add(text)
	return b
}

func (b *Block) String() string {
	return b.Buffer.String()
}

func GetLeveledString(level int, input string) string {
	return strconv.Itoa(level) + " " + input
}
