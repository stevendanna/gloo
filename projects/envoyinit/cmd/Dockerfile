FROM quay.io/solo-io/envoy-gloo:2aa832e6fc87137d2e830b45bca2143d0af37b12

COPY envoyinit-linux-amd64 /usr/local/bin/envoyinit

ENTRYPOINT ["/usr/bin/dumb-init", "--", "/usr/local/bin/envoyinit"]
CMD []