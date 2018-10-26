package blockmap

import(
    "testing"
    "fmt"
)

func TestInsert(t *testing.T) {
    genesisBlock := Block{ PrevHash: "GENESIS", Nonce:"1" , MinerId:"james"}
    bm := NewBlockMap(genesisBlock)
    block1 := Block{ PrevHash: "1234", Nonce:"1" , MinerId:"james"}
    block2 := Block{ PrevHash: "1235", Nonce:"2" , MinerId:"james"}
    block3 := Block{ PrevHash: "126", Nonce:"3" , MinerId:"james"}
    bm.Insert(block1)
    bm.Insert(block2)
    bm.Insert(block3)
    fmt.Println("block map",bm.GetMap())
    fmt.Println("longest chain",bm.GetLongestChain())
}
