.PHONY: docker
docker:
	@rm webook || true
	@GOOS=linux go build -tags=k8s -o webook .
	@docker rmi -f xiaoxin/webook:0.0.2
	@docker build -t xiaoxin/webook:0.0.2 .