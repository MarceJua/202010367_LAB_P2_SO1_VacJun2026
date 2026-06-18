# Etapa de construcción
FROM rust:1.79 AS builder
WORKDIR /usr/src/app
COPY Cargo.toml Cargo.lock ./
COPY src ./src
# Compilar en modo release para máxima velocidad
RUN cargo build --release

# Etapa de ejecución (imagen ultraligera)
FROM debian:bookworm-slim
# Instalamos las librerías de OpenSSL necesarias para reqwest
RUN apt-get update && apt-get install -y libssl-dev ca-certificates && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=builder /usr/src/app/target/release/api-rust .
EXPOSE 8080
CMD ["./api-rust"]
