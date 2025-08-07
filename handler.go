package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

func handleConn(conn net.Conn) {

	// Draw linux
	if _, err := conn.Write(initialMsg); err != nil {
		fmt.Println("Failed to send welcome:", err)
		return
	}

	// Loop until a valid username
	name := ""
	for {
		conn.Write([]byte("[ENTER YOUR NAME]: "))
		nameBuf := make([]byte, 1024)
		n, err := conn.Read(nameBuf)
		if err != nil {
			fmt.Println("Failed to read username:", err)
			return
		}
		name = strings.TrimSpace(string(nameBuf[:n]))

		if len(name) == 0 {
			continue
		}

		if _, exists := users[name]; exists {
			conn.Write([]byte("Name already taken.\n"))
			continue
		}

		users[name] = conn
		break
	}

	defer closeConn(name)

	// Send chat history
	for _, msg := range messages {
		conn.Write([]byte(msg))
	}

	// Broadcast join message
	joinMsg := fmt.Sprintf("%s has joined our chat...\n", name)
	messages = append(messages, joinMsg)

	// Message loop
	buf := make([]byte, 1024)
	for {
		n, err := conn.Read(buf)
		if err != nil || n == 1 {
			continue
		}

		text := strings.TrimSpace(string(buf[:n]))

		timestamp := time.Now().Format("2006-01-02 15:04:05")

		message := fmt.Sprintf("[%s][%s]: %s\n", timestamp, name, text)

		messages = append(messages, message)
	}
}

func closeConn(name string) {
	mu.Lock()
	defer mu.Unlock()

	conn := users[name]
	delete(users, name) // remove from map
	conn.Close()        // close the tcp conn

	leaveMsg := fmt.Sprintf("%s has left our chat...\n", name)
	messages = append(messages, leaveMsg)
}

// broadcast the most recent message
func startBroadcaster() {
	lastMsgCount := 0
	for {
		time.Sleep(100 * time.Millisecond) // avoid busy loop

		mu.Lock()
		if len(messages) > lastMsgCount {
			newMsg := messages[len(messages)-1]
			for _, user := range users {
				user.Write([]byte(newMsg))
			}
			lastMsgCount = len(messages)
		}
		mu.Unlock()
	}
}
