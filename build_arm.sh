set -e

GOOS=linux GOARCH=arm64 go build *.go -o cabot_linux_arm64

for i in feed ping karma weather; do
	(
	cd plugins_examples/$i
	GOOS=linux GOARCH=arm64 go build -o ../../plugins/${i}_linux_arm64 ${i}.go
	)
done
