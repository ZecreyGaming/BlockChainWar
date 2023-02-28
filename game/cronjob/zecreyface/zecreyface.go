package zecreyface

import (
	"fmt"
	zecreyface "github.com/Zecrey-Labs/zecrey-marketplace-go-sdk/sdk"
)

type Client struct {
	z            *zecreyface.Client
	nftPrefix    string
	acccountName string
	seed         string
	l2pk         string
	collectionId int64
}

func GetClient(accountName, seed, nftPrefix string, collectionId int64) (*Client, error) {
	z, err := zecreyface.NewClient(accountName, seed)
	if err != nil {
		return nil, err
	}
	_, l2pk, seed := z.GetMyInfo()
	//Id, err := zecreyface.GetDefaultCollectionId(accountName)
	//if err != nil {
	//	return nil, err
	//}
	return &Client{
		z:            z,
		acccountName: accountName,
		seed:         seed,
		l2pk:         l2pk,
		collectionId: collectionId,
		nftPrefix:    nftPrefix}, nil
}

func (c *Client) MintNft(collectionId int64, toAccountName string, nftName string, nftDescription string) (*zecreyface.RespCreateAsset, error) {
	result, err := zecreyface.UploadMedia("./game/media/MedalOfVictory.png")
	if err != nil {
		return nil, err
	}
	nftInfo, err := c.z.MintNft(collectionId, toAccountName,
		fmt.Sprintf("https://res.cloudinary.com/zecrey/image/upload/%s", result.PublicId), nftName,
		nftDescription, result.PublicId,
		"[]", "[]", "[]")
	if err != nil {
		return nil, err
	}
	return nftInfo, nil
}

func (c *Client) SignMessage(message string) (string, error) {
	_, _, seed := c.z.GetMyInfo()
	return zecreyface.SignMessage(seed, message)
}

func VerifyMessage(l2publicKey, eddsaSig, rawMessage string) (bool, error) {
	return zecreyface.VerifyMessage(l2publicKey, eddsaSig, rawMessage)
}

func (c *Client) GetAccountWinNfts(collectionId int64, accountName string) ([]*zecreyface.HauaraNftInfo, error) {
	result, err := zecreyface.GetCollectionAccountNftsByIregex(collectionId, accountName, c.nftPrefix)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *Client) GetCollectionWinNfts(collectionId int64) ([]*zecreyface.HauaraNftInfo, error) {
	result, err := zecreyface.GetCollectionNftsByIregex(collectionId, c.nftPrefix)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func GetAccountInfo(accountName string) (*zecreyface.RespGetAccountByAccountName, error) {
	return zecreyface.GetAccountByAccountName(accountName)
}
