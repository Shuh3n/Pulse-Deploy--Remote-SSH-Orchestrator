package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

const counterFile = "counter.txt"
const logFile = "worker.log"

func main() {
	// 1. Leer contador
	count := 0
	data, err := ioutil.ReadFile(counterFile)
	if err == nil {
		content := strings.TrimSpace(string(data))
		if n, err := strconv.Atoi(content); err == nil {
			count = n
		}
	}

	// 2. Incrementar
	count++

	// 3. Guardar contador
	err = ioutil.WriteFile(counterFile, []byte(strconv.Itoa(count)), 0644)
	if err != nil {
		log.Fatalf("Error escribiendo contador: %v", err)
	}

	// 4. Escribir al log y consola
	entry := fmt.Sprintf("[%d] - %s\n", count, time.Now().Format(time.RFC3339))
	
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Error abriendo log: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(entry); err != nil {
		log.Fatalf("Error escribiendo entrada de log: %v", err)
	}

	fmt.Print(entry)
	
	// Mantener el proceso vivo para el streaming de logs (opcional para simular un servicio)
	for {
		time.Sleep(5 * time.Second)
		fmt.Printf("Worker running... Last count: %d\n", count)
	}
}
