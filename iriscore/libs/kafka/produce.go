package kafka

import (
	"encoding/json"

	log "goweb/iriscore/libs/logrus"

	"github.com/Shopify/sarama"
)

type KafkaProducer struct {
	producer sarama.AsyncProducer
	topic    string
	kaddr    []string
	running  bool
}

var kproducer KafkaProducer

func Obj() *KafkaProducer {
	return &kproducer
}

func (k *KafkaProducer) InitKafka(kafkaAddrs []string, topos string, ishash bool, logger *log.Logger) (err error) {
	sarama.Logger = logger
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.NoResponse
	if ishash {
		config.Producer.Partitioner = sarama.NewHashPartitioner
	} else {
		//config.Producer.Partitioner = sarama.NewRandomPartitioner
		config.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	}
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	k.producer, err = sarama.NewAsyncProducer(kafkaAddrs, config)
	k.kaddr = kafkaAddrs
	k.running = true
	k.topic = topos
	go k.handleSuccess()
	go k.handleError()
	return
}

func (k *KafkaProducer) Stop() {
	k.running = false
}

func (k *KafkaProducer) handleSuccess() {
	var (
		pm *sarama.ProducerMessage
	)
	for k.running {
		pm = <-k.producer.Successes()
		if pm != nil {
			//log.Debugf("producer message success, partition:%d offset:%d key:%v valus:%s", pm.Partition, pm.Offset, pm.Key, pm.Value)
		} else {
			log.Errorf("error kafka produce vid:%s msg:%v", pm.Key, pm.Value)
		}

	}
}

func (k *KafkaProducer) handleError() {
	var (
		err *sarama.ProducerError
	)
	for k.running {
		err = <-k.producer.Errors()
		if err != nil {
			log.Errorf("producer message error, partition:%d offset:%d key:%v valus:%s error(%v)", err.Msg.Partition, err.Msg.Offset, err.Msg.Key, err.Msg.Value, err.Err)
		}
	}
}

func (k *KafkaProducer) PushDataToKafkaLB(key string, v interface{}) (err error) {
	var (
		vBytes []byte
	)
	if vBytes, err = json.Marshal(v); err != nil {
		log.Warnf("PushData4Kafka  json.Marshal err:%v", err)
		return
	}
	k.producer.Input() <- &sarama.ProducerMessage{Topic: k.topic, Key: sarama.ByteEncoder([]byte(key)), Value: sarama.ByteEncoder(vBytes)}
	return
}

func (k *KafkaProducer) PushDataToKafkaPoll(v interface{}) (err error) {
	var (
		vBytes []byte
	)
	if vBytes, err = json.Marshal(v); err != nil {
		return
	}
	kproducer.producer.Input() <- &sarama.ProducerMessage{Topic: k.topic, Value: sarama.ByteEncoder(vBytes)}
	return
}
