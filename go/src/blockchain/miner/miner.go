package main

import (
    "fmt"
    "os"
    "net"
    "bufio"
    "strings"
    "blockchain/miner/minerlib"
)

type miner struct {
}

type M interface {
}

var (
    HostIpPort string
)


func main() {
    HostIpPort = os.Args[1]
    ln, err := net.Listen("tcp", HostIpPort)
    if(err != nil){
	fmt.Println("failed to listen tcp",err)
        return
    }

    conn, err := ln.Accept()
    if(err != nil){
        fmt.Println("failed to accept messages",err)
        return
    }
    minerlib, err := minerlib.Initialize(conn) // TODO do I need to pass in connection?
    if(err != nil){
        fmt.Println("failed to init minerlib",err)
        return
    }

    var reply string
    for {
        // will listen for message to process ending in newline (\n)
	conn, err := ln.Accept()
	rawMessage, err := bufio.NewReader(conn).ReadString('\n')
        if(err != nil){
	    fmt.Println(err)
	    return
	}
	// output message received
        msg := string(rawMessage)
	fmt.Print("Message Received:", msg)

        if(strings.Contains(msg,"RFSINIT")){
	    fmt.Print("instruction RFSINIT")
	    reply = minerlib.CheckConnected()
	}

	// sample process for string received
	reply = strings.ToLower(reply) // all messages are lowercase
	// send new string back to client
	conn.Write([]byte(reply+ "\n"))
    }
}

