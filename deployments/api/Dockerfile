# build stage
FROM golang:1.25 AS builder

ARG GH_TOKEN

# set working directory
WORKDIR /app

# copy source code
COPY . .

# configure git to use ssh instead of https
ENV GOPRIVATE=github.com/brxyxn/*
#RUN mkdir -p -m 0700 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts
#RUN git config --global url."git@github.com:".insteadOf "https://github.com/"

RUN if [ ! -z "$GH_TOKEN" ]; then \
      git config --global url."https://user:${GH_TOKEN}@github.com".insteadOf "https://github.com"; \
    else \
      mkdir -p -m 0700 ~/.ssh && \
      ssh-keyscan github.com >> ~/.ssh/known_hosts && \
      git config --global url."git@github.com:".insteadOf "https://github.com/"; \
    fi

# build binary
RUN --mount=type=ssh CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ./bin/enginecare-api ./cmd/server/main.go

FROM alpine:3.22 AS runner

WORKDIR /app
COPY --from=builder /app/bin/enginecare-api .

EXPOSE 4000

CMD ["./enginecare-api"]
