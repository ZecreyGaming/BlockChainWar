# syntax=docker/dockerfile:1.0.0-experimental

FROM golang:alpine AS builder
LABEL stage=gobuilder
COPY . /block-chain-war
WORKDIR /block-chain-war/
RUN chmod 777 ./wait-for-it.sh
ENV CGO_ENABLED 0
ENV PATH="${PATH}:${GOPATH}/bin"
ENV GOPROXY https://goproxy.cn,direct
ENV GO111MODULE=on

ENV GOPRIVATE="*zecrey-crypto/,*zecrey-eth-rpc/"
ENV GONOPROXY="*zecrey-crypto/,*zecrey-eth-rpc/"
ENV GONOSUMDB="*zecrey-crypto/,*zecrey-eth-rpc/"

RUN apk update && apk upgrade && apk add --no-cache make gcc  git libc-dev openssh-client ca-certificates tzdata protoc && rm -rf /var/cache/apk/*

RUN mkdir ~/.ssh
RUN ssh-keyscan github.com >> ~/.ssh/known_hosts
RUN git config --global url."git@github.com:zecrey-labs".insteadOf "https://github.com/zecrey-labs"
RUN --mount=type=ssh,id=github go build -ldflags "-X main.CodeVersion=`git describe --tags` -X main.GitCommitHash=`git rev-parse --short HEAD` -linkmode=external -extldflags=-static" -o main
ENTRYPOINT ["./main"]