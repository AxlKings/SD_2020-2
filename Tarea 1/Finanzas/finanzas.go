package main

import (
	"fmt"
	"log"
	"strconv"
	"os"
	"os/signal"
	"syscall"
	"time"
	"encoding/csv"
	"encoding/json"
	"github.com/streadway/amqp"
)
//Funcion utilizada para printear errores
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
// Formato en Json en el cual se guardaran los datos recibidos por logistica
type PedidoJson struct {
   Id string `json:"id"`
   Estado string `json:"estado"`
   IdCamion int32 `json:"idCamion"`
   TipoCamion string `json:"tipoCamion"`
   Intentos int32 `json:"intentos"`
   Valor int32 `json:"valor"`
   TipoPedido string `json:"tipoPedido"`
}
//Print que se realizara al final del programa (al apretar CTL + C)
func printFinal () {
    println("")
    log.Printf("Proceso terminado.")
    fmt.Printf("Costos: %v \nGanancias: %v \nPerdidas: %v \nTotal: %v \nEnvios Totales : %v \nEnvios Entregados: %v \nEnvios No Entregados: %v\n",Costos,Ganancias,Perdidas,Total,EnviosTotales,EnviosEntregados,EnviosNoEntregados)  
}
//Variables globales que se utilizarán para guardar los datos de la sesión
var Costos float64
var Ganancias float64
var Perdidas float64
var Total float64
var EnviosTotales int32
var EnviosEntregados int32
var EnviosNoEntregados int32
func main() {
	//Conexion con RabbitMQ
	conn, err := amqp.Dial("amqp://test:test@10.6.40.220:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()
	//Creacion de canal con rabbitMQ
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()
	//Crear cola con rabbitMQ con nombre finanzas
	q, err := ch.QueueDeclare(
		"finanzas", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")
	//Lectura de mensajes en cola
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)
	//Verificacion de la existencia de registro, si no existe lo crea, si existe lo abre.
	var path = "registro.csv"
	file, errf := os.OpenFile(path, os.O_APPEND|os.O_WRONLY,os.ModeAppend)
	if errf != nil {
		file, errf = os.Create(path)
		if errf != nil {
			log.Fatalf("Error al crear los registros csv")
		}
		writer := csv.NewWriter(file)
		tags := []string {"Fecha","IdPaquete","Intentos","Valor","Estado","GananciaOPerdidas"}
		_ = writer.Write(tags)
		writer.Flush()
	}
	file.Close()
	//Funcion que recibira los paquete de logistica y calculara las variables relevantes
	go func() {
		Costos = 0
		Ganancias = 0
		Total = 0
		Perdidas = 0
		var mensajazo PedidoJson
		for d := range msgs {
			//Se decodifica el mensaje recibido
			json.Unmarshal(d.Body,&mensajazo)
			//Se abre el archivo de registro
			file, errf := os.OpenFile("registro.csv", os.O_APPEND|os.O_WRONLY,os.ModeAppend)
			if errf != nil {
				log.Fatalf("Error al abrir el archivo de registros")
			}
			//Se comienza el proceso de escritura
			writer := csv.NewWriter(file)
			//Id Estado IdCamion TipoCamion Intentos Valor TipoPedido
			hora:= time.Now()
			escribir := []string {hora.Format("2006-01-02 15:04:05"),mensajazo.Id,strconv.Itoa(int(mensajazo.Intentos)),strconv.Itoa(int(mensajazo.Valor)),mensajazo.Estado,strconv.Itoa(int(mensajazo.Valor - (mensajazo.Intentos - 1)*10))}
			writer.Write(escribir)
			writer.Flush()
			file.Close()
			// Se termina el proceso de escritura
			// Se realiza el calculo de variables a partir de las reglas de negocio
			//Costos Ganancias Perdidas Total EnviosTotales EnviosEntregados EnviosNoEntregados
			aux := float64((mensajazo.Intentos - 1)*10)
			realValor := float64(mensajazo.Valor)
			if(mensajazo.TipoPedido == "Normal" && mensajazo.Estado == "No Recibido"){
					realValor = 0
			}
			if(mensajazo.TipoPedido == "Prioritario" && mensajazo.Estado == "No Recibido"){
					realValor = realValor * 0.3
			}
			if(mensajazo.TipoPedido == "Prioritario" && mensajazo.Estado == "Recibido"){
					realValor += realValor * 0.3
			}
			Costos += aux
			if (realValor - aux < 0) {
				Perdidas += realValor - aux
			}else{
				Ganancias += realValor - aux
			}
			Total += realValor - aux
			EnviosTotales += 1
			if mensajazo.Estado == "Recibido"{
				EnviosEntregados += 1
			} else {
				EnviosNoEntregados +=1
			}
			log.Printf("Recibidos los datos del paquete con Id: %s", mensajazo.Id)
		}
	}()
	//Se captura la señal de final de programa de CTRL+C, con el objetivo de printear los resultados de la sesion
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		printFinal()
		os.Exit(1)
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}