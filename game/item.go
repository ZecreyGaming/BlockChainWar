package game

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"time"

	"github.com/solarlune/resolv"
)

type ItemType uint8

const (
	itemPixelR = 15

	ItemAccelerator ItemType = iota
	ItemTag                  = "ITEM"

	AcceleratorTag = "Accelerator"

	KindsOfItems = 1
)

var (
	ItemTagMap = map[ItemType]string{
		ItemAccelerator: AcceleratorTag,
	}

	ItemTagMapReverse = map[string]ItemType{
		AcceleratorTag: ItemAccelerator,
	}

	Accelerator = Item{
		Type:      ItemAccelerator,
		Name:      AcceleratorTag,
		Thumbnail: "https://res.cloudinary.com/zecrey/image/upload/v1665310283/1054990_rocket_spacecraft_spaceship_icon_gtia85.jpg",
	}
	ItemMap = map[ItemType]Item{
		ItemAccelerator: Accelerator,
	}

	AllItems = []Item{
		Accelerator,
	}
)

type Item struct {
	Type      ItemType `json:"type"`
	Name      string   `json:"name"`
	Thumbnail string   `json:"thumbnail"`
}

type ItemObject struct {
	Id   uint32
	X    float64
	Y    float64
	Item Item
}

func (p *ItemObject) Serialize() []byte {
	bytesBuffer := bytes.NewBuffer(make([]byte, 0))
	binary.Write(bytesBuffer, binary.BigEndian, p.Id)
	binary.Write(bytesBuffer, binary.BigEndian, uint8(p.Item.Type))
	x, y := space2MapXY(p.Center())
	binary.Write(bytesBuffer, binary.BigEndian, x)
	binary.Write(bytesBuffer, binary.BigEndian, y)
	return bytesBuffer.Bytes()
}

func (i *ItemObject) Center() (float64, float64) {
	return i.X + float64(itemPixelR), i.Y + float64(itemPixelR)
}

func itemIdToTag(Id uint32) string {
	return fmt.Sprintf("%s_%d", ItemTag, Id)
}

func itemTagsToId(tags []string) uint32 {
	var id uint32

	for _, tag := range tags {
		if tag[:len(ItemTag)] == ItemTag {
			fmt.Sscanf(tag, "%s_%d", &id)
		}
	}
	return id
}

func (g *Game) TryAddItem() {
	if g.GameStatus != GameRunning || rand.Intn(g.cfg.ItemFrameChance) != 1 {
		return
	}
	x, y := g.Map.RandomSpaceXY()
	g.space.Add(resolv.NewObject(x, y, float64(2*itemPixelR), float64(2*itemPixelR), ItemTag, ItemTagMap[ItemAccelerator]))
	item := &ItemObject{
		Id:   uint32(time.Now().UnixMilli()),
		X:    x,
		Y:    y,
		Item: ItemMap[ItemAccelerator],
	}

	g.Items.LoadOrStore(item.Id, item)
}
