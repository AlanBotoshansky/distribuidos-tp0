import socket
import logging
import signal
import multiprocessing as mp
import communication.protocol as protocol
import common.utils as utils

JOIN_PROCESS_TIMEOUT = 1

class Server:
    def __init__(self, port, listen_backlog, total_agencies):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._shutdown_requested = None
        self.connections = []
        self._total_agencies = total_agencies
        self._processes = []
        self._lottery_done_or_shutdown = None
        
        signal.signal(signal.SIGTERM, self.__handle_signal)
        logging.info('action: setup_signal | result: success | signal: SIGTERM')

    def __handle_signal(self, signalnum, frame):
        """
        Signal handler for graceful shutdown
        """
        if signalnum == signal.SIGTERM:
            logging.info('action: signal_received | result: success | signal: SIGTERM')
            if self._shutdown_requested is not None and self._lottery_done_or_shutdown is not None:
                self._shutdown_requested.set()
                self._lottery_done_or_shutdown.set()
                self.__cleanup()
            
    def run(self):
        """
        Multi-process server loop
        
        Server that accepts connections and handles each in a separate process.
        The loop will continue until a SIGTERM signal is received.
        """
        with mp.Manager() as manager:
            self._shutdown_requested = manager.Event()
            finished_agencies = manager.dict()
            winning_bets_by_agency = manager.dict()
            self._lottery_done_or_shutdown = manager.Event()
            file_lock = manager.Lock()
            while not self._shutdown_requested.is_set():
                try:
                    client_sock = self.__accept_new_connection()
                    process = mp.Process(target=handle_client_connection, args=(self._shutdown_requested, self._total_agencies, client_sock, finished_agencies, winning_bets_by_agency, self._lottery_done_or_shutdown, file_lock), daemon=True)
                    process.start()
                    self._processes.append(process)
                    self._processes = [p for p in self._processes if p.is_alive()]
                except OSError as e:
                    logging.error(f"action: accept_connection | result: fail | error: {e}")
                
    def __cleanup(self):
        """
        Cleanup server resources during shutdown
        """
        logging.info('action: shutting_down | result: in_progress')
        
        for conn in self.connections:
            close_client_connection(conn)
        self.connections.clear()
        
        try:
            self._server_socket.shutdown(socket.SHUT_RDWR)
            self._server_socket.close()
            logging.info('action: close_server_socket | result: success')
        except OSError as e:
            logging.error(f"action: close_server_socket | result: fail | error: {e}")
            
        for p in self._processes:
            p.join(timeout=JOIN_PROCESS_TIMEOUT)
                
        logging.info('action: shutting_down | result: success')

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
    
def close_client_connection(client_sock):
    """
    Close a client connection
    """
    try:
        client_sock.shutdown(socket.SHUT_RDWR)
        client_sock.close()
        logging.info('action: close_client_socket | result: success')
    except OSError as e:
        logging.error(f"action: close_client_socket | result: fail | error: {e}")

def safe_store_bets(bets, file_lock):
    """
    Process-safe wrapper for store_bets
    """
    with file_lock:
        utils.store_bets(bets)
        
def safe_load_bets(file_lock):
    """
    Process-safe wrapper for load_bets
    """
    with file_lock:
        return list(utils.load_bets())

def handle_client_connection(shutdown_requested, total_agencies, client_sock, finished_agencies, winning_bets_by_agency, lottery_done_or_shutdown, file_lock):
    """
    Worker function to handle a client connection in a separate process
    """
    while not shutdown_requested.is_set():
        try:
            packet = protocol.receive_packet(client_sock)
            message = protocol.deserialize_packet(packet)
        except (OSError, ConnectionError) as e:
            logging.error(f"action: receive_message | result: fail | error: {e}")
            break
        except ValueError as e:
            logging.error(f"action: deserialize_packet | result: fail | error: {e}")
            break
        
        if isinstance(message, list) and all(isinstance(bet, utils.Bet) for bet in message):
            handle_bets(message, client_sock, file_lock)
        elif isinstance(message, protocol.FinishedSendingBetsMessage):
            handle_finished_sending_bets(message, total_agencies, finished_agencies, winning_bets_by_agency, lottery_done_or_shutdown, file_lock)
        elif isinstance(message, protocol.LotteryWinnersRequestMessage):
            handle_lottery_winners_request(message, shutdown_requested, client_sock, winning_bets_by_agency, lottery_done_or_shutdown)
            break

    close_client_connection(client_sock)

def handle_bets(bets, client_sock, file_lock):
    """
    Handle bets received from a client

    Function that receives a list of bets and stores them in a file
    """
    bets_stored = False
    try:
        safe_store_bets(bets, file_lock)
        logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')
        bets_stored = True
    except OSError as e:
        logging.info(f'action: apuesta_recibida | result: fail | cantidad: {len(bets)} | error: {e}')
        
    try:
        if bets_stored:
            protocol.send_message(client_sock, protocol.BetsConfirmationMessage(protocol.BetsConfirmationResult.OK))
        else:
           protocol.send_message(client_sock, protocol.BetsConfirmationMessage(protocol.BetsConfirmationResult.ERROR))
        logging.info(f'action: bets_confirmation_sent | result: success | cantidad: {len(bets)}')
    except OSError as e:
        logging.error(f"action: bets_confirmation_sent | result: fail | cantidad: {len(bets)} | error: {e}")

def handle_finished_sending_bets(finished_sending_bets_message, total_agencies, finished_agencies, winning_bets_by_agency, lottery_done_or_shutdown, file_lock):
    """
    Handle finished sending bets message with multiprocessing synchronization
    """
    agency_id = finished_sending_bets_message.id_agencia
    logging.info(f'action: total_apuestas_recibidas | result: success | id_agencia: {agency_id}')
    finished_agencies[agency_id] = None
    if len(finished_agencies) == total_agencies:
        do_lottery(winning_bets_by_agency, lottery_done_or_shutdown, file_lock)

def do_lottery(winning_bets_by_agency, lottery_done_or_shutdown, file_lock):
    """
    Do the lottery with multiprocessing synchronization
    """
    try:
        for bet in safe_load_bets(file_lock):
            if utils.has_won(bet):
                winning_bets = winning_bets_by_agency.get(bet.agency, [])
                winning_bets.append(bet)
                winning_bets_by_agency[bet.agency] = winning_bets
        lottery_done_or_shutdown.set()
        logging.info("action: sorteo | result: success")
    except OSError as e:
        logging.error(f"action: sorteo | result: fail | error: {e}")

def handle_lottery_winners_request(lottery_winners_request_message, shutdown_requested, client_sock, winning_bets_by_agency, lottery_done_or_shutdown):
    """
    Handle lottery winners request with multiprocessing synchronization
    """
    agency_id = lottery_winners_request_message.id_agencia
    logging.info(f"action: lottery_winners_requested | result: success | agency_id: {agency_id}")
    lottery_done_or_shutdown.wait()
    if shutdown_requested.is_set():
        logging.info(f"action: lottery_winners_cancelled | result: success | agency_id: {agency_id}")
        return
    winning_bets = winning_bets_by_agency.get(agency_id, [])
    winners_dnis = [bet.document for bet in winning_bets]
    lottery_winners_response_message = protocol.LotteryWinnersResponseMessage(winners_dnis)
    try:
        protocol.send_message(client_sock, lottery_winners_response_message)
        logging.info(f"action: lottery_winners_sent | result: success | agency_id: {agency_id} | n_winners: {len(winners_dnis)}")
    except OSError as e:
        logging.error(f"action: lottery_winners_sent | result: fail | error: {e}")