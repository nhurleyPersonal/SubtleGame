package server

var RootTemplRaw = `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>HTMX WebSocket Example</title>
    <script src="https://unpkg.com/htmx.org@1.8.5"></script>
    <script src="https://unpkg.com/htmx-ext-ws@2.0.0/ws.js"></script>

    <style>
        .player-name {
            display: inline-block;
            background-color: #f0f0f0;
            padding: 5px 10px;
            border-radius: 4px;
            font-weight: bold;
        }
    </style>
</head>
<body hx-ext="ws" ws-connect="wss://{{.Host}}/startGame">
    <h1>HTMX WebSocket Example</h1>
    {{with .Players}}
	<input type="text" id="playername" placeholder="Enter your name">
    <input type="hidden" id="playerID" value="{{.PlayerID}}">
    <button onclick="setName()">Set Word</button>
        <table>
            <tr>
                <th>Player Name</th>
                <th>Player Word</th>
            </tr>
            {{range .}}
            <div id="player-{{.ID}}" ws-target="player-{{.ID}}" hx-swap="innerHTML">
                <tr>
                    <td>{{.ID}}</td>
                    <td>Data 3</td>
                </tr>
            </div>
        </table>
    
    {{end}}
    
    <script>
        var ws = new WebSocket("ws://" + window.location.host + "/startGame");

        ws.onopen = function() {
            console.log("WebSocket connection established");
        };

        ws.onmessage = function(event) {
            console.log("Message received from server:", event.data);
            var message = JSON.parse(event.data);
            if (message.type === "gameState") {
                var gameState = JSON.parse(message.body);
                
                document.getElementById("counterTwo").innerHTML = "CountTwo: " + gameState.countTwo;
				} else {
                var targetElement = document.getElementById(message.target);
                if (targetElement) {
                    targetElement.innerHTML = message.body;
                }
            }
        };

        ws.onerror = function(error) {
            console.error("WebSocket error:", error);
        };

        ws.onclose = function() {
            console.log("WebSocket connection closed");
        };

        function sendMessage(type) {
            console.log("Sending message:", type); // Debugging log
            var message = {
                type: type,
                body: ""
            };
            ws.send(JSON.stringify(message));
        }

        function setWord(word) {
            console.log("Sending message:", word); // Debugging log
            var message = {
                type: "setWord",
                body: word
            };
            ws.send(JSON.stringify(message));
        }

        function setName(id) {
            var word = document.getElementById("word").value;
            var playerID = document.getElementById("playerID").value;

            console.log("Sending message:", name); // Debugging log
            var message = {
                type: "setName",
                body: name
            };
            ws.send(JSON.stringify(message));
        }

		function sendLetterMessage(type) {
            console.log("Sending message:", type); // Debugging log
            var message = {
                type: type,
                body: ""
            };
            ws.send(JSON.stringify(message));
        }
    </script>
</body>
</html>`
