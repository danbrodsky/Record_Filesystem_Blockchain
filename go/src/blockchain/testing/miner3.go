package testing

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
)

type M interface {
	MakeKnown(addr string, reply *int) error
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
	*reply = 1
	fmt.Println("dialing:")
	return nil
}


func main() {

	miner := new(Miner)
	//miner.PeerMinersAddrs = append(miner.PeerMinersAddrs, "127.0.0.1:2000")
	//// load json into miner
	//file, e := os.Open(os.Args[1])
	//if e == nil {
	//	decoder := gob.NewDecoder(file)
	//	e = decoder.Decode(miner)
	//} else { log.Fatal("file error:", e)}
	//file.Close()


	// Open RPC server for other miners
	rpc.Register(miner)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", "localhost:3333")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)


	// connect to all known miners
	for _, addr := range miner.PeerMinersAddrs {
		client, err := rpc.DialHTTP("tcp", addr)
		if err == nil {
			fmt.Println("dialed")
			// make this miner known to the other miner
			var reply int
			client.Go("miner.MakeKnown", addr, reply, nil)
			fmt.Println(reply)
			minerConnections[addr] = client

		} else {			fmt.Println("not dialed")
			log.Println("dialing:", err)}
	}
}

