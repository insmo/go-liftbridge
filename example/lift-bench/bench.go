package main

import (
	"fmt"
	"math/rand"
	"time"

	lift "github.com/liftbridge-io/go-liftbridge"
	"github.com/liftbridge-io/go-liftbridge/liftbridge-grpc"
	"github.com/nats-io/go-nats"
	"golang.org/x/net/context"
)

const (
	msgSize = 10
	numMsgs = 1000000
)

var keys = [][]byte{[]byte("foo"), []byte("bar"), []byte("baz"), []byte("qux")}

func main() {
	addrs := []string{"localhost:9292"}
	client, err := lift.Connect(addrs)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	if err := client.CreateStream(context.Background(), "bar", "bar-stream", lift.MaxReplication()); err != nil {
		if err != lift.ErrStreamExists {
			panic(err)
		}
	}

	conn, err := nats.DefaultOptions.Connect()
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ackInbox := "ack"
	acked := 0
	ch := make(chan struct{})

	sub, err := conn.Subscribe(ackInbox, func(m *nats.Msg) {
		acked++
		if acked >= numMsgs {
			ch <- struct{}{}
		}
	})
	if err != nil {
		panic(err)
	}
	sub.SetPendingLimits(-1, -1)

	msg := make([]byte, msgSize)

	start := time.Now()
	for i := 0; i < numMsgs; i++ {
		m := lift.NewMessage(msg, lift.MessageOptions{
			Key:       keys[rand.Intn(len(keys))],
			AckInbox:  ackInbox,
			AckPolicy: proto.AckPolicy_ALL,
		})
		if err := conn.Publish("bar", m); err != nil {
			panic(err)
		}
	}

	<-ch
	elapsed := time.Since(start)
	fmt.Printf("Elapsed: %s, Msgs: %d, Msgs/sec: %f\n", elapsed, numMsgs, numMsgs/elapsed.Seconds())
}
