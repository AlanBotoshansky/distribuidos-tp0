package common

import (
	"net"
	"os"
	"strconv"
	"time"

	"github.com/op/go-logging"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/communication"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            uint32
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
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

// GetBetInfo Get the bet information from the environment variables
func (c *Client) getBetInfo() (string, string, string, string, uint32) {
	nombre := os.Getenv("NOMBRE")
	apellido := os.Getenv("APELLIDO")
	documento := os.Getenv("DOCUMENTO")
	nacimiento := os.Getenv("NACIMIENTO")
	numeroStr := os.Getenv("NUMERO")
	numeroInt, err := strconv.ParseUint(numeroStr, 10, 32)
	if err != nil {
		log.Errorf("action: parse_numero | result: fail | error: %v", err)
		numeroInt = 0
	}

	return nombre, apellido, documento, nacimiento, uint32(numeroInt)
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

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		err := c.createClientSocket()
		if err != nil {
			return
		}

		nombre, apellido, dni, nacimiento, numero := c.getBetInfo()
		log.Infof("action: apuesta_enviada | result: in_progress | dni: %v | numero: %v", dni, numero)
		betMessage := communication.NewBetMessage(c.config.ID, nombre, apellido, dni, nacimiento, uint32(numero))
		betMessageBytes := communication.SerializeBet(betMessage)
		err = communication.SendPacket(c.conn, betMessageBytes)

		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
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

		log.Infof("action: receive_packet | result: success | client_id: %v | msg: %v",
			c.config.ID,
			packet,
		)

		msg, err := communication.DeserializePacket(packet)

		if err != nil {
			log.Errorf("action: deserialize_packet | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}

		betConfirmation, ok := msg.(communication.BetConfirmationMessage)
		if ok {
			log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v", dni, numero)
		}

		if betConfirmation.Result == communication.BetConfirmationResultOk {
			log.Infof("action: apuesta_almacenada | result: success | dni: %v | numero: %v", dni, numero)
		} else {
			log.Infof("action: apuesta_almacenada | result: fail | dni: %v | numero: %v", dni, numero)
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
}
