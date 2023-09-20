FROM gdartifactory1.jfrog.io/docker-dcu-local/dcu-alpine3.15:3.3
COPY build/hashserve /app/
WORKDIR /app
ENTRYPOINT ["/app/hashserve"]