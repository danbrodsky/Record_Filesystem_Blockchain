package miningtask

import(
    "testing"
    "fmt"
    "blockchain/miner/blockmap"
)

func TestMining(t *testing.T) {
    block := blockmap.Block{ PrevHash: "1234", Nonce:"1" , MinerId:"james"}
    minedblock := computeBlock(block, 5)
    fmt.Println(minedblock)
}

