/**
 * Wedding Countdown Timer
 * Days remaining until May 30, 2026
 */

(function () {
  "use strict";

  let weddingDate = new Date("2026-05-30T17:30:00-04:00");
  let timerEl;

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }

  function init() {
    timerEl = document.getElementById("countdown-timer");
    if (!timerEl) return;
    update();
  }

  function update() {
    let now = new Date();
    let diff = weddingDate - now;

    if (diff <= 0) {
      timerEl.textContent = "Today is the day!";
      return;
    }

    let days = Math.ceil(diff / (1000 * 60 * 60 * 24));

    timerEl.innerHTML =
      '<span class="font-bold text-gray-700">' +
      days +
      "</span> days to go";
  }
})();
