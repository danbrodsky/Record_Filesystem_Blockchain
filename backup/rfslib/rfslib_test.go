package rfslib

import(
    "testing"
)

var (
        rfs RFS
)

func TestInitialize(t *testing.T) {
    localMiner := "127.0.0.1:8080"
    remoteMiner := "127.0.0.1:9090"
    rfs,err := Initialize(localMiner,remoteMiner)
    _= rfs
    CloseConnection()
    if(err != nil){
	t.Errorf("error")
     }
}

