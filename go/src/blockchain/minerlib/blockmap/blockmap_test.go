package blockmap

// DO NOT MODIFY THE CONFIG FILE THE TESTS WILL BREAK!!!
// DO NOT MODIFY THE CONFIG FILE THE TESTS WILL BREAK!!!
// DO NOT MODIFY THE CONFIG FILE THE TESTS WILL BREAK!!!
// DO NOT MODIFY THE CONFIG FILE THE TESTS WILL BREAK!!!

import(
    "testing"
    "blockchain/minerlib"
    "io/ioutil"
    "encoding/json"
    "log"
    "blockchain/rfslib"
    "fmt"
)

var(
    configs minerlib.Settings
)

func TestInsert(t *testing.T) {
    plan, e := ioutil.ReadFile("configs.json")
    if e == nil {
        err := json.Unmarshal(plan, &configs)
        if(err != nil){
            log.Fatal("error reading json:", err)
        }
    } else {
        log.Fatal("file error:", e)
    }
    blockCh := make(chan *Block,5)
    genesisBlock := Block{ PrevHash: "GENESIS", Nonce:"GENESIS" , MinerId:"GENESIS"}
    bm := Initialize(configs,genesisBlock)
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
    if(len(bm.GetMap()) != 6){
	t.Fail()
    }
    c := bm.GetLongestChain()
    if(c[0].Depth != 5  ||
	c[1].Depth != 4 ||
	c[2].Depth != 3 ||
	c[3].Depth != 2 ||
	c[4].Depth != 1 ||
	c[5].Depth != 0 ){
	fmt.Println("TestInsert:","depth missmatch")
	t.Fail()
    }
    if(c[0].PrevHash != GetHash(c[1]) ||
	c[1].PrevHash != GetHash(c[2]) ||
	c[2].PrevHash != GetHash(c[3]) ||
	c[3].PrevHash != GetHash(c[4]) ||
	c[4].PrevHash != GetHash(c[5])){
	fmt.Println("TestInsert:","prevhash missmatch")
        t.Fail()
    }
    if(!BHashEndsWithZeros(bm.GetLongestChain()[1], configs.PowPerNoOpBlock)){
	t.Fail()
    }
}


func TestStop(t *testing.T) {
    blockCh := make(chan *Block,5)
    genesisBlock := Block{ PrevHash: "GENESIS", Nonce:"GENESIS" , MinerId:"GENESIS"}
    bm := Initialize(configs,genesisBlock)
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
    if(len(bm.GetMap()) != 1){
	fmt.Println("TestStop:","map should be empty")
        t.Fail()
    }
}

func TestReads(t *testing.T) {
    plan, e := ioutil.ReadFile("configs.json")
    if e == nil {
        err := json.Unmarshal(plan, &configs)
        if(err != nil){
            log.Fatal("error reading json:", err)
        }
    } else {
        log.Fatal("file error:", e)
    }
    blockCh := make(chan *Block,5)
    genesisBlock := Block{ PrevHash: "GENESIS", Nonce:"GENESIS" , MinerId:"GENESIS"}
    bm := Initialize(configs,genesisBlock)
    PrepareMining()
    op := minerlib.Op{Op:"touch", Fname:"a.txt", SeqNum: 123, MinerId:"james"}
    op1 := minerlib.Op{Op:"touch", Fname:"b.txt", SeqNum: 123, MinerId:"james"}
    rec1 := rfslib.Record{}
    copy(rec1[:], "hi how ya doing today? fine? so am I :):D:D:D1")
    rec2 := rfslib.Record{}
    copy(rec2[:], "hi how ya doing today? fine? so am I :)2")
    rec3 := rfslib.Record{}
    copy(rec3[:], "hi how ya doing today? fine? so am I :):D:D3")
    rec4 := rfslib.Record{}
    copy(rec4[:], "hi how ya doing today? fine? so am I :):D4")

    op2 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec1 ,SeqNum: 123, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op, op1, op2, op2},"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock([]minerlib.Op{op2},"james",blockCh)
    <-blockCh
    op8 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec4 ,SeqNum: 123, MinerId:"james"}
    op3 := minerlib.Op{Op:"touch", Fname:"c.txt", SeqNum: 123, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op3,op8},"james",blockCh)
    <-blockCh
    op7 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec3 ,SeqNum: 123, MinerId:"james"}
    op4 := minerlib.Op{Op:"append", Fname:"b.txt", Rec:rec2 ,SeqNum: 456, MinerId:"james"}
    op5 := minerlib.Op{Op:"append", Fname:"c.txt", Rec:rec3 ,SeqNum: 789, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op2,op5},"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock([]minerlib.Op{op4,op4,op7},"james",blockCh)
    <-blockCh
    op6 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec2 ,SeqNum: 123, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op4,op5,op6},"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    ls := bm.LS()
    if(ls["a.txt"] != 7 || ls["b.txt"] != 3 || ls["c.txt"] != 2){
	fmt.Println("TestReads:", "ls fail")
	t.Fail()
    }
    cat := bm.Cat("a.txt")
    if(cat[0] != rec1 ||
       cat[1] != rec1 ||
       cat[2] != rec1 ||
       cat[3] != rec4 ||
       cat[4] != rec1 ||
       cat[5] != rec3 ||
       cat[6] != rec2){
	fmt.Println("TestReads:", "cat fail")
	t.Fail()
    }

    tail := bm.Tail(2,"a.txt")
    if(tail[0] != rec3 ||
       tail[1] != rec2){
        fmt.Println("TestReads:", "tail fail")
        t.Fail()
    }

    head := bm.Head(5,"a.txt")
    if(head[0] != rec1 ||
       head[1] != rec1 ||
       head[2] != rec1 ||
       head[3] != rec4 ||
       head[4] != rec1){
        fmt.Println("TestReads:", "head fail")
        t.Fail()
    }

    coins := bm.CountCoins("james")
    expectedCoins := 0 - int(configs.NumCoinsPerFileCreate)*3 - 12 + int(configs.MinedCoinsPerNoOpBlock)*4 + int(configs.MinedCoinsPerOpBlock)*6
    if(coins != expectedCoins){
	    fmt.Println("TestReads:", "CountCoinsFail:", coins, "expected:", expectedCoins)
	    t.Fail()
    }
}

func TestConfirms(t *testing.T) {
    plan, e := ioutil.ReadFile("configs.json")
    if e == nil {
        err := json.Unmarshal(plan, &configs)
        if(err != nil){
            log.Fatal("error reading json:", err)
        }
    } else {
        log.Fatal("file error:", e)
    }
    blockCh := make(chan *Block,5)
    genesisBlock := Block{ PrevHash: "GENESIS", Nonce:"GENESIS" , MinerId:"GENESIS"}
    bm := Initialize(configs,genesisBlock)
    PrepareMining()
    op := minerlib.Op{Op:"touch", Fname:"a.txt", SeqNum: 123, MinerId:"james"}
    op1 := minerlib.Op{Op:"touch", Fname:"b.txt", SeqNum: 123, MinerId:"james"}
    rec1 := rfslib.Record{}
    copy(rec1[:], "hi how ya doing today? fine? so am I :):D:D:D1")
    rec2 := rfslib.Record{}
    copy(rec2[:], "hi how ya doing today? fine? so am I :)2")
    rec3 := rfslib.Record{}
    copy(rec3[:], "hi how ya doing today? fine? so am I :):D:D3")
    rec4 := rfslib.Record{}
    copy(rec4[:], "hi how ya doing today? fine? so am I :):D4")

    op2 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec1 ,SeqNum: 123, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op, op1, op2, op2},"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock([]minerlib.Op{op2},"james",blockCh)
    <-blockCh
    op8 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec4 ,SeqNum: 123, MinerId:"james"}
    op3 := minerlib.Op{Op:"touch", Fname:"c.txt", SeqNum: 123, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op3,op8},"james",blockCh)
    <-blockCh
    if(bm.CheckIfFileExists("c.txt") != true){
        fmt.Println("TestConfirms:", "c.txt should exist")
        t.Fail()
    }
    fsize := bm.CheckFileSize("a.txt")
    if(fsize != 4){
	fmt.Println("TestConfirms:", "a.txt should have a size of 4 but is the size:", fsize)
        t.Fail()
    }

    op7 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec3 ,SeqNum: 123, MinerId:"james"}
    op4 := minerlib.Op{Op:"append", Fname:"b.txt", Rec:rec2 ,SeqNum: 123, MinerId:"james"}
    op5 := minerlib.Op{Op:"append", Fname:"c.txt", Rec:rec3 ,SeqNum: 123, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op2,op5},"james",blockCh)
    <-blockCh
    go bm.MineAndAddBlock([]minerlib.Op{op4,op4,op7},"james",blockCh)
    <-blockCh
    op6 := minerlib.Op{Op:"append", Fname:"a.txt", Rec:rec2 ,SeqNum: 123, MinerId:"james"}
    go bm.MineAndAddBlock([]minerlib.Op{op4,op5,op6},"james",blockCh)
    <-blockCh

    /////////////////////////////// 0 Confirms blocks
    ls := bm.LS()
    if(ls["a.txt"] != 3 || ls["b.txt"] != 0 || ls["c.txt"] != 0 ){
	fmt.Println("TestConfirms0:", "ls fail")
        t.Fail()
    }
    cat := bm.Cat("a.txt")
    if(len(cat) != 3 || cat[0] != rec1 ||
       cat[1] != rec1 ||
       cat[2] != rec1){
        fmt.Println("TestConfirms0:", "cat fail")
        t.Fail()
    }

    tail := bm.Tail(2,"a.txt")
    if(len(tail) != 2 || tail[0] != rec1 ||
       tail[1] != rec1){
        fmt.Println("TestConfirms0:", "tail fail")
        t.Fail()
    }

    head := bm.Head(5,"a.txt")
    if(len(head) != 3 || head[0] != rec1 ||
       head[1] != rec1 ||
       head[2] != rec1){
        fmt.Println("TestConfirms0:", "head fail")
        t.Fail()
    }

    coins := bm.CountCoins("james")
    expectedCoins := 0 - int(configs.NumCoinsPerFileCreate)*3 - 3 + int(configs.MinedCoinsPerNoOpBlock)*0 + int(configs.MinedCoinsPerOpBlock)*6
    if(coins != expectedCoins){
            fmt.Println("TestConfirms0:", "TestConfirms0:", coins, "expected:", expectedCoins)
            t.Fail()
    }

    ////////////////////////////////// 1 Confirm block
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    ls = bm.LS()
    if(ls["a.txt"] != 4 || ls["b.txt"] != 0 || ls["c.txt"] != 0){
        fmt.Println("TestConfirms1:", "ls fail")
        t.Fail()
    }
    cat = bm.Cat("a.txt")
    if(len(cat) != 4 || cat[0] != rec1 ||
       cat[1] != rec1 ||
       cat[2] != rec1 ||
       cat[3] != rec4){
        fmt.Println("TestConfirms1:", "cat fail")
        t.Fail()
    }

    tail = bm.Tail(2,"a.txt")
    if(len(tail) != 2 || tail[0] != rec1 || tail[1] != rec4){
        fmt.Println("TestConfirms1:", "tail fail")
        t.Fail()
    }

    head = bm.Head(5,"a.txt")
    if(len(head) != 4 || head[0] != rec1 ||
       head[1] != rec1 ||
       head[2] != rec1 ||
       head[3] != rec4){
        fmt.Println("TestConfirms1:", "head fail")
        t.Fail()
    }

    coins = bm.CountCoins("james")
    expectedCoins = 0 - int(configs.NumCoinsPerFileCreate)*3 - 4 + int(configs.MinedCoinsPerNoOpBlock)*1 + int(configs.MinedCoinsPerOpBlock)*6
    if(coins != expectedCoins){
            fmt.Println("TestConfirms1:", "TestConfirms0:", coins, "expected:", expectedCoins)
            t.Fail()
    }

    ////////////////////////////////// 2 Confirm block
    if(bm.CheckIfFileExists("c.txt") != true){
        fmt.Println("TestConfirms:", "c.txt should exist")
	t.Fail()
    }
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    ls = bm.LS()
    if(ls["a.txt"] != 5 || ls["b.txt"] != 0 || ls["c.txt"] != 1){
        fmt.Println("TestConfirms2:", "ls fail")
        t.Fail()
    }
    cat = bm.Cat("a.txt")
    if(len(cat) != 5 || cat[0] != rec1 ||
       cat[1] != rec1 ||
       cat[2] != rec1 ||
       cat[3] != rec4 ||
       cat[4] != rec1){
        fmt.Println("TestConfirms2:", "cat fail")
        t.Fail()
    }

    tail = bm.Tail(2,"a.txt")
    if(len(tail) != 2 || tail[0] != rec4 ||
       tail[1] != rec1){
        fmt.Println("TestConfirms2:", "tail fail")
        t.Fail()
    }

    head = bm.Head(5,"a.txt")
    if(len(head) != 5 || head[0] != rec1 ||
       head[1] != rec1 ||
       head[2] != rec1 ||
       head[3] != rec4 ||
       head[4] != rec1){
        fmt.Println("TestConfirms2:", "head fail")
        t.Fail()
    }

    coins = bm.CountCoins("james")
    expectedCoins = 0 - int(configs.NumCoinsPerFileCreate)*3 - 6 + int(configs.MinedCoinsPerNoOpBlock)*2 + int(configs.MinedCoinsPerOpBlock)*6
    if(coins != expectedCoins){
            fmt.Println("TestConfirms1:", "TestConfirms0:", coins, "expected:", expectedCoins)
            t.Fail()
    }

    ////////////////////////////////// 3 Confirm block
    go bm.MineAndAddBlock(nil,"james",blockCh)
    <-blockCh
    ls = bm.LS()
    if(ls["a.txt"] != 6 || ls["b.txt"] != 2 || ls["c.txt"] != 1){
        fmt.Println("TestConfirms3:", "ls fail")
        t.Fail()
    }
    cat = bm.Cat("a.txt")
    if(len(cat) != 6 || cat[0] != rec1 ||
       cat[1] != rec1 ||
       cat[2] != rec1 ||
       cat[3] != rec4 ||
       cat[4] != rec1 ||
       cat[5] != rec3){
        fmt.Println("TestConfirms2:", "cat fail")
        t.Fail()
    }

    tail = bm.Tail(2,"a.txt")
    if(len(tail) != 2 || tail[0] != rec1 ||
       tail[1] != rec3){
        fmt.Println("TestConfirms2:", "tail fail")
        t.Fail()
    }

    head = bm.Head(5,"a.txt")
    if(len(head) != 5 || head[0] != rec1 ||
       head[1] != rec1 ||
       head[2] != rec1 ||
       head[3] != rec4 ||
       head[4] != rec1){
        fmt.Println("TestConfirms2:", "head fail")
        t.Fail()
    }

    coins = bm.CountCoins("james")
    expectedCoins = 0 - int(configs.NumCoinsPerFileCreate)*3 - 9 + int(configs.MinedCoinsPerNoOpBlock)*3 + int(configs.MinedCoinsPerOpBlock)*6
    if(coins != expectedCoins){
            fmt.Println("TestConfirms1:", "TestConfirms0:", coins, "expected:", expectedCoins)
            t.Fail()
    }
}
