NAME = $(notdir $(shell pwd))

build:
	go build -o ${NAME} -v main.go

docker:
	docker build -t ${NAME} .

clean:
	rm -rf ${NAME}