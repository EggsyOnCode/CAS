package main

import (
	"fmt"
	"log"

	"github.com/EggsyOnCode/CAS/p2p"
)

func main() {
	tr := p2p.NewTCPTransporter(":3000")
	// this syntax means that the err recevied from the listen func ; test the con on it
	if err:=tr.ListenAndAccept(); err!=nil{
		log.Fatal(err)
	}
	fmt.Println("hello world")

	//we are blocking the thread here ; why?
	// we are blocking to keep the server alive  and keep listening
	select{}
}
