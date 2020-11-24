package main

import(
        "math/rand"
        "strconv"
        "time"
        "fmt"
        "os"
        "encoding/csv"
        "log"
        "context"
        "google.golang.org/grpc"
        pb "github.com/Pantuflaa/grpc/producto/pb"
)

// Struct para guardar los datos de los pedidos
type infoPedido struct {
   Id string
   producto string
   tipo string
   valor int32
   origen string
   destino string
   prioritario int32
   intentos int32
   seguimiento int32
}

// Funcion que revisar errores
func check(err error, message string){
    if err != nil{
        log.Fatal(message, err)
    }
}

// Funcion que simula el comportamiento de un camion, desde que pide un paquete hasta que vuelve a la central
func camion(tipo string, tiempo int32, delivery int32, idCamion int32){
    var data []string
    s := strconv.Itoa(int(idCamion))
    nameFile := s+tipo+".csv"           // Se crea el nombre del archivo a escribir
    arch, error := os.Open(nameFile)
    if(error != nil){                   // Si no existe se crea
        data = []string{"id-paquete", "tipo", "valor", "origen", "destino", "intentos", "fecha-entrega"}
        arch, error = os.Create(nameFile)
        check(error, "Error al crear el archivo")
        csvwriter := csv.NewWriter(arch)
        _ = csvwriter.Write(data)
        csvwriter.Flush()
    }
    defer arch.Close()
    for true{                            // Comienza el ciclo de actividades del camion
        var puerto string
        var tryPaquete bool = true      // Variables para saber si el camion debe seguir intentando entregar
        var tryPaquete2 bool = true     // el paquete respectivo.
        var entregado bool = false      // Variables para saber si el respectivo paquete fue entregado
        var entregado2 bool = false
        if(idCamion == 1){               // Se define a que puerto se comunicara
            puerto = "10.6.40.218:9002"
        } else if(idCamion == 2){
            puerto = "10.6.40.218:9003"
        } else{
            puerto = "10.6.40.218:9004"
        }
        var prioritario int32 = 0       // Si recibe un paquete retail, se vuelve 1 (true)
										// porque puede llevar prioritarios en la siguiente entrega
		var conn *grpc.ClientConn
		conn, errk := grpc.Dial(puerto, grpc.WithInsecure())
		if errk != nil{
			log.Fatalf("No se pudo conectar: %S", errk)
		}
		defer conn.Close()
		c := pb.NewPeticionClient(conn)
		infoCamion := &pb.ActEstado{    // Se crea la estructura para realizar actualizacion de estado
			Seguimiento: prioritario,
			Estado: tipo,
			IdCamion: idCamion,
			}
		paquete := &pb.InfoPaquete{     // Se crea un paquete vacio
			Id: "",
			Tipo: "",
			Valor: -1,
			Origen: "",
			Destino: "",
			Intentos: 0,
			Seguimiento: 0,
			}
		for true{                       // Se pide el primer paquete hasta recibir uno
			paquete, _ = c.PedirPaquete(context.Background(), infoCamion)
			if(paquete.Valor != -1){
				break
			}
		}
		log.Printf("El camion %v recibio el paquete %v", idCamion, paquete.Seguimiento)
		paquete2 := paquete
		paquete2, _ = c.PedirPaquete(context.Background(), infoCamion)  // Se intenta pedir el segundo paquete
		if(paquete2.Valor == -1){       // Si no lo recibe, espera el tiempo definido al inicio de la ejecucion
			time.Sleep(time.Duration(tiempo) * time.Second)
			paquete2, _ = c.PedirPaquete(context.Background(), infoCamion)
			if(paquete2.Valor != -1){
				log.Printf("El camion %v recibio el paquete %v y se fue", idCamion, paquete2.Seguimiento)
				if(paquete2.Tipo == "Retail"){
					prioritario = 1
				}
			} else{
				log.Printf("El camion %v no recibio un segundo paquete y se fue", idCamion)
				tryPaquete2 = false // Ya que no hay segundo paquete no intenta entregarlo
			}
		} else{
			log.Printf("El camion %v recibio el paquete %v y se fue", idCamion, paquete2.Seguimiento)
		}
		if(paquete.Tipo == "Retail"){
			prioritario = 1
		}
		if(paquete.Valor < paquete2.Valor){ // Se verifica el valor de los paquetes para entregar
			aux := paquete                   // primero el que genere mayor valor
			paquete = paquete2
			paquete2 = aux
		}
		for tryPaquete || tryPaquete2{ // Se itera las veces necesarias para que el camion intente entregar
			if(tryPaquete){            // ambos paquetes
				time.Sleep(time.Duration(delivery) * time.Second)
				tryPaquete = !entregarPaquete(paquete) // Se intenta entregar el paquete
				paquete.Intentos += 1
				log.Printf("El camion %v intento entregar el paquete %v", idCamion, paquete.Seguimiento)
				if(tryPaquete){
					tryPaquete = verificarPaquete(paquete) // Se verifica si debe seguir intentando
				} else{
					entregado = true
				}
			}
			// Analogamente para el paquete 2
            if(tryPaquete2){
                time.Sleep(time.Duration(delivery) * time.Second)
                tryPaquete2 = !entregarPaquete(paquete2)
                paquete2.Intentos += 1
                log.Printf("El camion %v intento entregar el paquete %v", idCamion, paquete2.Seguimiento)
                if(tryPaquete2){
                    tryPaquete2 = verificarPaquete(paquete2)
                } else{
                    entregado2 = true
                }
            }
        }

        // Se crea la estructura para notificar a logistica la informacion de la entrega
        // del paquete 1 y escribir en el archivo del camion
        var estado string = "No recibido"
        if(entregado){
            estado = "Recibido"
        }
        envio := &pb.Envio{
            Seguimiento: paquete.Seguimiento,
            Estado: estado,
            Intentos: paquete.Intentos,
            }

        c.EnviarPaquete(context.Background(), envio)
        if(entregado){
            aux := time.Now()
            date := aux.Format("2006-01-02 15:04:05")
            data = []string{paquete.Id, paquete.Tipo, strconv.Itoa(int(paquete.Valor)), paquete.Origen, paquete.Destino, strconv.Itoa(int(paquete.Intentos)), date}
        } else{
            data = []string{paquete.Id, paquete.Tipo, strconv.Itoa(int(paquete.Valor)), paquete.Origen, paquete.Destino, strconv.Itoa(int(paquete.Intentos)), "0"}
        }
        arch1, error := os.OpenFile(nameFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
        check(error, "Error al abrir el archivo")
        defer arch1.Close()
        w := csv.NewWriter(arch1)
        w.Write(data)
        w.Flush()

        // Analogamente para el paquete 2
        if(paquete2.Valor != -1){
            var estado2 string = "No recibido"
            if(entregado2){
                estado2 = "Recibido"
            }
            envio2 := &pb.Envio{
                Seguimiento: paquete2.Seguimiento,
                Estado: estado2,
                Intentos: paquete2.Intentos,
            }
            c.EnviarPaquete(context.Background(), envio2)
            if(entregado2){
                aux2 := time.Now()
                date2 := aux2.Format("2006-01-02 15:04:05")
                data = []string{paquete2.Id, paquete2.Tipo, strconv.Itoa(int(paquete2.Valor)), paquete2.Origen, paquete2.Destino, strconv.Itoa(int(paquete2.Intentos)), date2}
            } else{
                data = []string{paquete2.Id, paquete2.Tipo, strconv.Itoa(int(paquete2.Valor)), paquete2.Origen, paquete2.Destino, strconv.Itoa(int(paquete2.Intentos)), "0"}
            }
            arch2, error2 := os.OpenFile(nameFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
            check(error2, "Error al abrir el archivo")
            defer arch2.Close()
            w2 := csv.NewWriter(arch2)
			w2.Write(data)
			w2.Flush()
            log.Printf("Volvio el camion %v a la central, el paquete %v fue %v y el paquete %v fue %v", idCamion, paquete.Seguimiento, estado, paquete2.Seguimiento, estado2)
        } else{
            log.Printf("Volvio el camion %v a la central, el paquete %v fue %v", idCamion, paquete.Seguimiento, estado)
        }
    }
}

// Funcion para verificar si se debe seguir intentando entregar un paquete
func verificarPaquete(paquete *pb.InfoPaquete)(bool){
    if(paquete.Tipo == "Normal" || paquete.Tipo == "Prioritario"){
        if(paquete.Valor <= 10*(paquete.Intentos-1) || paquete.Intentos > 2){
            return false
        }
    } else{
        if(paquete.Intentos == 3){
            return false
        }
    }
    return true
}

// Funcion para simular el intento de entrega de un paquete, con 80% de probabilidad de exito
func entregarPaquete(paquete *pb.InfoPaquete)(bool){
    prob := rand.Intn(100)
    if(prob < 80){
        return true
    } else{
        return false
    }
}

// Flujo principal donde se piden los tiempos y se invocan los 3 camiones
func main(){
    var wait int32
    fmt.Println("Ingrese la cantidad de segundos que esperara por el segundo paquete:  ")
    fmt.Scanln(&wait)
    var delivery int32
    fmt.Println("Ingrese la cantidad de segundos que demorara un envio:  ")
    fmt.Scanln(&delivery)
    go camion("Normal", wait, delivery, int32(1))
    go camion("Retail", wait, delivery, int32(2))
    camion("Retail", wait, delivery, int32(3))
}