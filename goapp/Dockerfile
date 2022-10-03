FROM golang:latest

WORKDIR /app

COPY . .

# ENV PATH=$PATH:/usr/bin

RUN go mod download
RUN go build -o myApp1 cmd/web/*.go
EXPOSE 8080

CMD [ "sh", "docker-run.sh" ]