package communication

import (
	"encoding/binary"
	"fmt"
	"net"
)

const (
	LenMessageSize  = 2
	MessageTypeSize = 1
	HeaderSize      = LenMessageSize + MessageTypeSize

	IdAgenciaSize  = 4
	NombreSize     = 32
	ApellidoSize   = 32
	DniSize        = 8
	NacimientoSize = 10
	NumeroSize     = 4
)

const (
	MessageTypeBet             = 1
	MessageTypeBetConfirmation = 2
)

const (
	BetConfirmationResultOk    = 0
	BetConfirmationResultError = 1
)

type BetMessage struct {
	IdAgencia  uint32
	Nombre     string
	Apellido   string
	Dni        string
	Nacimiento string
	Numero     uint32
}

func NewBetMessage(idAgencia uint32, nombre, apellido, dni, nacimiento string, numero uint32) BetMessage {
	return BetMessage{
		IdAgencia:  idAgencia,
		Nombre:     nombre,
		Apellido:   apellido,
		Dni:        dni,
		Nacimiento: nacimiento,
		Numero:     numero,
	}
}

func SerializeBet(bet BetMessage) []byte {
	nombreBytes := []byte(bet.Nombre)
	apellidoBytes := []byte(bet.Apellido)
	dniBytes := []byte(bet.Dni)
	nacimientoBytes := []byte(bet.Nacimiento)

	messageSize := 0
	messageSize += IdAgenciaSize
	messageSize += NombreSize
	messageSize += ApellidoSize
	messageSize += DniSize
	messageSize += NacimientoSize
	messageSize += NumeroSize

	buffer := make([]byte, HeaderSize+messageSize)
	pos := 0

	binary.BigEndian.PutUint16(buffer[pos:pos+LenMessageSize], uint16(messageSize))
	pos += LenMessageSize

	buffer[pos] = MessageTypeBet
	pos += MessageTypeSize

	binary.BigEndian.PutUint32(buffer[pos:pos+IdAgenciaSize], bet.IdAgencia)
	pos += IdAgenciaSize

	copy(buffer[pos:pos+NombreSize], nombreBytes)
	pos += NombreSize

	copy(buffer[pos:pos+ApellidoSize], apellidoBytes)
	pos += ApellidoSize

	copy(buffer[pos:pos+DniSize], dniBytes)
	pos += DniSize

	copy(buffer[pos:pos+NacimientoSize], nacimientoBytes)
	pos += NacimientoSize

	binary.BigEndian.PutUint32(buffer[pos:pos+NumeroSize], uint32(bet.Numero))

	return buffer
}

func SendPacket(conn net.Conn, packet []byte) error {
	bytesWritten := 0
	for bytesWritten < len(packet) {
		n, err := conn.Write(packet[bytesWritten:])
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	return nil
}

type BetConfirmationMessage struct {
	Result uint8
}

func NewBetConfirmationMessage(result uint8) BetConfirmationMessage {
	return BetConfirmationMessage{
		Result: result,
	}
}

func ReceivePacket(conn net.Conn) ([]byte, error) {
	bufPacket := make([]byte, 0)
	for len(bufPacket) < HeaderSize {
		headerBytes := make([]byte, HeaderSize-len(bufPacket))
		n, err := conn.Read(headerBytes)
		if err != nil {
			return nil, err
		}
		bufPacket = append(bufPacket, headerBytes[:n]...)
	}

	messageSize := int(binary.BigEndian.Uint16(bufPacket[:LenMessageSize]))

	for len(bufPacket) < HeaderSize+messageSize {
		messageBytes := make([]byte, HeaderSize+messageSize-len(bufPacket))
		n, err := conn.Read(messageBytes)
		if err != nil {
			return nil, err
		}
		bufPacket = append(bufPacket, messageBytes[:n]...)
	}

	return bufPacket, nil
}

type InvalidMessageTypeError struct {
	MessageType uint8
}

func (e *InvalidMessageTypeError) Error() string {
	return fmt.Sprintf("invalid message type: %d", e.MessageType)
}

func DeserializePacket(packet []byte) (interface{}, error) {
	messageType := uint8(packet[LenMessageSize])
	switch messageType {
	case MessageTypeBetConfirmation:
		return DeserializeBetConfirmation(packet), nil
	default:
		return 0, &InvalidMessageTypeError{MessageType: messageType}
	}
}

func DeserializeBetConfirmation(packet []byte) BetConfirmationMessage {
	return BetConfirmationMessage{
		Result: packet[HeaderSize],
	}
}
