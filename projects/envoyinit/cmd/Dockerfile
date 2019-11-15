FROM quay.io/solo-io/envoy-gloo:0e9ac1a697e7b551878fa91b24e0ee81d84c943f

COPY envoyinit-linux-amd64 /usr/local/bin/envoyinit

ENTRYPOINT ["/usr/bin/dumb-init", "--", "/usr/local/bin/envoyinit"]
CMD []