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
<<<<<<< HEAD
        Op string
=======
    Op string
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
	K int
	Fname string
	Rec rfslib.Record
	MinerId string
<<<<<<< HEAD
        // random string to make ops unique
        Id string
=======
    SeqNum int
>>>>>>> 48658500cf8f7842c4336279188a8d016bd794a1
}

type File struct {

}

func ValidateBlock() bool {
    return true
}

func ValidateOp() bool {
    return true
}


