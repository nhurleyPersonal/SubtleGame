let selectedBox;
let selectedBoxIndex;
let currentBoxIndex = 0;
let playerNameDivs;

function showKeyboard() {
  //   const hiddenInput = document.getElementById("hiddenInput");
  //   hiddenInput.focus();
  //   hiddenInput.addEventListener("blur", () => {
  //     // Optionally, refocus if needed
  //     setTimeout(() => hiddenInput.focus(), 0);
  //   });
}

// SELECTING AND FILLING BOX LOGIC

// Define the keydown event handler
function handleKeydown(event) {
  const key = event.key;
  const lettersContainer = selectedBox.querySelector(".letters-container");
  const letterBoxes = lettersContainer.querySelectorAll(".letter");

  // Only process single character keys
  if (key.length === 1 && /[a-zA-Z]/.test(key)) {
    // Fill the current box with the typed letter
    letterBoxes[currentBoxIndex].innerText = key.toUpperCase();
    letterBoxes[currentBoxIndex].classList.add("letter-filled");

    // Remove highlight from the current box

    if (currentBoxIndex === letterBoxes.length - 1) {
      sendButton = document.querySelector(".send-word-button");
      sendButton.classList.add("send-word-button-ready");
    }

    if (currentBoxIndex < letterBoxes.length - 1) {
      letterBoxes[currentBoxIndex].classList.remove("letter-highlight");
      currentBoxIndex = currentBoxIndex + 1;
    }

    // Highlight the next box
    letterBoxes[currentBoxIndex].classList.add("letter-highlight");
    for (let i = 0; i < letterBoxes.length; i++) {
      if (i !== currentBoxIndex) {
        letterBoxes[i].classList.remove("letter-highlight");
      }
    }
  }

  if (key === "Backspace") {
    sendButton = document.querySelector(".send-word-button");
    sendButton.classList.remove("send-word-button-ready");
    letterBoxes[currentBoxIndex].classList.remove("letter-filled");
    letterBoxes[currentBoxIndex].classList.remove("letter-highlight");

    if (
      currentBoxIndex === letterBoxes.length - 1 &&
      letterBoxes[currentBoxIndex].innerText !== ""
    ) {
      letterBoxes[currentBoxIndex].innerText = "";
    } else {
      if (currentBoxIndex > 0) {
        currentBoxIndex = currentBoxIndex - 1;
      }
      letterBoxes[currentBoxIndex].innerText = "";
    }

    letterBoxes[currentBoxIndex].classList.add("letter-highlight");
  }
}

function selectPlayer() {
  selfPlayer = document.querySelectorAll(".self-player-container");
  selfPlayer.forEach((player) => {
    player.classList.add("player-container-clicked");
  });
  selectedBox = selfPlayer[0];
  showKeyboard();
  selectedBoxIndex = 0;
  currentBoxIndex = 0;
  const lettersContainer = selectedBox.querySelector(".letters-container");
  const letterBoxes = lettersContainer.querySelectorAll(".letter");
  letterBoxes[currentBoxIndex].classList.add("letter-highlight");
}

function playerRowClicked(row, index) {
  if (row.querySelector(".player-name").innerText === currentPlayer.Name) {
    return;
  }
  selectedBox = row;
  showKeyboard();
  selectedBoxIndex = index;
  currentBoxIndex = 0;
  const playerNameDivs = document.querySelectorAll(".player-container");
  playerNameDivs.forEach((div) => {
    div.classList.remove("player-container-clicked");
  });
  selectedBox.classList.add("player-container-clicked");
  const lettersContainer = selectedBox.querySelector(".letters-container");
  const letterBoxes = lettersContainer.querySelectorAll(".letter");
  letterBoxes[currentBoxIndex].classList.add("letter-highlight");
}

// BUILDING PLAYER ROWS

function deleteCurrentPlayersViews() {
  const playerList = document.getElementById("player-list");
  playerList.innerHTML = "";
  playerNameDivs.forEach((div) => {
    div.remove();
  });
}

function buildSelfPlayerGameView(playerInfo) {
  // Check if the player is the current player

  var playerList = document.getElementById("player-list");

  // Create a new player container
  var playerContainer = document.createElement("div");
  playerContainer.className = "player-container";

  // Create player name element
  var playerName = document.createElement("h2");
  playerName.className = "self-player-name";
  playerName.innerText = playerInfo.Name || "Unknown";
  playerContainer.appendChild(playerName);

  // Create letters container
  var lettersContainer = document.createElement("div");
  lettersContainer.className = "letters-container";
  for (let i = 0; i < 5; i++) {
    var letter = document.createElement("div");
    letter.className = "letter letter-filled-self";
    letter.innerText = playerInfo.Word[i];
    lettersContainer.appendChild(letter);
  }
  playerContainer.appendChild(lettersContainer);
  var spacer = document.createElement("div");
  spacer.className = "send-button-spacer";
  playerContainer.appendChild(spacer);
  // Append the new player container to the player list
  playerList.appendChild(playerContainer);

  // Attach event listener to the new player container
  playerContainer.addEventListener("click", function (event) {
    event.stopPropagation(); // Prevent the document click event from firing
    playerRowClicked(playerContainer);
  });
}

function buildPlayerItem(playerInfo) {
  // Check if the player is the current player

  var playerList = document.getElementById("player-list");

  // Create a new player container
  var playerContainer = document.createElement("div");
  playerContainer.className = "player-container";

  // Create player name element
  var playerName = document.createElement("h2");
  playerName.className = "player-name";
  playerName.innerText = playerInfo.Name || "Unknown";
  playerContainer.appendChild(playerName);

  // Create letters container
  let lettersContainerVerticalStack = document.createElement("div");
  lettersContainerVerticalStack.className = "letters-vertical-stack-container";

  var lettersContainer = document.createElement("div");
  lettersContainer.className = "letters-container";
  for (let i = 0; i < 5; i++) {
    var letter = document.createElement("div");
    letter.className = "letter";
    lettersContainer.appendChild(letter);
  }
  lettersContainerVerticalStack.appendChild(lettersContainer);
  playerContainer.appendChild(lettersContainerVerticalStack);

  // Send button
  sendButtonContainer = document.createElement("div");
  sendButtonContainer.className = "send-word-button-container";
  playerContainer.appendChild(sendButtonContainer);
  sendButtonSpan = document.createElement("span");
  sendButtonSpan.className = "material-symbols-outlined send-word-button";
  sendButtonSpan.innerText = "Send";
  sendButtonSpan.addEventListener("click", function () {
    console.log("Sending word");
    guessWord();
  });
  sendButtonContainer.appendChild(sendButtonSpan);

  // Append the new player container to the player list
  playerList.appendChild(playerContainer);

  // Attach event listener to the new player container
  playerContainer.addEventListener("click", function (event) {
    event.stopPropagation(); // Prevent the document click event from firing
    playerRowClicked(playerContainer);
  });
}

function buildPlayerLobbyView(playerInfo) {
  // Check if the player is the current player
  if (playerInfo.id === currentPlayer.id) {
    return;
  }

  var playerList = document.getElementById("player-list");
  // Create a new player container
  var playerContainer = document.createElement("div");
  playerContainer.className = "player-container";

  var playerNameBox = document.createElement("div");
  playerNameBox.className = "player-name-checkmark-box";
  playerContainer.appendChild(playerNameBox);

  // Create player name element
  var playerName = document.createElement("h2");
  playerName.className = "player-name";
  playerName.innerText = playerInfo.Name || "Unknown";
  var checkmark = document.createElement("span");
  checkmark.className = "checkmark";
  checkmark.innerText = "✓";
  playerNameBox.appendChild(checkmark);
  playerNameBox.appendChild(playerName);

  // Append the new player container to the player list
  playerList.appendChild(playerContainer);

  if (playerInfo.Ready) {
    indicatePlayerReady(playerInfo.Name);
  }
}

// EVENT LISTENERS

function attachEventListeners() {
  playerNameDivs = document.querySelectorAll(".player-container");
  playerNameDivs.forEach((div, index) => {
    div.addEventListener("click", function (event) {
      event.stopPropagation(); // Prevent the document click event from firing
      playerRowClicked(div, index);
    });
  });

  document.addEventListener("click", function () {
    playerNameDivs.forEach((div) => {
      div.classList.remove("player-container-clicked");
    });
  });

  document.addEventListener("keydown", handleKeydown);
}

function removeEventListeners() {
  document.removeEventListener("keydown", handleKeydown);
  //   playerNameDivs.forEach((div, index) => {
  //     div.removeEventListener("click", function (event) {
  //       event.stopPropagation();
  //       playerRowClicked(div, index);
  //     });
  //   });
}
