# syntax=docker/dockerfile:1

# ---- build stage ----
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache module downloads before copying source (layer caching).
COPY go.mod go.sum ./
RUN go mod download

COPY . ./

# Build a fully static binary (CGO disabled, no libc dependency).
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/hatch ./cmd/hatch

# ---- run stage ----
FROM scratch

COPY --from=build /bin/hatch /bin/hatch

EXPOSE 8080

ENTRYPOINT ["/bin/hatch"]
