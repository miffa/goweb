package kafka

import (
	"time"

	"github.com/Shopify/sarama"
	log "goweb/iriscore/libs/logrus"
	"github.com/wvanbergen/kafka/consumergroup"
)

const (
	OFFSETS_PROCESSING_TIMEOUT_SECONDS = 10 * time.Second
	OFFSETS_COMMIT_INTERVAL            = 10 * time.Second
)

type Caller func([]byte) error

type KafkaConsumer struct {
	cgprocessor *consumergroup.ConsumerGroup
	quitKafkaCg chan struct{}
	cgname      string
	topic       string
	zkroot      string
	zkaddrs     []string
}

func NewKafkaConsumer(cge, toc, zroot string, zks []string) *KafkaConsumer {
	return &KafkaConsumer{cgname: cge, topic: toc, zkaddrs: zks, zkroot: zroot}
}

func (k *KafkaConsumer) Run(process Caller, logger *log.Logger) error {
	log.Infof("start topic:%s consumer", k.topic)
	log.Infof("consumer group name:%s", k.cgname)
	sarama.Logger = logger
	//sarama.Logger = log.StandardLogger()
	config := consumergroup.NewConfig()
	config.Offsets.Initial = sarama.OffsetNewest
	config.Offsets.ProcessingTimeout = OFFSETS_PROCESSING_TIMEOUT_SECONDS
	config.Offsets.CommitInterval = OFFSETS_COMMIT_INTERVAL
	config.Zookeeper.Chroot = k.zkroot
	kafkaTopics := []string{k.topic}
	var err error
	k.cgprocessor, err = consumergroup.JoinConsumerGroup(k.cgname, kafkaTopics, k.zkaddrs, config)
	if err != nil {
		return err
	}
	go func() {
		for err := range k.cgprocessor.Errors() {
			log.Errorf("consumer error(%v)", err)
		}
	}()

	k.quitKafkaCg = make(chan struct{}) //synchronous chan
	go func() {
		for {
			select {
			case msg := <-k.cgprocessor.Messages():
				log.Debugf("deal with topic:%s, partitionId:%d, Offset:%d, Key:%s msg:%s", msg.Topic, msg.Partition, msg.Offset, msg.Key, msg.Value)
				process(msg.Value)
				//RETRY:
				//if err:= process(msg.Value);err!=nil{
				//     goto RETRY
				//}
				k.cgprocessor.CommitUpto(msg)
			case <-k.quitKafkaCg:
				k.cgprocessor.Closed()
				log.Infof("consumer group exit %s %s %s", k.cgname, kafkaTopics, k.zkaddrs)
				return
			}
		}
	}()
	return nil
}

func (k *KafkaConsumer) Quit() {
	k.quitKafkaCg <- struct{}{}
}
