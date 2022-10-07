package game

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type ItemType uint8

const (
	itemPixelR = 5

	Accelerator ItemType = iota
	ItemTag              = "ITEM"

	AcceleratorTag = "Accelerator"

	KindsOfItems = 1
)

var (
	ItemTagMap = map[ItemType]string{
		Accelerator: AcceleratorTag,
	}

	ItemTagMapReverse = map[string]ItemType{
		AcceleratorTag: Accelerator,
	}

	AllItem = map[ItemType]Item{
		Accelerator: Item{
			Type:      Accelerator,
			Name:      AcceleratorTag,
			Thumbnail: "https://i.imgur.com/8ZQ2Z9M.png",
		},
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
	binary.Write(bytesBuffer, binary.BigEndian, p.X)
	binary.Write(bytesBuffer, binary.BigEndian, p.Y)
	return bytesBuffer.Bytes()
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
