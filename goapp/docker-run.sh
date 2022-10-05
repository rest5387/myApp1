#!/usr/bin/env bash
# run myApp1 script for docker environment
# myApp1 is already build in docker container build stage
# can see .dockerfile [RUN] section.

# wait for DB service start
./wait-for-it.sh postgres:5432 -t 30
./wait-for-it.sh neo4j:7687 
./wait-for-it.sh redis:6379

./myApp1