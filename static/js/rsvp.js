// RSVP Form JavaScript — Per-Member Attending

(function () {
  "use strict";

  const mealOptions = [
    "Roasted Boneless Chicken Breast",
    "Grilled Brandt Farms 10z NY Strip",
    "Roasted Cauliflower Al Pastor (GF-V)",
  ];

  function toggleMealRow(row, show) {
    const mealRow = row.querySelector(".attendee-meal-row");
    if (!mealRow) return;
    mealRow.classList.toggle("hidden", !show);
  }

  function gatherFormData(form) {
    const rows = document.querySelectorAll("#attendee-rows .attendee-row");
    const attendees = [];
    let anyAttending = false;
    let allAttending = true;

    rows.forEach((row) => {
      const name = (row.querySelector(".attendee-name")?.textContent || "").trim();
      const attendingRadio = row.querySelector('.attendee-attending:checked');
      const decliningRadio = row.querySelector('.attendee-declining:checked');
      const attending = attendingRadio !== null;
      const meal = attending
        ? (row.querySelector(".attendee-meal")?.value || "").trim()
        : "";

      if (attending) anyAttending = true;
      if (!attending) allAttending = false;

      attendees.push({ name, attending, meal });
    });

    const attendingCount = attendees.filter((a) => a.attending).length;

    return {
      guest_id: form.dataset.guestId,
      attending: anyAttending,
      party_size: attendingCount,
      attendees,
      special_requests:
        document.getElementById("special-requests")?.value || "",
    };
  }

  function validateAttendees(attendees) {
    // Check that every member has responded
    for (let i = 0; i < attendees.length; i++) {
      const row = document.querySelectorAll("#attendee-rows .attendee-row")[i];
      const hasChoice =
        row?.querySelector('.attendee-attending:checked') !== null ||
        row?.querySelector('.attendee-declining:checked') !== null;
      if (!hasChoice) {
        return `Please select attending or declining for ${attendees[i].name || "guest " + (i + 1)}.`;
      }
    }

    // Validate meal for attending members
    for (const attendee of attendees) {
      if (!attendee.attending) continue;
      if (!attendee.meal) {
        return `Please select a meal for ${attendee.name}.`;
      }
      if (
        !mealOptions.some(
          (m) => m.toLowerCase() === attendee.meal.toLowerCase(),
        )
      ) {
        return `Please choose a valid meal option for ${attendee.name}.`;
      }
    }

    return "";
  }

  async function handleFormSubmit(e) {
    e.preventDefault();

    const form = e.target;
    const submitButton = document.getElementById("submit-button");

    const payload = gatherFormData(form);
    const attendeeError = validateAttendees(payload.attendees);
    if (attendeeError) {
      alert(attendeeError);
      return;
    }

    submitButton.disabled = true;
    const originalText = submitButton.textContent;
    submitButton.textContent = "Submitting...";

    try {
      const response = await fetch("/api/rsvp/submit", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });

      const result = await response.json();
      if (response.ok && result.success) {
        window.location.href = `/rsvp/success?attending=${encodeURIComponent(result.attending)}`;
      } else {
        throw new Error(result.error || "Failed to submit RSVP");
      }
    } catch (error) {
      console.error("RSVP submission error:", error);
      alert(
        "Sorry, there was an error submitting your RSVP. Please try again.",
      );
      submitButton.disabled = false;
      submitButton.textContent = originalText;
    }
  }

  function init() {
    const form = document.getElementById("rsvp-form");
    if (!form) return;

    // Per-row attending toggle: show/hide meal select
    const rows = document.querySelectorAll("#attendee-rows .attendee-row");
    rows.forEach((row) => {
      const radios = row.querySelectorAll('input[type="radio"]');
      radios.forEach((radio) => {
        radio.addEventListener("change", function () {
          toggleMealRow(row, this.value === "yes");
        });
      });
    });

    form.addEventListener("submit", handleFormSubmit);
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
