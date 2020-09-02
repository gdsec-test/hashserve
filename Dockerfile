FROM alpine:3.9
RUN apk update && \
        apk --no-cache add ca-certificates \
        openjdk8 \
        bash
WORKDIR /app

COPY . /app/

ENTRYPOINT ["/app/hashserve"]