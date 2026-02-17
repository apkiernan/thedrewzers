// RSVP Form JavaScript

(function () {
  "use strict";

  const mealOptions = ["chicken", "steak", "vegetarian"];

  function toggleAttendingDetails(show) {
    const details = document.getElementById("attending-details");
    if (!details) return;
    details.classList.toggle("hidden", !show);
  }

  function mealLabel(meal) {
    if (!meal) return "";
    return meal.charAt(0).toUpperCase() + meal.slice(1);
  }

  function currentAttendees() {
    const rows = document.querySelectorAll("#attendee-rows .attendee-row");
    return Array.from(rows).map((row) => ({
      name: (row.querySelector(".attendee-name")?.value || "").trim(),
      meal: (row.querySelector(".attendee-meal")?.value || "").trim().toLowerCase(),
    }));
  }

  function createAttendeeRow(index, attendee) {
    const row = document.createElement("div");
    row.className = "attendee-row grid grid-cols-1 md:grid-cols-2 gap-3";
    row.dataset.index = String(index);

    const nameInput = document.createElement("input");
    nameInput.type = "text";
    nameInput.className =
      "attendee-name w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent";
    nameInput.placeholder = `Guest ${index + 1} name`;
    nameInput.value = attendee.name || "";

    const mealSelect = document.createElement("select");
    mealSelect.className =
      "attendee-meal w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent bg-white";

    const placeholder = document.createElement("option");
    placeholder.value = "";
    placeholder.textContent = "Select a meal";
    mealSelect.appendChild(placeholder);

    mealOptions.forEach((meal) => {
      const option = document.createElement("option");
      option.value = meal;
      option.textContent = mealLabel(meal);
      if ((attendee.meal || "").toLowerCase() === meal) {
        option.selected = true;
      }
      mealSelect.appendChild(option);
    });

    row.appendChild(nameInput);
    row.appendChild(mealSelect);
    return row;
  }

  function updateAttendeeRows(partySize) {
    const container = document.getElementById("attendee-rows");
    if (!container) return;

    const count = Number.parseInt(partySize, 10);
    if (!Number.isFinite(count) || count < 1) return;

    const existing = currentAttendees();
    container.innerHTML = "";

    for (let i = 0; i < count; i++) {
      const attendee = existing[i] || { name: "", meal: "" };
      container.appendChild(createAttendeeRow(i, attendee));
    }
  }

  function gatherFormData(form) {
    const formData = new FormData(form);
    const attending = formData.get("attending") === "yes";

    const data = {
      guest_id: form.dataset.guestId,
      attending,
      party_size: 0,
      attendees: [],
      special_requests: "",
    };

    if (attending) {
      const attendees = currentAttendees();
      data.attendees = attendees;
      data.party_size = attendees.length;
      data.special_requests = formData.get("special_requests") || "";
    }

    return data;
  }

  function validateAttendees(attendees) {
    if (attendees.length === 0) {
      return "Please add at least one attending guest.";
    }

    for (let i = 0; i < attendees.length; i++) {
      const attendee = attendees[i];
      if (!attendee.name) {
        return `Please enter a name for guest ${i + 1}.`;
      }
      if (!attendee.meal) {
        return `Please select a meal for ${attendee.name || `guest ${i + 1}`}.`;
      }
      if (!mealOptions.includes(attendee.meal)) {
        return `Please choose a valid meal option for ${attendee.name}.`;
      }
    }

    return "";
  }

  async function handleFormSubmit(e) {
    e.preventDefault();

    const form = e.target;
    const submitButton = document.getElementById("submit-button");

    const attending = form.querySelector('input[name="attending"]:checked');
    if (!attending) {
      alert("Please select whether you will be attending.");
      return;
    }

    const payload = gatherFormData(form);
    if (payload.attending) {
      const attendeeError = validateAttendees(payload.attendees);
      if (attendeeError) {
        alert(attendeeError);
        return;
      }
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
        window.location.href = "/rsvp/success";
      } else {
        throw new Error(result.error || "Failed to submit RSVP");
      }
    } catch (error) {
      console.error("RSVP submission error:", error);
      alert("Sorry, there was an error submitting your RSVP. Please try again.");
      submitButton.disabled = false;
      submitButton.textContent = originalText;
    }
  }

  function init() {
    const form = document.getElementById("rsvp-form");
    if (!form) return;

    const attendingRadios = form.querySelectorAll('input[name="attending"]');
    attendingRadios.forEach((radio) => {
      radio.addEventListener("change", function () {
        toggleAttendingDetails(this.value === "yes");
      });
    });

    const partySizeSelect = document.getElementById("party-size-select");
    if (partySizeSelect) {
      partySizeSelect.addEventListener("change", function () {
        updateAttendeeRows(this.value);
      });
    }

    form.addEventListener("submit", handleFormSubmit);

    const checkedRadio = form.querySelector('input[name="attending"]:checked');
    if (checkedRadio) {
      toggleAttendingDetails(checkedRadio.value === "yes");
      if (checkedRadio.value === "yes" && partySizeSelect) {
        updateAttendeeRows(partySizeSelect.value);
      }
    }
  }

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
