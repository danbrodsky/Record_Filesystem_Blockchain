/*

This package specifies the application's interface to the distributed
records system (RFS) to be used in project 1 of UBC CS 416 2018W1.

You are not allowed to change this API, but you do have to implement it.

*/

package rfslib

import (
    "fmt"
    "net"
	"net/rpc"
	"time"
    "math/rand"

    "github.com/DistributedClocks/GoVector/govec"
    "github.com/DistributedClocks/GoVector/govec/vrpc"
)

type Op struct {
    Op string
    K int
    Fname string
    Rec Record
    MinerId string
    SeqNum int
}
var (
    GovecOptions = govec.GetDefaultLogOptions()
    RfsLogger *govec.GoLog
    serial string
    letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456790123456790123456790123456790")
    minerConnection *rpc.Client
    minerId string
)

func randSeq(n int) string {
    b := make([]rune, n)
    for i := range b {
        b[i] = letters[rand.Intn(len(letters))]
    }
    return string(b)
}


type LsReply struct{
    Err   error
    Files map[string]int
}

// A Record is the unit of file access (reading/appending) in RFS.
type Record [512]byte

type RecordsReply struct{
    Err     error
    Records []Record
}

type AppendReply struct {
	RecordNum int
	Err error
}


////////////////////////////////////////////////////////////////////////////////////////////
// <ERROR DEFINITIONS>

// These type definitions allow the application to explicitly check
// for the kind of error that occurred. Each API call below lists the
// errors that it is allowed to raise.
//
// Also see:
// https://blog.golang.org/error-handling-and-go
// https://blog.golang.org/errors-are-values

// Contains minerAddr
type DisconnectedError string

func (e DisconnectedError) Error() string {
	return fmt.Sprintf("RFS: Disconnected from the miner [%s]", string(e))
}

// Contains filename. The *only* constraint on filenames in RFS is
// that must be at most 64 bytes long.
type BadFilenameError string

func (e BadFilenameError) Error() string {
	return fmt.Sprintf("RFS: Filename [%s] has the wrong length", string(e))
}

// Contains filename
type FileDoesNotExistError string

func (e FileDoesNotExistError) Error() string {
	return fmt.Sprintf("RFS: Cannot open file [%s] in D mode as it does not exist locally", string(e))
}

// Contains filename
type FileExistsError string

func (e FileExistsError) Error() string {
	return fmt.Sprintf("RFS: Cannot create file with filename [%s] as it already exists", string(e))
}

// Contains filename
type FileMaxLenReachedError string

func (e FileMaxLenReachedError) Error() string {
	return fmt.Sprintf("RFS: File [%s] has reached its maximum length", string(e))
}

// </ERROR DEFINITIONS>
////////////////////////////////////////////////////////////////////////////////////////////

type RecordsFileSystem struct {
}

var (
    MinerAddr string
    LocalAddr string
    initFlag = false
    Conn net.Listener
    MinerConn net.Conn
)



// Represents a connection to the RFS system.
type RFS interface {
	// Creates a new empty RFS file with name fname.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileExistsError
	// - BadFilenameError
	CreateFile(fname string) (err error)

	// Returns a slice of strings containing filenames of all the
	// existing files in RFS.
	//
	// Can return the following errors:
	// - DisconnectedError
	ListFiles() (fnames []string, err error)

	// Returns the total number of records in a file with filename
	// fname.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileDoesNotExistError
	TotalRecs(fname string) (numRecs uint16, err error)

	// Reads a record from file fname at position recordNum into
	// memory pointed to by record. Returns a non-nil error if the
	// read was unsuccessful. If a record at this index does not yet
	// exist, this call must block until the record at this index
	// exists, and then return the record.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileDoesNotExistError
	ReadRec(fname string, recordNum uint16, record *Record) (err error)

	// Appends a new record to a file with name fname with the
	// contents pointed to by record. Returns the position of the
	// record that was just appended as recordNum. Returns a non-nil
	// error if the operation was unsuccessful.
	//
	// Can return the following errors:
	// - DisconnectedError
	// - FileDoesNotExistError
	// - FileMaxLenReachedError
	AppendRec(fname string, record *Record) (recordNum uint16, err error)
}

// The constructor for a new RFS object instance. Takes the miner's
// IP:port address string as parameter, and the localAddr which is the
// local IP:port to use to establish the connection to the miner.
//
// The returned rfs instance is singleton: an application is expected
// to interact with just one rfs at a time.
//
// This call should only succeed if the connection to the miner
// succeeds. This call can return the following errors:
// - Networking errors related to localAddr or minerAddr
func Initialize(localAddr string, minerAddr string) (rfs RFS, err error) {
	rand.Seed(time.Now().UnixNano())
	serial = randSeq(10)
	MinerAddr = minerAddr
	LocalAddr = localAddr
	RfsLogger = govec.InitGoVector("client", "./logs/rfslogfile" , govec.GetDefaultConfig())
	err = checkIfConnected(minerAddr)
	if(err != nil){
	    return nil, err
	} else{
	    rfs = RecordsFileSystem{}
            return rfs,nil
	}
}

func checkIfConnected(minerAddr string) error{
    fmt.Println("Checking if connected")
    client, err := vrpc.RPCDial("tcp", minerAddr, RfsLogger, GovecOptions)
    if err == nil {
        // make this miner known to the other miner
        var result string
        err := client.Call("Miner.IsConnected", serial,&result)
	if result != "disconnected" {
		minerConnection = client
		minerId = result
	    return nil
	} else {
	    return DisconnectedError(minerAddr)
	}
        fmt.Println(err)
    } else {fmt.Println("dialing:", err)}
    return err
}


func CloseConnection(){
	MinerConn.Close()
}

func checkIfFileExists(fname string) (bool, error) {
	var reply *bool
	err := minerConnection.Call("Miner.CheckForFile", fname, &reply)
	if err != nil {
		return false, err
	}
	return *reply, err
}

// Creates a new empty RFS file with name fname.
//
// Can return the following errors:
// - DisconnectedError
// - FileExistsError
// - BadFilenameError
func (rfs RecordsFileSystem) CreateFile(fname string) (err error){
	newOp := Op{"touch", -1, fname, Record{}, minerId, 0}
	var reply error
	err = minerConnection.Call("Miner.Touch", newOp, &reply)
	if err != nil {
		return err
	}
	if reply != nil {
		return reply
	}
    return nil
}

// Returns a slice of strings containing filenames of all the
// existing files in RFS.
//
// Can return the following errors:
// - DisconnectedError
func (rfs RecordsFileSystem) ListFiles() (fnames []string, err error){
    client, err := vrpc.RPCDial("tcp", MinerAddr, RfsLogger, GovecOptions)
    if err == nil {
        // make this miner known to the other miner
	var rep LsReply
	op := Op{Op:"ls"}
        err := client.Call("Miner.Ls", op, &rep)
        if(rep.Err == nil){
	    files := []string{}
            for k := range rep.Files{
                files = append(files, k)
            }
            return files,nil
        } else {
            return nil, rep.Err
        }
        fmt.Println(err)
    } else {fmt.Println("dialing:", err)}
    return nil, err //STUB TODO
}

// Returns the total number of records in a file with filename
// fname.
//
// Can return the following errors:
// - DisconnectedError
// - FileDoesNotExistError
func (rfs RecordsFileSystem) TotalRecs(fname string) (numRecs uint16, err error){
    return 1, nil //STUB TODO
}

// Reads a record from file fname at position recordNum into
// memory pointed to by record. Returns a non-nil error if the
// read was unsuccessful. If a record at this index does not yet
// exist, this call must block until the record at this index
// exists, and then return the record.
//
// Can return the following errors:
// - DisconnectedError
// - FileDoesNotExistError
func (rfs RecordsFileSystem) ReadRec(fname string, recordNum uint16, record *Record) (err error){
    return nil
}

// Appends a new record to a file with name fname with the
// contents pointed to by record. Returns the position of the
// record that was just appended as recordNum. Returns a non-nil
// error if the operation was unsuccessful.
//
// Can return the following errors:
// - DisconnectedError
// - FileDoesNotExistError
// - FileMaxLenReachedError
func (rfs RecordsFileSystem) AppendRec(fname string, record *Record) (recordNum uint16, err error){
	newOp := Op{"append", -1, fname, *record, minerId, 0}
	var reply AppendReply
	err = minerConnection.Call("Miner.Append", newOp, &reply)
	if err != nil {
		return 0, err
	}
	if reply.Err != nil {
		return 0, reply.Err
	}
	return uint16(reply.RecordNum), nil

}
