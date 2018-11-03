package main

import (
	"blockchain/minerlib"
	"blockchain/minerlib/blockmap"
	"blockchain/rfslib"
	"encoding/json"
	"fmt"
	"github.com/DistributedClocks/GoVector/govec"
	"github.com/DistributedClocks/GoVector/govec/vrpc"
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"os"
	"sort"
	"time"
)


type M interface {
	MakeKnown(addr string, reply *int) error
	ReceiveOp(operation minerlib.Op, reply *int) error
	ReceiveBlock(payload Payload, reply *int) error
	GetPreviousBlock(prevHash string, block *blockmap.Block) error
}

var (
    GovecOptions = govec.GetDefaultLogOptions()
    miner *Miner
    MinerLogger *govec.GoLog
    Configs minerlib.Settings
    blockStartTime int
    maxOps = 1
    confirmationCheckDelay time.Duration = 3
    confirmationTimeout time.Duration = 30
)

type Miner struct {
	Connections map[string]*rpc.Client
	WaitingOps map[int]minerlib.Op
	BlockMap blockmap.BlockMap
	IncomingOps chan []minerlib.Op
}

type Payload struct {
	ReturnAddr string
	Block blockmap.Block
}

type AppendReply struct {
	RecordNum int
	Err error
}

// returns 1 if miner is connected else 0
func (miner *Miner) IsConnected(clientId string ,res *string) error {
     // fmt.Println("client connection id:", clientId)
     // fmt.Println(len(miner.Connections))
     if len(miner.Connections) != 0 {
        *res = Configs.MinerID
     } else{
        *res = "disconnected"
     }

     return nil
}



func (miner *Miner) Ls(op minerlib.Op, reply *rfslib.LsReply) error {
     if(len(miner.Connections) == 0){
	reply.Err = rfslib.DisconnectedError(Configs.MinerID)
     } else{
	reply.Files = miner.BlockMap.LS()
     }

     return nil
}

func (miner *Miner) Cat(op minerlib.Op, reply *rfslib.RecordsReply) error {
     if(len(miner.Connections) == 0){
        reply.Err = rfslib.DisconnectedError(Configs.MinerID)
     } else if(!miner.BlockMap.CheckIfFileExists(op.Fname)){
	reply.Err = rfslib.FileDoesNotExistError(op.Fname)
     } else{
        reply.Records = miner.BlockMap.Cat(op.Fname)
     }
     return nil
}

func (miner *Miner) Tail(op minerlib.Op, reply *rfslib.RecordsReply) error {
     if(len(miner.Connections) == 0){
        reply.Err = rfslib.DisconnectedError(Configs.MinerID)
     } else if(!miner.BlockMap.CheckIfFileExists(op.Fname)){
        reply.Err = rfslib.FileDoesNotExistError(op.Fname)
     } else{
        reply.Records = miner.BlockMap.Tail(op.K,op.Fname)
     }
     return nil
}

func (miner *Miner) Head(op minerlib.Op, reply *rfslib.RecordsReply) error {
     if len(miner.Connections) == 0 {
        reply.Err = rfslib.DisconnectedError(Configs.MinerID)
     } else if !miner.BlockMap.CheckIfFileExists(op.Fname) {
        reply.Err = rfslib.FileDoesNotExistError(op.Fname)
     } else{
        reply.Records = miner.BlockMap.Head(op.K, op.Fname)
     }
     return nil
}

func (miner *Miner) Touch(op minerlib.Op, reply *error) error {
	op.SeqNum = int(time.Now().UnixNano())

	err := miner.BlockMap.ValidateOp(op)
	if err != nil {
		switch err.(type) {
		case rfslib.FileDoesNotExistError:
			*reply = rfslib.FileDoesNotExistError(op.Fname)
		case rfslib.FileExistsError:
			*reply = rfslib.FileExistsError(op.Fname)
		case rfslib.BadFilenameError:
			*reply = rfslib.BadFilenameError(op.Fname)
		}
		return nil
	}

	miner.ReceiveOp(op, nil)

	for true {
		time.Sleep(confirmationCheckDelay * time.Second)

		select {

		case <-time.After(confirmationTimeout * time.Second):
			fmt.Println("sending touch op again &&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
			miner.ReceiveOp(op, nil)

		default:
			if miner.BlockMap.CheckIfOpIsConfirmed(op) == 1 {
				return nil
			}
			if miner.BlockMap.CheckIfOpIsConfirmed(op) == -1 {
				return rfslib.FileExistsError(op.Fname)
			}
		}
	}
	return nil
}

func (miner *Miner) Append(op minerlib.Op, reply *AppendReply) error {
	op.SeqNum= int(time.Now().UnixNano())

	err := miner.BlockMap.ValidateOp(op)
	if err != nil {
		fmt.Println(err)
		switch err.(type) {
		case rfslib.FileDoesNotExistError:
			*reply = AppendReply{-1,rfslib.FileDoesNotExistError(op.Fname)}
		case rfslib.FileExistsError:
			*reply = AppendReply{-1,rfslib.FileExistsError(op.Fname)}
		case rfslib.BadFilenameError:
			*reply = AppendReply{-1,rfslib.BadFilenameError(op.Fname)}
		}
		return nil
	}

	miner.ReceiveOp(op, nil)

	for true {
		time.Sleep(confirmationCheckDelay * time.Second)

		select {

		case <-time.After(confirmationTimeout * time.Second):
			fmt.Println("timeout, sending op again")
			miner.ReceiveOp(op, nil)
		default:
			if miner.BlockMap.CheckIfOpIsConfirmed(op) == 1 {
				fmt.Println("Append op confirmed")
				*reply = AppendReply{miner.BlockMap.GetRecordPosition(op.SeqNum, op.Fname), nil}
				return nil
			}
			if miner.BlockMap.CheckIfOpIsConfirmed(op) == -1 {
				*reply = AppendReply{0, rfslib.FileExistsError(op.Fname)}
				return nil
			}
		}
	}
	return nil
}


func (miner *Miner) MakeKnown(addr string, reply *int) error {
	if _, ok := miner.Connections[addr]; !ok {
		MinerLogger.LogLocalEvent("MakeKnown Called", GovecOptions)
		client, err := vrpc.RPCDial("tcp", addr, MinerLogger, GovecOptions)
		if err == nil {
			miner.Connections[addr] = client
			// fmt.Println("connecting client addr: " + addr)
		} else {
			log.Println("dialing:", err)
		}
	}
	return nil
}

func (miner *Miner) ReceiveOp(operation minerlib.Op, reply *int) error {

	if _, ok := miner.WaitingOps[operation.SeqNum]; !ok {
		// op not received yet, store and flood it
		miner.WaitingOps[operation.SeqNum] = operation

		for _, conn := range miner.Connections {
			conn.Go("Miner.ReceiveOp", operation, nil, nil)
		}

		// reset timeout counter
		if blockStartTime == -1 {
			blockStartTime = time.Now().Second()
		}

		// block timeout or block full, send what you have
		if Configs.GenOpBlockTimeout < uint8(time.Now().Second() - blockStartTime) || len(miner.WaitingOps) >= maxOps {
			// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ creating new op block @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
			blockStartTime = -1
			var newOps []minerlib.Op
			for _, o := range miner.WaitingOps {
				// fmt.Println(o.Op)
				newOps = append(newOps, o)
			}
			// fmt.Println(operation.Op)

			// fmt.Println(newOps[0].Op)

			sort.Slice(newOps, func(i, j int) bool {
				return newOps[i].SeqNum < newOps[j].SeqNum
			})

			// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ creating new op block @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
			// fmt.Println(newOps)
			miner.IncomingOps <- newOps

			//miner.WaitingOps = make(map[int]minerlib.Op)

			// TODO: Add listener to check for op in longest chain with # blocks in front of it

		}
	}
	return nil
}

func (miner *Miner) ReceiveBlock(payload Payload, reply *int) error{
	// fmt.Println("return address in payload: ", payload.ReturnAddr)

	// if miner is behind, get previous BlockMap.Map until caught up
	if _, ok := miner.BlockMap.Map[payload.Block.PrevHash]; !ok {

		// add missing BlockMap.Map to a temp store in case they're fake
		missingBlocks := make([]blockmap.Block, 0)
		missingBlocks = append( missingBlocks, payload.Block)
		fmt.Println(missingBlocks)
		var prevBlock *blockmap.Block
		prevBlock = &payload.Block
		for !ok {
			// fmt.Println(prevBlock)
			miner.Connections[payload.ReturnAddr].Call("Miner.GetPreviousBlock", prevBlock.PrevHash, &prevBlock)
			// fmt.Println("return block")
			// fmt.Println(prevBlock)
			missingBlocks = append(missingBlocks, *prevBlock)

			_, ok = miner.BlockMap.Map[prevBlock.PrevHash]
		}
		fmt.Println(missingBlocks)
		// insert all the blocks in missingBlocks into the blockmap
		for i := range missingBlocks {
			fmt.Println(missingBlocks[len(missingBlocks)-i-1])
			err := miner.BlockMap.ValidateOps(missingBlocks[len(missingBlocks)-i-1].Ops)
			if err == nil {
				fmt.Println("Inserting catch up block %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
				miner.BlockMap.Insert(missingBlocks[len(missingBlocks)-i-1])
				for _, conn := range miner.Connections {
					conn.Go("Miner.ReceiveBlock", Payload{Configs.IncomingMinersAddr, payload.Block}, nil, nil)
				}			} else {
				missingBlocks = make([]blockmap.Block, 0)
				break
			}
		}

	} else if _, ok := miner.BlockMap.Map[blockmap.GetHash(payload.Block)]; !ok {
		fmt.Println("single block insert #####################################################################################################################################")

		err := miner.BlockMap.ValidateOps(payload.Block.Ops)
		if err == nil {
			fmt.Println("Inserting single block %%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%%")
			miner.BlockMap.Insert(payload.Block)
			for _, conn := range miner.Connections {
				conn.Go("Miner.ReceiveBlock", Payload{Configs.IncomingMinersAddr, payload.Block}, nil, nil)
			}
		}
		// fmt.Println("current block chain state:",miner.BlockMap.GetLongestChain())
		// send the block to connected miners
		// fmt.Println(miner.Connections)
	}

	return nil
}

func (miner *Miner) GetPreviousBlock(prevHash string, block *blockmap.Block) error {

	if _, ok := miner.BlockMap.Map[prevHash]; ok {
		// fmt.Println(miner.BlockMap.Map[prevHash])
		*block = miner.BlockMap.Map[prevHash]
	}
	return nil
}

func rpcServer() {
    // fmt.Println("Starting rpc server")
    miner = new(Miner)
    MinerLogger = govec.InitGoVector(Configs.MinerID, "./logs/minerlogfile" + Configs.MinerID, govec.GetDefaultConfig())
    miner.BlockMap = blockmap.Initialize(Configs, blockmap.Block{ PrevHash: "GENESIS", Nonce:"GENESIS" , MinerId:"GENESIS"})
    miner.WaitingOps = make(map[int]minerlib.Op)
    miner.Connections = make(map[string]*rpc.Client)
    miner.IncomingOps = make(chan []minerlib.Op)
    server := rpc.NewServer()
    server.Register(miner)
    l, e := net.Listen("tcp", Configs.IncomingMinersAddr)
    if e != nil {
        log.Fatal("listen error:", e)
    }
    vrpc.ServeRPCConn(server, l, MinerLogger, GovecOptions)
}

func rpcClient(){
    // fmt.Println("Starting rpc client")
    for _, addr := range Configs.PeerMinersAddrs {
        client, err := vrpc.RPCDial("tcp", addr, MinerLogger, GovecOptions)
        if err == nil {
        	// make this miner known to the other miner
	    	var result int
	    	err := client.Go("Miner.MakeKnown", Configs.IncomingMinersAddr, &result, nil)
	    	if err != nil {
				fmt.Println(err)
			}
	    	miner.Connections[addr] = client
	    	// fmt.Println("addr added: ", addr)
        } else {log.Println("dialing:", err)}
    }
}

func handleBlocks () {

	// create a noop block and start mining for a nonce
	completeBlock := make(chan blockmap.Block)
	go miner.BlockMap.MineAndAddNoOpBlock(Configs.MinerID, completeBlock)
	// fmt.Println("mining block now")

	waitingBlocks := make([][]minerlib.Op, 0)

	opBeingMined := false

	for true {
		if len(miner.Connections) == 0 {
			time.Sleep( 1 * time.Second)
			fmt.Println("disconnected")
			continue
		}
		// fmt.Println("waiting for block")
		select {
		// no miners connected, do nothing

		// receive a newly mined block, flood it and start mining noop
		case cb := <-completeBlock:
			fmt.Println("mined block received")
			// fmt.Println("miner connections: ", len(miner.Connections))
			// fmt.Println(miner.Connections)

			// remove ops in this block from waitingOps
			for sn, o := range miner.WaitingOps {
				for _, ob := range cb.Ops {
					if o == ob {
						delete(miner.WaitingOps, sn)
					}
				}
			}
			for _, conn := range miner.Connections {
				var reply int
				conn.Go("Miner.ReceiveBlock", Payload{Configs.IncomingMinersAddr, cb}, &reply, nil)
			}
			// start on noop or queued op block right away
			if len(waitingBlocks) == 0 {
				go miner.BlockMap.MineAndAddNoOpBlock(Configs.MinerID, completeBlock)
				opBeingMined = false
			} else {
				go miner.BlockMap.MineAndAddOpBlock(waitingBlocks[0],Configs.MinerID, completeBlock)
				// fmt.Println("mining block now")
				waitingBlocks = waitingBlocks[1:]
				fmt.Println("waiting blocks: ", waitingBlocks)
				opBeingMined = true
			}

		// receive a new order to mine a block, select whether this block waits or goes forward
		case ib := <-miner.IncomingOps:
			waitingBlocks = append(waitingBlocks, ib)
			// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ added op block to waiting blocks @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
			// noop block being mined, stop it and start this op block
			if !opBeingMined {
				// fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@ starting to mine op block @@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@")
				go miner.BlockMap.MineAndAddOpBlock(waitingBlocks[0],Configs.MinerID, completeBlock)
				waitingBlocks = waitingBlocks[1:]
				opBeingMined = true
			}
		// optimization: if longest chain changes then stop mining and recreate the current block with ops in longest chain not included
		// case rb := <- miner.redoBlock:
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
