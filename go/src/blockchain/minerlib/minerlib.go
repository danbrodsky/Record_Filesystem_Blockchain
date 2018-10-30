package minerlib

import (
    "blockchain/rfslib"
)


type Settings struct {
    MinedCoinsPerOpBlock   uint8  `json:"MinedCoinsPerOpBlock"`
    MinedCoinsPerNoOpBlock uint8  `json:"MinedCoinsPerNoOpBlock"`
    NumCoinsPerFileCreate  uint8  `json:"NumCoinsPerFileCreate"`
    GenOpBlockTimeout      uint8  `json:"GenOpBlockTimeout"`
    GenesisBlockHash       string `json:"GenesisBlockHash"`
    PowPerOpBlock          uint8  `json:"PowPerOpBlock"`
    PowPerNoOpBlock        uint8  `json:"PowPerNoOpBlock"`
    ConfirmsPerFileCreate  uint8  `json:"ConfirmsPerFileCreate"`
    ConfirmsPerFileAppend  uint8  `json:"ConfirmsPerFileAppend"`
    MinerID             string   `json:"MinerID"`
    PeerMinersAddrs     []string `json:"PeerMinersAddrs"`
    IncomingMinersAddr  string   `json:"IncomingMinersAddr"`
    OutgoingMinersIP    string   `json:"OutgoingMinersIP"`
    IncomingClientsAddr string   `json:"IncomingClientsAddr"`
}

// possible ops {ls,cat,tail,head,append,touch}
type Op struct {
    Op string
    K int
    Fname string
    Rec rfslib.Record
    MinerId string
    SeqNum int
}

type File struct {

}

func ValidateBlock() bool {
    return true
}

func ValidateOp() bool {
    return true
}


