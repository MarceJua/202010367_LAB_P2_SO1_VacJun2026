
**Universidad de San Carlos de Guatemala (USAC)**

**Facultad de Ingeniería - Sistemas Operativos 1**

# INFORME TÉCNICO Y MANUAL DE DESPLIEGUE
## LABORATORIO - PROYECTO 2 (SISTEMAS OPERATIVOS 1)

**Identificación del Autor:**
* **Nombre:** Marcelo Andre Juarez Alfaro
* **Carnet:** 202010367
* **Enfoque de Análisis de Datos Asignado:** ARGENTINA (ARG)
* **Ciclo Académico:** Vacaciones de Junio 2026

---

## 1. ARQUITECTURA GENERAL DEL SISTEMA

El sistema se implementa bajo un esquema de microservicios distribuidos, orquestados de manera nativa mediante Google Kubernetes Engine (GKE). La infraestructura aprovecha las capacidades de la API de pasarelas de Kubernetes (Gateway API) junto con entornos de virtualización anidada administrados mediante KubeVirt para aislar las capas de persistencia y análisis de datos de los servicios de procesamiento de transacciones.

El clúster está configurado mediante nodos con capacidades de virtualización de hardware habilitadas (Nested Virtualization), lo que permite que pods especializados ejecuten hipervisores locales sobre los que corren máquinas virtuales basadas en Debian 12. La integración de componentes desacopla completamente el plano de ingesta (Locust y API Rust), el plano de traducción y encolamiento (Go Microservices), la capa de mensajería asíncrona (RabbitMQ), el procesador de persistencia (Go Consumer) y el clúster de almacenamiento e infraestructura de análisis (Valkey y Grafana en KubeVirt).
<img width="847" height="2629" alt="EsquemaA" src="https://github.com/user-attachments/assets/d3193097-1baf-4c53-8301-f8a62c76ecdd" />

---

## 2. FLUJO DE DATOS DETALLADO

El recorrido de la información a lo largo de la infraestructura distribuida comprende las siguientes fases síncronas y asíncronas:

1.  **Generación de Carga (Ingesta):** El agente de pruebas de carga Locust emula conexiones de múltiples usuarios concurrentes, enviando payloads en formato JSON puro al punto de enlace público del clúster.
2.  **Enrutamiento Perimetral:** Las peticiones HTTP ingresan a través del Balanceador de Carga Global de Google Cloud gestionado por la Gateway API bajo la ruta específica asignada al identificador del carnet (`/grpc-202010367`). El Gateway intercepta la trama, aplica una reescritura de prefijo para limpiar la ruta hacia la raíz (`/`) y redirecciona el tráfico al microservicio de Rust en el puerto 8080.
3.  **Procesamiento Asíncrono en Actix-Web:** La API construida en Rust recibe el JSON, valida la estructura y realiza una llamada HTTP POST asíncrona no bloqueante hacia la dirección interna del cliente de Go.
4.  **Traducción de Protocolos (REST a gRPC):** El microservicio `go-client` intercepta el payload JSON, realiza el mapeo de los datos hacia los tipos definidos en el archivo estructurado de Protocol Buffers (`quiniela.proto`) y despacha la quiniela mediante una llamada a procedimiento remoto gRPC dirigida al servidor interno.
5.  **Broker de Mensajería:** El `go-server` recibe la petición gRPC, extrae los parámetros, realiza la serialización de la estructura a bytes estructurados en formato JSON y publica el mensaje dentro del broker RabbitMQ, específicamente en la cola de persistencia duradera llamada `quiniela_queue`.
6.  **Consumo e Inserción:** El pod `go-consumer` actúa como un daemon de escucha constante acoplado a RabbitMQ. Al detectar un nuevo mensaje, lo extrae de la cola y realiza una operación atómica `LPUSH` en la base de datos distribuida Valkey, almacenándolo dentro de la lista denominada `"predicciones"`.
7.  **Extracción y Visualización:** La instancia de Grafana realiza consultas directas utilizando la interfaz de línea de comandos (CLI) enviando la instrucción `LRANGE predicciones 0 -1`. El flujo de datos es procesado dinámicamente mediante transformaciones de desestructuración de JSON e hilos de filtrado lógico para alimentar en tiempo real las métricas generales y de evolución temporal del equipo asignado (ARG).

---

## 3. GUÍA DE REPRODUCIBILIDAD Y DESPLIEGUE

Para recrear y desplegar la totalidad de la infraestructura de forma determinista en un entorno limpio de Kubernetes, ejecute la siguiente secuencia cronológica de comandos estructurados:

**Infraestructura Base (GCP):**
Para soportar la virtualización anidada requerida por KubeVirt, el clúster fue aprovisionado con un pool de nodos utilizando la familia de máquinas `n1-standard-2`, garantizando el cumplimiento de la arquitectura base:

<img width="1312" height="823" alt="gcp2" src="https://github.com/user-attachments/assets/05143cd1-01f8-40dd-9ce6-58658a099302" />
<img width="1304" height="348" alt="gcp1" src="https://github.com/user-attachments/assets/c3dfd1a9-e50c-4c88-b71a-e7a36ea8f972" />


### Paso 1: Inicialización de la Capa de Virtualización (KubeVirt)
Asegúrese de contar con los operadores de KubeVirt instalados en el clúster y aplique el entorno de almacenamiento de base de datos:
```bash
kubectl apply -f k8s/valkey-vm.yaml
```

### Paso 2: Despliegue de los Componentes del Backend

Instancie el broker de mensajería y los tres microservicios desacoplados en Go y Rust:

```bash
kubectl apply -f k8s/rabbitmq.yaml
kubectl apply -f k8s/go-microservices.yaml
kubectl apply -f k8s/go-consumer.yaml
kubectl apply -f k8s/rust-api.yaml
```

### Paso 3: Activación del Plano de Control de Escalado y Perímetro

Instale las reglas de ruteo avanzadas de la Gateway API y el autoescalador por hardware de la capa de ingesta:

```bash
kubectl apply -f k8s/rust-hpa.yaml
kubectl apply -f k8s/gateway.yaml
```

### Paso 4: Aprovisionamiento del Sistema de Visualización

Despliegue la máquina virtual de KubeVirt dedicada al plano de monitoreo:

```bash
kubectl apply -f k8s/grafana-vm.yaml
```

## 4. CONFIGURACIONES TÉCNICAS CLAVE

<img width="3004" height="1214" alt="EsquemaC" src="https://github.com/user-attachments/assets/1389b2a5-3b53-4208-a0a2-a804a7a9d110" />

### A. Configuración de Gateway API

Se implementó el estándar moderno de Gateway API utilizando la clase avanzada de Balanceador de Carga Global Externo Administrado de Google (`gke-l7-global-external-managed`). Esto permite el uso de filtros de reescritura de URL (`URLRewrite`) a nivel de capa 7. El manifiesto mapea las solicitudes que coinciden con el prefijo `/grpc-202010367` y las transforma limpiamente a `/` antes de transferirlas al back-end de Rust en el puerto `8080`.

### B. Comunicación REST y gRPC

- **REST:** La comunicación perimetral e interna entre la API de Rust y el microservicio `go-client` se maneja mediante peticiones síncronas HTTP POST con payloads formateados en JSON a través de la librería `reqwest`.
- **gRPC:** La comunicación entre `go-client` y `go-server` se realiza mediante llamadas a procedimientos remotos de alto rendimiento sobre conexiones TCP persistentes en el puerto `50051`. El contrato de servicios está estrictamente definido en un archivo `.proto` compilado utilizando `protoc-gen-go-grpc`, abstrayendo la serialización binaria binaria optimizada.

### C. Uso de RabbitMQ

Se configuró un despliegue del broker con la imagen especializada `rabbitmq:3-management` para contar con visibilidad de colas. El servidor gRPC declara de forma segura y duradera una cola indexada bajo el identificador `quiniela_queue`. Las publicaciones emplean un contexto transaccional con un tipo de contenido explícito `application/json`, asegurando que las ráfagas entrantes permanezcan en memoria regulada o disco duradero en caso de retrasos en el subproceso consumidor.

### D. Configuración del Autoescalado (HPA)

El escalado horizontal automático se vinculó directamente al deployment `api-rust`. Para habilitar la capacidad métrica de cómputo del clúster, se definieron asignaciones de recursos límites y requerimientos fijos (`requests.cpu: 100m` y `limits.cpu: 200m`) dentro del manifiesto del contenedor. El objeto HPA evalúa la métrica provista por el `metrics-server` y está parametrizado para escalar el deployment desde un mínimo de 1 pod hasta un umbral máximo de 3 réplicas concurrentes en el momento exacto en que la utilización promedio de CPU sobrepasa el límite estricto del 30%.

### E. Registro Privado Zot y Gestión de OCI Artifacts

- **Gestión de Imágenes:** El registro privado Zot fue desplegado y configurado a nivel de arquitectura en red sobre una instancia dedicada externa ejecutando servicios HTTP en el puerto `8080`. Para la resolución nativa de contenedores e inyección del plano de microservicios, las imágenes de ejecución del backend fueron migradas e integradas bajo repositorios privados en Google Artifact Registry (`us-central1-docker.pkg.dev`), garantizando el cumplimiento de las políticas restrictivas de seguridad de las firmas TLS exigidas por los nodos de Google Kubernetes Engine (GKE).
- **OCI Artifacts:** El servidor de Zot maneja la persistencia y distribución de archivos independientes (no imágenes) tratando los documentos bajo la especificación abierta de artefactos OCI. Utilizando la utilidad CLI binaria `oras`, se implementó y documentó el flujo completo de empaquetado, carga y descarga del archivo de datos transaccionales `partido.json` mediante canales de comunicación directa:
    - *Comando de Publicación (Push):* `oras push --plain-http 34.136.113.161:8080/quiniela-data:v1 partido.json`
    - *Comando de Descarga (Pull):* `oras pull --plain-http 34.136.113.161:8080/quiniela-data:v1`
 
**Evidencia de Ejecución (Gateway, Autoescalado y OCI Artifacts):**
La siguiente bitácora de terminal demuestra la correcta asignación de la IP pública del Gateway, el objetivo de CPU establecido en el HPA, y las pruebas exitosas de empaquetado (push) y descarga (pull) del JSON como artefacto OCI en Zot:
<img width="1213" height="728" alt="comandos" src="https://github.com/user-attachments/assets/2728a3f6-4d6c-4f2f-afea-6fa26228fe86" />


## 5. DESPLIEGUE DE VIRTUALIZACIÓN (KUBEVIRT)

El núcleo de persistencia y análisis gráfico del sistema saca provecho de la abstracción de hipervisores nativos gobernados por Kubernetes a través del operador KubeVirt.

<img width="1954" height="570" alt="EsquemaB" src="https://github.com/user-attachments/assets/22b5a4b8-4a2d-4306-9898-65691746efe4" />

- **Valkey (Base de Datos):** Definido mediante un objeto de tipo `VirtualMachine` parametrizado en estado de ejecución constante (`runStrategy: Always`). Utiliza un disco virtual aprovisionado desde una imagen inmutable de Debian 12 (`quay.io/containerdisks/debian:12`). La red se configura en modo `masquerade: {}` perforando de forma segura y bidireccional el puerto `6379`. Mediante instrucciones estructuradas en el script del `cloud-init`, la máquina virtual arranca, instala el motor runtime de **containerd** de manera aislada, descarga la imagen oficial de Valkey y ejecuta el proceso del motor de datos utilizando comandos nativos del cliente de containerd (`ctr images pull` y `ctr run -d --net-host`), desactivando explícitamente el modo protegido para permitir tramas remotas enlazadas a la IP de la interfaz interna (`-bind 0.0.0.0 --protected-mode no`).
- **Grafana (Visualización):** Se despliega bajo un manifiesto de máquina virtual homólogo a Valkey, aislando el software de renderizado analítico en su propio sistema operativo Debian 12 virtualizado con 2 vCPUs y 2048M de RAM para evitar cuellos de botella en procesamiento concurrente. La interfaz de red de KubeVirt mapea el puerto `3000` (Web) y el puerto `22` (SSH para depuración y auditoría en tiempo real). Atendiendo a las especificaciones técnicas rigurosas de la cátedra, el aprovisionamiento interno evita por completo el uso de envoltorios aislados adicionales y ejecuta la suite de Grafana de forma nativa (Bare-Metal interno) sobre la máquina virtual Debian. Los comandos programados de automatización en el cloud-init actualizan los repositorios de paquetes de Debian, instalan las dependencias de fuentes del sistema (`adduser libfontconfig1 musl curl`), descargan el paquete binario oficial de distribución estática (`grafana_10.4.2_amd64.deb`), inyectan la instalación del software mediante el gestor del núcleo (`dpkg -i`) y registran el arranque automático gobernado por el gestor del sistema operativo (`systemctl enable/start grafana-server`).

**Evidencia de Ejecución de Máquinas Virtuales:**
Estado actual de los procesos de virtualización anidada corriendo de forma paralela en los nodos del clúster:
<img width="1219" height="776" alt="cali3" src="https://github.com/user-attachments/assets/e51d908b-ce5e-4307-9642-6598201afe0d" />


## 6. PRUEBAS DE CARGA Y ANÁLISIS DE RÉPLICAS

### A. Resultados de la Evaluación de Rendimiento con Locust

La infraestructura distribuida completa fue sometida a una prueba de estrés continua utilizando usuarios concurrentes parametrizados bajo el agente Locust, disparando payloads JSON transaccionales de quinielas hacia el endpoint balanceado de producción.

- **Peticiones Totales Despachadas:** 380 solicitudes HTTP POST exitosas.
- **Porcentaje de Error Registrado:** 0.00% de fallas de comunicación.
- **Tiempo de Respuesta Promedio:** 1044 ms.
- **Tiempo de Respuesta Mínimo Detectado:** 481 ms.
- **Tiempo de Respuesta Máximo Detectado:** 3078 ms.

 **Evidencia de Pruebas de Carga:**
 <img width="1211" height="691" alt="Locust1" src="https://github.com/user-attachments/assets/548b19d9-c411-486c-8f82-d629785d551f" />
 <img width="1211" height="691" alt="Locust2" src="https://github.com/user-attachments/assets/b542cf0d-6c0d-4e56-b938-1e32a802bdb1" />
<img width="1211" height="691" alt="Locust3" src="https://github.com/user-attachments/assets/58ea2311-548b-4fed-be42-8454d9c87919" />

Durante la inyección continua de transacciones, la monitorización en vivo del objeto Horizontal Pod Autoscaler (`kubectl get hpa`) arrojó una escalada progresiva en el uso de CPU de la API REST de Rust de un 0% basal hasta alcanzar picos estables de consumo medido. El HPA mantuvo el estado operacional dentro de los márgenes previstos de contención sin necesidad de forzar réplicas de pánico, validando el correcto dimensionamiento del pod.

### B. Análisis Comparativo de Rendimiento (1 Réplica vs 2 Réplicas)

Para auditar las capacidades de elasticidad y rendimiento de la capa lógica intermedia en Go, se ejecutaron pruebas de estrés comparativas aislando el comportamiento transaccional del clúster bajo dos escenarios específicos de escalado del plano del servidor de gRPC (`go-server`):

| **Métrica de Rendimiento Evaluada** | **Escenario A: 1 Réplica gRPC activa** | **Escenario B: 2 Réplicas gRPC balanceadas** |
| --- | --- | --- |
| **Tasa de Ingesta (Throughput)** | ~1.36 req/s estables bajo carga lineal. | ~2.84 req/s con procesamiento paralelo distribuido. |
| **Latencia de Red gRPC Promedio** | 1715 ms debido al encolamiento síncrono. | 760 ms al balancearse las tramas remotas en hilos distribuidos. |
| **Saturación de Cola (RabbitMQ)** | Presentó picos de retención transitoria en cola. | Flujo de vaciado inmediato; cola en 0 de forma constante. |
| **Resiliencia de Infraestructura** | Punto único de fallo; caídas de red elevan reintentos. | Alta disponibilidad nativa; conmutación inmediata ante fallos de pod. |

*Conclusión del Análisis:* El escalado a 2 réplicas optimiza drásticamente los tiempos de procesamiento global del sistema. Al duplicar las instancias del servidor gRPC que actúan como escritores de RabbitMQ, se elimina la congestión en las llamadas remotas internas del clúster, reduciendo los tiempos de procesamiento en más del 55% e incrementando el rendimiento general de ingesta masiva de datos hacia Valkey de forma lineal.

## 7. EVIDENCIA DE ENRUTAMIENTO Y LOGS DE MICROSERVICIOS

A continuación, se presenta la bitácora de ejecución en tiempo real de los comandos de calificación. Se demuestra la ingesta síncrona, traducción de protocolos y encolamiento de las transacciones a través de la API en Rust, los clientes Go y la retención en RabbitMQ:

<img width="1219" height="776" alt="cali1" src="https://github.com/user-attachments/assets/20d2c2b1-da3d-4b40-af79-1a12293d1417" />
<img width="1219" height="776" alt="cali2" src="https://github.com/user-attachments/assets/70e35523-fdfb-4bf9-8ede-80e66f7baa50" />

---

## 8. DASHBOARDS DE MONITOREO (ENFOQUE ARGENTINA)

Para cumplir con el requerimiento estricto de análisis de datos de la variante correspondiente (último dígito del carnet: 7), se extrajeron las métricas directamente desde Valkey. Mediante el uso de transformaciones JSON y agrupaciones en Grafana, se filtró la visualización temporal y las sumatorias exclusivamente para la etiqueta `ARG`.

**Métricas Generales (Máximos, Mínimos y Moda):**
<img width="1211" height="691" alt="Dashboard1" src="https://github.com/user-attachments/assets/be7f94f7-4dc2-4303-8443-1c79f69739bd" />
<img width="1211" height="691" alt="Dashboard2" src="https://github.com/user-attachments/assets/a47fcc1a-668c-4c49-a2f6-50eb6f2bf148" />

**Top de Usuarios y Predicciones Globales:**
<img width="1211" height="691" alt="Dashboard3" src="https://github.com/user-attachments/assets/f381d6a4-7811-498b-9439-27a5f8388aac" />

**Análisis Específico y Evolución Temporal (ARG):**
<img width="1211" height="691" alt="Dashoard5" src="https://github.com/user-attachments/assets/90f16be0-97da-4314-96cf-f85abe45bcd8" />

---

## 9. CONCLUSIONES

- **Ventajas de la Orquestación con Kubernetes:** Kubernetes y la plataforma de GKE proveen un entorno industrial inigualable para la gestión de microservicios distribuidos. La capacidad de declarar configuraciones mediante archivos estáticos declarativos, el autoescalado horizontal automatizado guiado por hardware (HPA) y la resiliencia nativa que sustituye de manera transparente los pods en estado de error garantizan la continuidad de las operaciones del negocio bajo patrones extremos de tráfico masivo.
- **Desventajas y Complejidades Técnicas:** El mayor reto detectado radica en el sobrecosto de abstracción que impone la virtualización anidada mediante operadores como KubeVirt. Correr un hipervisor que administra sistemas operativos completos sobre capas virtualizadas previas de contenedores introduce complejidades severas en el direccionamiento de redes lógicas (aislamientos por NAT/Masquerade) y restricciones drásticas de seguridad en los nodos nativos, obligando a implementar instalaciones quirúrgicas de software nativo interno para evadir fallos silenciosos de sockets de red en subprocesos anidados.
- **Evaluación Global de las Herramientas:** Las arquitecturas modernas basadas en el desacoplamiento de capas mediante colas asíncronas (RabbitMQ) combinadas con bases de datos en memoria optimizadas (Valkey) representan el estándar de oro para el diseño de software de alta concurrencia. La visibilidad de datos que provee la integración de transformaciones JSON nativas sobre canales CLI en Grafana cierra el ciclo de vida del dato de manera transparente, permitiendo transformar millones de registros crudos en tableros analíticos estratégicos con una latencia mínima.
