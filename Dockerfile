FROM golang:1.15

WORKDIR /go/src/auth_clean_architecture
ENV PATH="/go/bin:${PATH}"
ENV GO111MODULE=on
ENV CGO_ENABLED=1

RUN apt-get update && \
  apt-get install build-essential protobuf-compiler librdkafka-dev dsniff -y && \
  go get google.golang.org/protobuf/cmd/protoc-gen-go && \
  wget https://github.com/ktr0731/evans/releases/download/0.9.1/evans_linux_amd64.tar.gz && \
  tar -xzvf evans_linux_amd64.tar.gz && \
  mv evans ../bin && rm -f evans_linux_amd64.tar.gz

CMD ["tail", "-f", "/dev/null"]

EXPOSE 3000
