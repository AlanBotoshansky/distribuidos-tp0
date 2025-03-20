from enum import IntEnum
from common.utils import Bet

LEN_MESSAGE_SIZE = 2
MESSAGE_TYPE_SIZE = 1
HEADER_SIZE = LEN_MESSAGE_SIZE + MESSAGE_TYPE_SIZE

ID_AGENCIA_SIZE = 4
NOMBRE_SIZE = 32
APELLIDO_SIZE = 32
DNI_SIZE = 8
NACIMIENTO_SIZE = 10
NUMERO_SIZE = 4

RESULT_SIZE = 1

class MessageType(IntEnum):
    BETS = 1
    BETS_CONFIRMATION = 2
    FINISHED_SENDING_BETS = 3
    
class FinishedSendingBetsMessage:
    def __init__(self, id_agencia):
        self.id_agencia = id_agencia

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
    if message_type == MessageType.BETS:
        return deserialize_bets(packet[HEADER_SIZE:])
    elif message_type == MessageType.FINISHED_SENDING_BETS:
        return deserialize_finished_sending_bets(packet[HEADER_SIZE:])
    else:
        raise ValueError("Invalid message type")
    
def deserialize_bets(bets_message_bytes):
    pos = 0
    bytes_deserialized = 0
    bets = []
    
    id_agencia = int.from_bytes(bets_message_bytes[pos:pos + ID_AGENCIA_SIZE], byteorder="big")
    pos += ID_AGENCIA_SIZE
    bytes_deserialized += ID_AGENCIA_SIZE
    
    while bytes_deserialized < len(bets_message_bytes):
        nombre = bets_message_bytes[pos:pos + NOMBRE_SIZE].decode("utf-8")
        pos += NOMBRE_SIZE
        bytes_deserialized += NOMBRE_SIZE
        
        apellido = bets_message_bytes[pos:pos + APELLIDO_SIZE].decode("utf-8")
        pos += APELLIDO_SIZE
        bytes_deserialized += APELLIDO_SIZE
        
        dni = bets_message_bytes[pos:pos + DNI_SIZE].decode("utf-8")
        pos += DNI_SIZE
        bytes_deserialized += DNI_SIZE
        
        nacimiento = bets_message_bytes[pos:pos + NACIMIENTO_SIZE].decode("utf-8")
        pos += NACIMIENTO_SIZE
        bytes_deserialized += NACIMIENTO_SIZE
        
        numero = int.from_bytes(bets_message_bytes[pos:pos + NUMERO_SIZE], byteorder="big")
        pos += NUMERO_SIZE
        bytes_deserialized += NUMERO_SIZE
    
        bets.append(Bet(id_agencia, nombre, apellido, dni, nacimiento, numero))
    
    return bets

def deserialize_finished_sending_bets(finished_sending_bets_message_bytes):
    id_agencia = int.from_bytes(finished_sending_bets_message_bytes, byteorder="big")
    return FinishedSendingBetsMessage(id_agencia)

class BetsConfirmationResult(IntEnum):
    OK = 0
    ERROR = 1
    
class BetsConfirmationMessage:
    def __init__(self, result):
        self.result = result
    
    def serialize(self):
        """ Serializes the message to a bytearray """
        message_size = RESULT_SIZE.to_bytes(LEN_MESSAGE_SIZE, byteorder="big")
        message_type = MessageType.BETS_CONFIRMATION.to_bytes(MESSAGE_TYPE_SIZE, byteorder="big")
        result_bytes = self.result.to_bytes(RESULT_SIZE, byteorder="big")
        return message_size + message_type + result_bytes

def send_message(sock, message):
    """ Sends a message through a socket """
    packet = message.serialize()
    sock.sendall(packet)