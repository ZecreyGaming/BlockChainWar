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
		AVAX:  {6, 6},
		BNB:   {10, 8},
		MATIC: {8, 7},
		BTC:   {18, 10},
		ETH:   {14, 9},
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
	retval = append(retval, EdgeTag, ItemTag)
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
		if (x == cx && y == cy) || (x-1 == cx && y == cy) || (x+1 == cx && y == cy) || (x == cx && y-1 == cy) || (x == cx && y+1 == cy) {
			camp = c
			break
		}
	}
	return camp
}

func (c Camp) CenterCellIndex(row, col int) (int, int) {
	switch c {
	case ETH:
		return 35, 25
	case BNB:
		return 20, 4
	case AVAX:
		return 4, 4
	case MATIC:
		return 35, 4
	case BTC:
		return 4, 25
	default:
		return col / 5, row / 5
	}
}

func DecideCamp(msg string) Camp {
	for _, tag := range CampTagMap {
		if strings.Contains(strings.ToUpper(msg), tag) {
			return CampTagMapReverse[tag]
		}
	}
	return Empty
}
