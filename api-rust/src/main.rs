use actix_web::{post, web, App, HttpResponse, HttpServer, Responder};
use serde::{Deserialize, Serialize};

// Definimos la estructura exacta que enviará Locust
#[derive(Deserialize, Serialize, Debug)]
struct Prediction {
    home_team: String,
    away_team: String,
    home_goals: i32,
    away_goals: i32,
    username: String,
    timestamp: String,
}

// Ruta principal que recibirá las peticiones POST
#[post("/")]
async fn receive_prediction(prediction: web::Json<Prediction>) -> impl Responder {
    // Imprimimos el JSON recibido en la consola para verificar que funciona
    println!("[RUST] Nueva predicción recibida: {:?}", prediction);
    
    // Aquí (en el siguiente paso) agregaremos la lógica para enviarlo al cliente de Go
    
    // Respondemos con un 200 OK y devolvemos el mismo JSON
    HttpResponse::Ok().json(prediction.into_inner())
}

#[actix_web::main]
async fn main() -> std::io::Result<()> {
    println!(" Servidor Rust iniciando en http://127.0.0.1:8080");
    
    HttpServer::new(|| {
        App::new().service(receive_prediction)
    })
    .bind(("127.0.0.1", 8080))?
    .run()
    .await
}
