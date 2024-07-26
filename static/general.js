let selectedBox;
let selectedBoxIndex;
let currentBoxIndex = 0;
let playerNameDivs;

// Define the keydown event handler
function handleKeydown(event) {
  console.log(event);
  const key = event.key;
  const lettersContainer = selectedBox.querySelector(".letters-container");
  const letterBoxes = lettersContainer.querySelectorAll(".letter");

  //   if (key === "Tab") {
  //     currentBoxIndex = (currentBoxIndex + 1) % letterBoxes.length;
  //     selectedBox = playerNameDivs[selectedBoxIndex];
  //     playerRowClicked(selectedBox, selectedBoxIndex);
  //   }
  // Only process single character keys
  if (key.length === 1 && /[a-zA-Z0-9]/.test(key)) {
    // Fill the current box with the typed letter
    letterBoxes[currentBoxIndex].innerText = key;

    // Remove highlight from the current box
    letterBoxes[currentBoxIndex].classList.remove("letter-highlight");

    // Move to the next box
    currentBoxIndex = (currentBoxIndex + 1) % letterBoxes.length;

    // Highlight the next box
    letterBoxes[currentBoxIndex].classList.add("letter-highlight");
    for (let i = 0; i < letterBoxes.length; i++) {
      if (i !== currentBoxIndex) {
        letterBoxes[i].classList.remove("letter-highlight");
      }
    }
  }
}

function playerRowClicked(row, index) {
  selectedBox = row;
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

function deleteCurrentPlayers() {
  const playerList = document.getElementById("player-list");
  playerList.innerHTML = "";
  playerNameDivs.forEach((div) => {
    div.remove();
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
  var lettersContainer = document.createElement("div");
  lettersContainer.className = "letters-container";
  for (let i = 0; i < 6; i++) {
    var letter = document.createElement("div");
    letter.className = "letter";
    lettersContainer.appendChild(letter);
  }
  playerContainer.appendChild(lettersContainer);

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

  var playerList = document.getElementById("player-list");

  // Create a new player container
  var playerContainer = document.createElement("div");
  playerContainer.className = "player-container";

  // Create player name element
  var playerName = document.createElement("h2");
  playerName.className = "player-name";
  playerName.innerText = playerInfo.Name || "Unknown";
  playerContainer.appendChild(playerName);

  // Append the new player container to the player list
  playerList.appendChild(playerContainer);
}
