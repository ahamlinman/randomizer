deploy: Package.yaml
	aws cloudformation deploy \
		--template-file Package.yaml \
		--capabilities CAPABILITY_IAM \
		--stack-name $(RANDOMIZER_CF_STACK_NAME)

Package.yaml: Template.yaml dist
	aws cloudformation package \
		--template-file Template.yaml \
		--output-template-file Package.yaml \
		--s3-bucket $(RANDOMIZER_CF_S3_BUCKET)

.PHONY: dist # TODO: Make this not phony
dist:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 \
			 go build \
			 -ldflags='-s -w' \
			 -o dist/slack-lambda-handler \
			 ../cmd/slack-lambda-handler
