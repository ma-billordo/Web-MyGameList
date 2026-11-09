
async function listarJuegos() {
    const respuesta = await fetch("/library");
    if (!respuesta.ok) { //si el servidor contesta con un error, fetch cree que es una respuesta normal, no un error
        console.error("Error al pedir los juegos", respuesta.status);
        return;
    }
    const juegos = await respuesta.json();
    const lista = document.getElementById("lista-juegos");
    lista.innerHTML = "";

    for (const juego of juegos) {
        const li = document.createElement("li");

        const titulo = document.createElement("h3");
        titulo.className = "juego-titulo";
        titulo.textContent = juego.title;

        const descripcion = document.createElement("p");
        descripcion.className = "juego-descripcion";
        descripcion.textContent = juego.description;

        const boton = document.createElement("button");
        boton.className = "btn-borrar";
        boton.textContent = "Borrar";
        boton.setAttribute("data-id", juego.id_game);

        li.appendChild(titulo);
        li.appendChild(descripcion);
        li.appendChild(boton);
        lista.appendChild(li);
    }
}

const formGames = document.getElementById("form-juego");

async function crearJuego(evento) { //el parametro viene siempre por defecto en los eventListener, aca se le da un nombre
    evento.preventDefault(); //se previene el comportamiento por defecto del navegador

    const nuevoJuego = {
        title: document.getElementById("title").value,
        description: document.getElementById("description").value,
    };

    const respuesta = await fetch("/library", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(nuevoJuego),
    });

    if (!respuesta.ok) {
        alert(await respuesta.text());
        return;
    }
    
    formGames.reset();
    await listarJuegos();

}

const listaJuegos = document.getElementById("lista-juegos");

async function borrarJuego(evento){
    if (evento.target.tagName !== "BUTTON") {
        return;
    }
    const id = evento.target.dataset.id;
    const respuesta = await fetch(`/library/${id}`, {
        method: "DELETE",
    });

    if (!respuesta.ok){ 
        alert(await respuesta.text());
        return;
    }
    await listarJuegos()

}

document.addEventListener("DOMContentLoaded", listarJuegos);
formGames.addEventListener("submit", crearJuego);
listaJuegos.addEventListener("click", borrarJuego);