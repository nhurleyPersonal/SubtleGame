document.addEventListener("DOMContentLoaded", function () {
  if (document.getElementById("submit-button")) {
    const playerName = document.getElementById("submit-button");
    playerName.addEventListener("click", function (event) {
      console.log("Screen clicked");
    });
  }
});
