gosimpleproxy
=============

This program indended to be a simple proxy which listens to HTTP connections and
redirects all the traffic to the different port, depending on HTTP Host header.

If you are using  the proxy from the Docker container, do  not forget to specify
`host` as the network driver.
