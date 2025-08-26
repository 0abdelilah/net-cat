package main

import (
	"fmt"
	"net"
	"os"
	"sync"
)

var initialMsg = []byte(`
Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    .        ' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     -'       '
`)

var (
	mu       sync.Mutex
	users    = make(map[string]net.Conn)
	maxConns = 10
)

func main() {
	// Get PORT
	port := ":8989"
	if len(os.Args) == 2 {
		port = ":" + os.Args[1]
	} else if len(os.Args) != 1 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}

	// Listen for connections
	listener, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Started listening on localhost" + port)

	go startBroadcaster()

	// Accept connections
	for {
		if len(users) >= maxConns {
			continue
		}

		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		go handleConn(conn)
	}
}
