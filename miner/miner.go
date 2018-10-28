package main

import (
	"blockchain/miner/blockmap"
	"encoding/json"
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
	"sort"

	//"blockchain/miner/minerlib"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"time"
)


type M interface {
	MakeKnown(addr string, reply *int) error
	ReceiveOp(operation Op, reply *int) error
	ReceiveBlock(payload Payload, reply *int) error
	GetPreviousBlock(prevHash string, block *blockmap.Block) error
}

var (
    GovecOptions = govec.GetDefaultLogOptions()
    miner *Miner
    MinerLogger *govec.GoLog
    Configs Settings
    blockStartTime int
    maxOps = 10
)

type Miner struct {
	Connections map[string]*rpc.Client
	WaitingOps map[int]Op
	BlockMap blockmap.BlockMap
	IncomingOps chan []Op
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
	// Unix nano time to make ops unique
	seqNum int
}

type Payload struct {
	returnAddr string
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
	if _, ok := miner.WaitingOps[operation.seqNum]; !ok {
		// op not received yet, store and flood it
		miner.WaitingOps[operation.seqNum] = operation

		for _, conn := range miner.Connections {
			conn.Go("Miner.ReceiveOp", operation, nil, nil)
		}

		// reset timeout counter
		if blockStartTime == -1 {
			blockStartTime = time.Now().Second()
		}

		// block timeout or block full, send what you have
		if Configs.GenOpBlockTimeout < uint8(time.Now().Second() - blockStartTime) || len(miner.WaitingOps) >= maxOps {
			blockStartTime = -1

			newOps := make([]Op, len(miner.WaitingOps))
			for _, o := range miner.WaitingOps {
				newOps = append(newOps, o)
			}

			sort.Slice(newOps, func(i, j int) bool {
				return newOps[i].seqNum < newOps[j].seqNum
			})

			miner.IncomingOps <- newOps

			miner.WaitingOps = make(map[int]Op)
		}
	}
	return nil
}

func (miner *Miner) ReceiveBlock(payload Payload, reply *int) error {

	// if miner is behind, get previous BlockMap.Map until caught up
	if _, ok := miner.BlockMap.Map[payload.block.PrevHash]; !ok {

		// add missing BlockMap.Map to a temp store in case they're fake
		missingBlocks := make([]blockmap.Block, 0)
		missingBlocks = append( missingBlocks, payload.block)

		var prevBlock *blockmap.Block
		for !ok {
			miner.Connections[payload.returnAddr].Call("Miner.GetPreviousBlock", payload.block.PrevHash, &prevBlock)
			missingBlocks = append( missingBlocks, *prevBlock)

			// Validate the block, if it fails then dump missing blocks and break
			err := miner.BlockMap.Insert(*prevBlock)
			if err != nil {
				switch err.(type) {
				case blockmap.BlockNotValidError:
					missingBlocks = make([]blockmap.Block, 0)
					break
				// TODO: Add more error handling if necessary
				}
			}
			_, ok = miner.BlockMap.Map[prevBlock.PrevHash]
		}

		// insert all the blocks in missingBlocks into the blockmap
		for _, b := range missingBlocks {
			// each block already checked, no need to check for errors again
			miner.BlockMap.Insert(b)
		}

	} else {
		miner.BlockMap.Insert(payload.block)
	}
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

func rpcServer() {
    fmt.Println("Starting rpc server")
    miner = new(Miner)
    MinerLogger = govec.InitGoVector(Configs.MinerID, "./logs/minerlogfile" + Configs.MinerID, govec.GetDefaultConfig())
    miner.BlockMap = blockmap.NewBlockMap(blockmap.Block{ PrevHash: "GENESIS", Nonce:"GENESIS" , MinerId:"GENESIS"})
    miner.WaitingOps = make(map[int]Op)
    miner.Connections = make(map[string]*rpc.Client)
    server := rpc.NewServer()
    server.Register(miner)
    l, e := net.Listen("tcp", Configs.IncomingMinersAddr)
    if e != nil {
        log.Fatal("listen error:", e)
    }
    vrpc.ServeRPCConn(server, l, MinerLogger, GovecOptions)
}

func rpcClient(){
    fmt.Println("Starting rpc client")
    for _, addr := range Configs.PeerMinersAddrs {
        client, err := vrpc.RPCDial("tcp", addr, MinerLogger, GovecOptions)
        if err == nil {
            // make this miner known to the other miner
	    var result int
	    err := client.Go("Miner.MakeKnown", addr, &result, nil)
	    fmt.Println(err)
            miner.Connections[addr] = client
        } else {log.Println("dialing:", err)}
    }
}

func handleBlocks () {

	// create a noop block and start mining for a nonce
	completeBlock := make(chan *blockmap.Block)
	miner.BlockMap.MineAndAddBlock(nil, Configs.MinerID, completeBlock)

	waitingBlocks := make([][]Op, 0)

	opBeingMined := false

	for true {
		select {
		// receive a newly mined block, flood it and start mining noop
		case cb := <-completeBlock:
			for _, conn := range miner.Connections {
				conn.Go("Miner.ReceiveBlock", Payload{Configs.IncomingMinersAddr, *cb}, nil, nil)
			}
			// start on noop or queued op block right away
			if (len(waitingBlocks) == 0) {
				miner.BlockMap.MineAndAddBlock(nil, Configs.MinerID, completeBlock)
				opBeingMined = false
			} else {
				miner.BlockMap.MineAndAddBlock(waitingBlocks[0],Configs.MinerID, completeBlock)
				waitingBlocks = waitingBlocks[1:]
				opBeingMined = true
			}

		// receive a new order to mine a block, select whether this block waits or goes forward
		case ib := <-miner.IncomingOps:
			// TODO: Determine whether to mine a new block
			waitingBlocks = append(waitingBlocks, ib)

			// noop block being mined, stop it and start this op block
			if !opBeingMined {
				miner.BlockMap.MineAndAddBlock(waitingBlocks[0],Configs.MinerID, completeBlock)
				waitingBlocks = waitingBlocks[1:]
				opBeingMined = true
			}

		// TODO: Client logic goes here
		}
	}
}


func main() {
    // load json settings
    plan, e := ioutil.ReadFile(os.Args[1])
    if e == nil {
		err := json.Unmarshal(plan, &Configs)
		if err != nil {
	    	log.Fatal("error reading json:", err)
        }
    } else {
        log.Fatal("file error:", e)
    }
    go rpcServer()
    time.Sleep(3000 * time.Millisecond)
    go rpcClient()
    time.Sleep(3000 * time.Millisecond)

    // send this thread to manage the state machine
	handleBlocks()

}

