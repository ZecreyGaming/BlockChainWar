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
	c, err := GetClient(name, seedKey, nftPrefix, 6)
	if err != nil {
		t.Fatal(err)
	}
	eddsaSig, err := c.SignMessage(rawMessage)
	if err != nil {
		t.Fatal(err)
	}
	b, err := VerifyMessage(c.l2pk, eddsaSig, rawMessage)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, true, b)
}
