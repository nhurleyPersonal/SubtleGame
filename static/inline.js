console.log("HERE@!!!");
document.addEventListener("DOMContentLoaded", function () {
  console.log("HERE2");
  if (document.getElementById("submit-button")) {
    console.log("HERE3");
    const playerName = document.getElementById("submit-button");
    playerName.addEventListener("click", function (event) {
      console.log("Screen clicked");
    });
  }
});
