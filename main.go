package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type SSHConfig struct {
	IP       string `json:"ip"`
	Port     string `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ServiceConfig struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func main() {
	r := gin.Default()

	// Servir archivos estáticos
	r.StaticFile("/", "./static/index.html")

	// 1. Prueba de conexión SSH
	r.POST("/api/ssh-test", func(c *gin.Context) {
		var config SSHConfig
		if err := c.ShouldBindJSON(&config); err != nil {
			c.JSON(400, gin.H{"error": "Configuración inválida"})
			return
		}

		client, err := connectSSH(config)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error de conexión: " + err.Error()})
			return
		}
		defer client.Close()

		c.JSON(200, gin.H{"status": "¡Conectado exitosamente!"})
	})

	// 2. Subir y Desplegar ZIP
	r.POST("/api/deploy", func(c *gin.Context) {
		ip := c.PostForm("ip")
		port := c.PostForm("port")
		user := c.PostForm("user")
		password := c.PostForm("password")
		remotePath := c.PostForm("path")

		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "Fallo al subir el archivo"})
			return
		}

		tempPath := filepath.Join(os.TempDir(), file.Filename)
		if err := c.SaveUploadedFile(file, tempPath); err != nil {
			c.JSON(500, gin.H{"error": "Error al guardar archivo temporal"})
			return
		}
		defer os.Remove(tempPath)

		config := SSHConfig{IP: ip, Port: port, User: user, Password: password}
		client, err := connectSSH(config)
		if err != nil {
			c.JSON(500, gin.H{"error": "Conexión SSH fallida: " + err.Error()})
			return
		}
		defer client.Close()

		if err := transferFile(client, tempPath, remotePath); err != nil {
			c.JSON(500, gin.H{"error": "Fallo en transferencia: " + err.Error()})
			return
		}

		// Descomprimir remoto
		dir := path.Dir(remotePath)
		unzipCmd := fmt.Sprintf("unzip -o %s -d %s", remotePath, dir)
		output, err := executeCommand(client, unzipCmd)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error al descomprimir: " + err.Error() + "\nSalida: " + output})
			return
		}

		// INTENTO DE COMPILACIÓN AUTOMÁTICA
		executeCommand(client, fmt.Sprintf("cd %s && [ ! -f go.mod ] && go mod init deployapp", dir))

		binaryName := "app_bin"
		buildCmd := fmt.Sprintf("cd %s && go build -o %s .", dir, binaryName)
		buildOutput, err := executeCommand(client, buildCmd)

		fullBinaryPath := path.Join(dir, binaryName)
		executeCommand(client, fmt.Sprintf("chmod +x %s", fullBinaryPath))

		lsOutput, _ := executeCommand(client, fmt.Sprintf("ls -F %s", dir))

		statusMsg := "¡Desplegado correctamente!"
		if err != nil {
			statusMsg = "Desplegado, pero la compilación falló. ¿Tenés Go instalado en la VM?"
		} else {
			statusMsg = "¡Desplegado, compilado y listo para arrancar!"
		}

		c.JSON(200, gin.H{
			"status":       statusMsg,
			"binary_path":  fullBinaryPath,
			"build_log":    buildOutput,
			"files_in_dir": lsOutput,
		})
	})

	// 3. Generar y Activar Servicio Systemd
	r.POST("/api/service", func(c *gin.Context) {
		var req struct {
			SSH     SSHConfig     `json:"ssh"`
			Service ServiceConfig `json:"service"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Solicitud inválida"})
			return
		}

		serviceTmpl := `[Unit]
Description=Servicio {{.Name}}
After=network.target

[Service]
ExecStart={{.Path}}
WorkingDirectory={{dir .Path}}
Restart=always
RestartSec=2s
StandardOutput=append:{{dir .Path}}/worker.log
StandardError=append:{{dir .Path}}/worker.log

[Install]
WantedBy=multi-user.target
`
		funcMap := template.FuncMap{"dir": path.Dir}
		tmpl, err := template.New("service").Funcs(funcMap).Parse(serviceTmpl)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error en el template"})
			return
		}

		var b bytes.Buffer
		if err := tmpl.Execute(&b, req.Service); err != nil {
			c.JSON(500, gin.H{"error": "Error ejecutando template"})
			return
		}

		client, err := connectSSH(req.SSH)
		if err != nil {
			c.JSON(500, gin.H{"error": "Conexión SSH fallida"})
			return
		}
		defer client.Close()

		remoteServicePath := fmt.Sprintf("/etc/systemd/system/%s.service", req.Service.Name)
		tempServicePath := fmt.Sprintf("/tmp/%s.service", req.Service.Name)

		localTempFile := filepath.Join(os.TempDir(), req.Service.Name+".service")
		if err := os.WriteFile(localTempFile, b.Bytes(), 0644); err != nil {
			c.JSON(500, gin.H{"error": "Error al crear archivo temporal local: " + err.Error()})
			return
		}
		defer os.Remove(localTempFile)

		if err := transferFile(client, localTempFile, tempServicePath); err != nil {
			c.JSON(500, gin.H{"error": "No se pudo subir el archivo al directorio /tmp de la VM: " + err.Error()})
			return
		}

		moveCmd := fmt.Sprintf("sudo mv %s %s && sudo chown root:root %s && sudo chmod 644 %s",
			tempServicePath, remoteServicePath, remoteServicePath, remoteServicePath)

		output, err := executeCommand(client, moveCmd)
		if err != nil {
			c.JSON(500, gin.H{"error": "Error de permisos con SUDO. \nIMPORTANTE: Ejecutá en la terminal de tu VM:\n\necho \"$(whoami) ALL=(ALL) NOPASSWD:ALL\" | sudo tee /etc/sudoers.d/$(whoami)\n\nError detalle: " + err.Error() + "\nSalida: " + output})
			return
		}

		executeCommand(client, "sudo systemctl daemon-reload")
		executeCommand(client, fmt.Sprintf("sudo systemctl enable %s", req.Service.Name))

		c.JSON(200, gin.H{"status": "¡Servicio creado y activado exitosamente!"})
	})

	// 4. Señales de Control del Dashboard
	r.POST("/api/control", func(c *gin.Context) {
		var req struct {
			SSH    SSHConfig `json:"ssh"`
			Name   string    `json:"name"`
			Action string    `json:"action"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Solicitud inválida"})
			return
		}

		client, err := connectSSH(req.SSH)
		if err != nil {
			c.JSON(500, gin.H{"error": "Conexión SSH fallida"})
			return
		}
		defer client.Close()

		cmd := fmt.Sprintf("sudo systemctl %s %s", req.Action, req.Name)
		output, err := executeCommand(client, cmd)
		if err != nil && req.Action != "status" {
			c.JSON(500, gin.H{"error": "Comando fallido: " + err.Error() + "\nSalida: " + output})
			return
		}

		c.JSON(200, gin.H{"status": "¡Acción realizada!", "output": output})
	})

	// 5. Streamer de Logs (WebSocket)
	r.GET("/ws/logs", func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}
		defer ws.Close()

		var config struct {
			IP       string `json:"ip"`
			Port     string `json:"port"`
			User     string `json:"user"`
			Password string `json:"password"`
			LogPath  string `json:"log_path"`
		}
		if err := ws.ReadJSON(&config); err != nil {
			return
		}

		sshConf := SSHConfig{IP: config.IP, Port: config.Port, User: config.User, Password: config.Password}
		client, err := connectSSH(sshConf)
		if err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("Conexión SSH fallida: "+err.Error()))
			return
		}
		defer client.Close()

		session, err := client.NewSession()
		if err != nil {
			return
		}
		defer session.Close()

		stdout, _ := session.StdoutPipe()
		cmd := fmt.Sprintf("tail -f %s", config.LogPath)
		if err := session.Start(cmd); err != nil {
			ws.WriteMessage(websocket.TextMessage, []byte("Fallo al ejecutar tail"))
			return
		}

		buf := make([]byte, 1024)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				break
			}
			if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				break
			}
		}
	})

	fmt.Println("Servidor corriendo en http://localhost:8080")
	r.Run(":8080")
}

func connectSSH(config SSHConfig) (*ssh.Client, error) {
	if config.Port == "" {
		config.Port = "22"
	}
	sshConfig := &ssh.ClientConfig{
		User: config.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(config.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	return ssh.Dial("tcp", config.IP+":"+config.Port, sshConfig)
}

func executeCommand(client *ssh.Client, cmd string) (string, error) {
	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	session.Stderr = &b
	if err := session.Run(cmd); err != nil {
		return b.String(), err
	}
	return b.String(), nil
}

func transferFile(client *ssh.Client, localPath, remotePath string) error {
	f, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer f.Close()

	session, err := client.NewSession()
	if err != nil {
		return err
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stderr = &b

	w, err := session.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer w.Close()
		io.Copy(w, f)
	}()

	if err := session.Run(fmt.Sprintf("cat > %s", remotePath)); err != nil {
		if b.Len() > 0 {
			return fmt.Errorf("%w: %s", err, b.String())
		}
		return err
	}
	return nil
}
