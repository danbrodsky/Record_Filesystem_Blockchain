package blockmap

import (
    "fmt"
    "crypto/md5"
    "encoding/hex"
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


type Block struct{
    PrevHash string
    Ops []string
    Nonce string
    MinerId string
    Depth int
}

func NewBlockMap(genesisBlock Block) (blockmap BlockMap) {
    // TODO check if block is genesis
    blockmap = BlockMap{}
    genesisBlock.Depth =  0
    blockmap.tailBlock = genesisBlock
    blockmap.genesisBlock = genesisBlock
    blockmap.Map = make(map[string]Block)
    blockmap.Map[getHash(genesisBlock)] = genesisBlock
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
func getHash(block Block) string{
     h := md5.New()
     h.Write([]byte(fmt.Sprintf("%v", block)))
     return hex.EncodeToString(h.Sum(nil))
}

// Inserts a block in the block map
// Precondition Block should be valid that is all fields should be filled
// and the the previous hash should exist
// also the hash of the block should end with some number of 0s
func (bm *BlockMap) Insert(block Block) (err error){
    if(!BHashEndsWithZeros(block,4)){ // TODO set env variable
	return BlockNotValidError(getHash(block))
    }
    if _, ok := bm.Map[block.PrevHash]; ok {
        bm.Map[getHash(block)] = block
	bm.tailBlock = block
	fmt.Println("tail:", bm.tailBlock)
	return nil
    } else {
	return PrevHashDoesNotExistError(block.PrevHash)
    }
}

func (bm *BlockMap) GetMap() (map[string]Block){
    return bm.Map
}

func BHashEndsWithZeros(block Block, numZeros int) bool{
    hash := getHash(block)
    for i:= len(hash) - 1; i > len(hash)-1 -numZeros ; i--{
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
func (bm *BlockMap) MineAndAddBlock(ops []string, minerId string, blockCh chan *Block){
    block := Block{ PrevHash: getHash(bm.tailBlock),
		    Ops:ops,
		    MinerId:minerId,
		    Depth: bm.tailBlock.Depth+1 }
    minedBlock := ComputeBlock(block , 4) // TODO set numZeros
    if(minedBlock != nil){
        bm.Insert(*minedBlock)
	blockCh <-minedBlock
    } else{
	blockCh <-nil
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




