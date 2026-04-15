# 🚀 Go Deployer

¡Bienvenido al **Deployer Automático**! Esta herramienta está diseñada para simplificar el despliegue de aplicaciones escritas en Go. Gestiona todo el ciclo de vida de tu aplicación, desde la transferencia hasta la ejecución como servicio del sistema, directamente desde un panel de control intuitivo.

## ✨ Características

* **Prueba de Conexión SSH:** Verifica la conectividad con tu servidor antes de iniciar cualquier operación.
* **Despliegue con un Clic:** Sube un archivo `.zip`, el deployer lo transfiere, lo descomprime y lo compila en la VM remota automáticamente.
* **Gestión de Servicios Systemd:** Genera y activa servicios de `systemd` de forma automática para que tu aplicación se ejecute como un demonio del sistema.
* **Control Total:** Inicia, detén, reinicia y consulta el estado de tus servicios directamente desde la interfaz.
* **Logs en Tiempo Real:** Visualiza lo que sucede en tu servidor mediante WebSockets (transmisión en vivo de `tail -f`).

## 🛠️ Stack Tecnológico

* **Backend:** [Go](https://go.dev/) con el framework [Gin](https://gin-gonic.com/).
* **SSH & WebSockets:** `golang.org/x/crypto/ssh` para la ejecución remota y `gorilla/websocket` para la transmisión de logs.
* **Frontend:** HTML5, JavaScript y CSS (Tailwind CSS recomendado).

## 🚀 Cómo empezar

### Requisitos Previos

1.  **En tu máquina local:** Tener Go instalado para ejecutar el servidor del deployer.
2.  **En la VM remota (Destino):**
    * Tener `unzip` instalado: `sudo apt install unzip -y`
    * Tener `go` instalado (en caso de requerir compilación remota).
    * **IMPORTANTE (Permisos):** El usuario SSH necesita permisos de `sudo` sin contraseña para gestionar servicios. Puedes configurarlo ejecutando:
        ```bash
        echo "$(whoami) ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/$(whoami)
        ```

### Instalación y Ejecución

1.  **Clonar el repositorio:**
    ```bash
    git clone https://github.com/Shuh3n/Pulse-Deploy--Remote-SSH-Orchestrator.git
    cd parcial2-deployer
    ```

2.  **Descargar dependencias:**
    ```bash
    go mod tidy
    ```

3.  **Ejecutar la aplicación:**
    ```bash
    go run main.go
    ```

4.  **Acceder a la interfaz:**
    Abre tu navegador en: `http://localhost:8080`

## 📁 Estructura del Proyecto

```text
├── main.go          # Servidor principal, rutas y lógica SSH
├── static/          # Archivos del frontend (HTML, CSS, JS)
├── uploads/         # Almacenamiento temporal de archivos ZIP
├── worker/          # Aplicación de ejemplo para gestionar
└── go.mod           # Dependencias del proyecto
