let ws;
let currentPlayersIDs = [];
let currentPlayer;
let currentPlayers = [];
let currentPlayersMap = {}; // I sincerely apologize for this monstrosity

async function joinGameServer(name, serverID) {
  try {
    let protocol = window.location.protocol === "https:" ? "wss://" : "ws://";
    const wsUrl = `${protocol}${window.location.host}/ws?serverID=${serverID}`;
    ws = new WebSocket(wsUrl);
  } catch (error) {
    console.error("WebSocket creation error:", error);
  }

  ws.onopen = function () {
    console.log("WebSocket connection established");
    var message = {
      type: "joinGame",
      body: { name, serverID },
    };
    ws.send(JSON.stringify(message));
    var playerName = document.createElement("h2");
    playerName.innerText = playerInfo.message || "Unknown";
    document.body.appendChild(playerName);
  };

  ws.onmessage = function (event) {
    console.log("Message received from server:", event.data);
    var message = JSON.parse(event.data);
    (messageHandlers[message.type] || messageHandlers.default)(message);
    // Create player name element
    var playerName = document.createElement("h2");
    playerName.innerText = playerInfo.message.data || "Unknown";
    document.body.appendChild(playerName);
  };

  ws.onerror = function (error) {
    console.error("WebSocket error:", error);
  };

  ws.onclose = function () {
    console.log("WebSocket connection closed");
    sessionStorage.removeItem("ws");
    var playerName = document.createElement("h2");
    playerName.innerText = "Reconnecting..." || "Unknown";
    document.body.appendChild(playerName);
    if (reconnectAttempts < maxReconnectAttempts) {
      setTimeout(() => {
        reconnectAttempts++;
        joinGameServer(name, serverID);
      }, reconnectInterval);
    } else {
      console.log("Max reconnection attempts reached");
    }
  };
}

let reconnectInterval = 5000; // 5 seconds
let reconnectAttempts = 0;
let maxReconnectAttempts = 12; // 1 minute

var messageHandlers = {
  gameStarted: function (message) {
    currentPlayers = message.body;
    afterStartGame();
  },

  wordSet: function (message) {
    currentPlayer.Word = inputWord;
    indicatePlayerReady(message.body.player);
  },

  guessResults: function (message) {
    console.log("MESSAGE", message);
    body = JSON.parse(message.body);
    writeGuessResults(body.completelyCorrect, body.partiallyCorrect);
  },

  correctWord: function (message) {
    console.log("MESSAGE", message);
    let completelyCorrect = [];
    for (let i = 0; i < inputWord.length; i++) {
      completelyCorrect.push(i);
    }
    writeGuessResults(completelyCorrect, []);
  },

  invalidWord: function (message) {
    showInvalidWord();
  },

  whoami: function (message) {
    var playerInfo = JSON.parse(message.body);
    document.getElementById("player-name").innerText =
      playerInfo.Name || "Unknown";
    currentPlayer = playerInfo;
  },

  currentPlayers: function (message) {
    // Check if message.body is already an object
    if (typeof message.body === "string") {
      currentPlayers = JSON.parse(message.body);
    } else {
      currentPlayers = message.body;
    }
    deleteCurrentPlayersViews();
    currentPlayers.forEach((player) => {
      if (!player.Ready) {
        document
          .querySelector(".start-game-button")
          .classList.remove("start-game-button-ready");
      }
      buildPlayerLobbyView(player);
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
