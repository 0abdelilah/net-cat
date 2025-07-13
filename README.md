# Net-Cat 🐾

A Go-based reimplementation of the classic NetCat (`nc`) utility in a server-client architecture. Net-Cat enables TCP-based group chat functionality, supporting up to 10 simultaneous clients with persistent chat history and real-time user notifications.

---

## 🎯 Objectives

- Maintain a log of messages per session.

---

## 🚀 Features

- Real-time group chat over a TCP server.
- TCP connection (1 server : 10 concurent clients)
- Chat messages with timestamp and username  
  `e.g. [2020-01-20 15:48:41][Lee]: Hello!`
- Message history sent to newly joined clients
- Notifications when users join/leave
- Error handling on both client and server

---

## 🔧 Usage

### Start Server

```bash
$ go run .         # Defaults to port 8989
$ go run . 2525    # Custom port
