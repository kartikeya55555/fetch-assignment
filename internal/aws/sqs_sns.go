package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	awsg "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/kartikeya55555/fetch-assignment/internal/models"
)

// SQSClient interface
type SQSClient interface {
	EnsureQueue() error
	SendMessage(receipt models.Receipt) error
	GetMessages() ([]*sqs.Message, error)
	DeleteMessage(receiptHandle *string) error
}

// SNSClient interface
type SNSClient interface {
	EnsureTopic() error
	Publish(message string) error
}

type sqsClientImpl struct {
	svc       *sqs.SQS
	queueName string
	queueURL  string
}

type snsClientImpl struct {
	svc       *sns.SNS
	topicName string
	topicARN  string
}

// NewSQSClient creates an SQS client
func NewSQSClient(region, endpoint, queueName string) SQSClient {
	sess := session.Must(session.NewSession(&awsg.Config{
		Region:     awsg.String(region),
		Endpoint:   awsg.String(endpoint),
		DisableSSL: awsg.Bool(true),
	}))
	svc := sqs.New(sess)
	return &sqsClientImpl{
		svc:       svc,
		queueName: queueName,
	}
}

// NewSNSClient creates an SNS client
func NewSNSClient(region, endpoint, topicName string) SNSClient {
	sess := session.Must(session.NewSession(&awsg.Config{
		Region:     awsg.String(region),
		Endpoint:   awsg.String(endpoint),
		DisableSSL: awsg.Bool(true),
	}))
	svc := sns.New(sess)
	return &snsClientImpl{
		svc:       svc,
		topicName: topicName,
	}
}

func (c *sqsClientImpl) EnsureQueue() error {
	for i := 1; i <= 5; i++ {
		out, err := c.svc.CreateQueue(&sqs.CreateQueueInput{
			QueueName: awsg.String(c.queueName),
		})
		if err == nil {
			c.queueURL = *out.QueueUrl
			log.Printf("SQS queue ready: %s -> %s\n", c.queueName, c.queueURL)
			return nil
		}
		log.Printf("Failed to create queue (attempt %d): %v\n", i, err)
		time.Sleep(3 * time.Second)
	}
	return fmt.Errorf("could not create queue after 5 attempts")
}

func (c *sqsClientImpl) SendMessage(receipt models.Receipt) error {
	if c.queueURL == "" {
		return fmt.Errorf("queue is not initialized")
	}
	data, err := json.Marshal(receipt)
	if err != nil {
		return err
	}
	_, err = c.svc.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    awsg.String(c.queueURL),
		MessageBody: awsg.String(string(data)),
	})
	return err
}

func (c *sqsClientImpl) GetMessages() ([]*sqs.Message, error) {
	if c.queueURL == "" {
		return nil, fmt.Errorf("queue is not initialized")
	}
	out, err := c.svc.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            awsg.String(c.queueURL),
		MaxNumberOfMessages: awsg.Int64(10),
		WaitTimeSeconds:     awsg.Int64(5),
	})
	if err != nil {
		return nil, err
	}
	return out.Messages, nil
}

func (c *sqsClientImpl) DeleteMessage(receiptHandle *string) error {
	_, err := c.svc.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      awsg.String(c.queueURL),
		ReceiptHandle: receiptHandle,
	})
	return err
}

func (s *snsClientImpl) EnsureTopic() error {
	out, err := s.svc.CreateTopic(&sns.CreateTopicInput{
		Name: awsg.String(s.topicName),
	})
	if err != nil {
		return err
	}
	s.topicARN = *out.TopicArn
	log.Printf("SNS topic ready: %s -> %s\n", s.topicName, s.topicARN)
	return nil
}

func (s *snsClientImpl) Publish(message string) error {
	if s.topicARN == "" {
		return fmt.Errorf("topic is not initialized")
	}
	_, err := s.svc.Publish(&sns.PublishInput{
		TopicArn: awsg.String(s.topicARN),
		Message:  awsg.String(message),
	})
	return err
}
