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
	return
     }

    fmt.Println("list files")

    ls,err := rfs.ListFiles()
     if(err != nil){
        t.Errorf("error")
     }
     fmt.Println(ls)

     fmt.Println("creating file")
     err = rfs.CreateFile("test")
     if err != nil {
         fmt.Println(err)
     }
     var byteArray [512]byte
     copy(byteArray[:], "test record")
     var testRecord Record = byteArray
    fmt.Println("appending to file")

    recNum, err := rfs.AppendRec("test", &testRecord)
    if err != nil {
        fmt.Println(err)
    }
    fmt.Println(recNum)

    fmt.Println("list files")

    ls,err = rfs.ListFiles()
    if(err != nil){
        t.Errorf("error")
    }
    fmt.Println(ls)
}

