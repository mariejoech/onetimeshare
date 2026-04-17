#include .env
#export $(shell sed 's/=.*//' .env)
build:
	 go run .
