Sebastian Campos 201773517-1
Axel Reyes 201773502-3

Para que funcione debe estar ejecutadose la mv dist78 que simula logistica.

Para simular cliente, primero debe entrar a la carpeta "cliente". Dentro de esta carpeta deben haber archivos csv que contengan los productos que enviara el cliente. El formato de esos archivos deben estar de la misma forma que los archivos de prueba entregados por los ayudantes.

Al comenzar el programa, debe elegir el tipo de cliente que sera pyme o retail). Despues, se le pedira el nombree del archivo donde estaran los productos, el nombre debe incluir la extension (ej: "pyme.csv"). Luego, se le dara dos opciones para consultar codigos de seguimiento, automatica (la cual elgira un codigo de seguimiento entregaado al cliente al azar y se lo enviara al servidor) o "a mano" (podra ingresar el codigo que desee, si el codigo no existe recibira una respuesta vacia).

Para evitar problemas, si quiere ejecutar mas de dos clientes a la vez, estos deben trabajar con archivos csv distintos.

Para correr el codigo, ejecute make dentro de la carpeta cliente.

Hay una posibilidad de que la conexion no se realice correctamente. En ese caso, muy probablemente se deba al firewall por lo que deberá desactivarlo con el siguiente comando:

sudo systemctl stop firewalld

Supuestos
- Se asume que tanto pyme como retail intentan max 3 veces en caso de que no hayan ganancias, ya que en el enunciado dice: "Retail intenta siempre max 3 veces entregar" y "Pyme Máximo 2 reintentos o hasta que no haya gananci"", siendo que reintentos = intentos -1.