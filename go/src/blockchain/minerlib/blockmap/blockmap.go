package blockmap

import (
    "fmt"
    "crypto/md5"
    "encoding/hex"
    "math/rand"
    "blockchain/minerlib"
    "blockchain/rfslib"
)

type BlockMap struct {
    tailBlock Block
    genesisBlock Block
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

func Initialize(settings minerlib.Settings, genesisBlock Block) (blockmap BlockMap){
    Configs = settings
    blockmap = BlockMap{}
    genesisBlock.Depth = 0
    blockmap.tailBlock = genesisBlock
    blockmap.genesisBlock = genesisBlock
    blockmap.Map = make(map[string]Block)
    blockmap.Map[GetHash(genesisBlock)] = genesisBlock
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
    if(!BHashEndsWithZeros(block, Configs.PowPerOpBlock)){ // TODO set env variable
	return BlockNotValidError(GetHash(block))
    }
    if _, ok := bm.Map[block.PrevHash]; ok {
        bm.Map[GetHash(block)] = block
	bm.tailBlock = block
//	fmt.Println("tail:", bm.tailBlock)
	return nil
    } else {
	return PrevHashDoesNotExistError(block.PrevHash)
    }
}

func (bm *BlockMap) updateLongest(block Block) {
    if block.Depth == bm.tailBlock.Depth {
        if rand.Intn(2) == 1 {
            bm.tailBlock = block
        }
    }
    if block.Depth > bm.tailBlock.Depth {
        bm.tailBlock = block
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
    bm.tailBlock = block
}

// Mines a block and puts it in the block chain
// ops is the operation 
// minerId is the miner of the miner
func (bm *BlockMap) MineAndAddBlock(ops []minerlib.Op, minerId string, blockCh chan *Block){
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
}


// gets the longest chain of the blockchain with the first array to be the most recent
// block in the map and the last to be the genesis block
func (bm *BlockMap) GetLongestChain() ([]Block){
    var blockChain []Block
    var currBlock = bm.tailBlock
    for currBlock.PrevHash != bm.genesisBlock.PrevHash {
        blockChain = append(blockChain, currBlock)
	currBlock = bm.Map[currBlock.PrevHash]
    }
    blockChain = append(blockChain, bm.genesisBlock)
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
                       fs[op.Fname]++
                    }
		case "touch":
		    fs[op.Fname] = 0
		}
	    }
	}
    }
    return fs
}

func (bm *BlockMap) Cat(fname string) []rfslib.Record{
    bc := bm.GetLongestChain()
    f := []rfslib.Record{}
     for i := len(bc)-1 ; i >= 0 ; i--{
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
    for i := 0 ; i < len(bc) ; i++{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
            for n := len(bc[i].Ops) -1 ; n >= 0 ; n--{
		op := bc[i].Ops[n]
                if(op.Op == "append" && op.Fname == fname){
                    f = append(f, op.Rec)
                }
		if(len(f) == k){
		    break
		}
           }
        }
    }
    return f
}

func (bm *BlockMap) Head(k int,fname string) []rfslib.Record{
    bc := bm.GetLongestChain()
    f := []rfslib.Record{}
     for i := len(bc)-1 ; i >= 0 ; i--{
        if(bc[i].Ops != nil && len(bc[i].Ops) != 0){
            for _,op := range bc[i].Ops{
                if(op.Op == "append" && op.Fname == fname){
                    f = append(f, op.Rec)
	        }
		if(len(f) == k){
                    break
                }
           }
        }
    }
    return f
}

