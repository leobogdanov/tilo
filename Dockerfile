FROM golang:1.8

RUN curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

COPY . /go/src/github.com/leobogdanov/tilo

WORKDIR /go/src/github.com/leobogdanov/tilo

RUN dep ensure

RUN go install

CMD ["tilo"]