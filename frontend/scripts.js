class Event {
    constructor(type, payload) {
        this.type = type;
        this.payload = payload;
    }
}

class SendMessageEvent {
    constructor(message, from) {
        this.message = message;
        this.from = from;
    }
}

class NewMessageEvent {
    constructor(sendMessage, sent) {
        this.sendMessage = sendMessage;
        this.sent = sent;
    }
}

class LoginData {
    constructor(username, password) {
        this.username = username;
        this.password = password;
    }
}

const loginForm = document.getElementById('chat-login');

const chatContainer = document.getElementById('chatContainer');
const messageInput = document.getElementById('messageInput');
const sendButton = document.getElementById('sendButton');

loginForm.addEventListener('submit', function(event) {
    event.preventDefault();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    const url = 'login';
    userLogin = Object.assign(new LoginData, {username, password})

    const requestOptions = {
        method: 'POST', 
        headers: {'Content-Type': 'application/json'}, 
        body: JSON.stringify(userLogin)
    };

    fetch(url, requestOptions)
        .then(response => {
            if (!response.ok) {
                throw new Error(`Network response was not ok, status code: ${response.status}`);
            }
                return response.json(); 
        })
        .then(data => {
            if (data.otp != undefined) {
                connectWebsocketsOTP(data.otp);
            } else {
                throw new Error(`No OTP in response data: ${data}`);
            }
        })
        .catch(error => {
            alert('Fetch error:', error);
        });
}); 

sendButton.addEventListener('click', sendMessage);

messageInput.addEventListener('keydown', function(event) {
    if (event.key === 'Enter') {
        event.preventDefault();
        sendMessage();
    }
});

let connection;

function connectWebsocketsOTP(otp) {
    connection = new WebSocket("wss://" + window.location.host + "/ws?otp=" + otp);
    let loginWrapper = document.getElementById("login-wrapper");
    let chatWrapper = document.getElementById("chat-wrapper");

    connection.onopen = () => {
        let header = document.getElementById("ws-status");
        header.style.color = "blue";
        header.innerHTML = "WS: connected";

        loginWrapper.style.display = "none";
        chatWrapper.style.display = "flex";
    }

    connection.onclose = () => {
        let header = document.getElementById("ws-status");
        header.style.color = "red";
        header.innerHTML = "WS: not connected";   

        loginWrapper.style.display = "flex";
        chatWrapper.style.display = "none";
    }

    connection.onmessage = (evt) => {
        let eventData = JSON.parse(evt.data);

        let event = Object.assign(new Event(), eventData);  

        routeEvent(event);
    }
}

function routeEvent(event) {
    if (event.type == "") {
        console.log("No event type specified");
        return;
    }

    switch(event.type) {
        case "new_message":
            const messageEvent = Object.assign(new NewMessageEvent, event.payload)

            appendMessage(messageEvent);
            break;
        default:
            console.log("Not supported event type");
    }
}

function appendMessage(event) {
    let message = `${new Date().toLocaleString()} - ${event.from}: ${event.message}`;

    const messageElement = document.createElement('div');
    messageElement.className = 'message';
    messageElement.textContent = message;

    chatContainer.appendChild(messageElement);
    chatContainer.scrollTop = chatContainer.scrollHeight;
}

function sendEvent(type, payload) {
    let event = new Event(type, payload);

    connection.send(JSON.stringify(event));
}

function sendMessage() {
    const message = messageInput.value;
    
    if (message.trim() === '') {
        return;
    }

    if (connection != undefined) {        
        let outgoingEvent = new SendMessageEvent(message, "test")

        sendEvent("send_message", outgoingEvent);    

        messageInput.value = '';
    }
    
}

