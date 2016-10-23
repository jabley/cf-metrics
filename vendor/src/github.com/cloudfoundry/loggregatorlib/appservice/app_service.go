package appservice

import (
	"crypto/sha1"
	"fmt"
)

type AppService struct {
	AppId string
	Url   string
}

func (a AppService) Id() string {
	hash := sha1.Sum([]byte(a.Url))
	return fmt.Sprintf("%x", hash)
}
