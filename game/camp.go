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

	BTCInitialLen   = 6
	ETHInitialLen   = 5
	BNBInitialLen   = 4
	AVAXInitialLen  = 3
	MATICInitialLen = 2

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

func initCamp(i, j, r, c int) Camp {
	if i >= 0 && i < ETHInitialLen && j >= 0 && j < ETHInitialLen {
		return ETH
	}

	if i >= 0 && i < BNBInitialLen && j < c && j >= c-BNBInitialLen {
		return BNB
	}

	if i >= (r-BTCInitialLen)/2 && i < (r+BTCInitialLen)/2 && j >= (c-BTCInitialLen)/2 && j < (c+BTCInitialLen)/2 {
		return BTC
	}

	if i >= r-AVAXInitialLen && i < r && j >= 0 && j < AVAXInitialLen {
		return AVAX
	}

	if i >= r-MATICInitialLen && i < r && j >= c-MATICInitialLen && j < c {
		return MATIC
	}
	return Empty
}

func (c Camp) Center(row, col int) (int, int) {
	switch c {
	case ETH:
		return col - 2, row - 2
	case BNB:
		return col / 2, 2
	case AVAX:
		return 2, 2
	case MATIC:
		return col - 2, 2
	case BTC:
		return 2, row - 2
	default:
		return col / 2, row / 2
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
