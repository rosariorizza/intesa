
let queryString = window.location.search;

let urlParams = new URLSearchParams(queryString);

let roomCode = urlParams.get("room");

if(!roomCode){
    console.log("Room not found")
}

const socket = new WebSocket(`ws://localhost:4200/play?room=${roomCode}`);
let on = false



function startTimer(){
    socket.send(1);
    on = true;
}

function stopTimer(){
    socket.send(2);
    on = false;
}

function resetTimer(){
    socket.send(3);
    on = false;
}

socket.onopen = () => {
    console.log("Connessione WebSocket aperta.");
    document.getElementById("roomCodeShare").textContent = `Condividi il codice per giocare in gruppo: ${roomCode}`;

};

socket.onmessage = (event) => {
    const {timerValue, wordValue} = JSON.parse(event.data);
    document.getElementById("timerValue").textContent = "Timer: " + timerValue;
    document.getElementById("wordValue").textContent = wordValue;

};

socket.onerror = (error) => {
    console.error("Errore WebSocket: " + error);
};

socket.onclose = (event) => {
    if (event.wasClean) {
        console.log("Connessione WebSocket chiusa in modo pulito.");
    } else {
        console.error("Connessione WebSocket chiusa in modo anomalo.");
    }
};

document.addEventListener("keydown", async function(event) {
    if (event.keyCode === 32) { // 32 è il codice del tasto space bar
        if(!on){
            startTimer();
        } else {
            stopTimer();
        }
    }
});

