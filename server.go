package main

import (
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
	"time"
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
[ENTER YOUR NAME]:`)

var (
	mu    sync.Mutex
	conns = make(map[net.Conn]bool)
)

func main() {
	if len(os.Args) != 2 {
		fmt.Println("[USAGE]: ./TCPChat $port")
		return
	}
	port := ":" + os.Args[1]

	messages := make(map[int]string)
	var msgIndex int

	listener, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Started listening on localhost" + port)

	for len(conns) <= 10 {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Connection error:", err)
			continue
		}

		mu.Lock()
		conns[conn] = true
		mu.Unlock()

		go handleConn(conn, messages, &msgIndex)
	}
}

func handleConn(conn net.Conn, messages map[int]string, msgIndex *int) {
	defer closeConn(conn)

	// Ask for name
	if _, err := conn.Write(initialMsg); err != nil {
		fmt.Println("Failed to send welcome:", err)
		return
	}

	nameBuf := make([]byte, 1024)
	n, err := conn.Read(nameBuf)
	if err != nil {
		fmt.Println("Failed to read username:", err)
		return
	}
	name := strings.TrimSpace(string(nameBuf[:n]))

	// Broadcast join message
	joinMsg := fmt.Sprintf("%s has joined our chat...\n", name)
	broadcast([]byte(joinMsg), conn)

	// Send chat history
	mu.Lock()
	for j := 0; j <= *msgIndex; j++ {
		conn.Write([]byte(messages[j]))
	}
	mu.Unlock()

	// Message loop
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf("[%s][%s]: %s", timestamp, name, string(buf[:n]))

		mu.Lock()
		*msgIndex++
		messages[*msgIndex] = message
		mu.Unlock()

		fmt.Print(message)
		broadcast([]byte(message), conn)
	}

	// Leaving message
	leaveMsg := fmt.Sprintf("%s has left our chat...\n", name)
	broadcast([]byte(leaveMsg), conn)
}

func closeConn(conn net.Conn) {
	mu.Lock()
	delete(conns, conn)
	mu.Unlock()
	conn.Close()
}

func broadcast(message []byte, sender net.Conn) {
	mu.Lock()
	defer mu.Unlock()
	for conn := range conns {
		if conn != sender {
			conn.Write(message)
		}
	}
}
