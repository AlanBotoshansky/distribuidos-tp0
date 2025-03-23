# TP0: Docker + Comunicaciones + Concurrencia

En el presente repositorio se provee un esqueleto básico de cliente/servidor, en donde todas las dependencias del mismo se encuentran encapsuladas en containers. Los alumnos deberán resolver una guía de ejercicios incrementales, teniendo en cuenta las condiciones de entrega descritas al final de este enunciado.

 El cliente (Golang) y el servidor (Python) fueron desarrollados en diferentes lenguajes simplemente para mostrar cómo dos lenguajes de programación pueden convivir en el mismo proyecto con la ayuda de containers, en este caso utilizando [Docker Compose](https://docs.docker.com/compose/).

## Instrucciones de uso
El repositorio cuenta con un **Makefile** que incluye distintos comandos en forma de targets. Los targets se ejecutan mediante la invocación de:  **make \<target\>**. Los target imprescindibles para iniciar y detener el sistema son **docker-compose-up** y **docker-compose-down**, siendo los restantes targets de utilidad para el proceso de depuración.

Los targets disponibles son:

| target  | accion  |
|---|---|
|  `docker-compose-up`  | Inicializa el ambiente de desarrollo. Construye las imágenes del cliente y el servidor, inicializa los recursos a utilizar (volúmenes, redes, etc) e inicia los propios containers. |
| `docker-compose-down`  | Ejecuta `docker-compose stop` para detener los containers asociados al compose y luego  `docker-compose down` para destruir todos los recursos asociados al proyecto que fueron inicializados. Se recomienda ejecutar este comando al finalizar cada ejecución para evitar que el disco de la máquina host se llene de versiones de desarrollo y recursos sin liberar. |
|  `docker-compose-logs` | Permite ver los logs actuales del proyecto. Acompañar con `grep` para lograr ver mensajes de una aplicación específica dentro del compose. |
| `docker-image`  | Construye las imágenes a ser utilizadas tanto en el servidor como en el cliente. Este target es utilizado por **docker-compose-up**, por lo cual se lo puede utilizar para probar nuevos cambios en las imágenes antes de arrancar el proyecto. |
| `build` | Compila la aplicación cliente para ejecución en el _host_ en lugar de en Docker. De este modo la compilación es mucho más veloz, pero requiere contar con todo el entorno de Golang y Python instalados en la máquina _host_. |

### Servidor

Se trata de un "echo server", en donde los mensajes recibidos por el cliente se responden inmediatamente y sin alterar. 

Se ejecutan en bucle las siguientes etapas:

1. Servidor acepta una nueva conexión.
2. Servidor recibe mensaje del cliente y procede a responder el mismo.
3. Servidor desconecta al cliente.
4. Servidor retorna al paso 1.


### Cliente
 se conecta reiteradas veces al servidor y envía mensajes de la siguiente forma:
 
1. Cliente se conecta al servidor.
2. Cliente genera mensaje incremental.
3. Cliente envía mensaje al servidor y espera mensaje de respuesta.
4. Servidor responde al mensaje.
5. Servidor desconecta al cliente.
6. Cliente verifica si aún debe enviar un mensaje y si es así, vuelve al paso 2.

### Ejemplo

Al ejecutar el comando `make docker-compose-up`  y luego  `make docker-compose-logs`, se observan los siguientes logs:

```
client1  | 2024-08-21 22:11:15 INFO     action: config | result: success | client_id: 1 | server_address: server:12345 | loop_amount: 5 | loop_period: 5s | log_level: DEBUG
client1  | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:14 DEBUG    action: config | result: success | port: 12345 | listen_backlog: 5 | logging_level: DEBUG
server   | 2024-08-21 22:11:14 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:15 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°1
server   | 2024-08-21 22:11:15 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:20 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:20 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°2
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°3
client1  | 2024-08-21 22:11:25 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°3
server   | 2024-08-21 22:11:25 INFO     action: accept_connections | result: in_progress
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:30 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:30 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°4
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: success | ip: 172.25.125.3
server   | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | ip: 172.25.125.3 | msg: [CLIENT 1] Message N°5
client1  | 2024-08-21 22:11:35 INFO     action: receive_message | result: success | client_id: 1 | msg: [CLIENT 1] Message N°5
server   | 2024-08-21 22:11:35 INFO     action: accept_connections | result: in_progress
client1  | 2024-08-21 22:11:40 INFO     action: loop_finished | result: success | client_id: 1
client1 exited with code 0
```


## Parte 1: Introducción a Docker
En esta primera parte del trabajo práctico se plantean una serie de ejercicios que sirven para introducir las herramientas básicas de Docker que se utilizarán a lo largo de la materia. El entendimiento de las mismas será crucial para el desarrollo de los próximos TPs.

### Ejercicio N°1:
Definir un script de bash `generar-compose.sh` que permita crear una definición de Docker Compose con una cantidad configurable de clientes.  El nombre de los containers deberá seguir el formato propuesto: client1, client2, client3, etc. 

El script deberá ubicarse en la raíz del proyecto y recibirá por parámetro el nombre del archivo de salida y la cantidad de clientes esperados:

`./generar-compose.sh docker-compose-dev.yaml 5`

Considerar que en el contenido del script pueden invocar un subscript de Go o Python:

```
#!/bin/bash
echo "Nombre del archivo de salida: $1"
echo "Cantidad de clientes: $2"
python3 mi-generador.py $1 $2
```

En el archivo de Docker Compose de salida se pueden definir volúmenes, variables de entorno y redes con libertad, pero recordar actualizar este script cuando se modifiquen tales definiciones en los sucesivos ejercicios.

### Ejercicio N°2:
Modificar el cliente y el servidor para lograr que realizar cambios en el archivo de configuración no requiera reconstruír las imágenes de Docker para que los mismos sean efectivos. La configuración a través del archivo correspondiente (`config.ini` y `config.yaml`, dependiendo de la aplicación) debe ser inyectada en el container y persistida por fuera de la imagen (hint: `docker volumes`).


### Ejercicio N°3:
Crear un script de bash `validar-echo-server.sh` que permita verificar el correcto funcionamiento del servidor utilizando el comando `netcat` para interactuar con el mismo. Dado que el servidor es un echo server, se debe enviar un mensaje al servidor y esperar recibir el mismo mensaje enviado.

En caso de que la validación sea exitosa imprimir: `action: test_echo_server | result: success`, de lo contrario imprimir:`action: test_echo_server | result: fail`.

El script deberá ubicarse en la raíz del proyecto. Netcat no debe ser instalado en la máquina _host_ y no se pueden exponer puertos del servidor para realizar la comunicación (hint: `docker network`). `


### Ejercicio N°4:
Modificar servidor y cliente para que ambos sistemas terminen de forma _graceful_ al recibir la signal SIGTERM. Terminar la aplicación de forma _graceful_ implica que todos los _file descriptors_ (entre los que se encuentran archivos, sockets, threads y procesos) deben cerrarse correctamente antes que el thread de la aplicación principal muera. Loguear mensajes en el cierre de cada recurso (hint: Verificar que hace el flag `-t` utilizado en el comando `docker compose down`).

## Parte 2: Repaso de Comunicaciones

Las secciones de repaso del trabajo práctico plantean un caso de uso denominado **Lotería Nacional**. Para la resolución de las mismas deberá utilizarse como base el código fuente provisto en la primera parte, con las modificaciones agregadas en el ejercicio 4.

### Ejercicio N°5:
Modificar la lógica de negocio tanto de los clientes como del servidor para nuestro nuevo caso de uso.

#### Cliente
Emulará a una _agencia de quiniela_ que participa del proyecto. Existen 5 agencias. Deberán recibir como variables de entorno los campos que representan la apuesta de una persona: nombre, apellido, DNI, nacimiento, numero apostado (en adelante 'número'). Ej.: `NOMBRE=Santiago Lionel`, `APELLIDO=Lorca`, `DOCUMENTO=30904465`, `NACIMIENTO=1999-03-17` y `NUMERO=7574` respectivamente.

Los campos deben enviarse al servidor para dejar registro de la apuesta. Al recibir la confirmación del servidor se debe imprimir por log: `action: apuesta_enviada | result: success | dni: ${DNI} | numero: ${NUMERO}`.



#### Servidor
Emulará a la _central de Lotería Nacional_. Deberá recibir los campos de la cada apuesta desde los clientes y almacenar la información mediante la función `store_bet(...)` para control futuro de ganadores. La función `store_bet(...)` es provista por la cátedra y no podrá ser modificada por el alumno.
Al persistir se debe imprimir por log: `action: apuesta_almacenada | result: success | dni: ${DNI} | numero: ${NUMERO}`.

#### Comunicación:
Se deberá implementar un módulo de comunicación entre el cliente y el servidor donde se maneje el envío y la recepción de los paquetes, el cual se espera que contemple:
* Definición de un protocolo para el envío de los mensajes.
* Serialización de los datos.
* Correcta separación de responsabilidades entre modelo de dominio y capa de comunicación.
* Correcto empleo de sockets, incluyendo manejo de errores y evitando los fenómenos conocidos como [_short read y short write_](https://cs61.seas.harvard.edu/site/2018/FileDescriptors/).


### Ejercicio N°6:
Modificar los clientes para que envíen varias apuestas a la vez (modalidad conocida como procesamiento por _chunks_ o _batchs_). 
Los _batchs_ permiten que el cliente registre varias apuestas en una misma consulta, acortando tiempos de transmisión y procesamiento.

La información de cada agencia será simulada por la ingesta de su archivo numerado correspondiente, provisto por la cátedra dentro de `.data/datasets.zip`.
Los archivos deberán ser inyectados en los containers correspondientes y persistido por fuera de la imagen (hint: `docker volumes`), manteniendo la convencion de que el cliente N utilizara el archivo de apuestas `.data/agency-{N}.csv` .

En el servidor, si todas las apuestas del *batch* fueron procesadas correctamente, imprimir por log: `action: apuesta_recibida | result: success | cantidad: ${CANTIDAD_DE_APUESTAS}`. En caso de detectar un error con alguna de las apuestas, debe responder con un código de error a elección e imprimir: `action: apuesta_recibida | result: fail | cantidad: ${CANTIDAD_DE_APUESTAS}`.

La cantidad máxima de apuestas dentro de cada _batch_ debe ser configurable desde config.yaml. Respetar la clave `batch: maxAmount`, pero modificar el valor por defecto de modo tal que los paquetes no excedan los 8kB. 

Por su parte, el servidor deberá responder con éxito solamente si todas las apuestas del _batch_ fueron procesadas correctamente.

### Ejercicio N°7:

Modificar los clientes para que notifiquen al servidor al finalizar con el envío de todas las apuestas y así proceder con el sorteo.
Inmediatamente después de la notificacion, los clientes consultarán la lista de ganadores del sorteo correspondientes a su agencia.
Una vez el cliente obtenga los resultados, deberá imprimir por log: `action: consulta_ganadores | result: success | cant_ganadores: ${CANT}`.

El servidor deberá esperar la notificación de las 5 agencias para considerar que se realizó el sorteo e imprimir por log: `action: sorteo | result: success`.
Luego de este evento, podrá verificar cada apuesta con las funciones `load_bets(...)` y `has_won(...)` y retornar los DNI de los ganadores de la agencia en cuestión. Antes del sorteo no se podrán responder consultas por la lista de ganadores con información parcial.

Las funciones `load_bets(...)` y `has_won(...)` son provistas por la cátedra y no podrán ser modificadas por el alumno.

No es correcto realizar un broadcast de todos los ganadores hacia todas las agencias, se espera que se informen los DNIs ganadores que correspondan a cada una de ellas.

## Parte 3: Repaso de Concurrencia
En este ejercicio es importante considerar los mecanismos de sincronización a utilizar para el correcto funcionamiento de la persistencia.

### Ejercicio N°8:

Modificar el servidor para que permita aceptar conexiones y procesar mensajes en paralelo. En caso de que el alumno implemente el servidor en Python utilizando _multithreading_,  deberán tenerse en cuenta las [limitaciones propias del lenguaje](https://wiki.python.org/moin/GlobalInterpreterLock).

## Condiciones de Entrega
Se espera que los alumnos realicen un _fork_ del presente repositorio para el desarrollo de los ejercicios y que aprovechen el esqueleto provisto tanto (o tan poco) como consideren necesario.

Cada ejercicio deberá resolverse en una rama independiente con nombres siguiendo el formato `ej${Nro de ejercicio}`. Se permite agregar commits en cualquier órden, así como crear una rama a partir de otra, pero al momento de la entrega deberán existir 8 ramas llamadas: ej1, ej2, ..., ej7, ej8.
 (hint: verificar listado de ramas y últimos commits con `git ls-remote`)

Se espera que se redacte una sección del README en donde se indique cómo ejecutar cada ejercicio y se detallen los aspectos más importantes de la solución provista, como ser el protocolo de comunicación implementado (Parte 2) y los mecanismos de sincronización utilizados (Parte 3).

Se proveen [pruebas automáticas](https://github.com/7574-sistemas-distribuidos/tp0-tests) de caja negra. Se exige que la resolución de los ejercicios pase tales pruebas, o en su defecto que las discrepancias sean justificadas y discutidas con los docentes antes del día de la entrega. El incumplimiento de las pruebas es condición de desaprobación, pero su cumplimiento no es suficiente para la aprobación. Respetar las entradas de log planteadas en los ejercicios, pues son las que se chequean en cada uno de los tests.

La corrección personal tendrá en cuenta la calidad del código entregado y casos de error posibles, se manifiesten o no durante la ejecución del trabajo práctico. Se pide a los alumnos leer atentamente y **tener en cuenta** los criterios de corrección informados  [en el campus](https://campusgrado.fi.uba.ar/mod/page/view.php?id=73393).

## Resolución de los Ejercicios

### Ejercicio N°1:

#### Solución implementada
Se creó el script `generar-compose.sh` que permite generar un archivo de Docker Compose con una cantidad configurable de clientes. El script:

1. Recibe dos parámetros: nombre del archivo de salida y cantidad de clientes
2. Genera un encabezado para el archivo de Docker Compose que incluye el servicio del servidor
3. Genera las definiciones de los servicios para cada cliente (client1, client2, etc.)
4. Agrega la configuración de red necesaria para la comunicación

El script garantiza que:
- Se mantenga la convención de nombres (client1, client2, etc.)
- Los clientes se conecten a la misma red que el servidor
- Se establezcan las dependencias adecuadas (los clientes dependen del servidor)
- Cada cliente tenga un ID único configurado como variable de entorno

#### Cómo ejecutar
Para generar un archivo Docker Compose con 5 clientes:

```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

### Ejercicio N°2:

#### Solución implementada
Se modificó el script `generar-compose.sh` para incluir volúmenes que permiten inyectar los archivos de configuración en los contenedores sin necesidad de reconstruir las imágenes. 

Para ello se implementaron los siguientes volúmenes:
- Para el servidor: `./server/config.ini:/config.ini`
- Para cada cliente: `./client/config.yaml:/config.yaml`

Además, se agregó un .dockerignore para asegurar que los archivos de configuración no se incluyan en las imágenes.

### Ejercicio N°3:

#### Solución implementada
Se creó el script `validar-echo-server.sh` que permite probar el correcto funcionamiento del echo server utilizando `netcat` desde un contenedor Alpine. El script:

1. Define un mensaje de prueba que será enviado al servidor
2. Ejecuta un contenedor temporal Alpine en la misma red del servidor (`tp0_testing_net`)
3. Dentro del contenedor, utiliza `netcat` para:
   - Enviar el mensaje de prueba al servidor en el puerto 12345
   - Capturar la respuesta del servidor
4. Compara la respuesta con el mensaje original
   - Si coinciden, reporta éxito
   - Si no coinciden o hay un error, reporta fallo

#### Cómo ejecutar
Para validar el funcionamiento del echo server:

```bash
./validar-echo-server.sh
```

Si el servidor está funcionando correctamente, el script mostrará:
```
action: test_echo_server | result: success
```

En caso de error, mostrará:
```
action: test_echo_server | result: fail
```

### Ejercicio N°4:

#### Solución implementada
Se modificaron tanto el servidor como el cliente para implementar una terminación graceful cuando reciben la señal SIGTERM. Los cambios incluyeron:

##### Servidor:
1. Implementación de un manejador de señales que captura SIGTERM
2. Mecanismo para solicitar la finalización ordenada mediante un flag `_shutdown_requested`
3. Bucle principal que verifica periódicamente dicho flag y termina cuando está activo
4. Función de limpieza (`__cleanup`) que:
   - Cierra adecuadamente el socket del servidor
   - Registra mensajes de log sobre el cierre del socket

##### Cliente:
1. Implementación de un manejador de señales en una goroutine separada
2. Método `Shutdown()` que:
   - Cierra el canal de señalización para detener el envío de mensajes
   - Cierra apropiadamente la conexión de red si está activa
   - Registra mensajes de log sobre el cierre de la conexión

#### Cómo ejecutar
Para probar el cierre graceful:

1. Iniciar los contenedores:
```bash
make docker-compose-up
```

2. Envíar una señal SIGTERM a uno de los contenedores:
```bash
# Para el servidor
docker stop server

# Para un cliente específico
docker stop client1
```

3. Verificar los logs para confirmar el cierre ordenado:
```bash
make docker-compose-logs
```

Se deberían ver mensajes como:
```
server | action: signal_received | result: success | signal: SIGTERM
server | action: shutting_down | result: in_progress
server | action: close_server_socket | result: success
server | action: shutting_down | result: success
```

```
client1 | action: client_shutdown | result: in_progress | client_id: 1
client1 | action: close_connection | result: in_progress | client_id: 1
client1 | action: close_connection | result: success | client_id: 1
client1 | action: client_shutdown | result: success | client_id: 1
```

Cabe aclarar que en el caso del cliente, los logs de "close_connection" pueden no aparecer si el cliente no estaba conectado al momento de recibir la señal SIGTERM.

### Ejercicio N°5:

#### Solución implementada
Se modificó la lógica de negocio del cliente y servidor para implementar el caso de uso de "Lotería Nacional". Los cambios incluyeron:

##### Cliente:
1. Modificación del cliente para leer datos de apuesta desde variables de entorno:
   - Nombre, apellido, documento, nacimiento y número
2. Implementación del envío de la apuesta al servidor
3. Recepción de confirmación del servidor y registro en logs

##### Servidor:
1. Adaptación del servidor para recibir y procesar apuestas
2. Almacenamiento de la información mediante la función `store_bets(...)` provista
3. Envío de confirmaciones al cliente
4. Registro de logs de acuerdo al formato requerido

#### Protocolo de comunicación implementado
Se diseñó un protocolo binario con las siguientes características:

1. **Estructura de paquetes**:
   - **Header** (3 bytes): 
     - Tamaño del mensaje (2 bytes): Indica la longitud del payload
     - Tipo de mensaje (1 byte): Identifica el tipo de mensaje (BET=1, BET_CONFIRMATION=2)
   - **Payload**: Contenido específico según el tipo de mensaje

2. **Tipos de mensajes**:
   - **BET**: Envío de apuesta
     - ID de agencia (4 bytes)
     - Longitud del nombre (1 byte) + Nombre (variable)
     - Longitud del apellido (1 byte) + Apellido (variable)
     - DNI (8 bytes, string fijo)
     - Fecha de nacimiento (10 bytes, string fijo en formato yyyy-mm-dd)
     - Número apostado (4 bytes)
   - **BET_CONFIRMATION**: Confirmación de apuesta
     - Resultado (1 byte): OK=0, ERROR=1

3. **Manejo de errores**:
   - Detección de desconexiones durante envío/recepción
   - Manejo de errores de deserialización

4. **Prevención de problemas**:
   - Solución al problema de *short reads*: Lectura en bucle hasta completar el paquete
   - Solución al problema de *short writes*: Envío en bucle hasta completar la transmisión
   - Validación de tipos de mensajes

#### Cómo ejecutar
Para ejecutar el ejercicio con 5 agencias:

1. Generar el archivo docker-compose con las agencias:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

2. Iniciar los contenedores:
```bash
make docker-compose-up
```

3. Verificar los logs para confirmar el envío y almacenamiento de apuestas:
```bash
make docker-compose-logs
```

Se deberían ver mensajes como:
```
client1 | action: apuesta_enviada | result: success | dni: 30000001 | numero: 1001
server  | action: apuesta_almacenada | result: success | dni: 30000001 | numero: 1001
```

### Ejercicio N°6:

#### Solución implementada
Se modificaron los clientes para procesar y enviar múltiples apuestas a la vez en formato batch, acortando tiempos de transmisión y procesamiento. Los cambios incluyeron:

##### Cliente:
1. Lectura de archivos CSV para cargar las apuestas de cada agencia:
   - Implementación de la función `ReadBetsFromCSV()` que lee y procesa el archivo de apuestas
   - Montaje del archivo CSV correspondiente a cada agencia como un volumen
2. Envío de apuestas batchs:
   - Configuración del tamaño máximo de batch mediante el parámetro `batch.maxAmount` en `config.yaml`
   - Procesamiento de todas las apuestas del archivo, dividiéndolas en batches según el tamaño configurado
3. Control periódico para detectar señales de shutdown durante la lectura del archivo

##### Servidor:
1. Adaptación para recibir y procesar múltiples apuestas en un mismo mensaje:
   - Modificación del protocolo para soportar múltiples registros en un mismo mensaje
   - Respuesta única para todo el batch indicando éxito o error
2. Implementación del log requerido que indica la cantidad de apuestas procesadas

##### Modificacion al script `generar-compose.sh`:
Agregado de volúmenes para montar los archivos CSV de cada agencia:
```bash
- ./.data/agency-$i.csv:/.data/agency-$i.csv
```

#### Cómo ejecutar

1. Asegurarse de tener los archivos CSV de apuestas en el directorio `.data/`:

2. Generar el archivo docker-compose con 5 agencias:
```bash
./generar-compose.sh docker-compose-dev.yaml 5
```

3. Iniciar los contenedores:
```bash
make docker-compose-up
```

4. Verificar los logs para confirmar el procesamiento por batchs:
```bash
make docker-compose-logs
```

Los logs deberían mostrar mensajes como:
```
client1 | action: apuesta_enviada | result: success | client_id: 1
server  | action: apuesta_recibida | result: success | cantidad: 140
```