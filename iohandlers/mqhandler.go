/*
 * ZDNS Copyright 2022 Regents of the University of Michigan
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not
 * use this MQ except in compliance with the License. You may obtain a copy
 * of the License at http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
 * implied. See the License for the specific language governing
 * permissions and limitations under the License.
 */

package iohandlers

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	log "github.com/sirupsen/logrus"
)

type MQInputHandler struct {
	path string
}

func NewMQInputHandler(path string) *MQInputHandler {
	return &MQInputHandler{
		path: path,
	}
}

func parseMQMessgage(name []byte) (parsedName string) {
	tmp := string(name)
	tmp = strings.TrimSpace(tmp)
	tmp = strings.TrimSuffix(tmp, "\n")
	return tmp
}

// FeedChannel 是用于接收数据
func (h *MQInputHandler) FeedChannel(in chan<- interface{}, wg *sync.WaitGroup) error {
	defer close(in)
	defer (*wg).Done()

	// open config file
	data, err := ioutil.ReadFile(h.path)
	if err != nil {
		log.Fatal("Open Rabbitmq Config File Error, %v", err)
	}

	config_slice := strings.Split(string(data), ",")
	user := config_slice[0]
	password := config_slice[1]
	url := config_slice[2]
	queue := config_slice[3]

	dial_string := fmt.Sprintf("amqp://%s:%s@%s/", user, password, url)

	conn, err := amqp.Dial(dial_string)

	if err != nil {
		log.Fatalf("Connect to Rabbitmq error: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Connect to Channel error: %v", err)
	}

	q, err := ch.QueueDeclare(
		queue, false, false, false, false, nil,
	)
	if err != nil {
		log.Fatalf("Declare Queue error: %v", err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to register a consumer: %v", err)
	}

	//var forever chan struct{}

	for d := range msgs {
		name := parseMQMessgage(d.Body)
		log.Printf("Received a message: %s", name)
		in <- name
	}

	return nil
}
