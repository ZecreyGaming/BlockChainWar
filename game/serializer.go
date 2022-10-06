package game

import (
	"encoding/json"
)

// Serializer implements the serialize.Serializer interface
type Serializer struct{}

// NewSerializer returns a new Serializer.
func NewSerializer() *Serializer {
	return &Serializer{}
}

// Marshal returns the JSON encoding of v.
func (s *Serializer) Marshal(v interface{}) ([]byte, error) {
	if g, ok := v.(*Game); ok {
		return g.Serialize()
	}
	switch v := v.(type) {
	case *Game:
		return v.Serialize()
	case *Player:
		return v.Serialize(), nil
	case []byte:
		// fmt.Println("bytes", v)
		return v, nil
	}
	return json.Marshal(v)
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
func (s *Serializer) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// GetName returns the name of the serializer.
func (s *Serializer) GetName() string {
	return "custome"
}
