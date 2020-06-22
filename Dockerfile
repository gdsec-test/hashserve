FROM alpine:latest
RUN apk update && apk --no-cache add ca-certificates
COPY hashserve app/
WORKDIR /app
ENTRYPOINT ["/app/hashserve"]