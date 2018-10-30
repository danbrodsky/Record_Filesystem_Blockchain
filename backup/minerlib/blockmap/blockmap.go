package blockmap

import (
    "fmt"
    "crypto/md5"
    "encoding/hex"
    "math/rand"
    "blockchain/minerlib"
    "blockchain/rfslib"
<<<<<<< HEAD
)

type BlockMap struct {
    tailBlock Block
    genesisBlock Block
=======
	"time"
)

type BlockMap struct {
    TailBlock Block
    GenesisBlock Block
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
    Map map[string]Block
}

type BM interface {
    Insert(block Block) (err error)
    GetMap() (map[string]Block)
    GetLongestChain() ([]Block)
}

var (
    Configs minerlib.Settings
)

type Block struct{
    PrevHash string
    Ops []minerlib.Op
    Nonce string
    MinerId string
    Depth int
}

<<<<<<< HEAD
func Initialize(settings minerlib.Settings, genesisBlock Block) (blockmap BlockMap){
    Configs = settings
    blockmap = BlockMap{}
    genesisBlock.Depth = 0
    blockmap.tailBlock = genesisBlock
    blockmap.genesisBlock = genesisBlock
    blockmap.Map = make(map[string]Block)
    blockmap.Map[GetHash(genesisBlock)] = genesisBlock
=======
func Initialize(settings minerlib.Settings, GenesisBlock Block) (blockmap BlockMap){
    Configs = settings
    blockmap = BlockMap{}
    GenesisBlock.Depth = 0
    blockmap.TailBlock = GenesisBlock
    blockmap.GenesisBlock = GenesisBlock
    blockmap.Map = make(map[string]Block)
    blockmap.Map[GetHash(GenesisBlock)] = GenesisBlock
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
    return blockmap
}

type PrevHashDoesNotExistError string

func (e PrevHashDoesNotExistError) Error() string {
    return fmt.Sprintf("Block-Map: Error hash does not exist in map [%s]", string(e))
}

type BlockNotValidError string

func (e BlockNotValidError) Error() string {
    return fmt.Sprintf("Block-Map: Error block does not end with continuous 0s [%s]", string(e))
}

// Gets the hash of the block
func GetHash(block Block) string{
     h := md5.New()
     h.Write([]byte(fmt.Sprintf("%v", block)))
     return hex.EncodeToString(h.Sum(nil))
}

// Inserts a block in the block map
// Precondition Block should be valid that is all fields should be filled
// and the the previous hash should exist
// also the hash of the block should end with some number of 0s
func (bm *BlockMap) Insert(block Block) (err error){
<<<<<<< HEAD
    if(block.Ops != nil && len(block.Ops) != 0 && !BHashEndsWithZeros(block, Configs.PowPerOpBlock)){ // TODO set env variable
	return BlockNotValidError(GetHash(block))
    } else if(!BHashEndsWithZeros(block, Configs.PowPerNoOpBlock)){
	return BlockNotValidError(GetHash(block))
    }
    if _, ok := bm.Map[block.PrevHash]; ok {
        bm.Map[GetHash(block)] = block
	bm.tailBlock = block
//	fmt.Println("tail:", bm.tailBlock)
	return nil
=======
    if block.Ops != nil && len(block.Ops) != 0 && !BHashEndsWithZeros(block, Configs.PowPerOpBlock) { // TODO set env variable
	return BlockNotValidError(GetHash(block))
    } else if !BHashEndsWithZeros(block, Configs.PowPerNoOpBlock) {
	return BlockNotValidError(GetHash(block))
    }
    if _, ok := bm.Map[block.PrevHash]; ok {
    	bm.Map[GetHash(block)] = block
		bm.updateLongest(block)
		fmt.Println("tail:", bm.TailBlock)
    	return nil
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
    } else {
	return PrevHashDoesNotExistError(block.PrevHash)
    }
}

func (bm *BlockMap) updateLongest(block Block) {
<<<<<<< HEAD
    if block.Depth == bm.tailBlock.Depth {
        if rand.Intn(2) == 1 {
            bm.tailBlock = block
        }
    }
    if block.Depth > bm.tailBlock.Depth {
        bm.tailBlock = block
=======
    if block.Depth == bm.TailBlock.Depth {
        if rand.Intn(2) == 1 {
            bm.TailBlock = block
        }
    }
    if block.Depth > bm.TailBlock.Depth {
        bm.TailBlock = block
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
    }
}


func (bm *BlockMap) GetMap() (map[string]Block){
    return bm.Map
}

func BHashEndsWithZeros(block Block, numZeros uint8) bool{
    hash := GetHash(block)
    for i:= len(hash) - 1; i > len(hash)-1 -int(numZeros) ; i--{
        if(hash[i] != '0'){
	    return false
	}
    }
    return true
}

func (bm *BlockMap) SetTailBlock(block Block){
<<<<<<< HEAD
    bm.tailBlock = block
=======
    bm.TailBlock = block
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
}

// Mines a block and puts it in the block chain
// ops is the operation 
// minerId is the miner of the miner
func (bm *BlockMap) MineAndAddBlock(ops []minerlib.Op, minerId string, blockCh chan *Block){
<<<<<<< HEAD
    block := Block{ PrevHash: GetHash(bm.tailBlock),
		    Ops:ops,
		    MinerId:minerId,
		    Depth: bm.tailBlock.Depth+1 }
    minedBlock := ComputeBlock(block , Configs.PowPerOpBlock) // TODO set numZeros
    if(minedBlock != nil){
        bm.Insert(*minedBlock)
	blockCh <-minedBlock
    } else{
	blockCh <-nil//most likely mining was stopped
    }
=======
	block := Block{ PrevHash: GetHash(bm.TailBlock),
		Ops:ops,
		MinerId:minerId,
		Depth: bm.TailBlock.Depth+1 }
	StopMining()
	time.Sleep(1 * time.Second)
	PrepareMining()
	fmt.Println("mining started")
	minedBlock := ComputeBlock(block , 4) // TODO set numZeros
	if(minedBlock != nil){
		bm.Insert(*minedBlock)
		blockCh <-minedBlock
	}
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
}


// gets the longest chain of the blockchain with the first array to be the most recent
// block in the map and the last to be the genesis block
func (bm *BlockMap) GetLongestChain() ([]Block){
    var blockChain []Block
<<<<<<< HEAD
    var currBlock = bm.tailBlock
    for currBlock.PrevHash != bm.genesisBlock.PrevHash {
        blockChain = append(blockChain, currBlock)
	currBlock = bm.Map[currBlock.PrevHash]
    }
    blockChain = append(blockChain, bm.genesisBlock)
=======
    var currBlock = bm.TailBlock
    for currBlock.PrevHash != bm.GenesisBlock.PrevHash {
        blockChain = append(blockChain, currBlock)
	currBlock = bm.Map[currBlock.PrevHash]
    }
    blockChain = append(blockChain, bm.GenesisBlock)
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
    return blockChain
}

func (bm *BlockMap) LS() map[string]int{
    bc := bm.GetLongestChain()
    fs := make(map[string]int)
    for i := len(bc)-1 ; i >= 0 ; i--{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0 ){
	    for _,op := range bc[i].Ops{
		switch op.Op{
		case "append":
		    if _, ok := fs[op.Fname]; ok {
		       if(int(Configs.ConfirmsPerFileAppend) <= i){
                           fs[op.Fname]++
		       }
                    }
		case "touch":
		    if(int(Configs.ConfirmsPerFileCreate) <= i){
		        fs[op.Fname] = 0
		    }
		}
	    }
	}
    }
    return fs
}

func (bm *BlockMap) Cat(fname string) []rfslib.Record{
    bc := bm.GetLongestChain()
    f := []rfslib.Record{}
     for i := len(bc)-1 ; i >= int(Configs.ConfirmsPerFileAppend) ; i--{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
            for _,op := range bc[i].Ops{
		if(op.Op == "append" && op.Fname == fname){
                    f = append(f, op.Rec)
		}
	   }
        }
    }
    return f
}

func (bm *BlockMap) Tail(k int,fname string) []rfslib.Record{
    bc := bm.GetLongestChain()
    f := []rfslib.Record{}
    for i := int(Configs.ConfirmsPerFileAppend) ; i < len(bc) ; i++{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
            for n := len(bc[i].Ops) -1 ; n >= 0 ; n--{
		op := bc[i].Ops[n]
                if(op.Op == "append" && op.Fname == fname){
                    f = append(f, op.Rec)
                }
		if(len(f) == k){
		    return reverse(f)
		}
           }
        }
    }
    return reverse(f)
}

func reverse(l []rfslib.Record) []rfslib.Record{
    revList := []rfslib.Record{}
    for i:=len(l)-1 ; i>= 0; i--{
         revList = append(revList, l[i])
    }
    return revList
}

func (bm *BlockMap) Head(k int,fname string) []rfslib.Record{
    bc := bm.GetLongestChain()
    f := []rfslib.Record{}
    for i := len(bc)-1 ; i >= int(Configs.ConfirmsPerFileAppend) ; i--{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
            for _,op := range bc[i].Ops{
                if(op.Op == "append" && op.Fname == fname){
                    f = append(f, op.Rec)
	        }
		if(len(f) == k){
                    return f
                }
           }
        }
    }
    return f
}

func (bm *BlockMap) CountCoins(minerId string) int{
    bc := bm.GetLongestChain()
    var coins int = 0
    var appends int
    var touches int
    for i := len(bc) - 1 ; i >= 0 ; i--{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
	    appends = 0
	    touches = 0
	    for _,op := range bc[i].Ops{
                switch op.Op{
                case "append":
		    if(int(Configs.ConfirmsPerFileAppend) <= i){
		        if(op.MinerId == minerId){
			    appends++
		        }
	            }
                case "touch":
		    if(int(Configs.ConfirmsPerFileCreate) <= i){
		        if(op.MinerId == minerId){
                            touches++
                        }
	            }
                }
            }
	    //fmt.Println("appens:", appends)
	    //fmt.Println("touches:", touches)
	    if(bc[i].MinerId == minerId){
		coins += int(Configs.MinedCoinsPerOpBlock)
	    }
	    coins = coins - appends - touches*int(Configs.NumCoinsPerFileCreate)
	    //fmt.Println("coins:", coins)
	} else {
	    if(bc[i].MinerId == minerId){
	        coins += int(Configs.MinedCoinsPerNoOpBlock)
            }
	}
    }
    return coins
}

func (bm *BlockMap) CheckIfFileExists(fname string) bool{
    bc := bm.GetLongestChain()
     for i := len(bc)-1 ; i >= 0 ; i--{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
            for _,op := range bc[i].Ops{
                if(op.Op == "touch" && op.Fname == fname){
                    return true
                }
           }
        }
    }
    return false
}

// num records in a file
func (bm *BlockMap) CheckFileSize(fname string) int{
    bc := bm.GetLongestChain()
    size := 0
     for i := len(bc)-1 ; i >= 0 ; i--{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
            for _,op := range bc[i].Ops{
                if(op.Op == "append" && op.Fname == fname){
                    size ++
                }
           }
        }
    }
    return size
}
