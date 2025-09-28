##
## Deploy the application binary into a lean image
##
FROM harbor.htymeta.com/gcr.io/distroless/cc-debian12:debug-nonroot

WORKDIR /app

COPY go-sip-app go-sip-app

USER nonroot:nonroot

ENTRYPOINT ["./go-sip-app"]