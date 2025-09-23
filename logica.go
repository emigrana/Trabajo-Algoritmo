package main

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	constCantFilasTablero    = 25
	constCantColumnasTablero = 29

	constCantColumnas = 2
	constY            = 0
	constX            = 1

	constCantColumnasOvni = 4
	constTipoOvni         = 0
	constOvniY            = 1
	constOvniX            = 2
	constEnDescenso       = 3

	constTiempoDeDisparoOvni   = 3
	constTiempoLiberarcionOvni = 10

	constSimboloVacío       = ""
	constSimboloNave        = "N"
	constSimboloDisparoNave = "*"
	constSimboloDisparoOvni = "."
	constSimboloOvniLider   = "L"
	constSimboloOvniComun   = "C"
	constSimboloBorde       = "X"

	constCantColumnasDisparos = 2
)

// Vector global con las direcciones posibles
var (
	quieto    = [constCantColumnas]int{0, 0}
	izquierda = [constCantColumnas]int{0, -1}
	derecha   = [constCantColumnas]int{0, 1}
	arriba    = [constCantColumnas]int{-1, 0}
	abajo     = [constCantColumnas]int{1, 0}
)

// Vector global con las direccion de la nave
var direccionNave [constCantColumnas]int

// Variable global que indica si se presiono la barra espaciadora lo que ejecuta un disparo de la nave
var disparoNave bool

// Función para enviar actualizaciones a los clientes
func generarEventos() {
	var (
		tablero [constCantFilasTablero][constCantColumnasTablero]string

		nave         [constCantColumnas]int
		disparosNave [][constCantColumnas]int

		ovnis         [][constCantColumnasOvni]int
		disparosOvnis [][constCantColumnas]int

		ultimaEjecucionDisparoOvni    time.Time
		ultimaEjecucionLiberacionOvni time.Time

		puntos int
		vidas  int
	)

	rand.Seed(time.Now().Unix())

	//Se inicializa variables
	ultimaEjecucionDisparoOvni = time.Now()
	ultimaEjecucionLiberacionOvni = time.Now()

	disparoNave = false

	vidas = 3

	// Se genera tablero por primera vez con los bordes
	tablero = generarTablero()

	// Se genera la nave (posición inicial) por primera vez
	nave, direccionNave = inicializarNave(constCantFilasTablero, constCantColumnasTablero)

	// Se generan los ovnis (posiciones iniciales) por primera vez
	ovnis = inicializarOvnis(constCantFilasTablero, constCantColumnasTablero)

	// Se actualiza nave y ovnis en el tablero por primera vez
	actualizarTablero(&tablero, nave, disparosNave, ovnis, disparosOvnis)

	for {
		// Se actualizan las posiciones de la nave según la dirección
		calcularNuevaPosicionNave(tablero, &nave, &direccionNave)

		// Se crea un nuevo disparo si corresponde
		crearDisparoNave(nave, &disparoNave, &disparosNave)

		//Cada "constTiempoDeDisparoOvni" segundos, se crea un disparo de un ovni
		if time.Since(ultimaEjecucionDisparoOvni) >= constTiempoDeDisparoOvni*time.Second {
			crearDisparoOvni(ovnis, &disparosOvnis)
			ultimaEjecucionDisparoOvni = time.Now()
		}

		//Cada "constTiempoLiberarcionOvni" segundos, se libera un obvni de la formación
		if time.Since(ultimaEjecucionLiberacionOvni) >= constTiempoLiberarcionOvni*time.Second {
			liberarOvni(ovnis)
			ultimaEjecucionLiberacionOvni = time.Now()
		}

		// Se calcula la nueva posición de los ovnis liberados
		calcularNuevaPosicionOvnisLiberados(ovnis)

		// Se calcula las nuevas posiciones de los disparos de la nave y de los ovnis
		calcularNuevasPosicionesDisparos(tablero, disparosNave, disparosOvnis)

		// Se verifica el estado del juego y eliminan elementos si corresponde
		if !verificarEstadoDeJuego(tablero, nave, &ovnis, &disparosNave, &disparosOvnis, &puntos) {
			// Si no tiene más vidas, se devuelve pantalla gameOver
			vidas--

			if vidas <= 0 {
				enviarGameOver(puntos)
				return
			}
		} else {
			if len(ovnis) == 0 {
				enviarWin(puntos)

				return
			}

			enviarActualizacionTexto(fmt.Sprint("Puntaje: ", puntos, ". Vidas: ", vidas))
		}

		//Se actualiza el tablero con los valores de la nave, ovnis y disparos en sus nuevas posiciones
		actualizarTablero(&tablero, nave, disparosNave, ovnis, disparosOvnis)

		// Se envía actualización de tablero al cliente para mostrar en pantalla
		enviarActualizacionTablero(tablero)

		// Espera un tiempo antes de generar un nuevo movimiento
		time.Sleep(150 * time.Millisecond)
	}
}

func generarTablero() [constCantFilasTablero][constCantColumnasTablero]string {
	var tablero [constCantFilasTablero][constCantColumnasTablero]string

	for f := 0; f < constCantFilasTablero; f++ {
		if f == 0 {
			for c := 0; c < constCantColumnasTablero; c++ {
				tablero[f][c] = "X"
			}
		} else if f != (constCantFilasTablero - 1) {
			tablero[f][0] = "X"
			tablero[f][(constCantColumnasTablero - 1)] = "X"
		} else {
			for c := 0; c < constCantColumnasTablero; c++ {
				tablero[f][c] = "X"
			}
		}
	}

	return tablero
}

func inicializarNave(cantFilasTablero int, cantColumnasTablero int) ([constCantColumnas]int, [constCantColumnas]int) {

	var (
		nave [2]int
	)

	nave[0] = cantFilasTablero - 3
	nave[1] = cantColumnasTablero / 2

	return nave, quieto
	//return [constCantColumnas]int{}, quieto  **CONSULTAR**
}

func inicializarOvnis(cantFilasTablero int, cantColumnasTablero int) [][constCantColumnasOvni]int {

	var (
		ovnis             [][constCantColumnasOvni]int
		ovnisvector       [4]int
		varPatronFilas    int = 7
		varPatronColumnas int = 2
		cantidadLideres   int = 0
		TipoOvniAleatorio int
		maxCantLideres    = 18
	)
	rand.Seed(time.Now().UnixNano())

	for f := 2; f < varPatronFilas; f++ {
		for c := varPatronColumnas; c < cantColumnasTablero-varPatronColumnas; c++ {

			if cantidadLideres < maxCantLideres {
				TipoOvniAleatorio = 1 + rand.Intn(2) //tipo de ovni 1 es lider, tipo 2 es comun
				if TipoOvniAleatorio == 1 {
					cantidadLideres += 1
				}
			} else {
				TipoOvniAleatorio = 2
			}

			/*if TipoOvniAleatorio == 1 {
				TipoOvniAleatorioL = "L"
			} else {
				TipoOvniAleatorioL = "C"
			} */

			ovnisvector[0] = TipoOvniAleatorio
			ovnisvector[1] = f
			ovnisvector[2] = c
			ovnisvector[3] = 0
			ovnis = append(ovnis, ovnisvector)

		}
		cantidadLideres = 0
		varPatronColumnas += 1
		maxCantLideres -= 4
	}

	return ovnis
}

func actualizarTablero(tablero *[constCantFilasTablero][constCantColumnasTablero]string,
	nave [constCantColumnas]int,
	disparosNave [][constCantColumnas]int,
	ovnis [][constCantColumnasOvni]int,
	disparosOvnis [][constCantColumnas]int) {

	var (
		posicionY, posicionX int
	)

	//Recorrer el vector ovni, y representarlos en tablero dependiendo su tipo
	for f := 0; f < len(ovnis); f++ {
		if ovnis[f][0] == 1 {
			posicionY = ovnis[f][1]
			posicionX = ovnis[f][2]

			tablero[posicionY][posicionX] = "L"
		} else {
			posicionY = ovnis[f][1]
			posicionX = ovnis[f][2]

			tablero[posicionY][posicionX] = "C"
		}
	}

	//representar nave
	posicionY = nave[0]
	posicionX = nave[1]
	tablero[posicionY][posicionX] = "N"

}

func calcularNuevaPosicionNave(tablero [constCantFilasTablero][constCantColumnasTablero]string,
	nave *[constCantColumnas]int, direccionNave *[constCantColumnas]int) {

	//actualizar posición nave

	nave[0] += direccionNave[0]
	nave[1] += direccionNave[1]

	//verificar que no sea un borde y volver si es necesario
	if nave[0] == 1 || nave[0] == 28 {
		nave[0] -= direccionNave[0]
	} else if nave[1] == 1 || nave[1] == 28 {
		nave[1] -= direccionNave[1]
	}

	direccionNave[0] = 0
	direccionNave[1] = 0

}

func crearDisparoNave(nave [constCantColumnas]int,
	disparoNave *bool,
	disparosNave *[][constCantColumnasDisparos]int) {

	//PROGRAMAR
}

func crearDisparoOvni(ovnis [][constCantColumnasOvni]int,
	disparosOvnis *[][constCantColumnasDisparos]int) {

	//PROGRAMAR
}

func calcularNuevasPosicionesDisparos(tablero [constCantFilasTablero][constCantColumnasTablero]string,
	disparosNave [][constCantColumnasDisparos]int,
	disparosOvnis [][constCantColumnasDisparos]int) {

	//PROGRAMAR
}

func verificarEstadoDeJuego(tablero [constCantFilasTablero][constCantColumnasTablero]string,
	nave [constCantColumnas]int,
	ovnis *[][constCantColumnasOvni]int,
	disparosNave *[][constCantColumnasDisparos]int,
	disparosOvnis *[][constCantColumnasDisparos]int,
	puntos *int) bool {

	//PROGRAMAR

	return true
}

func eliminarDisparo(slice [][constCantColumnasDisparos]int, coordenadaY int, coordenadaX int) [][2]int {
	var nuevoSlice [][constCantColumnasDisparos]int
	for f := 0; f < len(slice); f++ {
		if slice[f][constY] != coordenadaY &&
			slice[f][constX] != coordenadaX {
			nuevoSlice = append(nuevoSlice, slice[f])
		}
	}
	return nuevoSlice
}

func eliminarOvni(slice [][constCantColumnasOvni]int, coordenadaY int, coordenadaX int) [][4]int {
	var nuevoSlice [][constCantColumnasOvni]int
	for f := 0; f < len(slice); f++ {
		if slice[f][constOvniY] != coordenadaY ||
			slice[f][constOvniX] != coordenadaX {
			nuevoSlice = append(nuevoSlice, slice[f])
		}
	}
	return nuevoSlice
}

func liberarOvni(ovnis [][constCantColumnasOvni]int) {
	//PROGRAMAR
}

func calcularNuevaPosicionOvnisLiberados(ovnis [][constCantColumnasOvni]int) {
	//PROGRAMAR
}

//aaa
