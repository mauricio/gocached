package server

import (
	"encoding/binary"
	"fmt"
	"github.com/mauricio/gocached/store"
	"net"
	"io"
)

const (
	REQUEST_KEY  = byte(0x80)
	RESPONSE_KEY = byte(0x81)

	MESSAGE_GET    = byte(0x00)
	MESSAGE_SET    = byte(0x01)
	MESSAGE_DELETE = byte(0x04)

	RESULT_OK              = byte(0x0000)
	RESULT_NOT_FOUND       = byte(0x0001)
	RESULT_EXISTS          = byte(0x0002)
	RESULT_ITEM_NOT_STORED = byte(0x0005)

	ERROR_VALUE_TOO_LARGE                 = byte(0x0003)
	ERROR_INVALID_ARGUMENTS               = byte(0x0004)
	ERROR_INCREMENT_DECREMENT_NON_NUMERIC = byte(0x0006)
	ERROR_UNKNOWN                         = byte(0x0081)
	ERROR_OUT_OF_MEMORY                   = byte(0x0082)

	DEFAULT_ENDIANNESS = binary.BigEndian
)

type Server interface {
	Start() error
	Stop() error
}

type tcp_server struct {
	port    int
	host    string
	running bool
	storage store.Storage
	server  net.Listener
}

type memcached_request struct {
	magicNumber byte
	messageType byte
	keyLength uint16
	extrasLength byte
	dataType byte
	reservedField uint16
	totalMessageBody uint32
	opaque uint32
	cas uint64
}

type memcached_response struct {
	magicNumber byte
	messageType byte
	keyLength uint16
	extrasLength byte
	dataType byte
	status uint16
	bodyLength uint32
	opaque uint32
	cas uint64
}

func New(port int, host string, storage store.Storage) Server {
	return &tcp_server{
		port:    port,
		host:    host,
		storage: storage,
	}
}

func (s *tcp_server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	s.server = l

	if err == nil {
		s.running = true
		go s.AcceptClients()
	}

	return err
}

func (s *tcp_server) Stop() error {
	if s.running {
		s.running = false
		return s.server.Close()
	} else {
		return nil
	}
}

func (s *tcp_server) AcceptClients() {
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

func (s *tcp_server) HandleConnection(connection net.Conn) {
	defer connection.Close()

	for s.running {
		var request memcached_request
		err := binary.Read(connection, binary.BigEndian, &request)

		if err {
			fmt.Println("binary.Read failed:", err)
			break
		}

		extras := make([]byte, int32(request.extrasLength), int32(request.extrasLength))
		if request.extrasLength > 0 {
			err = io.ReadFull(connection, extras)
			if err != nil {
				fmt.Println("reading extras failed:", err)
				break
			}
		}

		keyBytes := make([]byte, request.keyLength, request.keyLength)
		err = io.ReadFull(connection, keyBytes)

		if err != nil {
			fmt.Println("reading key failed:", err)
			break
		}

		key := string(keyBytes)
		response := prepareResponse(&request)

		switch request.messageType {
		case MESSAGE_GET:
			if !s.handleGet(connection, response, key) {
				break
			}
		case MESSAGE_SET:
			if !s.handleSet(connection, &request, response, key) {
				break
			}
		case MESSAGE_DELETE:
			if !s.handleDelete(connection, response, key) {
				break
			}
		}


	}

}

func (s *tcp_server) handleGet( connection net.Conn, response *memcached_response, key string ) bool {
	var err error
	body, found := s.storage.Get(key)
	if found {
		response.status = RESULT_OK
		response.bodyLength = len(body)
		err = binary.Write(connection, DEFAULT_ENDIANNESS, &response)

		if err != nil {
			fmt.Println("failed to write response to client:", err)
			return false
		}

		_, err := connection.Write(body)

		if err != nil {
			fmt.Println("writting body failed:", err)
			return false
		}
	} else {
		response.status = RESULT_NOT_FOUND
		err = binary.Write(connection, DEFAULT_ENDIANNESS, &response)

		if err != nil {
			fmt.Println("failed to write response to client:", err)
			return false
		}
	}

	return true
}

func (s *tcp_server) handleSet(connection net.Conn, request *memcached_request, response *memcached_response, key string) bool {
	bodyLength := request.totalMessageBody - request.keyLength - request.extrasLength
	body := make([]byte, bodyLength, bodyLength)
	_, err := io.ReadFull(connection, body)

	if err != nil {
		fmt.Println("reading body contents failed:", err)
		return false
	}

	s.storage.Put(key, body)

	response.status = RESULT_OK
	err = binary.Write(connection, DEFAULT_ENDIANNESS, &response)

	if err != nil {
		fmt.Println("writting body failed:", err)
		return false
	}

	return true
}

func (s *tcp_server) handleDelete(connection net.Conn, response *memcached_response, key string) bool {
	s.storage.Delete(key)

	response.status = RESULT_OK
	err := binary.Write(connection, DEFAULT_ENDIANNESS, &response)

	if err != nil {
		fmt.Println("writting body failed:", err)
		return false
	}

	return true
}

func prepareResponse( request *memcached_request ) *memcached_response {

	return &memcached_response{
		magicNumber: RESPONSE_KEY,
		messageType: request.messageType,
		opaque: request.opaque,
		cas: request.cas,
	}

}
