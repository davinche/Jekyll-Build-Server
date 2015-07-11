FROM golang:1.4.2-wheezy

RUN apt-get update && apt-get install -y ruby-dev node
RUN gem install jekyll

COPY . /go/src/webserver
WORKDIR /go/src/webserver
RUN go get
RUN go build

EXPOSE 8080

CMD ["webserver"]
