package config

import (
	"time"
)

const (
	RABBITMQ_URL = "amqp://guest:guest@localhost:8081/" // change this base on your rabbitmq url
	REDIS_URL    = "localhost:6379"
	QUEUE_NAME   = "report_requests"

	// KEY_PREFIX_REPORT_STATUS is used to store report status in Redis
	KEY_PREFIX_REPORT_STATUS = "report:status:"
	// KEY_PREFIX_REPORT_DATA is used to store report result data in Redis
	KEY_PREFIX_REPORT_DATA = "report:data:"

	NUM_PRODUCER_REQUESTS = 10                     // Number of report requests to create
	NUM_WORKERS           = 3                      // Number of concurrent workers to process reports
	WORKER_TIMEOUT        = 5 * time.Second        // Timeout per worker for each task
	PUBLISH_INTERVAL      = 500 * time.Millisecond // Interval for producer to send messages
)
