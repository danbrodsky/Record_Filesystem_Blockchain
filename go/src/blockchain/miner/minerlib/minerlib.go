package minerlib


import (
    "net"
)

type Minerlib interface{
    CheckConnected()(string)
}

type Miner struct {

}

type Block struct {
PrevHash string
Ops []string
Nonce string
MinerId string
Depth int
}


var (
    connection net.Conn
)

func Initialize(conn net.Conn) (miner Miner, err error) {
    connection = conn
    miner = Miner{}
    return miner,err
}

func (miner Miner) CheckConnected()(string){
    return "true"
}

func ValidateBlock() bool {
    return true
}

func ValidateOp() bool {
    return true
}

