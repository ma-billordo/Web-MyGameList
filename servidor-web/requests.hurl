#-----------
# USUARIOS
#-----------

# Crear usuario 1 - el password no debe volver en claro
POST http://localhost:8080/users
Content-Type: application/json
{
    "name": "valen",
    "email": "valen@gmail.com",
    "password": "1234"
}
HTTP 201
[Asserts]
jsonpath "$.id_user" == 1
jsonpath "$.name" == "valen"
jsonpath "$.email" == "valen@gmail.com"
jsonpath "$.password" != "1234"

# Crear usuario 2
POST http://localhost:8080/users
Content-Type: application/json
{
    "name": "max",
    "email": "max@gmail.com",
    "password": "5678"
}
HTTP 201
[Asserts]
jsonpath "$.id_user" == 2
jsonpath "$.name" == "max"
jsonpath "$.password" != "5678"

# Listar usuarios - ningún elemento expone el password
GET http://localhost:8080/users
HTTP 200
[Asserts]
jsonpath "$" count == 2
jsonpath "$[0].password" not exists

# Obtener usuario con ID 1 - el password no se muestra
GET http://localhost:8080/users/1
HTTP 200
[Asserts]
jsonpath "$.id_user" == 1
jsonpath "$.name" == "valen"
jsonpath "$.password" not exists

# --- Errores de validación en POST /users (400) ---

# Nombre vacío - debe fallar con 400
POST http://localhost:8080/users
Content-Type: application/json
{
    "name": "",
    "email": "vacio@gmail.com",
    "password": "1234"
}
HTTP 400

# Email con formato inválido - debe fallar con 400
POST http://localhost:8080/users
Content-Type: application/json
{
    "name": "Juan",
    "email": "esto-no-es-un-email",
    "password": "1234"
}
HTTP 400

# Password de menos de 4 caracteres - debe fallar con 400
POST http://localhost:8080/users
Content-Type: application/json
{
    "name": "Juan",
    "email": "juan@gmail.com",
    "password": "123"
}
HTTP 400

# Crear usuario con email duplicado - debe fallar con 409
# (el INSERT fallido consume un valor de la secuencia de ids: el próximo usuario creado ya no será el 3, por eso más abajo se captura su id)
POST http://localhost:8080/users
Content-Type: application/json
{
    "name": "Juan",
    "email": "max@gmail.com",
    "password": "1234"
}
HTTP 409

# --- Errores de id en GET /users/{id} ---

# ID no numérico - debe fallar con 400
GET http://localhost:8080/users/abc
HTTP 400

# Obtener usuario que no existe - debe fallar con 404
GET http://localhost:8080/users/9999
HTTP 404

# --- PUT /users/{id} ---

# Actualizar el usuario 2 con el email del usuario 1 - debe fallar con 409
PUT http://localhost:8080/users/2
Content-Type: application/json
{
    "name": "max",
    "email": "valen@gmail.com"
}
HTTP 409

# Actualizar el usuario 2 conservando su propio email - NO es un duplicado, debe dar 200
PUT http://localhost:8080/users/2
Content-Type: application/json
{
    "name": "max",
    "email": "max@gmail.com"
}
HTTP 200
[Asserts]
jsonpath "$.id_user" == 2
jsonpath "$.email" == "max@gmail.com"

# Actualizar usuario con ID 1 cambiando nombre y email
PUT http://localhost:8080/users/1
Content-Type: application/json
{
    "name": "valen updated",
    "email": "valen2@gmail.com"
}
HTTP 200
[Asserts]
jsonpath "$.id_user" == 1
jsonpath "$.name" == "valen updated"
jsonpath "$.email" == "valen2@gmail.com"

# Verificar que el cambio quedó guardado en la base
GET http://localhost:8080/users/1
HTTP 200
[Asserts]
jsonpath "$.name" == "valen updated"
jsonpath "$.email" == "valen2@gmail.com"

# Actualizar con nombre vacío - debe fallar con 400
PUT http://localhost:8080/users/2
Content-Type: application/json
{
    "name": "",
    "email": "max@gmail.com"
}
HTTP 400

# Actualizar con email inválido - debe fallar con 400
PUT http://localhost:8080/users/2
Content-Type: application/json
{
    "name": "max",
    "email": "esto-no-es-un-email"
}
HTTP 400

# Actualizar un usuario que no existe (con body válido) - debe fallar con 404
PUT http://localhost:8080/users/9999
Content-Type: application/json
{
    "name": "invalid",
    "email": "invalid@gmail.com"
}
HTTP 404

# --- DELETE /users/{id} ---

# Eliminar usuario con ID 1
DELETE http://localhost:8080/users/1
HTTP 204

# El usuario eliminado ya no existe - debe dar 404
GET http://localhost:8080/users/1
HTTP 404

# Método no soportado sobre la colección - debe fallar con 405
DELETE http://localhost:8080/users
HTTP 405


#-----------
# JUEGOS
#-----------

# Crear juego 1 - el promedio arranca en 0
POST http://localhost:8080/library
Content-Type: application/json
{
    "title": "Counter Strike",
    "description": "Un juego de disparos - tactical shooter"
}
HTTP 201
[Asserts]
jsonpath "$.id_game" == 1
jsonpath "$.title" == "Counter Strike"
jsonpath "$.avg_score" == "0.00"

# Crear juego 2
POST http://localhost:8080/library
Content-Type: application/json
{
    "title": "Age of Empires 2",
    "description": "Un juego de administración de recursos"
}
HTTP 201
[Asserts]
jsonpath "$.id_game" == 2

# Crear juego 3
POST http://localhost:8080/library
Content-Type: application/json
{
    "title": "Grand Theft Auto V",
    "description": "Un juego de mundo abierto"
}
HTTP 201
[Asserts]
jsonpath "$.id_game" == 3
jsonpath "$.avg_score" == "0.00"

# Listar juegos
GET http://localhost:8080/library
HTTP 200
[Asserts]
jsonpath "$" count == 3

# Obtener juego por ID
GET http://localhost:8080/library/1
HTTP 200
[Asserts]
jsonpath "$.id_game" == 1
jsonpath "$.title" == "Counter Strike"
jsonpath "$.avg_score" == "0.00"

# --- Errores de validación en POST /library (400) ---

# Crear un juego sin título - debe fallar con 400
POST http://localhost:8080/library
Content-Type: application/json
{
    "title": "",
    "description": "Una descripcion"
}
HTTP 400

# Título de solo espacios - debe fallar con 400
POST http://localhost:8080/library
Content-Type: application/json
{
    "title": "     ",
    "description": "Una descripcion"
}
HTTP 400

# Descripción vacía - debe fallar con 400
POST http://localhost:8080/library
Content-Type: application/json
{
    "title": "Juego sin descripcion",
    "description": ""
}
HTTP 400

# --- Errores de id en GET /library/{id} ---

# ID no numérico - debe fallar con 400
GET http://localhost:8080/library/abc
HTTP 400

# Obtener juego inexistente - debe fallar con 404
GET http://localhost:8080/library/9999
HTTP 404

# --- PUT /library/{id} ---

# Actualizar juego por ID - la respuesta NO incluye avg_score
PUT http://localhost:8080/library/1
Content-Type: application/json
{
    "title": "Counter Strike 2",
    "description": "Un juego de disparos en primera persona"
}
HTTP 200
[Asserts]
jsonpath "$.id_game" == 1
jsonpath "$.title" == "Counter Strike 2"
jsonpath "$.avg_score" not exists

# Verificar que el cambio quedó guardado y que el promedio se conserva
GET http://localhost:8080/library/1
HTTP 200
[Asserts]
jsonpath "$.title" == "Counter Strike 2"
jsonpath "$.avg_score" == "0.00"

# Actualizar con título vacío - debe fallar con 400
PUT http://localhost:8080/library/1
Content-Type: application/json
{
    "title": "",
    "description": "Una descripcion"
}
HTTP 400

# Actualizar con descripción vacía - debe fallar con 400
PUT http://localhost:8080/library/1
Content-Type: application/json
{
    "title": "Counter Strike 2",
    "description": ""
}
HTTP 400

# Actualizar un juego que no existe (con body válido) - debe fallar con 404
PUT http://localhost:8080/library/9999
Content-Type: application/json
{
    "title": "Invalid",
    "description": "No existe"
}
HTTP 404

# --- DELETE /library/{id} ---

# Eliminar juego
DELETE http://localhost:8080/library/2
HTTP 204

# El juego eliminado ya no existe - debe dar 404
GET http://localhost:8080/library/2
HTTP 404

# Método no soportado sobre la colección - debe fallar con 405
DELETE http://localhost:8080/library
HTTP 405

#-----------
# CATÁLOGO
# Estado: usuario 2 existe (el 1 fue eliminado); juegos 1, 3 y 4 existen (el 2 fue eliminado).
#-----------

# Crear un registro en el catálogo del usuario 2 - sin score queda en null
POST http://localhost:8080/catalogues/2
Content-Type: application/json
{
    "id_game": 1,
    "status": "Jugando"
}
HTTP 201
[Asserts]
jsonpath "$.id_user" == 2
jsonpath "$.id_game" == 1
jsonpath "$.status" == "Jugando"
jsonpath "$.score" == null

# Agregar otro juego al catálogo del usuario 2
POST http://localhost:8080/catalogues/2
Content-Type: application/json
{
    "id_game": 3,
    "status": "Por jugar"
}
HTTP 201
[Asserts]
jsonpath "$.id_user" == 2
jsonpath "$.id_game" == 3

# Listar catálogo del usuario 2
GET http://localhost:8080/catalogues/2
HTTP 200
[Asserts]
jsonpath "$" count == 2

# Obtener registro específico
GET http://localhost:8080/catalogue/2/1
HTTP 200
[Asserts]
jsonpath "$.id_user" == 2
jsonpath "$.id_game" == 1
jsonpath "$.status" == "Jugando"
jsonpath "$.score" == null

# --- Errores en POST /catalogues/{idUser} ---

# El juego ya está en el catálogo del usuario (clave primaria repetida) - debe fallar con 409
POST http://localhost:8080/catalogues/2
Content-Type: application/json
{
    "id_game": 1,
    "status": "Jugando"
}
HTTP 409

# Crear registro con status inválido - debe fallar con 400
POST http://localhost:8080/catalogues/2
Content-Type: application/json
{
    "id_game": 1,
    "status": "status invalido"
}
HTTP 400

# Score fuera de rango por arriba (9) - debe fallar con 400
POST http://localhost:8080/catalogues/2
Content-Type: application/json
{
    "id_game": 1,
    "status": "Jugando",
    "score": 9
}
HTTP 400

# Juego inexistente (viene en el body) - debe fallar con 400
POST http://localhost:8080/catalogues/2
Content-Type: application/json
{
    "id_game": 9999,
    "status": "Jugando"
}
HTTP 400

# Usuario inexistente (viene en la URL) - debe fallar con 404
POST http://localhost:8080/catalogues/9999
Content-Type: application/json
{
    "id_game": 1,
    "status": "Jugando"
}
HTTP 404

# --- Errores de URL y de lectura (GET) ---

# ID de usuario no numérico - debe fallar con 400
GET http://localhost:8080/catalogues/abc
HTTP 400

# URL del registro sin el id del juego (3 partes en vez de 4) - debe fallar con 400
GET http://localhost:8080/catalogue/2
HTTP 400

# Obtener registro inexistente - debe fallar con 404
GET http://localhost:8080/catalogue/2/9999
HTTP 404

# Registro de un usuario inexistente - debe fallar con 404
GET http://localhost:8080/catalogue/9999/1
HTTP 404

# Listar catálogo de usuario inexistente - debe fallar con 404
GET http://localhost:8080/catalogues/9999
HTTP 404

# --- PUT /catalogue/{idUser}/{idGame} ---

# Actualizar registro
PUT http://localhost:8080/catalogue/2/1
Content-Type: application/json
{
    "score": 5,
    "status": "Completado"
}
HTTP 200
[Asserts]
jsonpath "$.id_user" == 2
jsonpath "$.id_game" == 1
jsonpath "$.score" == 5
jsonpath "$.status" == "Completado"

# El trigger recalculó el promedio del juego 1 (único puntaje: 5)
GET http://localhost:8080/library/1
HTTP 200
[Asserts]
jsonpath "$.avg_score" == "5.00"

# Score fuera de rango - debe fallar con 400
PUT http://localhost:8080/catalogue/2/1
Content-Type: application/json
{
    "score": 9,
    "status": "Completado"
}
HTTP 400

# Status inválido - debe fallar con 400
PUT http://localhost:8080/catalogue/2/1
Content-Type: application/json
{
    "score": 5,
    "status": "status invalido"
}
HTTP 400

# Registro inexistente (con body válido) - debe fallar con 404
PUT http://localhost:8080/catalogue/2/9999
Content-Type: application/json
{
    "score": 3,
    "status": "Jugando"
}
HTTP 404

# Usuario inexistente (con body válido) - debe fallar con 404
PUT http://localhost:8080/catalogue/9999/1
Content-Type: application/json
{
    "score": 3,
    "status": "Jugando"
}
HTTP 404

# --- El PUT reemplaza el registro completo ---

# Ponerle puntaje 4 al registro (2, 3) - antes era "score": null y "status": "Por jugar"
PUT http://localhost:8080/catalogue/2/3
Content-Type: application/json
{
    "score": 4,
    "status": "Jugando"
}
HTTP 200
[Asserts]
jsonpath "$.score" == 4

# El promedio del juego 3 pasó a 4.00
GET http://localhost:8080/library/3
HTTP 200
[Asserts]
jsonpath "$.avg_score" == "4.00"

# PUT sin score: el puntaje anterior se borra (queda en null)
PUT http://localhost:8080/catalogue/2/3
Content-Type: application/json
{
    "status": "Completado"
}
HTTP 200
[Asserts]
jsonpath "$.score" == null
jsonpath "$.status" == "Completado"

# Se guardó como null y el promedio del juego 3 volvió a 0.00
GET http://localhost:8080/catalogue/2/3
HTTP 200
[Asserts]
jsonpath "$.score" == null

GET http://localhost:8080/library/3
HTTP 200
[Asserts]
jsonpath "$.avg_score" == "0.00"

# --- Trigger de promedio con dos usuarios + cascada al borrar un usuario ---

# Crear un tercer usuario y capturar su id (no es 3: el POST duplicado de más arriba consumió un valor de la secuencia)
POST http://localhost:8080/users
Content-Type: application/json
{
    "name": "ana",
    "email": "ana@gmail.com",
    "password": "9999"
}
HTTP 201
[Captures]
user3_id: jsonpath "$.id_user"

# El usuario nuevo puntúa el juego 1 con 3
POST http://localhost:8080/catalogues/{{user3_id}}
Content-Type: application/json
{
    "id_game": 1,
    "status": "Jugando",
    "score": 3
}
HTTP 201
[Asserts]
jsonpath "$.id_user" == {{user3_id}}
jsonpath "$.score" == 3

# Promedio del juego 1 con los puntajes 5 y 3
GET http://localhost:8080/library/1
HTTP 200
[Asserts]
jsonpath "$.avg_score" == "4.00"

# Eliminar al usuario nuevo: su catálogo se borra en cascada
DELETE http://localhost:8080/users/{{user3_id}}
HTTP 204

# Su registro desapareció
GET http://localhost:8080/catalogue/{{user3_id}}/1
HTTP 404

# El juego sigue existiendo y el trigger recalculó el promedio (solo queda el 5)
GET http://localhost:8080/library/1
HTTP 200
[Asserts]
jsonpath "$.title" == "Counter Strike 2"
jsonpath "$.avg_score" == "5.00"

# --- DELETE /catalogue/{idUser}/{idGame} ---

# Eliminar registro del catálogo (el del juego 1, que tiene score 5)
DELETE http://localhost:8080/catalogue/2/1
HTTP 204

# El trigger también se dispara al borrar: sin puntajes el promedio vuelve a 0.00
GET http://localhost:8080/library/1
HTTP 200
[Asserts]
jsonpath "$.avg_score" == "0.00"

# Eliminar registro inexistente - debe fallar con 404
DELETE http://localhost:8080/catalogue/2/9999
HTTP 404

# --- Cascada al borrar un juego ---

# Eliminar un juego que está en un catálogo (el 3, registro del usuario 2)
DELETE http://localhost:8080/library/3
HTTP 204

# El registro del catálogo se borró en cascada
GET http://localhost:8080/catalogue/2/3
HTTP 404

# El juego ya no existe
GET http://localhost:8080/library/3
HTTP 404

# El catálogo del usuario 2 quedó vacío (el usuario sigue existiendo)
GET http://localhost:8080/catalogues/2
HTTP 200
[Asserts]
jsonpath "$" count == 0

# --- Método no soportado (405) ---
# DELETE sobre la colección del catálogo - debe fallar con 405
DELETE http://localhost:8080/catalogues/2
HTTP 405