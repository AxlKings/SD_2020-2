Sebastian Campos 201773517-1
Axel Reyes 201773502-3

Para simular servidor, primero debe entrar a la carpeta "servidor".

Se pusieron diversos logs para indicar lo que realiza el servidor en todo momento. El servidor al iniciar crea un archivo registros.csv, si el archivo ya existe, los registros se agregaran al archivo existente.

Para correr el codigo, ejecute make dentro de la carpeta servidor.

Hay una posibilidad de que la conexion no se realice correctamente. En ese caso, muy probablemente se deba al firewall por lo que deberá des activarlo con el siguiente comando:

sudo systemctl stop firewalld

Supuestos
- Se asume que tanto pyme como retail intentan max 3 veces en caso de que no hayan ganancias, ya que en el enunciado dice: "Retail intenta siempre max 3 veces entregar" y "Pyme Máximo 2 reintentos o hasta que no haya gananciaa", siendo que reintentos = intentos -1.