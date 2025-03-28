import socket
import logging
import signal
import communication.protocol as protocol
import common.utils as utils

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._shutdown_requested = False
        self.connections = []
        
        signal.signal(signal.SIGTERM, self.__handle_signal)
        logging.info('action: setup_signal | result: success | signal: SIGTERM')

    def __handle_signal(self, signalnum, frame):
        """
        Signal handler for graceful shutdown
        """
        if signalnum == signal.SIGTERM:
            logging.info('action: signal_received | result: success | signal: SIGTERM')
            self._shutdown_requested = True
            self.__cleanup()
            
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again.
        The loop will continue until a SIGTERM signal is received.
        """
        while not self._shutdown_requested:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError as e:
                if self._shutdown_requested:
                    break
                logging.error(f"action: accept_connection | result: fail | error: {e}")
            
    def __cleanup(self):
        """
        Cleanup server resources during shutdown
        """
        logging.info('action: shutting_down | result: in_progress')
        
        for conn in self.connections:
            self.__close_client_connection(conn)
        
        try:
            self._server_socket.shutdown(socket.SHUT_RDWR)
            self._server_socket.close()
            logging.info('action: close_server_socket | result: success')
        except OSError as e:
            logging.error(f"action: close_server_socket | result: fail | error: {e}")
                
        logging.info('action: shutting_down | result: success')
        
    def __close_client_connection(self, client_sock):
        """
        Close a client connection
        """
        try:
            client_sock.shutdown(socket.SHUT_RDWR)
            client_sock.close()
            if client_sock in self.connections:
                self.connections.remove(client_sock)
            logging.info('action: close_client_socket | result: success')
        except OSError as e:
            logging.error(f"action: close_client_socket | result: fail | error: {e}")

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:            
            packet = protocol.receive_packet(client_sock)
            bets = protocol.deserialize_packet(packet)
        except (OSError, ConnectionError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            self.__close_client_connection(client_sock)
            return
        except ValueError as e:
            logging.error(f"action: deserialize_packet | result: fail | error: {e}")
            self.__close_client_connection(client_sock)
            return
            
        try:
            utils.store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
            protocol.send_message(client_sock, protocol.BetsConfirmationMessage(protocol.BetsConfirmationResult.OK))
        except OSError as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: {len(bets)}')
            protocol.send_message(client_sock, protocol.BetsConfirmationMessage(protocol.BetsConfirmationResult.ERROR))
        finally:
            self.__close_client_connection(client_sock)

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        self.connections.append(c)
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
