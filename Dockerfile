FROM golang:1.17 AS build-env
WORKDIR /app
COPY . .
RUN go mod download
ENV GO111MODULE=on
ENV CGO_ENABLED=0
# The binary is renamed in the next build step.
RUN make binary BINARY=binary

FROM scratch
# Alternatively if one needs a shell for debugging...
#FROM gcr.io/distroless/base-debian10:debug
# Include certificate bundle so our binary can verify HTTPS authenticity.
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-env /app/binary /prme
CMD ["-h"]
ENTRYPOINT ["/prme"]

