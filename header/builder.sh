#!/bin/bash

protoc header.proto -I. --gogofaster_out=.

protoc-go-tags --dir=.