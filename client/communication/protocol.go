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
	MessageTypeBets             = 1
	MessageTypeBetsConfirmation = 2
)

const (
	BetsConfirmationResultOk    = 0
	BetsConfirmationResultError = 1
)

type Bet struct {
	Nombre     string
	Apellido   string
	Dni        string
	Nacimiento string
	Numero     uint32
}

func NewBet(nombre, apellido, dni, nacimiento string, numero uint32) Bet {
	return Bet{
		Nombre:     nombre,
		Apellido:   apellido,
		Dni:        dni,
		Nacimiento: nacimiento,
		Numero:     numero,
	}
}

func SerializeBet(bet Bet) []byte {
	nombreBytes := []byte(bet.Nombre)
	apellidoBytes := []byte(bet.Apellido)
	dniBytes := []byte(bet.Dni)
	nacimientoBytes := []byte(bet.Nacimiento)

	betSize := 0
	betSize += NombreSize
	betSize += ApellidoSize
	betSize += DniSize
	betSize += NacimientoSize
	betSize += NumeroSize

	buffer := make([]byte, betSize)
	pos := 0

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

type Bets struct {
	IdAgencia uint32
	Bets      []Bet
}

func NewBets(idAgencia uint32, bets []Bet) Bets {
	return Bets{
		IdAgencia: idAgencia,
		Bets:      bets,
	}
}

func SerializeBets(bets Bets) []byte {
	messageSize := 0
	messageSize += IdAgenciaSize

	betsBytes := make([]byte, 0)
	for _, bet := range bets.Bets {
		betBytes := SerializeBet(bet)
		betsBytes = append(betsBytes, betBytes...)
		messageSize += len(betBytes)
	}

	buffer := make([]byte, HeaderSize+messageSize)
	pos := 0

	binary.BigEndian.PutUint16(buffer[pos:pos+LenMessageSize], uint16(messageSize))
	pos += LenMessageSize

	buffer[pos] = MessageTypeBets
	pos += MessageTypeSize

	binary.BigEndian.PutUint32(buffer[pos:pos+IdAgenciaSize], bets.IdAgencia)
	pos += IdAgenciaSize

	copy(buffer[pos:], betsBytes)

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

type BetsConfirmationMessage struct {
	Result uint8
}

func NewBetsConfirmationMessage(result uint8) BetsConfirmationMessage {
	return BetsConfirmationMessage{
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
	case MessageTypeBetsConfirmation:
		return DeserializeBetsConfirmation(packet), nil
	default:
		return 0, &InvalidMessageTypeError{MessageType: messageType}
	}
}

func DeserializeBetsConfirmation(packet []byte) BetsConfirmationMessage {
	return BetsConfirmationMessage{
		Result: packet[HeaderSize],
	}
}
