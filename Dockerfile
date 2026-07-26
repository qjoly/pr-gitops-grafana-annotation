FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/pr-gitops-grafana-annotation ./cmd/pr-gitops-grafana-annotation

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/pr-gitops-grafana-annotation /pr-gitops-grafana-annotation
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/pr-gitops-grafana-annotation"]
