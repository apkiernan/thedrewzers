// Admin RSVP Edit — Meal row toggle based on attending radio

(function () {
  "use strict";

  function toggleMealRow(row, show) {
    var mealRow = row.querySelector(".attendee-meal-row");
    if (!mealRow) return;
    mealRow.classList.toggle("hidden", !show);
  }

  function init() {
    var rows = document.querySelectorAll("#edit-rsvp-form .attendee-row");
    rows.forEach(function (row) {
      var radios = row.querySelectorAll('input[type="radio"]');
      radios.forEach(function (radio) {
        radio.addEventListener("change", function () {
          toggleMealRow(row, this.value === "yes");
        });
      });
    });
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
