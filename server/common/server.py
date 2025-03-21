import socket
import logging
import signal
import communication.protocol as protocol
import common.utils as utils

SOCKET_TIMEOUT = 1
TOTAL_AGENCIES = 5

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._shutdown_requested = False
        self._finished_agencies = set()
        self._winning_bets_by_agency = {}
        self._lottery_done = False
        
        signal.signal(signal.SIGTERM, self.__handle_signal)
        logging.info('action: setup_signal | result: success | signal: SIGTERM')

    def __handle_signal(self, signalnum, frame):
        """
        Signal handler for graceful shutdown
        """
        if signalnum == signal.SIGTERM:
            logging.info('action: signal_received | result: success | signal: SIGTERM')
            self._shutdown_requested = True
            
    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again.
        The loop will continue until a SIGTERM signal is received.
        """
        try:
            self._server_socket.settimeout(SOCKET_TIMEOUT)
            while not self._shutdown_requested:
                try:
                    client_sock = self.__accept_new_connection()
                    self.__handle_client_connection(client_sock)
                except socket.timeout:
                    continue
                except OSError as e:
                    if self._shutdown_requested:
                        break
                    logging.error(f"action: accept_connection | result: fail | error: {e}")
        finally:
            self.__cleanup()
            
    def __cleanup(self):
        """
        Cleanup server resources during shutdown
        """
        logging.info('action: shutting_down | result: in_progress')
        
        try:
            self._server_socket.shutdown(socket.SHUT_RDWR)
            self._server_socket.close()
            logging.info('action: close_server_socket | result: success')
        except OSError as e:
            logging.error(f"action: close_server_socket | result: fail | error: {e}")
                
        logging.info('action: shutting_down | result: success')

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:            
            packet = protocol.receive_packet(client_sock)
            message = protocol.deserialize_packet(packet)
        except (OSError, ConnectionError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            client_sock.close()
            return
        except ValueError as e:
            logging.error(f"action: deserialize_packet | result: fail | error: {e}")
            client_sock.close()
            return
        
        if isinstance(message, list) and all(isinstance(bet, utils.Bet) for bet in message):
            self.__handle_bets(message, client_sock)
        elif isinstance(message, protocol.FinishedSendingBetsMessage):
            self.__handle_finished_sending_bets(message)
        elif isinstance(message, protocol.LotteryWinnersRequestMessage):
            self.__handle_lottery_winners_request(message, client_sock)
        
        client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def __handle_bets(self, bets, client_sock):
        """
        Handle bets received from a client

        Function that receives a list of bets and stores them in a file
        """
        try:
            utils.store_bets(bets)
            logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
            protocol.send_message(client_sock, protocol.BetsConfirmationMessage(protocol.BetsConfirmationResult.OK))
        except OSError as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: {len(bets)}')
            protocol.send_message(client_sock, protocol.BetsConfirmationMessage(protocol.BetsConfirmationResult.ERROR))

    def __handle_finished_sending_bets(self, finished_sending_bets_message):
        """
        Handle finished sending bets message

        Function that receives a message from a client indicating that
        all bets have been sent
        """
        agency_id = finished_sending_bets_message.id_agencia
        logging.info(f'action: total_apuestas_recibidas | result: success | id_agencia: {agency_id}')
        self._finished_agencies.add(agency_id)
        if len(self._finished_agencies) == TOTAL_AGENCIES:
            self._do_lottery()
            
    def _do_lottery(self):
        try:
            for bet in utils.load_bets():
                if utils.has_won(bet):
                    winning_bets = self._winning_bets_by_agency.get(bet.agency, [])
                    winning_bets.append(bet)
                    self._winning_bets_by_agency[bet.agency] = winning_bets
            self._lottery_done = True
            logging.info("action: sorteo | result: success")
        except OSError as e:
            logging.error(f"action: sorteo | result: fail | error: {e}")
    
    def __handle_lottery_winners_request(self, lottery_winners_request_message, client_sock):
        agency_id = lottery_winners_request_message.id_agencia
        logging.info(f"action: lottery_winners_requested | result: success | agency_id: {agency_id}")
        if not self._lottery_done:
            protocol.send_message(client_sock, protocol.LotteryWinnersResponseMessage(protocol.LotteryWinnersResponseStatus.NOT_READY))
            return
        winning_bets = self._winning_bets_by_agency.get(agency_id, [])
        winners_dnis = [bet.document for bet in winning_bets]
        lottery_winners_response_message = protocol.LotteryWinnersResponseMessage(protocol.LotteryWinnersResponseStatus.READY, winners_dnis)
        protocol.send_message(client_sock, lottery_winners_response_message)
        logging.info(f"action: lottery_winners_sent | result: success | agency_id: {agency_id} | n_winners: {len(winners_dnis)}")