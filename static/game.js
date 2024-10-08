let round = 0;
let inputWord = "";

function joinGame(name, lobbyID) {
  if (name === "") {
    document.getElementById("error-text").innerText = "Please enter a name!";
    document.getElementById("error-text").classList.add("error-text-displayed");
    return
  }
  if (lobbyID === "") {
    document.getElementById("error-text").innerText = "Please enter a valid lobby.";
    document.getElementById("error-text").classList.add("error-text-displayed");
    return
  }

  fetch(
    `/gamelobby?lobbyID=${encodeURIComponent(
      lobbyID
    )}&name=${encodeURIComponent(name)}`,
    {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    }
  )
    .then((response) => {
      if (response.ok) {
        return response.text();
      } else if (response.status === 404) {
        response.text().then((text) => {
          document.getElementById("error-text").innerText = text;
          document.getElementById("error-text").classList.add("error-text-displayed");
        })
      }
    })
    .then((html) => {
      if (html) {
        document.body.innerHTML = html;
        htmx.process(document.body);
        attachEventListeners();
        selectPlayer();
      }
    })
    .catch((error) => {
      console.error("Error:", error);
    });
}

function createLobby() {
  fetch("/createServer", {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
    },
  })
    .then((response) => response.text())
    .then((html) => {
      const createButton = document.getElementById("createGameButton");
      createButton.innerHTML = html;
      createButton.className = "server-create-button-clicked";

      const copyIcon = document.createElement("span");
      copyIcon.className = "material-symbols-outlined";
      copyIcon.innerText = "content_copy";
      copyIcon.style.cursor = "pointer";
      createButton.onclick = function () {
        navigator.clipboard
          .writeText(html.slice(html.indexOf(">") + 1, html.indexOf(">") + 7))
          .then(() => {
            document.getElementById("copiedClipboardText").classList.add("clicked")
          })
          .catch((error) => {
            console.error("Error copying to clipboard:", error);
          });
      };

      createButton.appendChild(copyIcon);
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
  buildSelfPlayerGameView(currentPlayer);
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
  attachEventListeners();
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
  currentPlayers.forEach((playerMap) => {
    if (playerMap.Name === player) {
      playerMap.Ready = true;
    }
  });
  if (currentPlayer.Name === player) {
    console.log("SETTING READY TO TRUE");
    currentPlayer.Ready = true;
  }
  let canSubmit = true;
  if (!currentPlayer.Ready) {
    canSubmit = false;
    document
      .querySelector(".start-game-button")
      .classList.remove("start-game-button-ready");
  }
  currentPlayers.forEach((playerMap) => {
    if (!playerMap.Ready) {
      canSubmit = false;
      document
        .querySelector(".start-game-button")
        .classList.remove("start-game-button-ready");
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
  invalidWord.classList.remove("invalid-word-error");
  invalidWord.classList.add("invalid-word-visible");
}

function guessWord() {
  const lettersContainer = selectedBox.querySelector(".letters-container");
  const letterBoxes = lettersContainer.querySelectorAll(".letter");
  inputWord = "";
  letterBoxes.forEach((letterBox) => {
    inputWord += letterBox.textContent;
    letterBox.textContent = "";
    letterBox.classList.remove("letter-highlight");
  });
  if (inputWord.length !== letterBoxes.length) {
    return;
  }

  targetName = selectedBox.querySelector(".player-name").innerHTML;
  targetPlayerId = currentPlayers.find(
    (player) => player.Name === targetName
  ).id;
  ws.send(
    JSON.stringify({
      type: "guessWord",
      body: {
        word: inputWord,
        selfName: currentPlayer.id,
        targetName: targetPlayerId,
      },
    })
  );
}

function writeGuessResults(completelyCorrect, partiallyCorrect) {
  const lettersContainer = selectedBox.querySelector(".letters-container");
  const letterBoxes = lettersContainer.querySelectorAll(".letter");
  const lettersContainerVert = selectedBox.querySelector(
    ".letters-vertical-stack-container"
  );
  if (completelyCorrect && completelyCorrect.length == letterBoxes.length) {
    for (let i = 0; i < letterBoxes.length; i++) {
      letterBoxes[i].innerText = inputWord.charAt(i);
      letterBoxes[i].classList.add("letter-filled");
      letterBoxes[i].classList.add("letter-correct");
      letterBoxes[i].classList.remove("letter-highlight");
    }
    while (lettersContainerVert.firstChild) {
      lettersContainerVert.removeChild(lettersContainerVert.firstChild);
    }
    selectedBox = null;
    return;
  }

  const guessContainer = document.createElement("div");
  guessContainer.className = "guess-results-container";
  lettersContainerVert.appendChild(guessContainer);

  for (let i = 0; i < 5; i++) {
    var letter = document.createElement("div");
    letter.className = "letter letter-filled";
    letter.textContent = inputWord[i];
    guessContainer.appendChild(letter);
  }

  if (completelyCorrect && completelyCorrect.length > 0) {
    completelyCorrect.forEach((indOfGuess) => {
      Array.from(guessContainer.children)[indOfGuess].classList.add(
        "letter-correct"
      );
    });
  }

  if (partiallyCorrect && partiallyCorrect.length > 0) {
    partiallyCorrect.forEach((indOfGuess) => {
      Array.from(guessContainer.children)[indOfGuess].classList.add(
        "letter-partially-correct"
      );
    });
  }

  lettersContainerVert.appendChild(guessContainer);
}

// I ralize how horrendous this is, hopefully this will be fixed when I refactor to htmx
// I hate this so much
function writePlayerScore(playerID) {
  let targetDiv;
  let targetPlayer = currentPlayers.find((player) => player.id === playerID);
  let targetPlayerName = targetPlayer.Name;
  let playerList = document.body.querySelector(".player-list");
  Array.from(playerList.children).forEach((playerDiv) => {
    let playerName = playerDiv.querySelector(".player-name");
    if (!playerName) {
      playerName = playerDiv.querySelector(".self-player-name").innerHTML;
    } else {
      playerName = playerDiv.querySelector(".player-name").innerHTML;
    }
    if (playerName === targetPlayerName) {
      targetDiv = playerDiv;
    }
  });

  targetDiv.querySelector(
    ".player-score"
  ).innerHTML = `Score (${targetPlayer.Score})`;
}
