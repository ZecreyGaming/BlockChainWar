package game

import (
	"strings"
)

const (
	sizeOfCellStateBits = 4
	campMaskLeft        = byte(0xF0)
	campMaskRight       = byte(0x0F)
)

type Camp uint8 // should convert to int4 when transfered to client

const (
	Empty Camp = iota
	BTC
	ETH
	BNB
	AVAX
	MATIC

	EmptyTag = "Empty"
	BTCTag   = "BTC"
	ETHTag   = "ETH"
	BNBTag   = "BNB"
	AVAXTag  = "AVAX"
	MATICTag = "MATIC"
)

var (
	CampTagMap = map[Camp]string{
		Empty: EmptyTag,
		BTC:   BTCTag,
		ETH:   ETHTag,
		BNB:   BNBTag,
		AVAX:  AVAXTag,
		MATIC: MATICTag,
	}

	CampTagMapReverse = map[string]Camp{
		EmptyTag: Empty,
		BTCTag:   BTC,
		ETHTag:   ETH,
		BNBTag:   BNB,
		AVAXTag:  AVAX,
		MATICTag: MATIC,
	}

	CampSizeMap = map[Camp][2]int{
		Empty: {0, 0},
		AVAX:  {10, 10},
		BNB:   {14, 10},
		MATIC: {10, 10},
		BTC:   {16, 10},
		ETH:   {17, 10},
	}
)

func getCollisionTags(camp Camp) (retval []string) {
	switch camp {
	case BTC:
		retval = []string{CampTagMap[ETH], CampTagMap[BNB], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	case ETH:
		retval = []string{CampTagMap[BNB], CampTagMap[BTC], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	case BNB:
		retval = []string{CampTagMap[ETH], CampTagMap[BTC], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	case AVAX:
		retval = []string{CampTagMap[ETH], CampTagMap[BNB], CampTagMap[BTC], CampTagMap[MATIC], CampTagMap[Empty]}
	case MATIC:
		retval = []string{CampTagMap[ETH], CampTagMap[BNB], CampTagMap[BTC], CampTagMap[AVAX], CampTagMap[Empty]}
	default:
		retval = []string{CampTagMap[BTC], CampTagMap[ETH], CampTagMap[BNB], CampTagMap[AVAX], CampTagMap[MATIC], CampTagMap[Empty]}
	}
	retval = append(retval, HorizontalEdgeTag, VerticalEdgeTag, EdgeTag)
	return
}

func removeCampTags(tags []string) []string {
	ret := []string{}
	for _, tag := range tags {
		if _, ok := CampTagMapReverse[tag]; ok {
			ret = append(ret, tag)
		}
	}
	return ret
}

func initCamp(x, y int) Camp {
	camp := Empty
	for c := range CampTagMap {
		if c == Empty {
			continue
		}
		cx, cy := c.CenterCellIndex(mapRow, mapColumn)
		if y >= cy-CampSizeMap[c][1]/2 && y < cy+CampSizeMap[c][1]/2 && x >= cx-CampSizeMap[c][0]/2 && x < cx+CampSizeMap[c][0]/2 {
			camp = c
			break
		}
	}
	return camp
}

func (c Camp) CenterCellIndex(row, col int) (int, int) {
	switch c {
	case ETH:
		return col - CampSizeMap[ETH][0]/2, row - CampSizeMap[ETH][1]/2
	case BNB:
		return col / 2, CampSizeMap[BNB][1] / 2
	case AVAX:
		return CampSizeMap[AVAX][0] / 2, CampSizeMap[AVAX][1] / 2
	case MATIC:
		return col - CampSizeMap[MATIC][0]/2, CampSizeMap[MATIC][1] / 2
	case BTC:
		return CampSizeMap[BTC][0] / 2, row - CampSizeMap[BTC][1]/2
	default:
		return col / 5, row / 5
	}
}

func DecideCamp(msg string) Camp {
	for _, tag := range CampTagMap {
		if strings.Contains(msg, tag) {
			return CampTagMapReverse[tag]
		}
	}
	return Empty
}
