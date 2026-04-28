#!/bin/bash

mkdir data

cat schema.sql | sqlite3 data/data.db

go run . import_users chapter_member_listing.csv

