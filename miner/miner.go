package main

import (
	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
	"blockchain/minerlib/blockmap"
	//"blockchain/miner/minerlib"
	"io/ioutil"
	"encoding/json"
	"fmt"
	"time"
	"log"
	"net"
	//"net/http"
	"net/rpc"
	"os"
)


type M interface {
	MakeKnown(addr string, reply *int) error
	ReceiveOp(operation Op, reply *int) error
	ReceiveBlock(payload Payload, reply *int) error
	GetPreviousBlock(prevHash string, block *blockmap.Block) error
}

var (
    GovecOptions govec.GoLogOptions = govec.GetDefaultLogOptions()
    miner *Miner
    MinerLogger *govec.GoLog
    Configs Settings
)

type Miner struct {
	Connections map[string]*rpc.Client
	WaitingOps map[string]string
	BlockMap blockmap.BlockMap
}

type Settings struct {
    MinedCoinsPerOpBlock   uint8  `json:"MinedCoinsPerOpBlock"`
    MinedCoinsPerNoOpBlock uint8  `json:"MinedCoinsPerNoOpBlock"`
    NumCoinsPerFileCreate  uint8  `json:"NumCoinsPerFileCreate"`
    GenOpBlockTimeout      uint8  `json:"GenOpBlockTimeout"`
    GenesisBlockHash       string `json:"GenesisBlockHash"`
    PowPerOpBlock          uint8  `json:"PowPerOpBlock"`
    PowPerNoOpBlock        uint8  `json:"PowPerNoOpBlock"`
    ConfirmsPerFileCreate  uint8  `json:"ConfirmsPerFileAppend"`
    MinerID             string   `json:"MinerID"`
    PeerMinersAddrs     []string `json:"PeerMinersAddrs"`
    IncomingMinersAddr  string   `json:"IncomingMinersAddr"`
    OutgoingMinersIP    string   `json:"OutgoingMinersIP"`
    IncomingClientsAddr string   `json:"IncomingClientsAddr"`
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
	fmt.Println("here1")
	if _, ok := miner.Connections[addr]; !ok {
		fmt.Println("here2")
		MinerLogger.LogLocalEvent("MakeKnown Called", GovecOptions)
		client, err := vrpc.RPCDial("tcp", addr, MinerLogger, GovecOptions)
		if err == nil {
			miner.Connections[addr] = client
		} else {
			log.Println("dialing:", err)
		}
	}
	return nil
}

func (miner *Miner) ReceiveOp(operation Op, reply *int) error {
	if _, ok := miner.WaitingOps[operation.op + operation.id]; !ok{
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

func rpcserver() {
    fmt.Println("Starting rpc server")
    miner = new(Miner)
    MinerLogger = govec.InitGoVector(Configs.MinerID, "./logs/minerlogfile" + Configs.MinerID, govec.GetDefaultConfig())
    miner.BlockMap = blockmap.NewBlockMap(blockmap.Block{ PrevHash: "GENESIS", Nonce:"GENESIS" , MinerId:"GENESIS"})
    miner.WaitingOps = make(map[string]string)
    miner.Connections = make(map[string]*rpc.Client)
    server := rpc.NewServer()
    server.Register(miner)
    l, e := net.Listen("tcp", Configs.IncomingMinersAddr)
    if e != nil {
        log.Fatal("listen error:", e)
    }
    vrpc.ServeRPCConn(server, l, MinerLogger, GovecOptions)
}

func rpcclient(){
    fmt.Println("Starting rpc client")
    for _, addr := range Configs.PeerMinersAddrs {
        client, err := vrpc.RPCDial("tcp", addr, MinerLogger, GovecOptions)
        if err == nil {
            // make this miner known to the other miner
	    var result int
	    err := client.Call("Miner.MakeKnown", addr, &result)
	    fmt.Println(err)
            miner.Connections[addr] = client
        } else {log.Println("dialing:", err)}
    }
}


func main() {
    // load json settings
    plan, e := ioutil.ReadFile(os.Args[1])
    if e == nil {
	err := json.Unmarshal(plan, &Configs)
	if(err != nil){
	    log.Fatal("error reading json:", err)
        }
    } else {
        log.Fatal("file error:", e)
    }
    go rpcserver()
    time.Sleep(3000 * time.Millisecond)
    go rpcclient()
    time.Sleep(3000 * time.Millisecond)
    // TODO: add a receiver for managing block mining and client operations
}

