# Build the browser dashboard and Go backend
FROM node:22-bookworm AS frontend
WORKDIR /app
COPY frontend/package*.json ./frontend/
RUN cd frontend && npm install
COPY frontend ./frontend
RUN cd frontend && npm run check && npm run build

FROM golang:1.25-bookworm AS backend
WORKDIR /app
COPY --from=frontend /app/frontend/dist ./frontend/dist
COPY . .
RUN go build -mod=vendor -o dmg-web .

# Small runtime image
FROM debian:bookworm-slim
WORKDIR /app
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=backend /app/dmg-web ./dmg-web
ENV PORT=5000
EXPOSE 5000
CMD ["./dmg-web"]
