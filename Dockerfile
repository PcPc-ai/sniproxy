from debian:12

copy sniproxy /sniproxy
copy config.yaml /sniproxy.conf
copy domains.csv /domains.csv
entrypoint ["/sniproxy", "-config", "/sniproxy.conf"]