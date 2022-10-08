package game

import "math/rand"

const (
	mapRow     = 30
	mapColumn  = 40
	cellWidth  = 20
	cellHeight = 20
	lineWidth  = 1
)

type Map struct {
	Cells []Camp `json:"cells"`
}

func NewMap() Map {
	return Map{
		Cells: []Camp{},
	}
}

func (m *Map) W() float64 {
	return float64(mapColumn * (cellWidth + lineWidth))
}

func (m *Map) H() float64 {
	return float64(mapRow * (cellHeight + lineWidth))
}

func (m *Map) Serialize() []byte {
	l := len(m.Cells) * sizeOfCellStateBits / 8
	res := make([]byte, l)
	offset := 0
	for i := 0; i < len(m.Cells); i += 2 {
		n := byte(m.Cells[i]<<4) & campMaskLeft
		if i+1 < len(m.Cells) {
			n = n | (byte(m.Cells[i+1]) & campMaskRight)
		}
		res[offset] = n
		offset++
	}
	// fmt.Println(hex.EncodeToString(res))
	return res
}

func (m *Map) Size() uint32 {
	return uint32(len(m.Cells) * sizeOfCellStateBits / 8)
}

func (m *Map) OutofMap(x, y float64) bool {
	return x < 0 || x > m.W() || y < 0 || y > m.H()
}

func (m *Map) RandomSpaceXY() (float64, float64) {
	x := rand.Intn(mapColumn)*(cellWidth+lineWidth) + edgeWidth
	y := rand.Intn(mapRow)*(cellHeight+lineWidth) + edgeWidth
	return float64(x), float64(y)
}
