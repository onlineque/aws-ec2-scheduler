FROM golang:1.20 AS build-stage
LABEL authors="online"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /aws-ec2-scheduler && \
    strip /aws-ec2-scheduler

FROM gcr.io/distroless/base-debian11 AS build-release-stage
WORKDIR /
COPY --from=build-stage /aws-ec2-scheduler /aws-ec2-scheduler
USER nonroot:nonroot
ENTRYPOINT ["/aws-ec2-scheduler"]
