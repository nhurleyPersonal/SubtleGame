let ws;
let currentPlayersIDs = [];
let currentPlayer;

async function joinGameServer(name, serverID) {
  try {
    const wsUrl = `ws://${window.location.host}/ws?serverID=${serverID}`;
    ws = new WebSocket(wsUrl);
  } catch (error) {
    console.error("WebSocket creation error:", error);
  }

  ws.onopen = function () {
    console.log("WebSocket connection established");
    var message = {
      type: "joinGame",
      body: { name, serverID }, // Use the name here
    };
    ws.send(JSON.stringify(message));
  };

  ws.onmessage = function (event) {
    console.log("Message received from server:", event.data);
    var message = JSON.parse(event.data);
    (messageHandlers[message.type] || messageHandlers.default)(message);
  };

  ws.onerror = function (error) {
    console.error("WebSocket error:", error);
  };

  ws.onclose = function () {
    console.log("WebSocket connection closed");
    sessionStorage.removeItem("ws");
  };
}

var messageHandlers = {
  navigate: function (message) {
    console.log("Navigating to:", message.body);
    var data = JSON.parse(message.body);
    var htmlContent = data.html;
    document.body.innerHTML = htmlContent;
  },

  gameStarted: function (message) {
    afterStartGame();
  },

  whoami: function (message) {
    var playerInfo = JSON.parse(message.body);
    document.getElementById("playerName").innerText =
      playerInfo.Name || "Unknown";
    currentPlayer = playerInfo;
  },

  playerJoined: function (message) {
    console.log("playerJoined", message);
    var playerInfo = JSON.parse(message.body);
    if (currentPlayersIDs.includes(playerInfo.id)) {
      return;
    }
    if (playerInfo.id === currentPlayer.id) {
      return;
    }
    currentPlayersIDs.push(playerInfo.id);
    buildPlayerItem(playerInfo);
  },

  currentPlayers: function (message) {
    var currentPlayers = JSON.parse(message.body);
    deleteCurrentPlayers();
    currentPlayers.map((player) => {
      if (currentPlayersIDs.includes(player.id)) {
        return;
      }
      if (player.id === currentPlayer.id) {
        return;
      }
      currentPlayersIDs.push(player.id);
      buildPlayerItem(player);
    });
  },

  lobbyFull: function (message) {
    var errorMessage = document.getElementById("errorMessage");
    if (!errorMessage) {
      errorMessage = document.createElement("div");
      errorMessage.id = "errorMessage";
      errorMessage.style.color = "red";
      errorMessage.style.display = "none";
      errorMessage.innerText = "Failed to join game, lobby full";
      document.body.appendChild(errorMessage);
    }
    errorMessage.style.display = "block";
  },

  default: function (message) {
    var targetElement = document.getElementById(message.target);
    if (targetElement) {
      var body = message.body;
      try {
        body = JSON.parse(message.body);
      } catch (e) {
        console.log("Body is not a JSON string:", message.body);
      }
      targetElement.innerHTML =
        typeof body === "string" ? body : JSON.stringify(body);
    }
  },
};

function sendMessage(type) {
  console.log("Sending message:", type); // Debugging log
  var message = {
    type: type,
    body: "",
  };
  window.ws.send(JSON.stringify(message));
}

function sendStartGame() {
  console.log("Starting game");
  var message = {
    type: "startGame",
    body: { leader: "leader" },
  };
  ws.send(JSON.stringify(message));
}

function setWord(word) {
  console.log("Sending message:", word); // Debugging log
  var message = {
    type: "setWord",
    body: word,
  };
  window.ws.send(JSON.stringify(message));
}

function sendLetterMessage(type) {
  console.log("Sending message:", type); // Debugging log
  var message = {
    type: type,
    body: "",
  };
  window.ws.send(JSON.stringify(message));
}

window.addEventListener("beforeunload", function () {
  if (ws) {
    ws.close();
  }
});
