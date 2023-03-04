#!/bin/bash
set -xe

GOOS=linux GOARCH=arm64 go build -o cabot_linux_arm64 *.go

for i in feed ping karma weather chatgpt; do
	echo $i
	(
	cd plugins_examples/$i
	GOOS=linux GOARCH=arm64 go build -o ../../plugins/${i}_linux_arm64 ${i}.go
	)
done
