package blockmap

import(
    "testing"
    "fmt"
)

func TestInsert(t *testing.T) {
    blockCh := make(chan *Block,5)
    genesisBlock := Block{ PrevHash: "GENESIS", Nonce:"1" , MinerId:"james"}
    bm := NewBlockMap(genesisBlock, "3H4K12G453F9PRA5WD0NR567")
    PrepareMining()
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    fmt.Println("block map",bm.GetMap())
    fmt.Println("longest chain",bm.GetLongestChain())
    fmt.Println("block end with 3 zeros:",  BHashEndsWithZeros(bm.GetLongestChain()[1],3))
    fmt.Println("block end with 4 zeros:",  BHashEndsWithZeros(bm.GetLongestChain()[1],4))
    fmt.Println("block end with 5 zeros:",  BHashEndsWithZeros(bm.GetLongestChain()[1],5))
    fmt.Println("block end with 6 zeros:",  BHashEndsWithZeros(bm.GetLongestChain()[1],6))
}


func TestStop(t *testing.T) {
    blockCh := make(chan *Block,5)
    genesisBlock := Block{ PrevHash: "GENESIS", Nonce:"1" , MinerId:"james"}
    bm := NewBlockMap(genesisBlock, "3H4K12G453F9PRA5WD0NR567")
    PrepareMining()
    go bm.MineAndAddBlock(nil,"james",blockCh)
    StopMining()
    <-blockCh
    PrepareMining()
    go bm.MineAndAddBlock(nil,"james",blockCh)
    StopMining()
    <-blockCh
    PrepareMining()
    go bm.MineAndAddBlock(nil,"james",blockCh)
    StopMining()
    <-blockCh
    fmt.Println("block map",bm.GetMap())
    fmt.Println("longest chain",bm.GetLongestChain())
}

