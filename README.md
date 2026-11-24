# Web-MyGameList

Servidor web creado por Valentina Bianchi y Maximo Billordo para la materia Programación Web, de la carrera Ingeniería de Sistemas de la UNICEN.

## Descripción

MyGameList es una aplicación para llevar un catálogo de los juegos favoritos de una persona. Para cada juego se guarda un título, una calificación (de 1 a 5 estrellas) y un estado (Por jugar, Jugando, Completado, Platinado).

## Modelo de datos

El esquema de la base de datos (PostgreSQL) está compuesto por tres tablas:

* **Users**: usuarios registrados en la aplicación (id, nombre, email, contraseña).
* **Library**: catálogo global de juegos disponibles (id, título, descripción, calificación promedio).
* **Catalogue**: tabla intermedia que representa la lista personal de cada usuario (qué juego tiene agregado, con qué puntaje y en qué estado).

Además hay dos vistas que simplifican las consultas de la interfaz:

* **UserInfo**: los datos de un usuario que se pueden mostrar (id, nombre y email).
* **CatalogueInfo**: cada registro del catálogo junto con el título del juego.

La calificación promedio de cada juego (`avg_score`) la mantiene un trigger de la base, que la recalcula cada vez que cambia un puntaje en `Catalogue`.

El esquema completo se encuentra en `db/schema/schema.sql`, y las consultas SQL utilizadas para generar el código de acceso a datos están en `db/queries/queries.sql`.
### Modelo relacional
![diagrama](diagrama.png)

## Requisitos

* Si no está instalado, usar el siguiente comando para instalar docker-compose: `sudo apt install docker-compose-v2`
* Verificar la instalación con: `docker compose version`

## Cómo ejecutar el servidor

Todo corre dentro de contenedores Docker (servidor y PostgreSQL), sin instalar Go ni PostgreSQL en la máquina host. Desde la raíz del proyecto:

```bash
cd servidor-web
make run
```

Esto construye la imagen del servidor, levanta PostgreSQL con el esquema de `db/schema/schema.sql`, genera el código con `sqlc` y `templ`, y deja el servidor escuchando en `http://localhost:8080`. El servidor se compila al arrancar el contenedor, por lo que puede tardar unos segundos en responder. Se detiene con `Ctrl+C`, y al hacerlo se eliminan los contenedores y los datos. También se puede usar `make clean`.

## Interfaz web (TP4, TP5 y TP6)

Al correr `make run` y abrir `http://localhost:8080` se muestra la **portada**, con enlaces a las páginas de juegos y de usuarios y un selector para entrar al catálogo de un usuario. Cada entidad tiene su propia página:

- **Juegos (`/library`)**: cargar un juego (título y descripción), ver la biblioteca como tarjetas con el puntaje promedio de cada juego y borrar juegos.
- **Usuarios (`/users`)**: registrar un usuario (nombre, email y contraseña), ver una tabla con nombre y email, ordenable por nombre, y borrar usuarios.
- **Catálogo (`/catalogue?user={id}`)**: el catálogo de un usuario, en tarjetas. Permite agregar un juego con su estado y su puntaje (en estrellas, opcional) y quitarlo del catálogo.

En el TP4 el frontend estaba hecho con HTML, CSS y JavaScript, y consumía la API REST del TP3. A partir del TP5 la interfaz se renderiza en el servidor: las vistas son componentes `templ` (en `views/`) y el estilo viene de Pico.css (cargado por CDN).

En el TP6 la interfaz pasó a ser reactiva con [HTMX](https://htmx.org/) 4.0, cargado por CDN en el layout. Crear o borrar un registro ya no recarga la página: los formularios y los botones envían la petición con atributos HTMX (`hx-post`, `hx-delete`, `hx-target`, `hx-swap`) y el servidor responde solo con el fragmento de HTML que cambia. La navegación entre páginas usa `hx-boost`, por lo que tampoco recarga la página completa. Sigue sin haber archivos JavaScript propios.

## Rutas

| Método y URL                         | Descripción                                                              | Respuestas                   |
| ------------------------------------ | ------------------------------------------------------------------------ | ---------------------------- |
| `GET /`                              | Portada: enlaces y selector de usuario                                   | 200, 404, 405, 500           |
| `GET /library`                       | Página de juegos                                                         | 200, 405, 500                |
| `POST /library`                      | Crea un juego y devuelve la lista actualizada (fragmento)                | 200, 400, 405, 500           |
| `DELETE /library/{id}`               | Elimina un juego y devuelve una respuesta vacía                          | 200, 400, 404, 405, 500      |
| `GET /users[?order=asc\|desc]`       | Página de usuarios; en un pedido de HTMX devuelve solo la tabla         | 200, 405, 500                |
| `POST /users`                        | Crea un usuario y devuelve la tabla actualizada (fragmento)              | 200, 400, 405, 409, 500      |
| `DELETE /users/{id}`                 | Elimina un usuario y devuelve una respuesta vacía                        | 200, 400, 404, 405, 500      |
| `GET /catalogue?user={id}`           | Página del catálogo de un usuario (redirige a `/` si falta el usuario)   | 200, 303, 404, 405, 500      |
| `POST /catalogue/{idUser}`           | Agrega un juego al catálogo y devuelve la lista actualizada (fragmento)  | 200, 400, 404, 405, 409, 500 |
| `DELETE /catalogue/{idUser}/{idGame}`| Quita un juego del catálogo y devuelve una respuesta vacía               | 200, 400, 404, 405, 500      |

Los errores de los formularios (`POST`) y de los borrados de juegos y de catálogo devuelven un fragmento HTML con el mensaje, que se muestra debajo del formulario o dentro de la tarjeta afectada. Los demás errores (por ejemplo, un id inválido en la URL) devuelven un mensaje de texto.

## Cómo correr los tests de la capa de datos

Todo el flujo de testing está automatizado con un `Makefile`, que se apoya en `docker compose` para levantar una base de datos PostgreSQL limpia, generar el código con `sqlc` y `templ`, compilar el proyecto y correr los tests, sin necesidad de tener PostgreSQL, Go ni `sqlc` instalados en la máquina host (todo corre dentro de contenedores Docker).

Para correr los tests, desde la raíz del proyecto:

```bash
cd servidor-web
make test
```

Esto ejecuta, en orden:

1. Limpieza de cualquier contenedor o volumen previo (`docker compose down -v`).
2. Construcción de la imagen donde corren los tests (`docker compose build test`).
3. Levantamiento del contenedor de PostgreSQL, esperando a que esté listo para aceptar conexiones (`healthcheck`).
4. Dentro de un contenedor: generación del código con `sqlc generate` y `templ generate`, compilación (`go build`) y ejecución de los tests (`go test ./... -v`).
5. Limpieza de los contenedores y volúmenes al finalizar, tanto si los tests pasaron como si fallaron.

## Decisiones de diseño (TP6)

- **Una página por entidad:** juegos, usuarios y catálogo tienen su propia página, y una portada reúne los accesos. Las rutas `/library` y `/users` atienden `GET` (muestran la página) y `POST` (crean) en la misma URL, con un `switch` por método.
- **Navegación con `hx-boost`:** el `<body>` lleva `hx-boost:inherited="true"`, así que los enlaces y el formulario de la portada se piden por HTMX y reemplazan el `<body>` sin recargar. Como esos pedidos también traen el header `HX-Request`, los handlers distinguen los pedidos parciales (por ejemplo, el clic en el encabezado de la tabla de usuarios) de los de navegación con `HX-Boosted`.
- **Crear:** el formulario usa `hx-post` y el handler ya no redirige: después de insertar vuelve a consultar la lista completa y devuelve solo ese componente, que HTMX coloca con `hx-swap="outerHTML"` en el contenedor de la lista.
- **Borrar:** el botón de cada elemento usa `hx-delete` con `hx-target` en la tarjeta o fila (`closest article` o `closest tr`) y `hx-swap="outerHTML"`. El servidor responde 200 con cuerpo vacío; no se usa 204 porque HTMX no intercambia contenido en esa respuesta. Se agregó `hx-confirm` para pedir confirmación, y al borrar un usuario se avisa que también se borra su catálogo.
- **Errores:** HTMX 4 intercambia también las respuestas 4xx y 5xx, por lo que un `http.Error` en texto plano reemplazaría la lista o la tarjeta. Por eso los errores que se muestran al usuario responden con el status correspondiente y un fragmento HTML (`WriteError` y el componente `ErrorMessage`), y el atributo `hx-status` lo dirige al lugar correcto: debajo del formulario, o en un contenedor propio dentro de cada tarjeta al borrar un juego o un registro del catálogo. En la tabla de usuarios, el error del borrado se muestra con un `alert` y la fila no se reemplaza (`hx-status` con `swap:none`).
- **Errores de URL o id inválido** quedan como `http.Error` en texto plano: la interfaz siempre genera URLs válidas, por lo que solo aparecerían si alguien modifica la URL a mano.
- **Formularios:** se limpian automáticamente al terminar de crear un registro correctamente (`hx-on:htmx:after:swap`); si hay un error, se conserva lo escrito.
- **Lista vacía:** el mensaje de "todavía no hay elementos" se renderiza siempre dentro del contenedor de cada lista, y una regla CSS en el layout (con `:has()`) lo oculta mientras haya alguna tarjeta o fila. Así el mensaje reaparece al borrar el último elemento sin recargar, y el servidor puede seguir respondiendo vacío al borrar. El CSS está en un `<style>` del layout, porque la aplicación no tiene un archivo `.css` propio.
- **Orden de la tabla de usuarios:** el estado del orden vive en la URL (`/users?order=asc|desc`), que el encabezado actualiza con `hx-push-url`. Al crear un usuario, el servidor lee el orden actual desde el header `HX-Current-URL` y mantiene la tabla ordenada. Las consultas ordenan por nombre y desempatan por id, para que el orden sea estable.
- **Catálogo sin login:** como no hay autenticación, el usuario se elige en la portada con un selector y la página del catálogo se identifica con `?user={id}`. El catálogo no se edita: para cambiar el estado o el puntaje de un juego se lo quita y se lo vuelve a agregar. El estado se valida en el servidor contra los cuatro valores permitidos, y el puntaje es opcional (de 1 a 5).
- **Borrados en cascada:** borrar un usuario o un juego elimina también sus registros del catálogo (`ON DELETE CASCADE`).
- **HTMX por CDN:** se usa la versión 4.0.0 con `integrity`.

## Limitaciones conocidas

- La primera vez la base puede tardar varios segundos en inicializarse.
- No hay autenticación ni autorización: cualquiera puede elegir a cualquier usuario y ver o modificar su catálogo. Las contraseñas se guardan en texto plano.
- No se pueden editar usuarios, juegos ni registros del catálogo desde la interfaz, solo crearlos y borrarlos.
- La tabla de usuarios solo se puede ordenar por nombre.
- No hay restricción de título único en `Library`.