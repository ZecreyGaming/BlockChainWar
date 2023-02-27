package zecreyface

import (
	"github.com/bmizerany/assert"
	"testing"
)

func Test_client_SignMessage(t *testing.T) {
	name := "alice"
	seedKey := "alice seed...."
	rawMessage := "eth"
	nftPrefix := "your companyName"
	c, err := Client(name, seedKey, nftPrefix)
	if err != nil {
		t.Fatal(err)
	}
	eddsaSig, err := c.SignMessage(rawMessage)
	if err != nil {
		t.Fatal(err)
	}
	b, err := c.VerifyMessage(eddsaSig, rawMessage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, b)
}
