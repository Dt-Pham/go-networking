# Create a root CA Key (optional, only if not using external CA) 
# Optionally, you can protect the root CA with a password by
# adding -aes256 or similar 
openssl genrsa -out ca-key.pem 4096

# Create root cert with the CA key
openssl req -key ca-key.pem -new -x509 -days 7300 -out ca-cert.pem

################ Server Key/Cert (without extensions) ###############
# Create a server/client 2048-bit RSA private key
openssl genrsa -out localhost-key.pem 2048

# Verify the private key
openssl rsa -check -in localhost-key.pem

# Create a CSR for the new cert
openssl req -new -key localhost-key.pem -out localhost-csr.pem

# validate the CSR
openssl req -in localhost-csr.pem -noout -text

# Create a signed server cert with CA key good for 365 days
openssl x509 -req -in localhost-csr.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -days 365 -out localhost-cert.pem

# verify the cert
openssl x509 -noout -text -in localhost-cert.pem

# Verify that the private key matches its geneated public certificate
openssl rsa -noout -modulus -in localhost-key.pem | openssl md5
openssl req -noout -modulus -in localhost-csr.pem | openssl md5
openssl x509 -noout -modulus -in localhost-cert.pem | openssl md5

################ Client Key/Cert (without extensions) ###############
# Use the following to do mutual authentication of the client from server

# Create a client 2048-bit RSA private key
openssl genrsa -out client-key.pem 2048

# Verify the private key
openssl rsa -check -in client-key.pem

# Create a CSR for the new cert
openssl req -new -key client-key.pem -out client-csr.pem

# validate the CSR
openssl req -in client-csr.pem -noout -text

# Create a signed client cert with CA key good for 365 days
openssl x509 -req -in client-csr.pem -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial -days 365 -out client-cert.pem

# verify the cert
openssl x509 -noout -text -in client-cert.pem

# Verify that the private key matches its geneated public certificate
openssl rsa -noout -modulus -in client-key.pem | openssl md5
openssl req -noout -modulus -in client-csr.pem | openssl md5
openssl x509 -noout -modulus -in client-cert.pem | openssl md5