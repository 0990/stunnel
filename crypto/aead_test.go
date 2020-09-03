package crypto

import (
	"fmt"
	"github.com/0990/stunnel/util"
	"github.com/sirupsen/logrus"
	"testing"
)

func Test_Aead(t *testing.T) {
	aead, err := util.CreateAesGcmAead(util.StringToAesKey("abcd", 32))
	if err != nil {
		logrus.Fatalln(err)
	}
	a := make([]byte, 1, 2)
	a[0] = 'a'
	fmt.Println(len(a[:0]), cap(a[:0]))
	//a := []byte("abc")
	nonce := make([]byte, aead.NonceSize())
	ret := aead.Seal(a[:0], nonce, a, nil)
	fmt.Println(ret, a)
}
