package p2p

import (
	"fmt"
	"net"
	"sync"
)

// TCPPeer rep the remote node with whom conn is estbalished over TCP
type TCPPeer struct {
	// the underlying connection obj rep the conn between server and peer
	conn net.Conn

	// if we dial the conn and retrive the conn -> outbound(since we are making the conn)->>true
	// if we accept conn and retrive the conn -> outbound(since we are making the conn)->>false ---> inbound connecttion
	outbound bool
}
type TCPTransportOpts struct {
	ListenAddr string
	//make them capitalized to make these properties public
	Handshakefunc Handshake
	Decoder       Decoder
}
type TCPTransporter struct {
	TCPTransportOpts
	Listener net.Listener
	mu       sync.RWMutex
	peers    map[net.Addr]Peer
}

type TempMsg struct{}

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
	}
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
	peer := TCPPeer{
		conn:     conn,
		outbound: true,
	}

	// if the handshake has not been established then perhaps disconnect?
	if err := tr.Handshakefunc(peer); err != nil {
		fmt.Printf("handshake failed with the remote peer ; %v", err)
	}

	//if the conn is succesfful then decode the data being sent
	msg := &Message{}
	for {
		if err := tr.Decoder.Decode(conn, msg); err != nil {
			fmt.Printf("error occurred in decoding %v", err.Error())
			continue
		}

		//setting the remote addr of the msg sender
		msg.From = conn.RemoteAddr()

		fmt.Printf("message: %+v\n", msg)
	}
}
