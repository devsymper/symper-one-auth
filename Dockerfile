FROM golang:1.23.7-alpine3.20 as build
ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux

RUN apk add --no-cache make git

WORKDIR /app

# Pulling dependencies
COPY ./Makefile ./go.* ./
RUN make deps

# Building stuff
COPY . /app

# Make sure you change the RELEASE_VERSION value before publishing an image.
# RUN RELEASE_VERSION=unspecified make build && ls -lh /app && ls -lh /app/bin || true

RUN go mod tidy && go build -o auth "./main.go" \
&& echo "----- FILES IN /app -----" \
&& ls -lh /app

# Always use alpine:3 so the latest version is used. This will keep CA certs more up to date.
FROM alpine:3
RUN adduser -D -u 1000 supabase

RUN apk add --no-cache ca-certificates

# Create /app and /src directories with proper permissions for supabase user
RUN mkdir -p /app /src && chown -R supabase:supabase /app /src
COPY --from=build /app/migrations /usr/local/etc/auth/migrations/
COPY --from=build /app/auth /usr/local/bin/auth
RUN ln -s /usr/local/bin/auth /usr/local/bin/gotrue
EXPOSE 9999
# Set working directory and switch to non-root user
WORKDIR /src
USER supabase
CMD ["auth"]
