#!/bin/bash

# Create the server CA cert.
openssl req -x509                                     \
  -newkey rsa:4096                                    \
  -noenc                                              \
  -days 3650                                          \
  -keyout server_ca_key.pem                           \
  -out server_ca_cert.pem                             \
  -subj /C=US/ST=CA/L=SVL/O=gRPC/CN=test-server_ca/   \
  -config ./openssl.cnf                               \
  -extensions test_ca                                 \
  -sha256

# Generate a server cert.
openssl genrsa -out server_key.pem 4096

openssl req -new                                    \
  -key server_key.pem                               \
  -out server_csr.pem                               \
  -subj /C=US/ST=CA/L=SVL/O=gRPC/CN=test-server1/   \
  -config ./openssl.cnf                             \
  -reqexts test_server

openssl x509 -req           \
  -in server_csr.pem        \
  -CAkey server_ca_key.pem  \
  -CA server_ca_cert.pem    \
  -days 3650                \
  -set_serial 1000          \
  -out server_cert.pem      \
  -extfile ./openssl.cnf    \
  -extensions test_server   \
  -sha256

openssl verify -verbose -CAfile server_ca_cert.pem server_cert.pem

rm *_csr.pem
