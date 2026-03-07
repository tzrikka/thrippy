#!/bin/bash

openssl genrsa -out thrippy_client_key_2.pem 4096

openssl req -new                                           \
  -key thrippy_client_key_2.pem                            \
  -out thrippy_client_csr_2.pem                            \
  -subj /C=US/ST=CA/O=Tzrikka/OU=Thrippy/CN=test-client-2/ \
  -config ./openssl.cnf                                    \
  -reqexts test_client

openssl x509 -req                  \
  -in thrippy_client_csr_2.pem     \
  -CA thrippy_client_ca_cert.pem   \
  -CAkey thrippy_client_ca_key.pem \
  -days 3650                       \
  -out thrippy_client_cert_2.pem   \
  -extfile ./openssl.cnf           \
  -extensions test_client          \
  -CAcreateserial                  \
  -sha256

openssl verify -verbose -CAfile thrippy_client_ca_cert.pem thrippy_client_cert_2.pem

rm *_csr.pem
