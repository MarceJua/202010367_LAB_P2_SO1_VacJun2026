from locust import HttpUser, task, between
import random
import datetime
import json

class QuinielaUser(HttpUser):
    # Tiempo de espera entre peticiones de cada usuario (entre 1 y 3 segundos)
    wait_time = between(1.0, 3.0)

    # Lista de equipos válidos según el archivo proto
    teams = ["GTM", "MEX", "BRA", "ARG", "ESP"]

    @task
    def send_prediction(self):
        home_team = random.choice(self.teams)
        away_team = random.choice(self.teams)

        # Evitar que juegue el mismo equipo
        while away_team == home_team:
            away_team = random.choice(self.teams)

        # Generar goles aleatorios
        home_goals = random.randint(0, 5)
        away_goals = random.randint(0, 5)

        # Simular nombres de usuario
        username = f"user_{random.randint(1, 10)}"

        # Timestamp actual en formato ISO 8601
        timestamp = datetime.datetime.now(datetime.timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")

        payload = {
            "home_team": home_team,
            "away_team": away_team,
            "home_goals": home_goals,
            "away_goals": away_goals,
            "username": username,
            "timestamp": timestamp
        }

        headers = {'Content-Type': 'application/json'}

        # Enviar la petición POST al endpoint raíz
        self.client.post("/grpc-202010367", data=json.dumps(payload), headers=headers)
