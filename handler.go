package main

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Message struct {
	Sender  string
	Content string
	Time    string
}

var messages []Message

// new messages with current timing

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

		if !IsLetter(name) {
			conn.Write([]byte("Invalid name.\n"))
			continue
		}

		mu.Lock()
		if _, exists := users[name]; exists {
			conn.Write([]byte("Name already taken.\n"))
			mu.Unlock()
			continue
		}

		users[name] = conn
		mu.Unlock()
		break
	}

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
		Time:    time.Now().Format("2006-01-02 15:04:05"),
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
		if err != nil || n == 0 {
			continue
		}

		text := strings.TrimSpace(string(buf[:n]))
		if text == "" {
			continue
		}

		if !IsLetter(text) {
			conn.Write([]byte("Invalid message.\n"))
			continue
		}

		newMsg := Message{
			Sender:  name,
			Content: fmt.Sprintf("[%s][%s]: %s\n", timestamp, name, text),
			Time:    timestamp,
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
	delete(users, name)
	conn.Close()

	messages = append(messages, Message{
		Sender:  name,
		Content: fmt.Sprintf("%s has left our chat...\n", name),
		Time:    time.Now().Format("2006-01-02 15:04:05"),
	})
}

func startBroadcaster() {
	lastMsgCount := 0
	for {
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		// inside startBroadcaster()
		if len(messages) > lastMsgCount {
			newMsg := messages[len(messages)-1]
			for uname, user := range users {
				if uname == newMsg.Sender {
					continue
				}
				// send the new message
				user.Write([]byte("\n" + newMsg.Content))

				// reprint prompt for user
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				prompt := fmt.Sprintf("[%s][%s]: ", timestamp, uname)

				user.Write([]byte(prompt))
			}
			lastMsgCount = len(messages)
		}
		mu.Unlock()
	}
}
