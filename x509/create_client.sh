#!/bin/bash

# Client CA
openssl req -x509                                           \
  -newkey rsa:4096                                          \
  -noenc                                                    \
  -days 3650                                                \
  -out thrippy_client_ca_cert.pem                           \
  -keyout thrippy_client_ca_key.pem                         \
  -subj /C=US/ST=CA/O=Tzrikka/OU=Thrippy/CN=test-client-ca/ \
  -config ./openssl.cnf                                     \
  -extensions test_ca                                       \
  -sha256

# Client certificate(s)
openssl genrsa -out thrippy_client_key.pem 4096

openssl req -new                                           \
  -key thrippy_client_key.pem                              \
  -out thrippy_client_csr.pem                              \
  -subj /C=US/ST=CA/O=Tzrikka/OU=Thrippy/CN=test-client-1/ \
  -config ./openssl.cnf                                    \
  -reqexts test_client

openssl x509 -req                  \
  -in thrippy_client_csr.pem       \
  -CA thrippy_client_ca_cert.pem   \
  -CAkey thrippy_client_ca_key.pem \
  -days 3650                       \
  -out thrippy_client_cert.pem     \
  -extfile ./openssl.cnf           \
  -extensions test_client          \
  -CAcreateserial                  \
  -sha256

openssl verify -verbose -CAfile thrippy_client_ca_cert.pem thrippy_client_cert.pem

rm *_csr.pem
