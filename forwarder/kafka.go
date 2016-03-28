package forwarder

import (
	"strings"
	"time"

	"github.com/Shopify/sarama"
	l "github.com/Sirupsen/logrus"
	"github.com/tsheasha/relayd/config"
)

func init() {
	RegisterForwarder("Kafka", newKafka)
}

// Kafka forwarder
type Kafka struct {
	BaseForwarder

	brokers []string
	conn    sarama.SyncProducer
	conf    *sarama.Config
}

// newKafka returns a new Kafka forwarder
func newKafka(
	initialBufferSize int,
	log *l.Entry) Forwarder {

	k := new(Kafka)
	k.name = "Kafka"

	k.log = log
	return k
}

// Configure the Kafka forwarder
func (k *Kafka) Configure(configMap map[string]interface{}) {
	k.conf = sarama.NewConfig()

	if v, exists := configMap["acks"]; exists {
		switch v.(string) {
		case "0":
			k.conf.Producer.RequiredAcks = sarama.NoResponse
		case "1":
			k.conf.Producer.RequiredAcks = sarama.WaitForLocal
		case "-1":
			k.conf.Producer.RequiredAcks = sarama.WaitForAll
		}
	}

	if v, exists := configMap["ack_timeout"]; exists {
		k.conf.Producer.Timeout = time.Duration(config.GetAsInt(v, 1000)) * time.Millisecond
	}

	if v, exists := configMap["batch_n"]; exists {
		k.conf.Producer.Flush.MaxMessages = config.GetAsInt(v, 100)
	}

	if v, exists := configMap["batch_t"]; exists {
		k.conf.Producer.Flush.Frequency = time.Duration(config.GetAsInt(v, 5)) * time.Second
	}

	if v, exists := configMap["brokers"]; exists {
		k.brokers = config.GetAsSlice(v)
	}

	if v, exists := configMap["close_timeout"]; exists {
		k.conf.Net.KeepAlive = time.Duration(config.GetAsInt(v, 1000)) * time.Millisecond
	}

	if v, exists := configMap["compression"]; exists {
		switch v.(string) {
		case "none":
			k.conf.Producer.Compression = sarama.CompressionNone
		case "gzip":
			k.conf.Producer.Compression = sarama.CompressionGZIP
		case "snappy":
			k.conf.Producer.Compression = sarama.CompressionSnappy
		}
	}

	if v, exists := configMap["retries"]; exists {
		k.conf.Producer.Retry.Max = config.GetAsInt(v, 5)
	}

	if v, exists := configMap["stagger"]; exists {
		k.conf.Producer.Retry.Backoff = time.Duration(config.GetAsInt(v, 1000)) * time.Millisecond
	}

	k.configureCommonParams(configMap)
}

// Run runs the forwarder main loop
func (k *Kafka) Run() {
	conn, err := sarama.NewSyncProducer(k.brokers, k.conf)
	if err != nil {
		k.log.Error("Failed to create Kafka producer ", err)
		return
	}

	k.conn = conn
	k.run(k.emitMsg)
}

func (k *Kafka) emitMsg(m []byte) bool {

	topic := strings.SplitN(string(m), ":", 1)[0]
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(m),
	}
	partition, offset, err := k.conn.SendMessage(msg)
	if err != nil {
		k.log.Error("Failed to send message to Kafka endpoint ", err)
		return false
	}

	k.log.Debug("Sent successfully to Kafka: ", partition, offset)
	return true
}
