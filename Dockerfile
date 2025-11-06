
FROM node:25 AS build

RUN npm install -g pnpm
RUN npm i -g typescript

WORKDIR /build/frontend
COPY frontend/package.json frontend/pnpm-lock.yaml ./
RUN pnpm install

COPY frontend ./
RUN pnpm run build


FROM golang:1.25-trixie AS go-build
RUN apt update && apt install -y \
    --no-install-recommends --no-install-suggests \
    build-essential \
    gcc \
    sqlite3 \
    libsqlite3-dev

WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend ./
ENV CGO_ENABLED=1
RUN go build -tags "sqlite" \
    -ldflags '-linkmode external -extldflags "-static"' \
    -o http /app/cmd/slotswapper/main.go
RUN sed -i 's|"frontendDir":.*|"frontendDir": "/app/frontend/dist"|g' /app/config.json

FROM scratch AS final

WORKDIR /app
COPY --from=go-build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /build/frontend/dist /app/frontend/dist
COPY --from=go-build /app/http /app/http
COPY --from=go-build /app/config.json /app/config.json
COPY --from=go-build /app/db/migrations/ /app/db/migrations/


EXPOSE 8080
ENTRYPOINT [ "/app/http" ]