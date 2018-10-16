package miningtask

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"
	"encoding/hex"
)



var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456790123456790123456790123456790")

func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func computeNonceSecretHash(numZeros int) (nonce string, hash string) {
	zeros := ""
	for i:= 0; i<numZeros; i++{
		fmt.Println("adding 0")
		zeros = zeros+"0"
	}
        rand.Seed(time.Now().UnixNano())
	for{
		h := md5.New()
		nonce = randSeq(20)
	        h.Write([]byte(nonce))
		hash = hex.EncodeToString(h.Sum(nil))
		fmt.Println(hash)
		if(nonce[len(nonce)-numZeros:len(nonce)] == zeros){
			return nonce,hash
		}
	}
}

