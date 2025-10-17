package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"text/template"
)

// Agrega un canal para enviar actualizaciones a los clientes
var updates = make(chan string)

func main() {
	// Ruta para servir la página principal
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	// Ruta para manejar la solicitud POST del evento de teclado
	http.HandleFunc("/keypress", keyPressHandler)

	// Agrega una nueva ruta para SSE
	http.HandleFunc("/updates", updatesHandler)

	// Agrega una nueva ruta para "game over"
	http.HandleFunc("/gameover", gameoverHandler)

	//Agrega una nueva ruta para "you win"
	http.HandleFunc("/win", winHandler)

	// NEW: handler para reiniciar el juego desde la página "YOU WIN"
	http.HandleFunc("/restart", restartHandler)

	// Inicia una goroutine para enviar actualizaciones a los clientes
	go generarEventos()

	fmt.Println("Abrir navegador e ingresar a http://localhost:8080/")

	// Inicia el servidor en el puerto 8080
	http.ListenAndServe(":8080", nil)
}

// Handler para "game over"
func gameoverHandler(w http.ResponseWriter, r *http.Request) {
	// Estructura para pasar datos a la plantilla
	type PageData struct {
		Points string
	}

	// Datos dinámicos que quieres pasar a la plantilla
	data := PageData{
		Points: r.URL.Query().Get("points"),
	}

	// Parsear la plantilla HTML
	tmpl, err := template.ParseFiles("gameover.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ejecutar la plantilla con los datos
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handler para "game over"
func winHandler(w http.ResponseWriter, r *http.Request) {
	// Estructura para pasar datos a la plantilla
	type PageData struct {
		Points string
	}

	// Datos dinámicos que quieres pasar a la plantilla
	data := PageData{
		Points: r.URL.Query().Get("points"),
	}

	// Parsear la plantilla HTML
	tmpl, err := template.ParseFiles("win.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Ejecutar la plantilla con los datos
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Handler para la conexión SSE
func updatesHandler(w http.ResponseWriter, r *http.Request) {
	// Establece las cabeceras para indicar que esta es una conexión SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// Mantén la conexión abierta
	for {
		// Espera a recibir una actualización desde el canal de actualizaciones
		update := <-updates

		// Escribe la actualización al cliente
		fmt.Fprintf(w, "data: %s\n\n", update)

		// Flushea el buffer para enviar la actualización inmediatamente
		w.(http.Flusher).Flush()
	}
}

// Handler para manejar la solicitud POST cuando el usuario presiona una tecla
func keyPressHandler(w http.ResponseWriter, r *http.Request) {
	// Leer los datos JSON enviados en la solicitud
	var data map[string]string
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error al leer los datos JSON", http.StatusBadRequest)
		return
	}

	// Obtener la tecla presionada del mapa de datos
	keyPressed, ok := data["key"]
	if !ok {
		http.Error(w, "Tecla no proporcionada", http.StatusBadRequest)
		return
	}

	// Actualizar la dirección según la tecla presionada
	switch keyPressed {
	case "ArrowRight":
		direccionNave = derecha
	case "ArrowLeft":
		direccionNave = izquierda
	case "ArrowUp":
		direccionNave = arriba
	case "ArrowDown":
		direccionNave = abajo
	case " ":
		disparoNave = true
	}

	// Responder con un mensaje de éxito
	w.WriteHeader(http.StatusOK)
}

// Se envía actualización de tablero al cliente
func enviarActualizacionTablero(tablero [constCantFilasTablero][constCantColumnasTablero]string) {
	// Convierte la matriz en una cadena JSON
	update, err := json.Marshal(tablero)
	if err != nil {
		fmt.Println("Error al convertir la matriz en JSON:", err)
	}

	// Envía la actualización a todos los clientes conectados
	updates <- string(update)
}

// Se envía actualización de texto al cliente
func enviarActualizacionTexto(text string) {
	// Envía la actualización a todos los clientes conectados
	updates <- "{\"is_text\": true, \"text\": \"" + text + "\"}"
}

func enviarGameOver(points int) {
	// Envía la actualización de game over a todos los clientes conectados
	texto := fmt.Sprint("{\"game_over\": true, \"points\": \"", points, "\"}")
	updates <- texto

	fmt.Println("Game Over. Points:", points)
}

func enviarWin(points int) {
	// Envía la actualización de game over a todos los clientes conectados
	texto := fmt.Sprint("{\"win\": true, \"points\": \"", points, "\"}")
	updates <- texto

	fmt.Println("Win. Points:", points)
}

// NEW: restartHandler
func restartHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Lanzar una nueva instancia de la lógica del juego
	go generarEventos()
	w.WriteHeader(http.StatusOK)
}
