
function sendSongs() {
    fetch("/songs", {
        method: "POST",
        headers: {
            "Content-Type": "application/json"
        },
        body: JSON.stringify({ "songname": "Despacito - Luis Fonsi ft. Daddy Yankee", "quality": "1080p" })
    });
}
