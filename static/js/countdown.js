/**
 * Wedding Countdown Timer
 * Live countdown to the ceremony at 5:30 PM ET on May 30, 2026
 */

(function () {
  "use strict";

  let weddingDate = new Date("2026-05-30T17:30:00-04:00");
  let timerEl;
  let intervalId;

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  function init() {
    timerEl = document.getElementById("countdown-timer");
    if (!timerEl) return;
    update();
    intervalId = setInterval(update, 1000);
  }

  function update() {
    let now = new Date();
    let diff = weddingDate - now;

    if (diff <= 0) {
      clearInterval(intervalId);
      timerEl.textContent = "Today is the day!";
      return;
    }

    let days = Math.floor(diff / (1000 * 60 * 60 * 24));
    let hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
    let minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
    let seconds = Math.floor((diff % (1000 * 60)) / 1000);

    timerEl.innerHTML =
      '<span class="font-bold text-gray-700">' +
      days +
      "</span> days " +
      '<span class="font-bold text-gray-700">' +
      hours +
      "</span> hours " +
      '<span class="font-bold text-gray-700">' +
      minutes +
      "</span> minutes " +
      '<span class="font-bold text-gray-700">' +
      seconds +
      "</span> seconds";
  }
})();
