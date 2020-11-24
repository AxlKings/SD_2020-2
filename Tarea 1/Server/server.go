package main

    import (
    "os"
    "strconv"
    "encoding/csv"
    "google.golang.org/grpc"
    "log"
    "encoding/json"
    "time"
    "context"
    "net"
    "sync"
    "github.com/streadway/amqp"
    pb "github.com/Pantuflaa/grpc/producto/pb"
)

type productoServer struct {
    pb.UnimplementedPeticionServer
}

// Funcion para que el cliente realice un pedido
func (s *productoServer) RealizarPeticion(ctx context.Context, objeto *pb.Objeto) (*pb.Serie, error) {
    log.Printf("Recibi el objeto con la siguiente id: %s", objeto.Id)

    // Se mapean todos los campos del paquete al struct que contiene la informacion de este
    // ademas de otros campos necesarios a futuro
    var info infoPedido
    info.Id = objeto.Id
    info.producto = objeto.Producto
    info.valor = objeto.Valor
    info.origen = objeto.Tienda
    info.destino = objeto.Destino
    SERIE = SERIE + 1
    info.seguimiento = SERIE
    info.prioritario = objeto.Prioritario
    info.intentos = 0
    if(info.prioritario == 0){ // Se utilizaron locks para el uso de variables compartidas entre hebras
        info.tipo = "Normal"
        normales.mux.Lock()
        normales.queue = append(normales.queue, info)
        normales.mux.Unlock()
    } else if(info.prioritario == -1){
        info.tipo = "Retail"
        retails.mux.Lock()
        retails.queue = append(retails.queue, info)
        retails.mux.Unlock()
    } else{
        info.tipo = "Prioritario"
        prioritarios.mux.Lock()
        prioritarios.queue = append(prioritarios.queue, info)
        prioritarios.mux.Unlock()
    }

    // Se crea el struct con la informacion necesaria para la consulta de seguimiento
    pedido := Pedido{
        Id: info.Id,
        estado: "En bodega",
        IdCamion: 0,
        tipoCamion: "",
        intentos: 0,
        valor: info.valor,
        tipoPedido: info.tipo,
	}
	estados.mux.Lock()
    estados.mapa[SERIE] = pedido
    estados.mux.Unlock()

    // Se escribe el registro en el archivo
    file, errf := os.OpenFile("registro.csv", os.O_APPEND|os.O_WRONLY,os.ModeAppend)
    if errf != nil {
        log.Fatalf("Error al abrir el archivo de registros")
    }
    writer := csv.NewWriter(file)
    hora:= time.Now()
    //fecha id tipo nombre valor origen destino seguimiento
    var seg string
    if(info.prioritario == -1){ // Si es de retail se escribe 0 como su numero de seguimiento
        seg = "0"
    } else{
        seg = strconv.Itoa(int(SERIE))
    }
    escribir := []string {hora.Format("2006-01-02 15:04:05"),info.Id,info.tipo,info.producto,strconv.Itoa(int(info.valor)),info.origen,info.destino, seg}
    writer.Write(escribir)
    writer.Flush()
    file.Close()
    return &pb.Serie{Serie: SERIE}, nil // Se retorna el numero de seguimiento
}

// Funcion para que el cliente realice una consulta de seguimiento
func (s *productoServer) PedirEstado(ctx context.Context, serie *pb.Serie) (*pb.Estado, error) {
    log.Printf("Recibi el siguiente numero de seguimiento: %v", serie.Serie)
    return &pb.Estado{Estado: estados.mapa[serie.Serie].estado}, nil
}

// Funcion para actualizar el estado de un paquete
func (s *productoServer) ActualizarEstado(ctx context.Context, info *pb.ActEstado) (*pb.Estado, error) {
    log.Printf("Se actualizo el estado del paquete %v a %v", info.GetSeguimiento(), info.Estado)
    return &pb.Estado{Estado: "Actualizacion exitosa"}, nil
}

// Funcion para que el camion le entregue la informacion del paquete que se intento entregar a logistica
// luego de volver a la central
func (s *productoServer) EnviarPaquete(ctx context.Context, paquete *pb.Envio) (*pb.Estado, error){
    estados.mux.Lock()
    pedido := estados.mapa[paquete.Seguimiento]
    pedido.estado = paquete.Estado
    pedido.intentos = paquete.Intentos
    estados.mapa[paquete.Seguimiento] = pedido
    estados.mux.Unlock()
    pedidos.mux.Lock()
    pedidos.queue = append(pedidos.queue, pedido) // Se agrega a la cola para que logistica lo envie a
    pedidos.mux.Unlock()                          // finanzas
    return &pb.Estado{Estado: "Recibido"}, nil
}

// Funcion para que los camiones pidan paquetes para entregar
func (s *productoServer) PedirPaquete(ctx context.Context, info *pb.ActEstado) (*pb.InfoPaquete, error) {
    var paquete infoPedido
    pedido := &pb.InfoPaquete{ // Se crea un paquete vacio para rellenarlo segun corresponda con
            Id: "",            // el paquete a entregar
            Tipo: "",
            Valor: -1,
            Origen: "",
            Destino: "",
            Intentos: 0,
            Seguimiento: 0,
        }
	if(info.Estado == "Normal"){ // Si el camion es tipo normal se intenta entregar un
		if(len(prioritarios.queue) != 0){  // paquete prioritario
            prioritarios.mux.Lock()
            paquete = prioritarios.queue[0]
            prioritarios.queue = prioritarios.queue[1:]
            prioritarios.mux.Unlock()
            log.Printf("Se entrego el paquete %v al camion tipo %v", paquete.Id, info.Estado)
            estados.mux.Lock()
            pedi2 := estados.mapa[paquete.seguimiento]
            pedi2.estado = "En camino"
            pedi2.IdCamion = info.IdCamion
            pedi2.tipoCamion = info.Estado
            estados.mapa[paquete.seguimiento] = pedi2
            estados.mux.Unlock()
            return &pb.InfoPaquete{
                Id: paquete.Id,
                Tipo: paquete.tipo,
                Valor: paquete.valor,
                Origen: paquete.origen,
                Destino: paquete.destino,
                Intentos: paquete.intentos,
                Seguimiento: paquete.seguimiento,
            }, nil
        } else if(len(normales.queue) != 0){ // Sino, se intenta entregar un paquete normal
            normales.mux.Lock()
            paquete = normales.queue[0]
            normales.queue = normales.queue[1:]
            defer normales.mux.Unlock()
            log.Printf("Se entrego el paquete %v al camion tipo %v", paquete.Id, info.Estado)
            estados.mux.Lock()
            pedi2 := estados.mapa[paquete.seguimiento]
            pedi2.estado = "En camino"
            pedi2.IdCamion = info.IdCamion
            pedi2.tipoCamion = info.Estado
            estados.mapa[paquete.seguimiento] = pedi2
            estados.mux.Unlock()
            return &pb.InfoPaquete{
                Id: paquete.Id,
                Tipo: paquete.tipo,
                Valor: paquete.valor,
                Origen: paquete.origen,
                Destino: paquete.destino,
                Intentos: paquete.intentos,
                Seguimiento: paquete.seguimiento,
            }, nil
        } else{ // En caso de que no haya pedidos disponibles, se entrega un paquete vacio
            return pedido, nil
        }
    } else{
        if(len(retails.queue) != 0){ // Si el camion es de Retail, se intenta entregar primero un
            retails.mux.Lock()       // paquete de retail
            paquete = retails.queue[0]
            retails.queue = retails.queue[1:]
            defer retails.mux.Unlock()
            log.Printf("Se entrego el paquete %v al camion tipo %v", paquete.Id, info.Estado)
            estados.mux.Lock()
            pedi2 := estados.mapa[paquete.seguimiento]
            pedi2.estado = "En camino"
            pedi2.IdCamion = info.IdCamion
            pedi2.tipoCamion = info.Estado
            estados.mapa[paquete.seguimiento] = pedi2
            estados.mux.Unlock()
            return &pb.InfoPaquete{
                Id: paquete.Id,
                Tipo: paquete.tipo,
				Valor: paquete.valor,
				Origen: paquete.origen,
                Destino: paquete.destino,
                Intentos: paquete.intentos,
                Seguimiento: paquete.seguimiento,
            }, nil
        } else if(info.Seguimiento == 1 && len(prioritarios.queue) != 0){ // Sino, se verifica si puede
            prioritarios.mux.Lock()                                    // llevar paquetes prioritarios
            paquete = prioritarios.queue[0]                            // y si hay disponibles
            prioritarios.queue = prioritarios.queue[1:]
            defer prioritarios.mux.Unlock()
            log.Printf("Se entrego el paquete %v al camion tipo %v", paquete.Id, info.Estado)
            estados.mux.Lock()
            pedi2 := estados.mapa[paquete.seguimiento]
            pedi2.estado = "En camino"
            pedi2.IdCamion = info.IdCamion
            pedi2.tipoCamion = info.Estado
            estados.mapa[paquete.seguimiento] = pedi2
            estados.mux.Unlock()
            return &pb.InfoPaquete{
                Id: paquete.Id,
                Tipo: paquete.tipo,
                Valor: paquete.valor,
                Origen: paquete.origen,
                Destino: paquete.destino,
                Intentos: paquete.intentos,
                Seguimiento: paquete.seguimiento,
            }, nil
        } else{
            return pedido, nil
        }
    }
}

// Struct con la informacion necesaria para el seguimiento de los pedidos
type Pedido struct {
   Id string
   estado string
   IdCamion int32
   tipoCamion string
   intentos int32
   valor int32
   tipoPedido string
}

// Struct con la informacion necesaria para finanzas, en formato JSON
type PedidoJson struct {
   Id string `json:"Id"`
   Estado string `json:"Estado"`
   IdCamion int32 `json:"IdCamion"`
   TipoCamion string `json:"TipoCamion"`
   Intentos int32 `json:"Intentos"`
   Valor int32 `json:"Valor"`
   TipoPedido string `json:"TipoPedido"`
}

// Struct con la informacion necesaria de los pedidos
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

// Struct para utilizar locks en las colas de pedidos
type syncQueue struct{
   queue []infoPedido
   mux sync.Mutex
}

// Struct para utilizar locks en la cola con la informacion para finanzas
type syncQueque struct{
   queue []Pedido
   mux sync.Mutex
}

// Struct para utilizar locks en el map que contiene la informacion necesaria
// para el seguimiento
type syncMap struct{
   mapa map[int32]Pedido
   mux sync.Mutex
}

var SERIE int32 = 0 // Variable para asignar los numeros de seguimiento
var prioritarios syncQueue
var normales syncQueue  // Se definen las colas y el map
var retails syncQueue
var estados syncMap
var pedidos syncQueque
func main() {
    estados.mapa = make(map[int32]Pedido)
    lis, err := net.Listen("tcp", ":9000") // Se escucha en el puerto 9000 para que cliente realice pedidos
    if err != nil {
        log.Fatalf("Fallo al escuchar en el puerto 9000: %v", err)
    }

    grpcServer := grpc.NewServer()

    pb.RegisterPeticionServer(grpcServer, &productoServer{})

    // Se intenta abrir el archivo
    var path = "registro.csv"
    file, errf := os.OpenFile(path, os.O_APPEND|os.O_WRONLY,os.ModeAppend)
    if errf != nil { // Si no existe, se crea
            file, errf = os.Create(path)
            if errf != nil {
                log.Fatalf("Error al crear los registros csv")
            }
            writer := csv.NewWriter(file)
            tags := []string {"Fecha","Id-Paquete","Tipo","Nombre","Valor","Origen","Destino","Seguimiento"}
            _ = writer.Write(tags)
            writer.Flush()
    }
    file.Close()
    go func() {  // Se escucha en el puerto 9001 para que cliente pida seguimiento
        lis2, erk := net.Listen("tcp", ":9001")
        if erk != nil {
                log.Fatalf("Murio el puerto 9001")
        }
        grpcServer2 := grpc.NewServer()
        pb.RegisterPeticionServer(grpcServer2, &productoServer{})
        if ermm := grpcServer.Serve(lis2); ermm !=nil{
                log.Fatalf("Murio Grpc del puerto 9001")
        }
	}()
	
	go func() { // Se escucha en el puerto 9002 para el camion 1
        lis2, erk := net.Listen("tcp", ":9002")
        if erk != nil {
                log.Fatalf("Murio el puerto 9002")
        }
        grpcServer2 := grpc.NewServer()
        pb.RegisterPeticionServer(grpcServer2, &productoServer{})
        if ermm := grpcServer.Serve(lis2); ermm !=nil{
                log.Fatalf("Murio Grpc del puerto 9002")
        }
    }()

    go func() { // Se escucha en el puerto 9002 para el camion 2
        lis2, erk := net.Listen("tcp", ":9003")
        if erk != nil {
                log.Fatalf("Murio el puerto 9003")
        }
        grpcServer2 := grpc.NewServer()
        pb.RegisterPeticionServer(grpcServer2, &productoServer{})
        if ermm := grpcServer.Serve(lis2); ermm !=nil{
                log.Fatalf("Murio Grpc del puerto 9003")
        }
    }()

    go func() { // Se escucha en el puerto 9002 para el camion 3
        lis2, erk := net.Listen("tcp", ":9004")
        if erk != nil {
                log.Fatalf("Murio el puerto 9004")
        }
        grpcServer2 := grpc.NewServer()
        pb.RegisterPeticionServer(grpcServer2, &productoServer{})
        if ermm := grpcServer.Serve(lis2); ermm !=nil{
                log.Fatalf("Murio Grpc del puerto 9004")
        }
    }()
    go func(){
        conn, err := amqp.Dial("amqp://test:test@10.6.40.220:5672") // Se crea la conexion
        if err != nil{                                              // asincronica con finanzas
            log.Fatalf("Fallo al conectar RabbitMQ")
        }
        ch, erri := conn.Channel() // Se crea el canal
        if erri != nil {
            log.Fatalf("No se pudo crear un canal para RABBITMQ")
        }
        q , errk := ch.QueueDeclare( // Se declara la cola para el envio de mensajes
                "finanzas",
                false,
                false,
                false,
                false,
                nil,
        )
        if errk != nil {
            log.Fatalf("Error al declarar la cola : %s", errk)
        }
        for {
            for len(pedidos.queue) == 0 { // Se espera a que haya un registro que enviar
                time.Sleep(time.Second)
            }
            pedidos.mux.Lock()
            elemento := pedidos.queue[0]
            pedidos.queue = pedidos.queue[1:]
            pedidos.mux.Unlock()
            jsonElemento := &PedidoJson{ // Se crea el JSON
				Id : elemento.Id,
				Estado : elemento.estado,
                IdCamion : elemento.IdCamion,
                TipoCamion : elemento.tipoCamion,
                Intentos : elemento.intentos,
                Valor: elemento.valor,
                TipoPedido : elemento.tipoPedido,
            }
            marshaleao, _ := json.Marshal(jsonElemento)
            err = ch.Publish( // Se envia el registro
                "",
                q.Name,
                false,
                false,
                amqp.Publishing{
                        ContentType : "application/json",
                        Body: marshaleao,
                })
        }

    }()
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("Failed to serve gRPC server over port 9000: %v", err)
    }

}