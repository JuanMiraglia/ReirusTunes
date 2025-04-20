function sendSongs() {
    let songList = document.getElementById("song-names").value.trim();

    if (!songList) {
        alert("Por favor ingresa al menos una canción.");
        return;
    }

    let songs = songList.split("\n").map(song => song.trim()).filter(song => song !== "");

    fetch("http://localhost:8080/songs", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ songs })
    })
        .then(response => {
            if (!response.ok) throw new Error("Respuesta inválida del servidor");
            return response.json();
        })
        .then(data => {
            alert(data.message); // "Todas las descargas han finalizado"

            // Descargar automáticamente los archivos
            if (Array.isArray(data.files)) {
                data.files.forEach(file => {
                    const link = document.createElement("a");
                    link.href = `http://localhost:8080/downloads/${encodeURIComponent(file)}`;
                    link.download = file;
                    document.body.appendChild(link);
                    link.click();
                    document.body.removeChild(link);
                });
            }
        })
        .catch(error => {
            console.error("Error:", error);
            alert("Ocurrió un error al procesar tu solicitud.");
        });
}
