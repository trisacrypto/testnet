#!/bin/sh

# Drop the existing tables
psql -d $RVASP_DATABASE_DSN -c "SELECT 'DROP TABLE IF EXISTS ' || tablename || ' CASCADE;' FROM pg_tables;"

# Initialize the database with the fixtures
/bin/rvasp initdb