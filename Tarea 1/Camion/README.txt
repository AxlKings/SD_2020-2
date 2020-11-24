Sebastian Campos 201773517-1
Axel Reyes 201773502-3

Para que funcione debe estar ejecutandose la mv dist78 que simula logistica.

Para simular camiones, primero debe entrar a la carpeta "camion".

Al comenzar el programa, debe elegir el tiempo de espera por el segundo paquete, luego se pedira el tiempo que se demoran los camiones en realizar cada entrega.

Se pusieron diversos logs para indicar lo que realizan los camiones. Los camiones al iniciar crearan un archivo.csv que tiene como nombre el id del camion y tipo de camion, si el archivo ya existe, los registros se agregaran al archivo existente.

Para correr el codigo, ejecute make dentro de la carpeta camion.

Hay una posibilidad de que la conexion no se realice correctamente. En ese caso, muy probablemente se deba al firewall por lo que deberá des activarlo con el siguiente comando:

sudo systemctl stop firewalld

Supuestos
- Se asume que tanto pyme como retail intentan max 3 veces en caso de que no hayan ganancias, ya que en el enunciado dice: "Retail intenta siempre max 3 veces entregar" y "Pyme Máximo 2 reintentos o hasta que no haya gananciaa", siendo que reintentos = intentos -1.