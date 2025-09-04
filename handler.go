package main

import (
	"fmt"
	"io"
	"net"
	"strings"
	"time"
)

type Message struct {
	Sender  string
	Content string
}

func IsLetter(text string) bool {
	valid := true
	for _, r := range text {
		if r < 32 || r > 126 {
			valid = false
			break
		}
	}
	return valid
}

func handleConn(conn net.Conn) {
	// Reject if group full
	if len(users) >= maxConns {
		conn.Write([]byte("Group is full"))
		conn.Close()
		return
	}

	// Send welcome
	if _, err := conn.Write(initialMsg); err != nil {
		fmt.Println("Failed to send welcome:", err)
		return
	}

	// Ask for valid username
	var name string
	for {
		conn.Write([]byte("[ENTER YOUR NAME]: "))
		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("Failed to read username:", err)
			conn.Close()
			return
		}
		name = strings.TrimSpace(string(buf[:n]))
		if len(name) == 0 || !IsLetter(name) {
			conn.Write([]byte("Invalid name.\n"))
			continue
		}

		mu.Lock()
		if _, exists := users[name]; exists {
			mu.Unlock()
			conn.Write([]byte("Name already taken.\n"))
			continue
		}
		users[name] = conn
		mu.Unlock()
		break
	}

	// Ensure leave message runs no matter what
	defer closeConn(name)

	// Send chat history
	mu.Lock()
	for _, msg := range messages {
		conn.Write([]byte(msg.Content))
	}
	mu.Unlock()

	// Broadcast join
	joinMsg := Message{
		Sender:  name,
		Content: fmt.Sprintf("%s has joined our chat...\n", name),
	}
	mu.Lock()
	messages = append(messages, joinMsg)
	mu.Unlock()

	// Message loop
	buf := make([]byte, 1024)
	for {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		prompt := fmt.Sprintf("[%s][%s]: ", timestamp, name)
		conn.Write([]byte(prompt))

		n, err := conn.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println(name, "disconnected")
			} else {
				fmt.Println("Read error:", err)
			}
			break // trigger defer closeConn(name)
		}
		if n == 0 {
			continue
		}

		text := strings.TrimSpace(string(buf[:n]))
		if text == "" || !IsLetter(text) {
			conn.Write([]byte("Invalid message.\n"))
			continue
		}

		newMsg := Message{
			Sender:  name,
			Content: fmt.Sprintf("[%s][%s]: %s\n", timestamp, name, text),
		}
		mu.Lock()
		messages = append(messages, newMsg)
		mu.Unlock()
	}
}

func closeConn(name string) {
	mu.Lock()
	defer mu.Unlock()

	conn := users[name]

	// broadcast leave BEFORE closing connection
	leaveMsg := Message{
		Sender:  name,
		Content: fmt.Sprintf("%s has left our chat...\n", name),
	}
	messages = append(messages, leaveMsg)

	delete(users, name)
	conn.Close()
}

func broadcastMessage() {
	lastMsgCount := 0
	for {
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		if len(messages) > lastMsgCount {
			newMsg := messages[len(messages)-1]

			for uname, user := range users {
				if uname == newMsg.Sender {
					continue
				}

				// move to new line, print message
				user.Write([]byte("\033[s" + "\n" + newMsg.Content))

				// reprint prompt and protect it
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				prompt := fmt.Sprintf("[%s][%s]: ", timestamp, uname)

				user.Write([]byte(prompt))
				user.Write([]byte("\033[u\033[2B"))
			}

			lastMsgCount = len(messages)
		}
		mu.Unlock()
	}
}
