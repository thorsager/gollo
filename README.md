# gollo
Simple Test web-app in go


## A bit of config
Gollo will bind to the IP expressed in the ENV var `SERVER_IP` on the port expressed in `SERVER_PORT`. if `SERVER_IP` is
not set, Gollo will bin do all available addresses, and if `SERVER_PORT` is not set it will bind to port **8080**.

Also Gollo will print out the message found in the ENV var `GOLLO_MESSAGE` in every response.
