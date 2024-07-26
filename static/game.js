let round = 0;

function joinGame(name, serverID) {
  console.log(name, serverID);
  fetch("/gamelobby", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ name: name }),
  })
    .then((response) => {
      if (response.ok) {
        return response.text(); // Get the HTML content
      } else {
        console.error("Failed to fetch from /gamelobby");
      }
    })
    .then((html) => {
      if (html) {
        document.body.innerHTML = html; // Write the HTML content to the DOM
        writeGameView("lobbyview");
        joinGameServer(name, serverID); // Reinitialize the WebSocket connection
        attachEventListeners();
      }
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}

function createLobby() {
  console.log("HERE");
  fetch("/createServer", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
  })
    .then((response) => response.text())
    .then((html) => {
      document.getElementById("createGameButton").outerHTML = html;
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}

function startGame() {
  beforeStartGame();
}

function beforeStartGame() {
  sendStartGame();
}

// Timer functions
function startTimer(duration, display) {
  let timer = duration,
    minutes,
    seconds;
  timerInterval = setInterval(function () {
    minutes = parseInt(timer / 60, 10);
    seconds = parseInt(timer % 60, 10);

    minutes = minutes < 10 ? "0" + minutes : minutes;
    seconds = seconds < 10 ? "0" + seconds : seconds;

    display.textContent = minutes + ":" + seconds;

    if (--timer < 0) {
      clearInterval(timerInterval);
    }
  }, 1000);
}

function afterStartGame() {
  round++;
  writeGameView("gameroom");
  const timerDisplay = document.getElementById("game-timer");
  const gameDuration = 19;
  startTimer(gameDuration, timerDisplay);
}

function writeGameView(templateId) {
  let clon = document.getElementById(templateId).content.cloneNode(true);
  var children = document.body.children;
  console.log(children);
  for (var i = 0; i < children.length; i++) {
    var child = children[i];
    child.remove();
  }
  document.body.appendChild(clon);
}
