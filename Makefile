# Makefile

export GOPATH := $(shell pwd)

NAME := renamer
RM := rm -rf
SRC := main.go models.go

all: build

build:
	echo $$GOPATH
	go get -d
	go build -o $(NAME) $(SRC)

re: clean all

clean:
	$(RM) $(NAME)

.PHONY: all build re clean
