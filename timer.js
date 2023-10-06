let timer;
let seconds = 0;
let minutes = 1;

let i = 0;
let words = [];
let words2 = [];
let on = false;

document.addEventListener("keydown", async function(event) {
    if (event.keyCode === 32) { // 32 è il codice del tasto space bar
        console.log("AAAAAAAaa")
        if(!on){
            await startTimer();
        } else {
            stopTimer();
        }
    }
});


async function startTimer() {

    if(!on){
        on = true;
        let response= await fetch("http://localhost:8000/");
        response = await response.json();
        words = Array.from(response);
        
        response = await fetch("http://localhost:8000/");
        response = await response.json();
        words2 = Array.from(response);
        i=0;

    } else return;


    if(i<10){
        document.getElementById("word").innerHTML = words[i];
        i++;
        timer = setInterval(updateTimer, 1000);
    } else{
        i=0;
        words = words2;

        let response = await fetch("http://localhost:8000/");
        response = await response.json();
        words2 = Array.from(response);

        document.getElementById("word").innerHTML = words[i];
        i++;
        timer = setInterval(updateTimer, 1000);
    }


}

function stopTimer() {
    on = false;
    clearInterval(timer);
}

function resetTimer() {
    on = false;
    clearInterval(timer);
    seconds = 1;
    minutes = 1;
    updateTimer();
}

function updateTimer() {
    if (seconds === 0) {
        if(minutes==0) {
            seconds = 0;
            minutes = 1;
            resetTimer()
            return;
        }
        seconds = 60;
        minutes--;
    }
    seconds--;
    

    const formattedTime = `${padNumber(minutes)}:${padNumber(seconds)}`;
    document.getElementById("timer").innerHTML = formattedTime;
}

function padNumber(num) {
    return num.toString().padStart(2, "0");
}
