FROM docker-dcu-local.artifactory.secureserver.net/alpine:3.9
RUN apk update && \
        apk --no-cache add ca-certificates \
        openjdk8 \
        bash
WORKDIR /app
# install custom root certificates
RUN mkdir -p /usr/local/share/ca-certificates/
COPY certs/* /usr/local/share/ca-certificates/
RUN update-ca-certificates

COPY hashserve /app/
RUN chmod +x /app/hashserve
ENTRYPOINT ["/app/hashserve"]