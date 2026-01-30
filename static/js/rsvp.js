// RSVP Form JavaScript

(function () {
  "use strict";

  // Toggle attending details visibility
  function toggleAttendingDetails(show) {
    const details = document.getElementById("attending-details");
    if (details) {
      if (show) {
        details.classList.remove("hidden");
      } else {
        details.classList.add("hidden");
      }
    }
  }

  // Update attendee name inputs based on party size
  function updateAttendeeNames(partySize) {
    const container = document.getElementById("attendee-names");
    if (!container) return;

    const currentInputs = container.querySelectorAll("input");
    const currentCount = currentInputs.length;
    const newCount = parseInt(partySize, 10);

    if (newCount > currentCount) {
      // Add more inputs
      for (let i = currentCount; i < newCount; i++) {
        const input = document.createElement("input");
        input.type = "text";
        input.name = "attendee_names[]";
        input.placeholder = `Guest ${i + 1} name`;
        input.className =
          "w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-300 focus:border-transparent";
        container.appendChild(input);
      }
    } else if (newCount < currentCount) {
      // Remove extra inputs
      for (let i = currentCount - 1; i >= newCount; i--) {
        container.removeChild(currentInputs[i]);
      }
    }
  }

  // Gather form data for submission
  function gatherFormData(form) {
    const formData = new FormData(form);
    const attending = formData.get("attending") === "yes";

    const data = {
      guest_id: form.dataset.guestId,
      attending: attending,
      party_size: attending ? parseInt(formData.get("party_size"), 10) : 0,
      attendee_names: [],
      dietary_restrictions: [],
      special_requests: "",
    };

    if (attending) {
      // Get attendee names
      const names = formData.getAll("attendee_names[]");
      data.attendee_names = names.filter((name) => name.trim() !== "");

      // Get dietary restrictions (split by newlines)
      const dietary = formData.get("dietary_restrictions");
      if (dietary) {
        data.dietary_restrictions = dietary
          .split("\n")
          .map((line) => line.trim())
          .filter((line) => line !== "");
      }

      // Get special requests
      data.special_requests = formData.get("special_requests") || "";
    }

    return data;
  }

  // Handle form submission
  async function handleFormSubmit(e) {
    e.preventDefault();

    const form = e.target;
    const submitButton = document.getElementById("submit-button");

    // Validate attending choice is selected
    const attending = form.querySelector('input[name="attending"]:checked');
    if (!attending) {
      alert("Please select whether you will be attending.");
      return;
    }

    // Disable submit button
    submitButton.disabled = true;
    const originalText = submitButton.textContent;
    submitButton.textContent = "Submitting...";

    try {
      const data = gatherFormData(form);

      const response = await fetch("/api/rsvp/submit", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (response.ok && result.success) {
        // Redirect to success page
        window.location.href = "/rsvp/success";
      } else {
        throw new Error(result.error || "Failed to submit RSVP");
      }
    } catch (error) {
      console.error("RSVP submission error:", error);
      alert(
        "Sorry, there was an error submitting your RSVP. Please try again."
      );
      submitButton.disabled = false;
      submitButton.textContent = originalText;
    }
  }

  // Initialize the form
  function init() {
    const form = document.getElementById("rsvp-form");
    if (!form) return;

    // Handle attending radio button changes
    const attendingRadios = form.querySelectorAll('input[name="attending"]');
    attendingRadios.forEach((radio) => {
      radio.addEventListener("change", function () {
        toggleAttendingDetails(this.value === "yes");
      });
    });

    // Handle party size changes
    const partySizeSelect = document.getElementById("party-size-select");
    if (partySizeSelect) {
      partySizeSelect.addEventListener("change", function () {
        updateAttendeeNames(this.value);
      });
    }

    // Handle form submission
    form.addEventListener("submit", handleFormSubmit);

    // Set initial state based on existing selection
    const checkedRadio = form.querySelector('input[name="attending"]:checked');
    if (checkedRadio) {
      toggleAttendingDetails(checkedRadio.value === "yes");

      // Initialize attendee names if attending
      if (checkedRadio.value === "yes" && partySizeSelect) {
        // Give a small delay to ensure DOM is ready
        setTimeout(() => {
          const currentInputs = document.querySelectorAll(
            "#attendee-names input"
          );
          const partySize = parseInt(partySizeSelect.value, 10);
          if (currentInputs.length < partySize) {
            updateAttendeeNames(partySize);
          }
        }, 0);
      }
    }
  }

  // Run initialization when DOM is ready
  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
