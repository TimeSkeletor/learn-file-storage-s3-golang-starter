.PHONY: run

set_aws:
	@echo "Downloading AWS Cli..."
	curl "https://awscli.amazonaws.com/AWSCLIV2.pkg" -o "AWSCLIV2.pkg"
	@echo "Installing AWS Cli..."
	sudo installer -pkg AWSCLIV2.pkg -target /
	@echo "Aws Cli installed at: "
	which aws
	@echo "version: "
	aws --version

installdepsmac:
	@echo "Installing dependencies"
	brew update
	brew install gcc
	go env -w CGO_ENABLED=1
	brew install ffmpeg
	brew install sqlite3

installdepslinux:
	@echo "Installing dependencies"
	sudo apt update
	sudo apt install gcc
	go env -w CGO_ENABLED=1
	sudo apt install ffmpeg
	sudo apt install sqlite3

setsampleimages:
	./samplesdownload.sh

run:
	go run .