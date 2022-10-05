package model

import "encoding/binary"

type Map struct {
	Row    uint32 `json:"row"`
	Column uint32 `json:"column"`

	CellWidth  uint32 `json:"cell_width"`
	CellHeight uint32 `json:"cell_height"`
	LineWidth  uint32 `json:"line_width"`

	Cells []Camp `json:"cells"`
}

func (m *Map) Serialize() []byte {
	l := 20 + len(m.Cells)*sizeOfCellStateBits/8
	res := make([]byte, l)
	binary.LittleEndian.PutUint32(res[0:4], m.Row)
	binary.LittleEndian.PutUint32(res[4:8], m.Column)
	binary.LittleEndian.PutUint32(res[8:12], m.CellWidth)
	binary.LittleEndian.PutUint32(res[12:16], m.CellHeight)
	binary.LittleEndian.PutUint32(res[16:20], m.LineWidth)
	offset := 20
	for i := 0; i < len(m.Cells); i += 2 {
		n := byte(m.Cells[i]<<4) & campMaskLeft
		if i+1 < len(m.Cells) {
			n = n | (byte(m.Cells[i+1]) & campMaskRight)
		}
		res[offset] = n
		offset++
	}
	return res
}

func (m *Map) Size() uint32 {
	return uint32(20 + len(m.Cells)*sizeOfCellStateBits/8)
}
