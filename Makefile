.PHONY: build-all
build-all:
	make build-ping
	make build-sms

.PHONY: build-ping
build-ping:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o ./bin/ping ./cmd/ping

.PHONY: build-sms
build-sms:
	GOARCH=amd64 GOOS=linux CGO_ENABLED=0 go build -o ./bin/sms ./cmd/sms

.PHONY: invoke-sms
invoke-sms:
	TOPIC_ARN=$$(cd terraform/sms && tflocal output -raw sns_topic_arn) \
	&& awslocal sns publish \
		--topic-arn $$TOPIC_ARN \
		--message 'This is a test of the SMS lambda function'

.PHONY: lint
lint:
	golangci-lint run

.PHONY: render-sms
render-sms:
	set -a \
	&& source .env \
	&& cd terraform/sms \
	&& envsubst < terraform.tfvars.tmpl > terraform.tfvars

.PHONY: start-ping
start-ping:
	docker compose up -d
	cd terraform/ping \
	&& tflocal init \
	&& AWS_DEFAULT_REGION=us-east-1 tflocal apply -auto-approve

.PHONY: start-sms
start-sms:
	docker compose up -d
	cd terraform/sms \
	&& tflocal init \
	&& AWS_DEFAULT_REGION=us-east-1 tflocal apply -auto-approve

.PHONY: stop-ping
stop-ping:
	cd terraform/ping \
	&& tflocal init \
	&& AWS_DEFAULT_REGION=us-east-1 tflocal destroy -auto-approve

.PHONY: stop-sms
stop-sms:
	cd terraform/sms \
	&& tflocal init \
	&& AWS_DEFAULT_REGION=us-east-1 tflocal destroy -auto-approve
