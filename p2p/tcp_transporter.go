package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

// TCPPeer rep the remote node with whom conn is estbalished over TCP
type TCPPeer struct {
	// the underlying connection obj rep the conn between server and peer
	net.Conn

	// if we dial the conn and retrive the conn -> outbound(since we are making the conn)->>true
	// if we accept conn and retrive the conn -> outbound(since we are making the conn)->>false ---> inbound connecttion
	outbound bool
	wg       *sync.WaitGroup
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
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
	}
}

// implements the CloseStream method of the Peer interface
func (tp *TCPPeer) CloseStream() {
	tp.wg.Done()
}
func (tp *TCPPeer) Send(b []byte) error {
	n, err := tp.Conn.Write(b)
	if err != nil {
		return err
	}

	log.Printf("%v bytes sent by %s", n, tp.Conn.RemoteAddr().String())

	return nil
}

// here we could've returned Transpoort interface but that would've been bad for teesting because we wouldn't have had access to Listener etc objects
func NewTCPTransporter(opts TCPTransportOpts) *TCPTransporter {
	return &TCPTransporter{
		TCPTransportOpts: opts,
		tranch:           make(chan RPC, 1024),
	}
}

// this syntax means that we cna only read from the channel not write ot it
// Consume() <- chan means that its receive only channel
// same is the case when its used in teh struct def
func (tr *TCPTransporter) Consume() <-chan RPC {
	return tr.tranch
}
func (tr *TCPTransporter) Addr() string {
	return tr.ListenAddr
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

	log.Printf("TCP listener listening on port: %s", tr.ListenAddr)
	return nil
}

func (tr *TCPTransporter) Close() error {
	return tr.Listener.Close()
}

func (tr *TCPTransporter) startAcceptLoop() {
	for {
		conn, err := tr.Listener.Accept()

		// when the listener is closed
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			fmt.Printf("tcp accept conn error %s\n", err)
		}

		fmt.Printf("the incoming tcp connec %v\n", conn)
		go tr.handleConn(conn, false)
	}

}


// implements the Dial interface of Transporter
func (s *TCPTransporter) Dial(address string) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return err
	}

	go s.handleConn(conn, true)
	return nil
}

func (tr *TCPTransporter) handleConn(conn net.Conn, outbound bool) {
	var err error

	// this is gonna happen at the end of func regardless of any state?
	defer func() {
		fmt.Printf("dropping ur connection %v\n", err)
		conn.Close()
	}()

	peer := &TCPPeer{
		Conn:     conn,
		outbound: outbound,
		wg:       &sync.WaitGroup{},
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
	for {
		rpc := RPC{}
		err = tr.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		rpc.From = conn.RemoteAddr()
		//setting the remote addr of the rpc sender
		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Printf("[%s] waiting till stream is done\n", conn.RemoteAddr().String())
			peer.wg.Wait()
			fmt.Printf("[%s] stream is done; continue reading normal loop\n", conn.RemoteAddr().String())
			continue
		}
		tr.tranch <- rpc
	}
}
