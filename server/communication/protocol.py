from enum import IntEnum

LEN_MESSAGE_SIZE = 2
MESSAGE_TYPE_SIZE = 1
HEADER_SIZE = LEN_MESSAGE_SIZE + MESSAGE_TYPE_SIZE

ID_AGENCIA_SIZE = 4
LEN_NOMBRE_SIZE = 1
LEN_APELLIDO_SIZE = 1
DNI_SIZE = 8
NACIMIENTO_SIZE = 10
NUMERO_SIZE = 4

RESULT_SIZE = 1

class MessageType(IntEnum):
    BET = 1
    BET_CONFIRMATION = 2

class BetMessage:
    def __init__(self, id_agencia, nombre, apellido, dni, nacimiento, numero):
        self.id_agencia = id_agencia
        self.nombre = nombre
        self.apellido = apellido
        self.dni = dni
        self.nacimiento = nacimiento
        self.numero = numero

def receive_packet(sock):
    """ Receives a packet from a socket """
    buf_packet = bytearray()
    while len(buf_packet) < HEADER_SIZE:
        header_bytes = sock.recv(HEADER_SIZE - len(buf_packet))
        if not header_bytes:
            raise ConnectionError("Connection closed while reading header")
        buf_packet.extend(header_bytes)
    
    message_size = int.from_bytes(buf_packet[:LEN_MESSAGE_SIZE], byteorder="big")
    
    while len(buf_packet) < HEADER_SIZE + message_size:
        message_bytes = sock.recv(HEADER_SIZE + message_size - len(buf_packet))
        if not message_bytes:
            raise ConnectionError("Connection closed while reading message")
        buf_packet.extend(message_bytes)
    
    return buf_packet


def deserialize_packet(packet):
    """ Deserializes a message from a packet (bytearray) """
    message_type = MessageType(packet[LEN_MESSAGE_SIZE])
    pos = HEADER_SIZE
    if message_type == MessageType.BET:
        id_agencia = int.from_bytes(packet[pos:pos + ID_AGENCIA_SIZE], byteorder="big")
        pos += ID_AGENCIA_SIZE
        
        nombre_size = int.from_bytes(packet[pos:pos + LEN_NOMBRE_SIZE], byteorder="big")
        pos += LEN_NOMBRE_SIZE
        nombre = packet[pos:pos + nombre_size].decode("utf-8")
        pos += nombre_size
        
        apellido_size = int.from_bytes(packet[pos:pos + LEN_APELLIDO_SIZE], byteorder="big")
        pos += LEN_APELLIDO_SIZE
        apellido = packet[pos:pos + apellido_size].decode("utf-8")
        pos += apellido_size
        
        dni = packet[pos:pos + DNI_SIZE].decode("utf-8")
        pos += DNI_SIZE
        
        nacimiento = packet[pos:pos + NACIMIENTO_SIZE].decode("utf-8")
        pos += NACIMIENTO_SIZE
        
        numero = int.from_bytes(packet[pos:pos + NUMERO_SIZE], byteorder="big")
        
        return BetMessage(id_agencia, nombre, apellido, dni, nacimiento, numero)
    else:
        raise ValueError("Invalid message type")

class BetConfirmationResult(IntEnum):
    OK = 0
    ERROR = 1
    
class BetConfirmationMessage:
    def __init__(self, result):
        self.result = result
    
    def serialize(self):
        """ Serializes the message to a bytearray """
        message_size = RESULT_SIZE.to_bytes(LEN_MESSAGE_SIZE, byteorder="big")
        message_type = MessageType.BET_CONFIRMATION.to_bytes(MESSAGE_TYPE_SIZE, byteorder="big")
        result_bytes = self.result.to_bytes(RESULT_SIZE, byteorder="big")
        return message_size + message_type + result_bytes

def send_message(sock, message):
    """ Sends a message through a socket """
    packet = message.serialize()
    sock.sendall(packet)