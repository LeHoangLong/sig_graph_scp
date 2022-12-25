#!/bin/bash
CSR_FILE=dev.csr
CFG_FILE=ssl_server.cfg
KEY_FILE=dev.key
CRT_FILE=dev.crt
CA_CRT_FILE=ca.crt
CA_KEY_FILE=ca.key

echo ${CSR_FILE}
openssl req -new -newkey rsa:4096 -nodes -keyout ${KEY_FILE} -out ${CSR_FILE} -config ${CFG_FILE} -extensions req_ext
openssl x509 -req -in ${CSR_FILE} -CA ${CA_CRT_FILE} -CAkey ${CA_KEY_FILE} -CAcreateserial -out ${CRT_FILE} -days 365 -extfile ${CFG_FILE} -extensions v3_req

cat dev.crt dev.key > both.crt