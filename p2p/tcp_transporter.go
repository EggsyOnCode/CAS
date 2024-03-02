package p2p

import (
	"fmt"
	"net"
)

// TCPPeer rep the remote node with whom conn is estbalished over TCP
type TCPPeer struct {
	// the underlying connection obj rep the conn between server and peer
	conn net.Conn

	// if we dial the conn and retrive the conn -> outbound(since we are making the conn)->>true
	// if we accept conn and retrive the conn -> outbound(since we are making the conn)->>false ---> inbound connecttion
	outbound bool
}

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr string
	//make them capitalized to make these properties public
	Handshakefunc Handshake
	Decoder       Decoder
	//onpeer method is called when a tranporter receives a new peer ; then it can regiser it , save it  or do whatever it can with it
	OnPeer func(Peer) error
}
type TCPTransporter struct {
	TCPTransportOpts
	Listener net.Listener
	// tranport rpc
	tranch chan RPC
}

// constructor for TCPPeer
func NewTPCPeer(conn net.Conn, outbound bool) *TCPPeer {
	return &TCPPeer{
		conn:     conn,
		outbound: outbound,
	}

}

// here we could've returned Transpoort interface but that would've been bad for teesting because we wouldn't have had access to Listener etc objects
func NewTCPTransporter(opts TCPTransportOpts) *TCPTransporter {
	return &TCPTransporter{
		TCPTransportOpts: opts,
		tranch:           make(chan RPC),
	}
}

// this syntax means that we cna only read from the channel not write ot it
func (tr *TCPTransporter) Consume() <-chan RPC {
	return tr.tranch
}

func (tr *TCPTransporter) ListenAndAccept() error {
	var err error
	//listen to the address speciified int tcp.listenAddr
	tr.Listener, err = net.Listen("tcp", tr.ListenAddr)
	if err != nil {
		fmt.Println(err.Error())
	}
	//launch a goroutine to handle these new connections
	go tr.startAcceptLoop()
	return nil
}

func (tr *TCPTransporter) startAcceptLoop() {
	for {
		conn, err := tr.Listener.Accept()
		if err != nil {
			fmt.Printf("tcp accept conn error %s\n", err)
		}

		fmt.Printf("the incoming tcp connec %v\n", conn)
		go tr.handleConn(conn)
	}

}

func (tr *TCPTransporter) handleConn(conn net.Conn) {
	var err error

	// this is gonna happen at the end of func regardless of any state?
	defer func() {
		fmt.Printf("dropping ur connection \n", err)
		conn.Close()
	}()

	peer := &TCPPeer{
		conn:     conn,
		outbound: true,
	}

	// if the handshake has not been established then perhaps disconnect?
	if err := tr.Handshakefunc(peer); err != nil {
		fmt.Printf("handshake failed with the remote peer ; %v", err)
	}
	// test the OnPeer method
	if tr.OnPeer != nil {
		if err = tr.OnPeer(peer); err != nil {
			return
		}
	}

	//if the conn is succesfful then decode the data being sent
	rpc := RPC{}
	for {
		err = tr.Decoder.Decode(conn, &rpc)
		if err != nil{
			return
		}
		//setting the remote addr of the rpc sender
		rpc.From = conn.RemoteAddr()
		tr.tranch <- rpc
	}
}
