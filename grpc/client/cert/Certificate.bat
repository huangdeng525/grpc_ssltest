@echo off
echo Generating certificate....
rem  Generate valid CA
openssl genrsa -passout pass:1234 -aes128 -out ca.key 2048
openssl req -passin pass:1234 -new -x509 -days 365 -key ca.key -out ca.crt -subj  "/C=CN/ST=SICHUAN/L=CHENGDU/O=Test/OU=andisat/CN=andisat"
rem  Generate valid Server Key/Cert
openssl genrsa -passout pass:1234 -aes128 -out server.key 2048
openssl req -passin pass:1234 -new -key server.key -out server.csr -subj  "/C=CN/ST=SICHUAN/L=CHENGDU/O=Test/OU=andisat/CN=andisat"
openssl x509 -req -passin pass:1234 -days 365 -in server.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out server.crt
rem  Remove passphrase from the Server Key
openssl rsa -passin pass:1234 -in server.key -out server.key
rem  Generate valid Client Key/Cert
openssl genrsa -passout pass:1234 -aes128 -out client.key 2048
openssl req -passin pass:1234 -new -key client.key -out client.csr -subj  "/C=CN/ST=SICHUAN/L=CHENGDU/O=Test/OU=andisat/CN=andisat"
openssl x509 -passin pass:1234 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out client.crt
rem  Remove passphrase from Client Key живЊ
openssl rsa -passin pass:1234 -in client.key -out client.key
echo Generate certificate finished....
@pause