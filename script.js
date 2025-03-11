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
        .then(response => response.json())
        .then(data => alert(data.message))
        .catch(error => console.error("Error:", error));
}
