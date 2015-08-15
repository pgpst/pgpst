FROM scratch

COPY dist/* /
COPY ./certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/pgpst"]
CMD ["--help"]
