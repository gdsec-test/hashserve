FROM alpine:latest
RUN apk update && \
        apk --no-cache add ca-certificates \
        openjdk8 \
        bash
WORKDIR /app

RUN mkdir -p /app/tmp/pdna

#COPY hashserve /app/
COPY . /app/

ENV JAVA_HOME /usr/lib/jvm/java-1.8-openjdk
RUN export JAVA_HOME

ENV LD_LIBRARY_PATH /usr/local/lib
RUN export LD_LIBRARY_PATH

ENV PDNACLASSPATH /app/pdna/bin/java
RUN export PDNACLASSPATH

ENV CLASSPATH /app/pdna/bin/java
RUN export CLASSPATH

ENV DOWNLOAD_FILE_LOC /app/tmp/pdna
RUN export DOWNLOAD_FILE_LOC

# Copy the appropriate .so files into /usr/lib/
RUN cp pdna/tmp/lib/PhotoDNAx64.so.1.72 /usr/lib/
ENTRYPOINT ["/app/hashserve"]