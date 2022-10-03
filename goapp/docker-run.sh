#!/usr/bin/env bash

# wait for DB service start
./wait-for-it.sh postgres:5432
./wait-for-it.sh neo4j:7687 
./wait-for-it.sh redis:6379

./myApp1