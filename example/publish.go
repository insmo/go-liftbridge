package main

import (
	"fmt"
	"github.com/nats-io/go-nats"
	"github.com/tylertreat/go-jetbridge"
	"strconv"
	"sync"
)

func main() {
	conn, err := nats.DefaultOptions.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Flush()
	defer conn.Close()

	ackInbox := "acks"
	var wg sync.WaitGroup

	sub, err := conn.Subscribe(ackInbox, func(m *nats.Msg) {
		ack, err := jetbridge.UnmarshalAck(m.Data)
		if err != nil {
			panic(err)
		}
		fmt.Println("ack:", ack.StreamSubject, ack.StreamName, ack.Offset, ack.MsgSubject)
		wg.Done()
	})
	if err != nil {
		panic(err)
	}
	defer sub.Unsubscribe()

	count := 5
	wg.Add(count)
	for i := 0; i < count; i++ {
		m := jetbridge.NewEnvelope([]byte("test"), []byte(strconv.Itoa(i)), ackInbox)
		if err := conn.Publish("foo", m); err != nil {
			panic(err)
		}
	}

	wg.Wait()
}