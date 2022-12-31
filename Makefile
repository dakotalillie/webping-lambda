.PHONY: build-all
build-all:
	make build-sms

.PHONY: build-sms
build-sms:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o ./bin/sms ./cmd/sms

.PHONY: invoke-sms
invoke-sms:
	TOPIC_ARN=$$(cd cmd/sms/terraform && tflocal output -raw sns_topic_arn) \
	&& awslocal sns publish \
		--topic-arn $$TOPIC_ARN \
		--message 'This is a test of the SMS lambda function'

.PHONY: start-sms
start-sms:
	docker-compose up -d
	cd cmd/sms/terraform && tflocal init && AWS_DEFAULT_REGION=us-east-1 tflocal apply -auto-approve

.PHONY: stop-sms
stop-sms:
	cd cmd/sms/terraform && tflocal init && AWS_DEFAULT_REGION=us-east-1 tflocal destroy -auto-approve

