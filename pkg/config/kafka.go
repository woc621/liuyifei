package config

import (
	"context"
	"fmt"

	"github.com/Shopify/sarama"
	"github.com/hpcloud/tail"
)

type KafkaProducer struct {
	Client sarama.SyncProducer
	Msg    *sarama.ProducerMessage
}

func GetClientKafka(address []string) (client sarama.SyncProducer, err error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true

	client, err = sarama.NewSyncProducer(address, config)
	if err != nil {
		fmt.Println("connect kafka err : ", err)
		return
	}
	return
}

func (kafkaclient KafkaProducer) SendToKafka(ctx context.Context,topic string,lineCh chan *tail.Line) {
	kafkamsg := &sarama.ProducerMessage{}
	kafkamsg.Topic = topic
	for {
		select {
		case <-ctx.Done():
			fmt.Printf("ctx.Done cancel")
			close(lineCh)
			return
		case data := <- lineCh:
			kafkamsg.Value=sarama.StringEncoder(data.Text)
			_, _, err := kafkaclient.Client.SendMessage(kafkamsg)
			if err != nil {
				fmt.Println("send msg failed, err:", err)
			}
		}

	}
	//fmt.Printf("pid:%v offset:%v\n", pid, offset)
}
