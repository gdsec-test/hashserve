FROM docker-dcu-local.artifactory.secureserver.net/dcu-alpine3.15:3.3
COPY build/hashserve /app/
WORKDIR /app
ENTRYPOINT ["/app/hashserve"]