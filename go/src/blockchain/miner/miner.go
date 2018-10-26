package main

import (
	"blockchain/miner/minerlib"
	"encoding/gob"
	"log"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
)


type M interface {
	MakeKnown(addr string, reply *int) error
	ReceiveOp(op string, reply *int) error
	ReceiveBlock(block minerlib.Block, reply *int) error
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
}

var (
	minerConnections = make(map[string]*rpc.Client)
	waitingOps = make(map[string]bool)
	blocks = make(map[string]*minerlib.Block)

	currentLongest *minerlib.Block
	genesisBlock *minerlib.Block

)


func (miner Miner) MakeKnown(addr string, reply *int) error {
	if _, ok := minerConnections[addr]; !ok {
		client, err := rpc.DialHTTP("tcp", addr)
		if err == nil {
			minerConnections[addr] = client
		} else {
			log.Println("dialing:", err)
		}
	}
	return nil
}

func (miner Miner) ReceiveOp(op string, reply *int) error {
	if _, ok := waitingOps[op]; !ok && minerlib.ValidateOp() {
		// op not received yet, store and flood it
		waitingOps[op] = true
		for _, conn := range minerConnections {
			conn.Go("Miner.ReceiveOp", op, nil, nil)
		}

		// TODO: Check timeout and maxOps and see if it's time to mine a new block
	}
	return nil
}

func (miner Miner) ReceiveBlock(returnAddr string, hash string, block minerlib.Block, reply *int) error {
	// TODO: Implement ValidateBlock

	// if miner is behind, get previous blocks until caught up
	if _, ok := blocks[block.PrevHash]; !ok {

		// add missing blocks to a temp store in case they're fake
		missingBlocks := make(map[string]*minerlib.Block)
		missingBlocks[hash] = &block

		for !ok {
			var prevBlock *minerlib.Block
			minerConnections[returnAddr].Call("Miner.SendPreviousBlock", block.PrevHash, &prevBlock)
			missingBlocks[block.PrevHash] = prevBlock
			block = *prevBlock
			_, ok = blocks[block.PrevHash]
		}

	}

	// TODO: change to blockmap insert
	if minerlib.ValidateBlock() {
		addToTree(block, hash)

		// send the block to connected miners
		for _, conn := range minerConnections {
			conn.Go("Miner.ReceiveBlock", block, nil, nil)
		}
	}
	return nil
}

func (miner Miner) SendPreviousBlock(hash string, block *minerlib.Block) error {

	*block = *blocks[hash]

	return nil
}

func addToTree(block minerlib.Block, hash string) {
	blocks[hash] = &block
	updateLongest(block)
}

func updateLongest(block minerlib.Block) {
	if block.Depth == currentLongest.Depth {
		if rand.Intn(2) == 1 {
			currentLongest = &block
		}
	}
	if block.Depth > currentLongest.Depth {
		currentLongest = &block
	}
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

	// TODO: add genesis block

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
			minerConnections[addr] = client

		} else {log.Println("dialing:", err)}
	}
}

