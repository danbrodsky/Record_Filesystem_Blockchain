package blockmap

import(
    "testing"
    "fmt"
)

func TestMining(t *testing.T) {
    block := Block{ PrevHash: "1234", Nonce:"1" , MinerId:"james"}
    minedblock := ComputeBlock(block, 2)
    fmt.Println(minedblock)
}



