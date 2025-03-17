package communication

import (
	"encoding/binary"
	"net"
)

const (
	LenTotalSize = 4
	LenMsgType   = 1
	LenResult    = 1

	LenNombre     = 2
	LenApellido   = 2
	LenDni        = 2
	LenNacimiento = 2
	LenNumero     = 4
)

const (
	MsgTypeBet = 1
)

const (
	ResultSuccess = 0
)

type BetMessage struct {
	Nombre     string
	Apellido   string
	Dni        string
	Nacimiento string
	Numero     int32
}

func NewBetMessage(nombre, apellido, dni, nacimiento string, numero int32) BetMessage {
	return BetMessage{
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

	totalSize := LenTotalSize + LenMsgType + LenResult
	totalSize += LenNombre + len(nombreBytes)
	totalSize += LenApellido + len(apellidoBytes)
	totalSize += LenDni + len(dniBytes)
	totalSize += LenNacimiento + len(nacimientoBytes)
	totalSize += LenNumero

	buffer := make([]byte, totalSize)
	pos := 0

	binary.BigEndian.PutUint32(buffer[pos:LenTotalSize], uint32(totalSize))
	pos += LenTotalSize

	buffer[pos] = MsgTypeBet
	pos += LenMsgType

	buffer[pos] = ResultSuccess
	pos += LenResult

	binary.BigEndian.PutUint16(buffer[pos:pos+LenNombre], uint16(len(nombreBytes)))
	pos += LenNombre
	copy(buffer[pos:pos+len(nombreBytes)], nombreBytes)
	pos += len(nombreBytes)

	binary.BigEndian.PutUint16(buffer[pos:pos+LenApellido], uint16(len(apellidoBytes)))
	pos += LenApellido
	copy(buffer[pos:pos+len(apellidoBytes)], apellidoBytes)
	pos += len(apellidoBytes)

	binary.BigEndian.PutUint16(buffer[pos:pos+LenDni], uint16(len(dniBytes)))
	pos += LenDni
	copy(buffer[pos:pos+len(dniBytes)], dniBytes)
	pos += len(dniBytes)

	binary.BigEndian.PutUint16(buffer[pos:pos+LenNacimiento], uint16(len(nacimientoBytes)))
	pos += LenNacimiento
	copy(buffer[pos:pos+len(nacimientoBytes)], nacimientoBytes)
	pos += len(nacimientoBytes)

	binary.BigEndian.PutUint32(buffer[pos:pos+LenNumero], uint32(bet.Numero))

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
