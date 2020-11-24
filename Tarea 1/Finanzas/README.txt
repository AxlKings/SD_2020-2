Sebastian Campos 201773517-1
Axel Reyes 201773502-3

Para simular finanzas, primero debe entrar a la carpeta "finanzas".

Se pusieron diversos logs para indicar que recibe los datos de los productos entregados por logistica. Finanzas al iniciar crea un archivo registro.csv, si el archivo ya existe, los registros se agregaran al archivo existente.
Para poder ver el resumen de las entregas, debe terminar el programa (Ctrl + C).

Para correr el codigo, ejecute make dentro de la carpeta finanzas.

Si se corre el servidor por mucho tiempo antes que finanzas, puede ser que no todos los paquetes se guarden debido a que la cola tiene un limite de espacio.

Hay una posibilidad de que la conexion no se realice correctamente. En ese caso, muy probablemente se deba al firewall por lo que deberá des activarlo con el siguiente comando:

sudo systemctl stop firewalld

Supuestos
- Se asume que tanto pyme como retail intentan max 3 veces en caso de que no hayan ganancias, ya que en el enunciado dice: "Retail intenta siempre max 3 veces entregar" y "Pyme Máximo 2 reintentos o hasta que no haya gananciaa", siendo que reintentos = intentos -1.