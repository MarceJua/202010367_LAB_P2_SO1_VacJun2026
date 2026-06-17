use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};

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

    let client = reqwest::Client::new();
    
    // Hacemos el POST al Microservicio 2 (Go Client) en el puerto 8081
    let res = client.post("http://localhost:8081/")
        .json(&prediction.into_inner())
        .send()
        .await;

    match res {
        Ok(response) => {
            println!("[RUST] Enviado con éxito a Go.");
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
    println!("Servidor Rust iniciando en http://127.0.0.1:8080");
    
    HttpServer::new(|| {
        App::new().service(receive_prediction)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
