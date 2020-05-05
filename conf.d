server {
listen 80;
server_name mahaporn.com;
return 308 https://www.mahaporn.com$request_uri;
}

server {
listen 443 ssl;
server_name mahaporn.com;
ssl_certificate /etc/nginx/certs/example.com.crt;
ssl_certificate_key /etc/nginx/certs/example.com.key;
ssl_dhparam /etc/nginx/certs/example.com.dhparam.pem;
return 308 https://www.mahaporn.com$request_uri;
}

server {
        listen 80;
        server_name mahaporn.com;
        return 301 $scheme://www.mahaporn.com$request_uri;
}