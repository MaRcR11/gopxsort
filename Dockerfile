FROM scratch
ARG TARGETPLATFORM
COPY $TARGETPLATFORM/sortmancy /usr/bin/sortmancy
WORKDIR /work
ENTRYPOINT ["/usr/bin/sortmancy"]
