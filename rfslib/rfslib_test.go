package rfslib

import(
    "testing"
    "fmt"
)

var (
        rfs RFS
)

func TestInitialize(t *testing.T) {
    localAddr := "127.0.0.1:8080"
    remoteMiner := "127.0.0.1:5050"
    rfs,err := Initialize(localAddr,remoteMiner)
    if(err != nil){
	t.Errorf("error")
     }
     ls,err := rfs.ListFiles()
     if(err != nil){
        t.Errorf("error")
     }
     fmt.Println(ls)

}

