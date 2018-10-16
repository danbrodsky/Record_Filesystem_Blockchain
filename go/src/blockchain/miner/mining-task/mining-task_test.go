package miningtask

import(
    "testing"
    "fmt"
)

func TestMining(t *testing.T) {
    nonce, hash := computeNonceSecretHash(4)
    fmt.Println(nonce,hash)
}

