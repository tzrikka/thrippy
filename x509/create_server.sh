#!/bin/bash

# Server CA
openssl req -x509                                           \
  -newkey rsa:4096                                          \
  -noenc                                                    \
  -days 3650                                                \
  -out thrippy_server_ca_cert.pem                           \
  -keyout thrippy_server_ca_key.pem                         \
  -subj /C=US/ST=CA/O=Tzrikka/OU=Thrippy/CN=test-server-ca/ \
  -config ./openssl.cnf                                     \
  -extensions test_ca                                       \
  -sha256

# Server certificate
openssl genrsa -out thrippy_server_key.pem 4096

openssl req -new                                         \
  -key thrippy_server_key.pem                            \
  -out thrippy_server_csr.pem                            \
  -subj /C=US/ST=CA/O=Tzrikka/OU=Thrippy/CN=test-server/ \
  -config ./openssl.cnf                                  \
  -reqexts test_server

openssl x509 -req                  \
  -in thrippy_server_csr.pem       \
  -CA thrippy_server_ca_cert.pem   \
  -CAkey thrippy_server_ca_key.pem \
  -days 3650                       \
  -out thrippy_server_cert.pem     \
  -extfile ./openssl.cnf           \
  -extensions test_server          \
  -CAcreateserial                  \
  -sha256

openssl verify -verbose -CAfile thrippy_server_ca_cert.pem thrippy_server_cert.pem

rm *_csr.pem
