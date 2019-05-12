STACK_CONFIG ?= config.json

ifeq (,$(wildcard $(STACK_CONFIG)))
    $(error $(STACK_CONFIG) is not found)
endif

StackName := $(shell cat $(STACK_CONFIG) | jq '.["StackName"]')
Region := $(shell cat $(STACK_CONFIG) | jq '.["Region"]')
CodeS3Bucket := $(shell cat $(STACK_CONFIG) | jq '.["CodeS3Bucket"]')
CodeS3Prefix := $(shell cat $(STACK_CONFIG) | jq '.["CodeS3Prefix"]')

# PARAMETERS := $(shell cat $(STACK_CONFIG) | grep -e LambdaRoleArn -e SecretArn -e DeepAlertStack | tr '\n' ' ')

LambdaRoleArn := LambdaRoleArn=$(shell cat $(STACK_CONFIG) | jq '.LambdaRoleArn | select(. != null)')
SecretArn := SecretArn=$(shell cat $(STACK_CONFIG) | jq '.["SecretArn"]')

TEMPLATE_FILE=template.yml
SAM_FILE=sam.yml

ifneq (, $(strip $(PARAMETERS)))
	PARAMETERS_OVERRIDES=--parameter-overrides $(PARAMETERS)
else
	PARAMETERS_OVERRIDES=
endif

all: deploy

test:
	go test -v

clean:
	rm build/main

build/main: *.go
	env GOARCH=amd64 GOOS=linux go build -o build/main

sam.yml: $(TEMPLATE_FILE) build/main
	aws cloudformation package \
		--region $(Region) \
		--template-file $(TEMPLATE_FILE) \
		--s3-bucket $(CodeS3Bucket) \
		--s3-prefix $(CodeS3Prefix) \
		--output-template-file $(SAM_FILE)

deploy: $(SAM_FILE)
	aws cloudformation deploy \
		--region $(Region) \
		--template-file $(SAM_FILE) \
		--stack-name $(StackName) \
		--capabilities CAPABILITY_IAM \
		--parameter-overrides $(LambdaRoleArn) $(SecretArn)
