package main

import (
	"blockchain/miner/blockmap"
	"blockchain/miner/minerlib"
	"encoding/gob"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
)


type M interface {
	MakeKnown(addr string, reply *int) error
	ReceiveOp(operation Op, reply *int) error
	ReceiveBlock(payload Payload, reply *int) error
	GetPreviousBlock(prevHash string, block *blockmap.Block) error
}

type Miner struct {
	MinedCoinsPerOpBlock int
	MinedCoinsPerNoOpBlock int
	NumCoinsPerFileCreate int
	GenOpBlockTimeout int
	GenesisBlockHash string
	PowPerOpBlock int
	PowPerNoOpBlock int
	ConfirmsPerFileCreate int
	ConfirmsPerFileAppend int
	MinerID string
	PeerMinersAddrs []string
	IncomingMinersAddr string
	OutgoingMinersIP string
	IncomingClientsAddr string

	Connections map[string]*rpc.Client
	WaitingOps map[string]string
	BlockMap blockmap.BlockMap
}

type Op struct {
	op string
	// random string to make ops unique
	id string
}

type Payload struct {
	returnAddr string
	hash string
	block blockmap.Block
}


func (miner *Miner) MakeKnown(addr string, reply *int) error {
	if _, ok := miner.Connections[addr]; !ok {
		client, err := rpc.DialHTTP("tcp", addr)
		if err == nil {
			miner.Connections[addr] = client
		} else {
			log.Println("dialing:", err)
		}
	}
	return nil
}

func (miner *Miner) ReceiveOp(operation Op, reply *int) error {
	if _, ok := miner.WaitingOps[operation.op + operation.id]; !ok && minerlib.ValidateOp() {
		// op not received yet, store and flood it
		miner.WaitingOps[operation.op + operation.id] = operation.op
		for _, conn := range miner.Connections {
			conn.Go("Miner.ReceiveOp", operation, nil, nil)
		}

		// TODO: Check timeout and maxOps and see if it's time to mine a new block
	}
	return nil
}

func (miner *Miner) ReceiveBlock(payload Payload, reply *int) error {

	// if miner is behind, get previous BlockMap.Map until caught up
	// TODO: move this portion to blockmap
	// ================================================================================================================
	if _, ok := miner.BlockMap.Map[payload.block.PrevHash]; !ok {

		// add missing BlockMap.Map to a temp store in case they're fake
		missingBlocks := make(map[string]blockmap.Block)
		missingBlocks[payload.hash] = payload.block

		for !ok {
			var prevBlock *blockmap.Block
			miner.Connections[payload.returnAddr].Call("Miner.GetPreviousBlock", payload.block.PrevHash, &prevBlock)
			missingBlocks[payload.block.PrevHash] = *prevBlock
			// TODO: Validate the block here, if it fails then dump missing blocks and break
			_, ok = miner.BlockMap.Map[prevBlock.PrevHash]
		}
		// TODO: Insert all the blocks in missingBlocks into the blockmap
	}
	// ================================================================================================================

	// TODO: change to blockmap insert
	miner.BlockMap.Insert(payload.block)

		// send the block to connected miners
	for _, conn := range miner.Connections {
		conn.Go("Miner.ReceiveBlock", payload, nil, nil)
	}

	return nil
}

func (miner *Miner) GetPreviousBlock(prevHash string, block *blockmap.Block) error {

	if _, ok := miner.BlockMap.Map[prevHash]; ok {
		*block = miner.BlockMap.Map[prevHash]
	}
	return nil
}

func main() {

	miner := new(Miner)

	// load json into miner
	file, e := os.Open(os.Args[1])
	if e == nil {
		decoder := gob.NewDecoder(file)
		e = decoder.Decode(miner)
	} else { log.Fatal("file error:", e)}
	file.Close()
	
	// create blockmap with the genesis block and associated hash
	miner.BlockMap = blockmap.NewBlockMap(blockmap.Block{nil, nil, nil, nil, 0}, miner.GenesisBlockHash)

	miner.WaitingOps = make(map[string]string)
	miner.Connections = make(map[string]*rpc.Client)

	// Open RPC server for other miners
	rpc.Register(miner)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", miner.IncomingMinersAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)

	// connect to all known miners
	for _, addr := range miner.PeerMinersAddrs {
		client, err := rpc.DialHTTP("tcp", addr)
		if err == nil {
			// make this miner known to the other miner
			client.Go("Miner.MakeKnown", addr, nil, nil)
			miner.Connections[addr] = client

		} else {log.Println("dialing:", err)}
	}

	// TODO: add a receiver for managing block mining and client operations
}

