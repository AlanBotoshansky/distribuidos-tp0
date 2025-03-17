package communication

import (
	"encoding/binary"
	"net"
)

const (
	LenMessageSize  = 2
	MessageTypeSize = 1
	HeaderSize      = LenMessageSize + MessageTypeSize

	IdAgenciaSize   = 4
	LenNombreSize   = 2
	LenApellidoSize = 2
	DniSize         = 8
	NacimientoSize  = 10
	NumeroSize      = 4
)

const (
	MessageTypeBet = 1
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
	messageSize += LenNombreSize + len(nombreBytes)
	messageSize += LenApellidoSize + len(apellidoBytes)
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

	binary.BigEndian.PutUint16(buffer[pos:pos+LenNombreSize], uint16(len(nombreBytes)))
	pos += LenNombreSize
	copy(buffer[pos:pos+len(nombreBytes)], nombreBytes)
	pos += len(nombreBytes)

	binary.BigEndian.PutUint16(buffer[pos:pos+LenApellidoSize], uint16(len(apellidoBytes)))
	pos += LenApellidoSize
	copy(buffer[pos:pos+len(apellidoBytes)], apellidoBytes)
	pos += len(apellidoBytes)

	copy(buffer[pos:pos+DniSize], dniBytes)
	pos += DniSize

	copy(buffer[pos:pos+NacimientoSize], nacimientoBytes)
	pos += NacimientoSize

	binary.BigEndian.PutUint32(buffer[pos:pos+NumeroSize], uint32(bet.Numero))

	return buffer
}

func SendMessage(conn net.Conn, data []byte) error {
	bytesWritten := 0
	for bytesWritten < len(data) {
		n, err := conn.Write(data[bytesWritten:])
		if err != nil {
			return err
		}
		bytesWritten += n
	}
	return nil
}
