#!/usr/bin/env bash

psql -c "drop schema public cascade; create schema public;"
