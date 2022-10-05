#!/bin/bash
# build and run script for local machine or VM environment.
# Be careful, you need to check DBs are available by yourself.

# build and run myApp1
go build -o myApp1 cmd/web/*.go && ./myApp1
