docker-build:
	docker build -t codeexecute:1.0.0 .

helm-install:
	helm upgrade --install chart chart

helm-uninstall:
	helm uninstall chart
