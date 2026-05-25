#!/bin/bash

mkdir data

cat schema.sql | sqlite3 data/data.db

go run . import_users users.csv
go run . initalize_election
