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

	constTiempoDeDisparoOvni   = 5
	constTiempoLiberarcionOvni = 3 //devolver a 3 y 10

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
	quieto               = [constCantColumnas]int{0, 0}
	izquierda            = [constCantColumnas]int{0, -1}
	derecha              = [constCantColumnas]int{0, 1}
	arriba               = [constCantColumnas]int{-1, 0}
	abajo                = [constCantColumnas]int{1, 0}
	Vidas            int = 3
	Puntos           int = 0
	velocidadJuego   int = 300
	VelocidadJuego   int = 300
	cantVidasCreadas int = 0
)

// Vector global con las direccion de la nave
var direccionNave [constCantColumnas]int

// Variable global que indica si se presiono la barra espaciadora lo que ejecuta un disparo de la nave
var disparoNave bool

// Inicializo Puntos y Vida para acumularlos entre niveles
func LeerEstado() (vidas, puntos, velocidadJuego int) {
	return Vidas, Puntos, VelocidadJuego
}

func GuardarEstado(vidas, puntos, velocidadJuego int) {
	Vidas = vidas
	Puntos = puntos
	VelocidadJuego = velocidadJuego
}

// Función para enviar actualizaciones a los clientes
func generarEventos() {
	var (
		tablero [constCantFilasTablero][constCantColumnasTablero]string

		nave         [constCantColumnas]int
		disparosNave [][constCantColumnas]int

		ovnis         [][constCantColumnasOvni]int
		disparosOvnis [][constCantColumnas]int

		vidasExtra [constCantColumnas]int

		ultimaEjecucionDisparoOvni    time.Time
		ultimaEjecucionLiberacionOvni time.Time

		vidas, puntos, velocidadJuego = LeerEstado()
	)

	rand.Seed(time.Now().Unix())

	vidas, puntos, velocidadJuego = LeerEstado()

	//Se inicializa variables
	ultimaEjecucionDisparoOvni = time.Now()
	ultimaEjecucionLiberacionOvni = time.Now()

	disparoNave = false

	// Se genera tablero por primera vez con los bordes
	tablero = generarTablero()

	// Se genera la nave (posición inicial) por primera vez
	nave, direccionNave = inicializarNave(constCantFilasTablero, constCantColumnasTablero)

	// Se generan los ovnis (posiciones iniciales) por primera vez
	ovnis = inicializarOvnis(constCantFilasTablero, constCantColumnasTablero)

	// Se actualiza nave y ovnis en el tablero por primera vez
	actualizarTablero(&tablero, nave, disparosNave, ovnis, disparosOvnis, &vidasExtra)

	for {
		select {
		case <-stopCurrent:
			return // salir si recibimos señal de stop
		default:
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

			// Se calcula las nuevas posicio|nes de los disparos de la nave y de los ovnis
			calcularNuevasPosicionesDisparos(tablero, disparosNave, disparosOvnis)
			crearVidas(puntos, vidas, ovnis, &vidasExtra, &cantVidasCreadas)
			calcularNuevaPosicionVidas(&vidasExtra)

			// Se verifica el estado del juego y eliminan elementos si corresponde
			if !verificarEstadoDeJuego(tablero, nave, &ovnis, &disparosNave, &disparosOvnis, &puntos, &vidasExtra, &vidas) {

				vidas--
				GuardarEstado(vidas, puntos, velocidadJuego) //guardo las vidas y los puntos globales (del anterior)
				// Si no tiene más vidas, se devuelve pantalla gameOver
				if vidas <= 0 {
					enviarGameOver(puntos)
					return
				}
			} else {
				if len(ovnis) == 0 {
					if velocidadJuego > 100 {
						velocidadJuego -= 50
					}

					GuardarEstado(vidas, puntos, velocidadJuego)
					enviarWin(puntos)

					disparosNave = nil
					disparosOvnis = nil
					//Se genera la nave y ovnis otra vez (posición inicial)
					nave, direccionNave = inicializarNave(constCantFilasTablero, constCantColumnasTablero)
					ovnis = inicializarOvnis(constCantFilasTablero, constCantColumnasTablero)

					//generarEventos()
					return
				}
				GuardarEstado(vidas, puntos, velocidadJuego)
				enviarActualizacionTexto(fmt.Sprint("Puntaje: ", puntos, ". Vidas: ", vidas))
			}

			//Se actualiza el tablero con los valores de la nave, ovnis y disparos en sus nuevas posiciones
			actualizarTablero(&tablero, nave, disparosNave, ovnis, disparosOvnis, &vidasExtra)

			// Se envía actualización de tablero al cliente para mostrar en pantalla
			enviarActualizacionTablero(tablero)

			GuardarEstado(vidas, puntos, velocidadJuego) //sin este aveces se tilda mucho tiempo, idk

			// Espera un tiempo antes de generar un nuevo movimiento

			time.Sleep(time.Duration(velocidadJuego) * time.Millisecond)
		}
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

	nave[constY] = cantFilasTablero - 3
	nave[constX] = cantColumnasTablero / 2

	return nave, quieto

}

func inicializarOvnis(cantFilasTablero int, cantColumnasTablero int) [][constCantColumnasOvni]int {

	var (
		ovnis             [][constCantColumnasOvni]int
		ovnisvector       [4]int
		varPatronFilas    int = 3 //2 //7
		varPatronColumnas int = 2 //2
		cantidadLideres   int = 0
		TipoOvniAleatorio int
		maxCantLideres    = 1 //18
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

			ovnisvector[constTipoOvni] = TipoOvniAleatorio
			ovnisvector[constOvniY] = f
			ovnisvector[constOvniX] = c
			ovnisvector[constEnDescenso] = 0
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
	disparosOvnis [][constCantColumnas]int, vidasExtra *[constCantColumnas]int) {

	var (
		posicionY, posicionX int
	)

	//Borrar los elementos
	for f := 1; f < (constCantFilasTablero - 1); f++ {
		for c := 1; c < (constCantColumnasTablero - 1); c++ {
			tablero[f][c] = ""
		}
	}

	//Recorrer el vector ovni, y representarlos en tablero dependiendo su tipo
	for f := 0; f < len(ovnis); f++ {
		if ovnis[f][constTipoOvni] == 1 {
			posicionY = ovnis[f][constOvniY]
			posicionX = ovnis[f][constOvniX]

			tablero[posicionY][posicionX] = "L"
		} else {
			posicionY = ovnis[f][constOvniY]
			posicionX = ovnis[f][constOvniX]

			tablero[posicionY][posicionX] = "C"
		}
	}

	//Representar vidas
	if vidasExtra[constX] != 0 {
		posicionX = vidasExtra[constX]
		posicionY = vidasExtra[constY]
		tablero[posicionY][posicionX] = "Q"
	}

	//Representar nave
	posicionY = nave[constY]
	posicionX = nave[constX]
	tablero[posicionY][posicionX] = "N"

	//Representar disparos nave
	for f := 0; f < len(disparosNave); f++ {
		posicionY = disparosNave[f][constY]
		posicionX = disparosNave[f][constX]
		tablero[posicionY][posicionX] = "*"
	}

	//Representar disparos ovnis

	for f := 0; f < len(disparosOvnis); f++ {
		posicionY = disparosOvnis[f][constY]
		posicionX = disparosOvnis[f][constX]
		tablero[posicionY][posicionX] = "."
	}
}

func calcularNuevaPosicionNave(tablero [constCantFilasTablero][constCantColumnasTablero]string,
	nave *[constCantColumnas]int, direccionNave *[constCantColumnas]int) {

	//actualizar posición nave
	filaAnterior := nave[constY]
	columnaAnterior := nave[constX]

	tablero[filaAnterior][columnaAnterior] = ""

	nave[constY] += direccionNave[constY]
	nave[constX] += direccionNave[constX]

	//verificar que no sea un borde y volver si es necesario
	if nave[constY] == 0 || nave[constY] == (constCantFilasTablero-1) {
		nave[constY] -= direccionNave[constY]
	} else if nave[constX] == 0 || nave[constX] == (constCantColumnasTablero-1) {
		nave[constX] -= direccionNave[constX]
	}

	//Devolver posicion nave a 0
	direccionNave[constY] = 0
	direccionNave[constX] = 0

}

func crearDisparoNave(nave [constCantColumnas]int,
	disparoNave *bool,
	disparosNave *[][constCantColumnasDisparos]int) {

	var (
		nuevoDisparo [constCantColumnas]int
	)

	if *disparoNave {
		nuevoDisparo[constY] = nave[constY] //-1
		nuevoDisparo[constX] = nave[constX]
		*disparosNave = append(*disparosNave, nuevoDisparo)
		*disparoNave = false
	}

}

func crearDisparoOvni(ovnis [][constCantColumnasOvni]int,
	disparosOvnis *[][constCantColumnasDisparos]int) {

	var (
		nuevoDisparo [constCantColumnas]int
	)
	rand.Seed(time.Now().UnixNano())
	ovniElegido := rand.Intn(len(ovnis))

	nuevoDisparo[constY] = ovnis[ovniElegido][constOvniY] // Y
	nuevoDisparo[constX] = ovnis[ovniElegido][constOvniX] // X
	*disparosOvnis = append(*disparosOvnis, nuevoDisparo)

}

func calcularNuevasPosicionesDisparos(tablero [constCantFilasTablero][constCantColumnasTablero]string,
	disparosNave [][constCantColumnasDisparos]int,
	disparosOvnis [][constCantColumnasDisparos]int) {

	for f := 0; f < len(disparosNave); f++ {
		disparosNave[f][constY] = disparosNave[f][constY] - 1
	}

	for f := 0; f < len(disparosOvnis); f++ {
		disparosOvnis[f][constY] = disparosOvnis[f][constY] + 1
	}
}

// Nueva Función que cada ciertos puntos nos da una vida
func crearVidas(puntos int, vidas int, ovnis [][constCantColumnasOvni]int, vidasExtra *[constCantColumnas]int, cantVidasCreadas *int) {
	var (
		cantPuntosNecesarias int = 50
	)
	/*Lo que buscamos hacer aquí es evitar que al tener 50/100/150... puntos, no se creen
	vidas de forma infinita, (hasta que tengamos un puntaje diferente a multiplo de 50
	por primero constatamos que tengamos más de 0 puntos, y cada vez que creamos una vida
	incrementamos el monto necesario para obtener una más de a 50 puntos*/
	if puntos != 0 {
		if *cantVidasCreadas == 0 {
			if len(ovnis) > 0 {
				if puntos%cantPuntosNecesarias == 0 {

					rand.Seed(time.Now().UnixNano())
					ovniElegido := rand.Intn(len(ovnis))

					vidasExtra[constX] = ovnis[ovniElegido][constX]
					//vidasExtra[constY] = ovnis[ovniElegido][constY]
					vidasExtra[constY] = 1
					(*cantVidasCreadas)++
				}
			}

		} else {
			cantPuntosNecesarias = (*cantVidasCreadas) * 50
			if len(ovnis) > 0 {
				if puntos%cantPuntosNecesarias == 0 {

					rand.Seed(time.Now().UnixNano())
					ovniElegido := rand.Intn(len(ovnis))

					vidasExtra[constX] = ovnis[ovniElegido][constX]
					//vidasExtra[constY] = ovnis[ovniElegido][constY]
					vidasExtra[constY] = 1
					(*cantVidasCreadas)++
				}
			}

		}
	}

}

// calcular nueva poscion de las vidas
func calcularNuevaPosicionVidas(vidasExtra *[constCantColumnas]int) {

	if vidasExtra[constX] != 0 {
		vidasExtra[constY]++
	}

}

func verificarEstadoDeJuego(tablero [constCantFilasTablero][constCantColumnasTablero]string,
	nave [constCantColumnas]int,
	ovnis *[][constCantColumnasOvni]int,
	disparosNave *[][constCantColumnasDisparos]int,
	disparosOvnis *[][constCantColumnasDisparos]int,
	puntos *int, vidasExtra *[constCantColumnas]int, vidas *int) bool {

	var (

	//disparosEliminar[][2]int
	)

	//Eliminar disparos de nave cuando tocan borde

	for f := 0; f < len(*disparosNave); {
		if (*disparosNave)[f][constY] == 0 {
			coordenadaY := (*disparosNave)[f][constY]
			coordenadaX := (*disparosNave)[f][constX]
			(*disparosNave) = eliminarDisparo(*disparosNave, coordenadaY, coordenadaX)

		} else {
			f++
		}
	}

	/* Eliminar las vidas que toquen el borde. Al no poder eliminarlas, las colocamos
	en la columna 0, donde el programa no las representa, y tampoco les acuatualiza
	la posicion. Esto hasta que se se "Cree otra vida" y se le reasignen los valores)*/
	if (*vidasExtra)[constY] == constCantFilasTablero-2 {
		(*vidasExtra)[constX] = 0
	}

	//si tocamos una vida que está en caída, sumarnos una vida y eliminar (esconder) la vida
	if (*vidasExtra)[constY] == nave[constY] && (*vidasExtra)[constX] == nave[constX] {
		(*vidas)++
		vidasExtra[constX] = 0
	}

	//Eliminar disparo ovni
	/*for f2 := 0; f2 < len(*disparosNave); {
		for f1 := 0; f1 < len(*disparosOvnis); f1++ {
			if (*disparosNave)[f2][constY] == (*disparosOvnis)[f1][constY]-1 && (*disparosNave)[f2][constX] == (*disparosOvnis)[f1][constX] {
				coordenadaY := (*disparosNave)[f2][constY]
				coordenadaX := (*disparosNave)[f2][constX]
				(*disparosNave) = eliminarDisparo(*disparosNave, coordenadaY, coordenadaX)
				break
			}

		}
		f2++
	}*/

	//elminar disparos de nave que toquen disparos de ovnis
	for f2 := len(*disparosNave) - 1; f2 >= 0; f2-- {
		if len(*disparosNave) == 0 {
			break
		}
		for f1 := len(*disparosOvnis) - 1; f1 >= 0; f1-- {
			if ((*disparosNave)[f2][constY] == (*disparosOvnis)[f1][constY]-1 && (*disparosNave)[f2][constX] == (*disparosOvnis)[f1][constX]) || ((*disparosNave)[f2][constY] == (*disparosOvnis)[f1][constY] && (*disparosNave)[f2][constX] == (*disparosOvnis)[f1][constX]) {
				coordenadaY := (*disparosNave)[f2][constY]
				coordenadaX := (*disparosNave)[f2][constX]
				(*disparosNave) = eliminarDisparo(*disparosNave, coordenadaY, coordenadaX)
			}
		}
	}
	//Eliminar ovnis que hayamos disparado hay error acá

	for f2 := len(*disparosNave) - 1; f2 >= 0; f2-- {

		for f := 0; f < len(*ovnis); f++ {
			if len(*disparosNave) != 0 {
				if (*ovnis)[f][constTipoOvni] == 2 && (*ovnis)[f][constEnDescenso] == 0 && (*disparosNave)[f2][constY] == (*ovnis)[f][constOvniY] && (*disparosNave)[f2][constX] == (*ovnis)[f][constOvniX] {
					coordenadaY := (*ovnis)[f][constOvniY]
					coordenadaX := (*ovnis)[f][constOvniX]
					(*ovnis) = eliminarOvni(*ovnis, coordenadaY, coordenadaX)

					coordenadaY = (*disparosNave)[f2][constY]
					coordenadaX = (*disparosNave)[f2][constX]
					(*disparosNave) = eliminarDisparo(*disparosNave, coordenadaY, coordenadaX)

					*puntos += 10
					break

				}
			}

		}
	}

	//Eliminar disparo ovni toca tablero
	for f := 0; f < len(*disparosOvnis); {
		if (*disparosOvnis)[f][constY] == constCantFilasTablero-1 {
			coordenadaY := (*disparosOvnis)[f][constY]
			coordenadaX := (*disparosOvnis)[f][constX]
			(*disparosOvnis) = eliminarDisparo(*disparosOvnis, coordenadaY, coordenadaX)
		} else {
			f++
		}

	}
	//Eliminar ovni en descenso que toca el borde
	for f := 0; f < len(*ovnis); {
		if (*ovnis)[f][constOvniY] == constCantFilasTablero-1 {
			coordenadaY := (*ovnis)[f][constOvniY]
			coordenadaX := (*ovnis)[f][constOvniX]
			(*ovnis) = eliminarOvni(*ovnis, coordenadaY, coordenadaX)
		} else {
			f++
		}

	}

	//retornar false si disparo ovni toca a la nave
	for f := 0; f < len(*disparosOvnis); f++ {
		if (*disparosOvnis)[f][constY] == nave[constY] && (*disparosOvnis)[f][constX] == nave[constX] {
			return false
		}
	}

	//Cambiar ovnis lideres por comunes al recibir disparo
	for f := len(*disparosNave) - 1; f >= 0; f-- {
		for f2 := 0; f2 < len(*ovnis); f2++ {
			if len(*disparosNave) != 0 {
				if (*ovnis)[f2][constTipoOvni] == 1 && (*ovnis)[f2][constEnDescenso] == 0 && (*disparosNave)[f][constY] == (*ovnis)[f2][constOvniY] && (*disparosNave)[f][constX] == (*ovnis)[f2][constOvniX] {

					(*ovnis)[f2][constTipoOvni] = 2

					coordenadaY := (*disparosNave)[f][constY]
					coordenadaX := (*disparosNave)[f][constX]
					(*disparosNave) = eliminarDisparo(*disparosNave, coordenadaY, coordenadaX)
					break
				}
			}

		}

	}

	//Si un ovni choca, restar una vida
	for f := 0; f < len(*ovnis); f++ {
		if (*ovnis)[f][constOvniY] == nave[constY] && (*ovnis)[f][constOvniX] == nave[constX] {
			coordenadaX := (*ovnis)[f][constOvniX]
			coordenadaY := (*ovnis)[f][constOvniY]

			(*ovnis) = eliminarOvni(*ovnis, coordenadaY, coordenadaX)
			return false
		}
	}

	return true
}

func eliminarDisparo(slice [][constCantColumnasDisparos]int, coordenadaY int, coordenadaX int) [][constCantColumnasDisparos]int {
	var nuevoSlice [][constCantColumnasDisparos]int
	for f := 0; f < len(slice); f++ {
		if !(slice[f][constY] == coordenadaY && slice[f][constX] == coordenadaX) {
			nuevoSlice = append(nuevoSlice, slice[f])
		}
	}
	return nuevoSlice
}

func eliminarOvni(slice [][constCantColumnasOvni]int, coordenadaY int, coordenadaX int) [][4]int {
	var nuevoSlice [][constCantColumnasOvni]int
	for f := 0; f < len(slice); f++ {
		if slice[f][constOvniY] != coordenadaY || slice[f][constOvniX] != coordenadaX {
			nuevoSlice = append(nuevoSlice, slice[f])
		}
	}
	return nuevoSlice
}

func liberarOvni(ovnis [][constCantColumnasOvni]int) {
	//Elegir ovni random
	rand.Seed(time.Now().UnixNano())
	ovniElegido := rand.Intn(len(ovnis))

	//Cambiar estado del ovni
	ovnis[ovniElegido][constEnDescenso] = 1
}

func calcularNuevaPosicionOvnisLiberados(ovnis [][constCantColumnasOvni]int) {
	for f := 0; f < len(ovnis); f++ {
		if ovnis[f][constEnDescenso] == 1 {
			ovnis[f][constOvniY]++
		}
	}

}

/*func ReiniciarJuegoCompleto() {
	Vidas = 3
	Puntos = 0
	VelocidadJuego = 300
	direccionNave = quieto
	disparoNave = false
	fake
}*/
