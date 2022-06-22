package models

import (
	"context"
	"log"

	"github.com/Shopify/sarama"
	"github.com/cloudevents/sdk-go/protocol/kafka_sarama/v2"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/client"
	"github.com/omnibuildplatform/omni-manager/util"
)

const (
	TopicImageStatus = "omni-repository-image-status"
	GroupID          = "omni-manager"
)

var (
	saramaConfig *sarama.Config
	brokers      []string
	receiver     *kafka_sarama.Consumer
	err          error
	clientItem   client.Client
)

type ImageBlockStatus string

const (
	ImageCreated     ImageBlockStatus = "obp.omni_repository.image.created"
	ImageDownloading ImageBlockStatus = "obp.omni_repository.image.downloading"
	ImageDownloaded  ImageBlockStatus = "obp.omni_repository.image.downloaded"
	ImageVerifying   ImageBlockStatus = "obp.omni_repository.image.verifying"
	ImageVerified    ImageBlockStatus = "obp.omni_repository.image.verified"
	ImagePushing     ImageBlockStatus = "obp.omni_repository.image.pushing"
	ImagePushed      ImageBlockStatus = "obp.omni_repository.image.pushed"
	ImageFailed      ImageBlockStatus = "obp.omni_repository.image.failed"
)

func RegisterEventLinstener() {
	saramaConfig = sarama.NewConfig()
	saramaConfig.Version = sarama.V2_8_1_0
	brokers = []string{util.GetConfig().MQ.KafkaServer}
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	go registerEventLinstener(TopicImageStatus, handleDownloadStatusEvent)
}

func registerEventLinstener(topicName string, hanldFunc interface{}) {

	receiver, err = kafka_sarama.NewConsumer(brokers, saramaConfig, GroupID, topicName)
	if err != nil {
		log.Printf("failed to create protocol: %s", err.Error())

	}
	defer receiver.Close(context.Background())
	clientItem, err = cloudevents.NewClient(receiver, client.WithPollGoroutines(1))
	if err != nil {
		log.Printf("failed to create client, %v", err)
	}
	err = clientItem.StartReceiver(context.Background(), hanldFunc)
	if err != nil {
		log.Printf("failed to start receiver(%s) error: %s", topicName, err)

	}
	log.Printf(" TOPIC :(%s) 停止监听。\n", topicName)
}
