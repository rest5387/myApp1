#!/bin/bash

sudo service postgresql start
neo4j start
redis-start
go build -o myApp1 cmd/web/*.go && ./myApp1