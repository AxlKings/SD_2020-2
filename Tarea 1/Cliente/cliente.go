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
//Slice que contendra todos los numeros de seguimiento de este cliente
var series []int32 = make([]int32, 0)
func main(){
    //Se pide el tipo de cliente a ejecutar
    var cliente int32
    var archivo string
    fmt.Println("Seleccione el tipo de cliente: ")
    fmt.Println("1.- Pyme")
    fmt.Println("2.- Retail")
    fmt.Scanln(&cliente)
    //Se pide el nombre del archivo CSV que contendran los producos
    fmt.Println("Ingrese el nombre del archivo CSV: ")
    fmt.Scanln(&archivo)
    //Rutina de go que se encargara de los pedidos
    go func (cliente int32, archivo string) {
        //Se pide el delay entre entegas
        var delay int64
        fmt.Println("Seleccione la cantidad de segundos entre pedido:  ")
        fmt.Scanln(&delay)
        //Se realiza la conexion grpc con logistica
        var conn *grpc.ClientConn
        conn, errk := grpc.Dial("10.6.40.218:9000", grpc.WithInsecure())
        if errk != nil {
            log.Fatalf("No se pudo conectar: %S", errk)
        }
        defer conn.Close()
        //Se crea un objetos con los metodos de logistica
        c := pb.NewPeticionClient(conn)
        //Se abre el archivo CSV
        csvFile, err := os.Open(archivo)
        if err != nil {
            fmt.Println(err)
        }
        fmt.Println("Successfully Opened CSV file")
        defer csvFile.Close()
        //Se define el puntero que leera el archivo CSV
        reader := csv.NewReader(csvFile)
        if _, err := reader.Read(); err != nil {
             panic(err)
        }
        //Se leen todas las lineas del archivo
        productos, er := reader.ReadAll()
        if er != nil {
            fmt.Println(er)
        }
        //Se itera entre todos los productos del archivo
        for _, line := range productos {
            //Se mapean los objetos del archivo csv mensaje de protocol buffer
            valo, _ := strconv.Atoi(line[2])
            valoo := int32(valo)
			var producto pb.Objeto
			if(cliente == 1){
                priori, _ := strconv.Atoi(line[5])
                priorii := int32(priori)
                producto = pb.Objeto {
                    Id: line[0],
                    Producto: line[1],
                    Valor: valoo,
                    Tienda: line[3],
                    Destino: line[4],
                    Prioritario: priorii,
                }
            } else{
                producto = pb.Objeto {
                        Id: line[0],
                        Producto: line[1],
                        Valor: valoo,
                        Tienda: line[3],
                        Destino: line[4],
                        Prioritario: -1,
                }
            }
            //Se ejecuta un metodo de logistica entregandole un mensaje de protocol buffer
            response, errr:= c.RealizarPeticion(context.Background(), &producto)
            if errr != nil{
                log.Fatalf("Error")
            }
            if(cliente == 1){
                log.Printf("Se ha recibido el numero de seguimiento %v para el paquete con id %v ", response.Serie, producto.Id)
            } else{
                log.Printf("Se ha enviado el paquete con id %v ", producto.Id)
            }
            //Se actualizan los numeros de seguimientos y se ejecuta el delay entre entregas
            series = append(series, int32(response.Serie))
            time.Sleep(time.Duration(delay) * time.Second)
        }
    }(cliente, archivo)
    //Loop que evita poder pedir numeros de seguimiento al servidor sin haber recibido uno
    for len(series) == 0 {
        time.Sleep(time.Second)
    }
    //Se realiza una conexion con logistica
    var conn2 *grpc.ClientConn
    conn2, errq := grpc.Dial("10.6.40.218:9001", grpc.WithInsecure())
    if errq != nil {
        log.Fatalf("could not connect: %S", errq)
    }
    defer conn2.Close()
    // Si el cliente es retail no puede enviar codigos de seguimiento, caso contrario si puede.
    if(cliente == 1){
        //Se crea objeto con los metodos de logistica
        s := pb.NewPeticionClient(conn2)
        var seguimiento int32
        var eleccion int32
        //Se le brindan dos opciones al usuario, meter los codigos de seguimiento al azar y de forma automatica o a mano.
        log.Printf("Puede elegir entre pedir los estados de codigo de conocimiento ingresando el numero o puede hacerlo de forma automatica con codigos al azar: ")
        log.Printf("1.- Automatico y al azar")
        log.Printf("2.- A mano")
        fmt.Scanln(&eleccion)
        if eleccion == 1 {
            for {
                //Se eligue un codigo al azar y se pide a logistica cada 4 segundos
                seguimiento = int32(rand.Intn(len(series)))
                log.Printf("Se pedira el estado del número de seguimiento: %v", series[seguimiento])
                estado, error := s.PedirEstado(context.Background(), &pb.Serie{Serie: series[seguimiento]})
                if error != nil {
					log.Fatalf("Mensaje no enviado")
                }
                log.Printf("Se ha recibido el estado %v para el número de seguimiento %v", estado.Estado,series[seguimiento])
                time.Sleep(time.Duration(4)*time.Second)
            }
        } else {

            for {
                // Se pide un codigo de seguimiento por pantalla y se envia a logistica
                log.Printf("Seleccione un codigo de seguimiento: ")
                fmt.Scanln(&seguimiento)
                estado, error := s.PedirEstado(context.Background(), &pb.Serie{Serie : seguimiento})
                if error != nil {
                    log.Fatalf("Falla al recibir estado")
                }
                log.Printf("Se ha recibido el estado %v para el numero de seguimiento %v", estado.Estado, seguimiento)
            }
        }
    } else{
        //Loop utilizado para que el codigo se ejecute si es que el cliente es retail
        for {
            time.Sleep(time.Duration(4000)*time.Second)
        }
    }
}