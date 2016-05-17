package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/mauricio/gocached/store"
)

const (
	requestKey  = byte(0x80)
	responseKey = byte(0x81)

	messageGet    = byte(0x00)
	messageSet    = byte(0x01)
	messageDelete = byte(0x04)

	resultOk            = byte(0x0000)
	resultNotFound      = byte(0x0001)
	resultExists        = byte(0x0002)
	resultItemNotStored = byte(0x0005)

	errorValueTooLarge                = byte(0x0003)
	errorInvalidArguments             = byte(0x0004)
	errorIncrementDecrementNotNumeric = byte(0x0006)
	errorUnknown                      = byte(0x0081)
	errorOutOfMemory                  = byte(0x0082)
)

var defautEndianness = binary.BigEndian

type Server interface {
	Start() error
	Stop() error
}

type tcpServer struct {
	port         int32
	host         string
	running      bool
	storage      store.Storage
	server       net.Listener
	runningMutex sync.Mutex
	serverMutex  sync.Mutex
}

type memcachedRequest struct {
	MagicNumber      byte
	MessageType      byte
	KeyLength        uint16
	ExtrasLength     byte
	DataType         byte
	ReservedField    uint16
	TotalMessageBody uint32
	Opaque           uint32
	Cas              uint64
}

type memcachedResponse struct {
	MagicNumber  byte
	MessageType  byte
	KeyLength    uint16
	ExtrasLength byte
	DataType     byte
	Status       uint16
	BodyLength   uint32
	Opaque       uint32
	Cas          uint64
}

func New(port int32, host string, storage store.Storage) Server {
	return &tcpServer{
		port:    port,
		host:    host,
		storage: storage,
	}
}

func (s *tcpServer) isRunning() bool {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()
	return s.running
}

func (s *tcpServer) setRunning(value bool) {
	s.runningMutex.Lock()
	defer s.runningMutex.Unlock()
	s.running = value
}

func (s *tcpServer) Start() error {
	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()
	if !s.isRunning() {
		l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
		s.server = l

		if err == nil {
			s.setRunning(true)
			go s.AcceptClients()
		}

		return err
	}

	return nil
}

func (s *tcpServer) Stop() error {
	s.serverMutex.Lock()
	defer s.serverMutex.Unlock()

	if s.isRunning() {
		s.setRunning(false)
		return s.server.Close()
	}

	return nil
}

func (s *tcpServer) AcceptClients() {
	for s.isRunning() {
		connection, err := s.server.Accept()
		if err == nil {
			go s.HandleConnection(connection)
		} else {
			fmt.Printf("Failed to accept client, stopping: %s\n", err.Error())
			s.Stop()
		}
	}
}

func (s *tcpServer) HandleConnection(connection net.Conn) {
	defer connection.Close()

	for s.isRunning() {
		var request memcachedRequest
		err := binary.Read(connection, binary.BigEndian, &request)

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("binary.Read failed:", err)
			break
		}

		extras := make([]byte, int32(request.ExtrasLength), int32(request.ExtrasLength))
		if request.ExtrasLength > 0 {
			_, err = io.ReadFull(connection, extras)
			if err != nil {
				fmt.Println("reading extras failed:", err)
				break
			}
		}

		keyBytes := make([]byte, request.KeyLength, request.KeyLength)
		_, err = io.ReadFull(connection, keyBytes)

		if err != nil {
			fmt.Println("reading key failed:", err)
			break
		}

		key := string(keyBytes)
		response := prepareResponse(&request)

		switch request.MessageType {
		case messageGet:
			if !s.handleGet(connection, response, key) {
				break
			}
		case messageSet:
			if !s.handleSet(connection, &request, response, key) {
				break
			}
		case messageDelete:
			if !s.handleDelete(connection, response, key) {
				break
			}
		}

	}

}

func (s *tcpServer) handleGet(connection net.Conn, response *memcachedResponse, key string) bool {
	var err error
	body, found := s.storage.Get(key)
	if found {
		response.Status = uint16(resultOk)
		response.BodyLength = uint32(len(body))
		err = binary.Write(connection, defautEndianness, response)

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
		response.Status = uint16(resultNotFound)
		err = binary.Write(connection, defautEndianness, response)

		if err != nil {
			fmt.Println("failed to write response to client:", err)
			return false
		}
	}

	return true
}

func (s *tcpServer) handleSet(connection net.Conn, request *memcachedRequest, response *memcachedResponse, key string) bool {
	bodyLength := request.TotalMessageBody - uint32(request.KeyLength) - uint32(request.ExtrasLength)
	body := make([]byte, bodyLength, bodyLength)
	_, err := io.ReadFull(connection, body)

	if err != nil {
		fmt.Println("reading body contents failed:", err)
		return false
	}

	s.storage.Put(key, body)

	response.Status = uint16(resultOk)
	err = binary.Write(connection, defautEndianness, response)

	if err != nil {
		fmt.Println("writting body failed:", err)
		return false
	}

	return true
}

func (s *tcpServer) handleDelete(connection net.Conn, response *memcachedResponse, key string) bool {
	s.storage.Delete(key)

	response.Status = uint16(resultOk)
	err := binary.Write(connection, defautEndianness, response)

	if err != nil {
		fmt.Println("writting body failed:", err)
		return false
	}

	return true
}

func prepareResponse(request *memcachedRequest) *memcachedResponse {

	return &memcachedResponse{
		MagicNumber: responseKey,
		MessageType: request.MessageType,
		Opaque:      request.Opaque,
		Cas:         request.Cas,
	}

}
