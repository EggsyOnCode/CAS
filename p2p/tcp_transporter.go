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
type TCPTransporter struct {
	ListenAddr string
	Listener   net.Listener

	mu    sync.RWMutex
	peers map[net.Addr]Peer
}

// here we could've returned Transpoort interface but that would've been bad for teesting because we wouldn't have had access to Listener etc objects
func NewTCPTransporter(listenAddr string) *TCPTransporter {
	return &TCPTransporter{
		ListenAddr: listenAddr,
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

		go tr.handleConn(&conn)
	}

}

func (tr *TCPTransporter) handleConn(conn *net.Conn) {
	peer := TCPPeer{
		conn: *conn,
		outbound: true,
	}
	

	fmt.Printf("the incoming tcp connec %v\n", peer)
}
