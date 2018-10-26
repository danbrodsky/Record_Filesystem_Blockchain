package miningtask

import (
	"crypto/md5"
	"fmt"
	"math/rand"
	"time"
	"encoding/hex"
	"blockchain/miner/blockmap"
)



var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456790123456790123456790123456790")

func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}

func computeBlock(block blockmap.Block, numZeros int) blockmap.Block{
	zeros := ""
	for i:= 0; i<numZeros; i++{
		zeros = zeros+"0"
	}
        rand.Seed(time.Now().UnixNano())
	for{
		block.Nonce = randSeq(15)
		hash := getHash(block)
		if(hash[len(hash)-numZeros:len(hash)] == zeros){
			fmt.Println(hash)
			return block
		}
	}
}

func getHash(block blockmap.Block) string{
     h := md5.New()
     h.Write([]byte(fmt.Sprintf("%v", block)))
     return hex.EncodeToString(h.Sum(nil))
}

