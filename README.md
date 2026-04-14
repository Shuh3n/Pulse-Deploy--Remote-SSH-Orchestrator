# 🚀 Parcial 2 - Go Deployer 

¡Bienvenido al **Deployer Automático**! Esta herramienta está diseñada para que desplegar tus aplicaciones escritas en Go sea un trámite. Olvidate de andar copiando archivos a mano y peleándote con la terminal de la VM; hacelo todo desde un dashboard fachero.

## ✨ Características

- **Prueba de Conexión SSH:** Verificá que llegás a tu servidor antes de intentar nada.
- **Deploy con un Clic:** Subí un archivo `.zip`, el deployer lo transfiere, lo descomprime y lo compila en la VM remota.
- **Gestión de Servicios Systemd:** Generá y activá servicios de `systemd` automáticamente para que tu app corra como un demonio.
- **Control Total:** Start, Stop, Restart y Status de tus servicios desde la interfaz.
- **Logs en Tiempo Real:** Mirá lo que está pasando en tu servidor mediante WebSockets (streaming de `tail -f`).

## 🛠️ Stack Tecnológico

- **Backend:** [Go](https://go.dev/) con el framework [Gin](https://gin-gonic.com/).
- **SSH & WebSockets:** `golang.org/x/crypto/ssh` para la magia remota y `gorilla/websocket` para los logs.

## 🚀 Cómo empezar

### Requisitos Previos

1. **En tu máquina local:** Tener Go instalado para correr el servidor del deployer.
2. **En la VM remota (Destino):**
   - Tener `unzip` instalado (`sudo apt install unzip`).
   - Tener `go` instalado (si querés que compile allá).
   - **IMPORTANTE (Permisos):** Para que el deployer pueda gestionar servicios, el usuario SSH necesita permisos de `sudo` sin contraseña. Podés activarlo corriendo esto en la VM:
     ```bash
     echo "$(whoami) ALL=(ALL) NOPASSWD:ALL" | sudo tee /etc/sudoers.d/$(whoami)
     ```

### Instalación y Ejecución

1. Cloná el repositorio:
   ```bash
   git clone https://github.com/santi/parcial2-deployer.git
   cd parcial2-deployer
   ```

2. Descargá las dependencias:
   ```bash
   go mod tidy
   ```

3. Corré la aplicación:
   ```bash
   go run main.go
   ```

4. Abrí tu navegador en: `http://localhost:8080`

## 📁 Estructura del Proyecto

- `main.go`: El corazón del servidor, las rutas de la API y la lógica de SSH.
- `static/`: Contiene el frontend (HTML/JS) de la aplicación.
- `worker.go`: (Si existe lógica separada para procesos pesados).
- `go.mod`: Definición del módulo y dependencias.

## 📝 Notas de Uso

- El deployer intenta inicializar un `go mod` si no encuentra uno en la carpeta de destino.
- Los servicios creados se guardan en `/etc/systemd/system/` con la configuración de reinicio automático (`Restart=always`).

---