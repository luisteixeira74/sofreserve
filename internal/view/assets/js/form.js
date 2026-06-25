// =====================
// UI STATE MANAGER
// =====================

const UI = {
  loadingOverlay: null,

  loading: {
    start() {
      let overlay = document.getElementById("loading-overlay");

      if (!overlay) {
        overlay = document.createElement("div");
        overlay.id = "loading-overlay";
        overlay.innerHTML = `
          <div class="loading-box">
            <span class="prompt">&gt;</span> processando...
          </div>
        `;
        document.body.appendChild(overlay);
      }

      overlay.style.display = "flex";
      UI.loadingOverlay = overlay;
    },

    stop() {
      const overlay = document.getElementById("loading-overlay");
      if (overlay) overlay.style.display = "none";
    },
  },
};

// =====================
// FORM HANDLING (CRITICAL)
// =====================

document.addEventListener("DOMContentLoaded", () => {
  const forms = document.querySelectorAll("form[data-ux='critical-submit']");

  forms.forEach((form) => {
    form.addEventListener("submit", (e) => {
      if (form.dataset.submitting === "true") {
        e.preventDefault();
        return;
      }

      form.dataset.submitting = "true";
      form.classList.add("loading");

      const btn = form.querySelector("[data-loading='true']");

      if (btn) {
        btn.disabled = true;
        btn.dataset.originalText = btn.innerText || btn.value;

        if (btn.tagName === "BUTTON") {
          btn.innerText = "executando...";
        } else {
          btn.value = "executando...";
        }
      }

      UI.loading.start();
    });
  });
});

// =====================
// SAFE RESET (IMPORTANT)
// =====================

// garante que nunca fica travado após redirect
window.addEventListener("pageshow", () => {
  UI.loading.stop();

  document.querySelectorAll("form").forEach((form) => {
    form.dataset.submitting = "false";
    form.classList.remove("loading");
  });
});

// =====================
// COPY LINK
// =====================

document.addEventListener("DOMContentLoaded", () => {
  const inputs = document.querySelectorAll(".copy-input");

  inputs.forEach((input) => {
    input.addEventListener("click", async () => {
      if (input.dataset.ready && input.dataset.ready !== "true") {
        return;
      }

      try {
        await navigator.clipboard.writeText(input.value);

        const feedback = input.parentElement.querySelector(".copy-feedback");

        if (feedback) {
          feedback.classList.add("show");

          setTimeout(() => {
            feedback.classList.remove("show");
          }, 1200);
        }
      } catch (err) {
        input.select();
        document.execCommand("copy");
      }
    });
  });
});

// =====================
// TYPING EFFECT
// =====================

document.addEventListener("DOMContentLoaded", () => {
  const inputs = document.querySelectorAll(".dynamic-link");

  inputs.forEach((input) => {
    input.dataset.ready = "false";
    input.disabled = true;

    const text = input.dataset.link || "";

    if (!text) {
      input.dataset.ready = "true";
      input.disabled = false;
      return;
    }

    input.value = "";

    let i = 0;

    setTimeout(() => {
      const interval = setInterval(() => {
        input.value += text[i];
        i++;

        if (i >= text.length) {
          clearInterval(interval);

          input.dataset.ready = "true";
          input.disabled = false;

          input.classList.remove("typing");
          input.onclick = () => input.select();
        }
      }, 20);
    }, 700);
  });
});

// =====================
// HERO FLOW
// =====================

const steps = document.querySelectorAll(".hero-flow .step");

if (steps.length > 0) {
  let i = 0;

  function activateStep(index) {
    steps.forEach((s) => s.classList.remove("active"));
    steps[index].classList.add("active");
  }

  setTimeout(() => {
    activateStep(0);

    setInterval(() => {
      i = (i + 1) % steps.length;
      activateStep(i);
    }, 3000);
  }, 6000);
}

// =====================
// UTILITIES
// =====================

function setSeats(value) {
  const input = document.getElementById("seats");
  if (input) input.value = value;
}

function changeSeats(delta) {
  const input = document.getElementById("seats");
  if (!input) return;

  let current = parseInt(input.value || 0, 10);

  current += delta;

  if (current < 1) current = 1;
  if (current > 999) current = 999;

  input.value = current;
}
