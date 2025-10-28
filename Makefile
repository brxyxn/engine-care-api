build:
	docker build -t engine-care-api:latest --file=./deployments/api/Dockerfile --ssh default="$(HOME)/.ssh/id_ed25519" .