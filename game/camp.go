package game

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

func (c Camp) Center(row, col int) (x int, y int) {
	x, y = row/2, col/2
	switch c {
	case ETH:
		x, y = ETHInitialLen/2, ETHInitialLen/2
	case BNB:
		x, y = col-BNBInitialLen/2, BNBInitialLen/2
	case AVAX:
		x, y = AVAXInitialLen/2, row-AVAXInitialLen/2
	case MATIC:
		x, y = col-MATICInitialLen/2, row-MATICInitialLen/2
	default:
	case BTC:
		x, y = row/2, col/2
	}
	return
}
