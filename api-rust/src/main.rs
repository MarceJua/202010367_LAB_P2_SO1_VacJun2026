use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};
use std::env;

#[derive(Deserialize, Serialize, Debug)]
struct Prediction {
    home_team: String,
    away_team: String,
    home_goals: i32,
    away_goals: i32,
    username: String,
    timestamp: String,
}

#[post("/")]
async fn receive_prediction(prediction: web::Json<Prediction>) -> impl Responder {
    println!("[RUST] JSON recibido de Locust, enviando a Go Client...");

    // Leemos la variable de entorno, si no existe, usamos localhost
    let go_client_url = env::var("GO_CLIENT_URL").unwrap_or_else(|_| "http://localhost:8081/".to_string());

    let client = reqwest::Client::new();
    
    let res = client.post(&go_client_url)
        .json(&prediction.into_inner())
        .send()
        .await;

    match res {
        Ok(response) => {
            println!("[RUST] Enviado con éxito a Go en {}.", go_client_url);
            HttpResponse::Ok().body(response.text().await.unwrap_or_default())
        }
        Err(e) => {
            println!("[RUST] Error al enviar a Go: {}", e);
            HttpResponse::InternalServerError().body("Error comunicando con Go Client")
        }
    }
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!("Servidor Rust iniciando en http://0.0.0.0:8080");
    
    HttpServer::new(|| {
        App::new().service(receive_prediction)
    })
    .bind(("0.0.0.0", 8080))? // Es mejor bindear a 0.0.0.0 para Docker
    .run()
    .await
}
