package server

import (
	"bytes"
	"encoding/binary"
	"net"
	"io"
	"fmt"
	"github.com/mauricio/gocached/store"
)

const (
	MESSAGE_HEADER_SIZE = 24

	REQUEST_KEY = 0x80
	RESPONSE_KEY = 0x81

	MESSAGE_GET = 0x00
	MESSAGE_SET = 0x01
	MESSAGE_DELETE = 0x04

	RESULT_OK = 0x0000
	RESULT_NOT_FOUND = 0x0001
	RESULT_EXISTS = 0x0002
	RESULT_ITEM_NOT_STORED = 0x0005

	ERROR_VALUE_TOO_LARGE = 0x0003
	ERROR_INVALID_ARGUMENTS = 0x0004
	ERROR_INCREMENT_DECREMENT_NON_NUMERIC = 0x0006
	ERROR_UNKNOWN = 0x0081
	ERROR_OUT_OF_MEMORY = 0x0082
)


type Server interface {
	Start() error
	Stop() error
}

type tcp_server struct {
	port int
	host string
	running bool
	storage store.Storage
	server net.Listener
}

func New(port int, host string, storage store.Storage) Server {
	return & tcp_server{
		port: port,
		host: host,
		storage: storage,
	}
}

func (s * tcp_server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	s.server = l

	if err == nil {
		s.running = true
		go s.AcceptClients()
	}

	return err
}

func (s * tcp_server) Stop() error {
	if s.running {
		s.running = false
		return s.server.Close()
	} else {
		return nil
	}
}

func (s * tcp_server) AcceptClients() {
	for s.running {
		connection, error := s.server.Accept()
		if error == nil {
			go s.HandleConnection(connection)
		} else {
			fmt.Errorf("Failed to accept client, stopping: %s\n", error.Error())
			s.running = false
		}
	}
}

func (s * tcp_server) HandleConnection(connection net.Conn) {
	defer connection.Close()
	buffer := make([]byte, MESSAGE_HEADER_SIZE)

	for s.running {

/**
| offset | description                                                           |
| 0      | magic number indicating if server or client packet                    |
| 1      | message type                                                          |
| 2-3    | size of the key in this message (if there is one)                     |
| 4      | extras length, some messages contain an extra field, that's it's size |
| 5      | data type, not in use                                                 |
| 6-7    | reserved field, not in use                                            |
| 8-11   | total message body size (this includes the key size as well)          |
| 12-15  | opaque field for operations that use it                               |
| 16-23  | CAS field for operations that use it                                  |
| 24-N   | bytes that symbolize the key that is being operated on                |
 */

		n, err := io.ReadFull(connection, buffer)

		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error reading from client:", err.Error())
			break
		} else {
			reader := bytes.NewReader(buffer)
			magic_number, _ := reader.ReadByte()
			message_type, _ := reader.ReadByte()
			key_size, _ := binary.ReadVarint(reader)


		}


	}


}
