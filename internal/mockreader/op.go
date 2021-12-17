package mockreader

import (
	"fmt"

	"github.com/chronos-tachyon/enumhelper"
)

type Op uint

const (
	UnknownOp Op = iota
	MarkOp
	StatOp
	ReadOp
	ReadAtOp
	SeekOp
	CloseOp
)

var mockReaderOpData = []enumhelper.EnumData{
	{GoName: "UnknownOp", Name: "<unknown>"},
	{GoName: "MarkOp", Name: "Mark"},
	{GoName: "StatOp", Name: "Stat"},
	{GoName: "ReadOp", Name: "Read"},
	{GoName: "ReadAtOp", Name: "ReadAt"},
	{GoName: "SeekOp", Name: "Seek"},
	{GoName: "CloseOp", Name: "Close"},
}

func (op Op) GoString() string {
	return enumhelper.DereferenceEnumData("Op", mockReaderOpData, uint(op)).GoName
}

func (op Op) String() string {
	return enumhelper.DereferenceEnumData("Op", mockReaderOpData, uint(op)).Name
}

var (
	_ fmt.GoStringer = Op(0)
	_ fmt.Stringer   = Op(0)
)
