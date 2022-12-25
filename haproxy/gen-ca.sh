#!/bin/bash
CSR_FILE=dev.csr
CFG_FILE=ssl_ca.cfg
KEY_FILE=dev.key
CRT_FILE=dev.crt
CA_CRT_FILE=ca.crt
CA_KEY_FILE=ca.key

openssl req -x509 -new -newkey rsa:2048 -nodes -keyout ${CA_KEY_FILE} -out ${CA_CRT_FILE} -days 365 -config ${CFG_FILE}