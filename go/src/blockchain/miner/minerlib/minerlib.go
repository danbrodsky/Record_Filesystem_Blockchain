package minerlib


import (
    "net"
)

type Minerlib interface{
    CheckConnected()(string)
}

type Miner struct {

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


