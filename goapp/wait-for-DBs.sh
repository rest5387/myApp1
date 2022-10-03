#!/usr/bin/env bash

# wait-for-postgres.sh
./wait-for-it postgres:5432
./wait-for-it neo4j:7687 
./wait-for-it redis:6379