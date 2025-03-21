package common

import (
	"encoding/csv"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/communication"
)

const RecordsToCheckShutdown = 500

const MillisecondsBetweenLotteryWinnersRequests = 5000

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID             uint32
	ServerAddress  string
	LoopAmount     int
	LoopPeriod     time.Duration
	BatchMaxAmount int
}

// Client Entity that encapsulates how
type Client struct {
	config       ClientConfig
	conn         net.Conn
	shutdownChan chan struct{}
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config:       config,
		shutdownChan: make(chan struct{}),
	}
	return client
}

func (c *Client) ReadBetsFromCSV() ([]communication.Bet, error) {
	filePath := fmt.Sprintf("../../.data/agency-%d.csv", c.config.ID)
	file, err := os.Open(filePath)
	if err != nil {
		log.Errorf("action: open_file | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return nil, err
	}
	defer file.Close()

	bets := make([]communication.Bet, 0)
	reader := csv.NewReader(file)
	recordCount := 0

	for {
		if recordCount%RecordsToCheckShutdown == 0 {
			select {
			case <-c.shutdownChan:
				log.Infof("action: read_bets_interrumpted | result: succes | client_id: %v", c.config.ID)
				return nil, fmt.Errorf("shutdown requested")
			default:
			}
		}

		bet_record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Errorf("action: read_bet_record | result: fail | client_id: %v | error: %v", c.config.ID, err)
			continue
		}

		recordCount += 1

		if len(bet_record) != 5 {
			log.Errorf("action: read_bet_record | result: fail | client_id: %v | error: invalid record length", c.config.ID)
			continue
		}
		nombre := bet_record[0]
		apellido := bet_record[1]
		dni := bet_record[2]
		nacimiento := bet_record[3]
		numero, err := strconv.ParseUint(bet_record[4], 10, 32)
		if err != nil {
			log.Errorf("action: read_bet_record | result: fail | client_id: %v | error: %v", c.config.ID, err)
			continue
		}
		bet := communication.NewBet(nombre, apellido, dni, nacimiento, uint32(numero))
		bets = append(bets, bet)
	}
	return bets, nil
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

// Shutdown Close the connection and signal the client to stop sending messages
func (c *Client) Shutdown() {
	log.Infof("action: client_shutdown | result: in_progress | client_id: %v", c.config.ID)
	close(c.shutdownChan)

	if c.conn != nil {
		log.Infof("action: close_connection | result: in_progress | client_id: %v", c.config.ID)
		c.conn.Close()
		log.Infof("action: close_connection | result: success | client_id: %v", c.config.ID)
		c.conn = nil
	}

	log.Infof("action: client_shutdown | result: success | client_id: %v", c.config.ID)
}

// StartClientLoop Send batchs of bets to the server
func (c *Client) StartClientLoop() {
	bets, err := c.ReadBetsFromCSV()
	if err != nil {
		return
	}

	select {
	case <-c.shutdownChan:
		log.Infof("action: client_interrumpted | result: success | client_id: %v", c.config.ID)
		return
	default:
	}

	betsSent := 0
	for betsSent < len(bets) {
		err := c.createClientSocket()
		if err != nil {
			return
		}

		log.Infof("action: apuesta_enviada | result: in_progress | client_id: %v", c.config.ID)
		batchBets := bets[betsSent:min(betsSent+c.config.BatchMaxAmount, len(bets))]
		bets := communication.NewBets(c.config.ID, batchBets)
		betsMessageBytes := communication.SerializeBets(bets)
		err = communication.SendPacket(c.conn, betsMessageBytes)

		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		betsSent += len(batchBets)

		packet, err := communication.ReceivePacket(c.conn)

		if c.conn != nil {
			log.Infof("action: close_connection_after_use | result: in_progress | client_id: %v", c.config.ID)
			c.conn.Close()
			log.Infof("action: close_connection_after_use | result: success | client_id: %v", c.config.ID)
			c.conn = nil
		}

		if err != nil {
			log.Errorf("action: receive_packet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_packet | result: success | client_id: %v",
			c.config.ID,
		)

		msg, err := communication.DeserializePacket(packet)

		if err != nil {
			log.Errorf("action: deserialize_packet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		betsConfirmation, ok := msg.(communication.BetsConfirmationMessage)
		if ok {
			log.Infof("action: apuesta_enviada | result: success | client_id: %v", c.config.ID)
		}

		if betsConfirmation.Result == communication.BetsConfirmationResultOk {
			log.Infof("action: apuesta_almacenada | result: success | client_id: %v", c.config.ID)
		} else {
			log.Infof("action: apuesta_almacenada | result: fail | client_id: %v", c.config.ID)
		}

		// Wait a time between sending one message and the next one
		// Use select to either wait for the period or for shutdown signal
		select {
		case <-time.After(c.config.LoopPeriod):
			// Continue to next iteration
		case <-c.shutdownChan:
			log.Infof("action: loop_interrupted | result: success | client_id: %v", c.config.ID)
			return
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)

	c.notifyFinishedSendingBets()

	c.getLotteryWinners()
}

func (c *Client) notifyFinishedSendingBets() {
	err := c.createClientSocket()
	if err != nil {
		return
	}
	defer func() {
		if c.conn != nil {
			log.Infof("action: close_connection_after_use | result: in_progress | client_id: %v", c.config.ID)
			c.conn.Close()
			log.Infof("action: close_connection_after_use | result: success | client_id: %v", c.config.ID)
			c.conn = nil
		}
	}()

	log.Infof("action: notificar_fin_apuestas | result: in_progress | client_id: %v", c.config.ID)

	finishedSendingBetsMessage := communication.NewFinishedSendingBetsMessage(c.config.ID)
	finishedSendingBetsMessageBytes := communication.SerializeFinishedSendingBetsMessage(finishedSendingBetsMessage)

	err = communication.SendPacket(c.conn, finishedSendingBetsMessageBytes)
	if err != nil {
		log.Errorf("action: notificar_fin_apuestas | result: fail | client_id: %v | error: %v", c.config.ID, err)
		return
	}

	log.Infof("action: notificar_fin_apuestas | result: success | client_id: %v", c.config.ID)
}

func (c *Client) getLotteryWinners() {
	for {
		err := c.createClientSocket()
		if err != nil {
			return
		}
		defer func() {
			if c.conn != nil {
				log.Infof("action: close_connection_after_use | result: in_progress | client_id: %v", c.config.ID)
				c.conn.Close()
				log.Infof("action: close_connection_after_use | result: success | client_id: %v", c.config.ID)
				c.conn = nil
			}
		}()

		log.Infof("action: consulta_ganadores | result: in_progress | client_id: %v", c.config.ID)

		lotteryWinnersRequestMessage := communication.NewLotteryWinnersRequestMessage(c.config.ID)
		lotteryWinnersRequestMessageBytes := communication.SerializeLotteryWinnersRequestMessage(lotteryWinnersRequestMessage)

		err = communication.SendPacket(c.conn, lotteryWinnersRequestMessageBytes)
		if err != nil {
			log.Errorf("action: consulta_ganadores | result: fail | client_id: %v | error: %v", c.config.ID, err)
			return
		}

		packet, err := communication.ReceivePacket(c.conn)

		if c.conn != nil {
			log.Infof("action: close_connection_after_use | result: in_progress | client_id: %v", c.config.ID)
			c.conn.Close()
			log.Infof("action: close_connection_after_use | result: success | client_id: %v", c.config.ID)
			c.conn = nil
		}

		if err != nil {
			log.Errorf("action: receive_packet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		log.Infof("action: receive_packet | result: success | client_id: %v",
			c.config.ID,
		)

		msg, err := communication.DeserializePacket(packet)

		if err != nil {
			log.Errorf("action: deserialize_packet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		lotteryWinnersResponseMessage, ok := msg.(communication.LotteryWinnersResponseMessage)
		if ok && lotteryWinnersResponseMessage.Status == communication.LotteryWinnersResponseStatusReady {
			log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(lotteryWinnersResponseMessage.WinnersDnis))
			break
		}

		select {
		case <-time.After(time.Millisecond * MillisecondsBetweenLotteryWinnersRequests):
			// Continue to next iteration
		case <-c.shutdownChan:
			log.Infof("action: consulta_ganadores_interrupted | result: success | client_id: %v", c.config.ID)
			return
		}
	}
}
