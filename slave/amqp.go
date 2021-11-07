package slave

import (
	"encoding/json"
	"fmt"
	"github.com/siddontang/go-log/log"
	"github.com/streadway/amqp"
)

type Client struct {
	conn *amqp.Connection
}

func NewClient(host string, port int32, user string, password string) (cli *Client, err error) {
	addr := fmt.Sprintf("amqp://%s:%s@%s:%d", user, password, host, port)
	conn, err := amqp.Dial(addr)
	if err != nil {
		log.Infof("amqp.Dial err:%s", addr)
		conn.Close()
		return
	}
	cli = &Client{conn: conn}
	return
}

func (cli *Client) pushMq(exchange string, routingKey string, body interface{}) (err error) {
	ch, err := cli.conn.Channel()
	if err != nil {
		return
	}

	data, err := json.Marshal(body)
	if err != nil {
		return
	}

	err = ch.Publish(
		exchange,
		routingKey, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(data),
		})
	log.Infof(" [x] Sent %s", body)
	return
}
