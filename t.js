const socket = new WebSocket('ws://localhost:8080/ws/');


// Connection opened
socket.addEventListener('open', function (event) {
    socket.send('Hello Server!');
});

// Listen for messages
socket.addEventListener('message', function (event) {
    let str = event.data;

    if (str.startsWith('$')) {
        str = str.substring(1);
        document.getElementById('state').innerHTML = str
        return;
    }

    console.log(event.data);
});

socket.addEventListener('error', function (event) {
    console.log('Error from server ', event.data);
});

socket.addEventListener('close', function (event) {
    console.log('Closed from server ', event.data);
});


const start = () => {
    fetch('http://localhost:8080/servers/1234/start/', {method: "POST"})
}

const stop = () => {
    fetch('http://localhost:8080/servers/1234/stop/', {method: "POST"})
}

const cmd = () => {
    var name = window.prompt("Enter command: ");
    if (name) {
    socket.send(name)
    }
}