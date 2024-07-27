let round = 0;
let inputWord = "";

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
        selectPlayer();
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
  let canSubmit = true;
  if (!currentPlayer.Ready) {
    canSubmit = false;
  }
  currentPlayers.forEach((player) => {
    if (!player.Ready) {
      canSubmit = false;
    }
  });

  if (canSubmit) {
    beforeStartGame();
  }
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
  buildPlayerItem(currentPlayer);
  currentPlayers.forEach((player) => {
    if (player.Name !== currentPlayer.Name) {
      buildPlayerItem(player);
    }
  });
}

function writeGameView(templateId) {
  removeEventListeners();
  let clon = document.getElementById(templateId).content.cloneNode(true);
  var children = document.body.children;
  //   attachEventListeners();
  for (var i = 0; i < children.length; i++) {
    var child = children[i];
    child.remove();
  }
  document.body.appendChild(clon);
}

function sendWord() {
  const lettersContainer = selectedBox.querySelector(".letters-container");
  const letterBoxes = lettersContainer.querySelectorAll(".letter");
  inputWord = "";
  letterBoxes.forEach((letterBox) => {
    inputWord += letterBox.textContent;
  });
  if (inputWord.length !== letterBoxes.length) {
    return;
  }
  ws.send(JSON.stringify({ type: "setWord", body: { word: inputWord } }));
}

function indicatePlayerReady(player) {
  currentPlayers.forEach((player) => {
    if (player.Name === player) {
      player.Ready = true;
    }
  });
  if (currentPlayer.Name === player) {
    currentPlayer.Ready = true;
  }
  let canSubmit = true;
  if (!currentPlayer.Ready) {
    canSubmit = false;
  }
  currentPlayers.forEach((player) => {
    if (!player.Ready) {
      canSubmit = false;
    }
  });

  if (canSubmit) {
    startButton = document.querySelector(".start-game-button");
    startButton.classList.add("start-game-button-ready");
  }

  const playerDivs = document.querySelectorAll(".player-name-checkmark-box");
  const firstPlayer = playerDivs[0];
  const firstPlayerDivName = firstPlayer.querySelector(".player-name");
  if (firstPlayerDivName.innerText === player) {
    const invalidWord = document.getElementById("invalid-word-error");
    invalidWord.classList.remove("invalid-word-visible");
    invalidWord.classList.add("invalid-word-error");
  }
  playerDivs.forEach((playerDiv) => {
    const playerDivName = playerDiv.querySelector(".player-name");
    if (playerDivName.innerText === player) {
      const checkmark = playerDiv.querySelector(".checkmark");
      checkmark.classList.add("checkmark-ready");
    }
  });
}

function showInvalidWord() {
  const playerDivs = document.querySelectorAll(".player-name-checkmark-box");
  const firstPlayer = playerDivs[0];
  const checkmark = firstPlayer.querySelector(".checkmark");
  checkmark.classList.remove("checkmark-ready");

  const invalidWord = document.getElementById("invalid-word-error");
  console.log(invalidWord);
  invalidWord.classList.remove("invalid-word-error");
  invalidWord.classList.add("invalid-word-visible");
}
