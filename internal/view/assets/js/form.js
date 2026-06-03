document.addEventListener("DOMContentLoaded", () => {
  const forms = document.querySelectorAll("form");

  forms.forEach((form) => {
    form.addEventListener("submit", (e) => {
      // evita múltiplos submits
      if (form.dataset.submitting === "true") {
        e.preventDefault();
        return;
      }

      form.dataset.submitting = "true";

      const btn = form.querySelector("[data-loading='true']");

      if (btn) {
        btn.disabled = true;
        btn.dataset.originalText = btn.innerText || btn.value;

        if (btn.tagName === "BUTTON") {
          btn.innerText = "Carregando...";
        } else {
          btn.value = "Carregando...";
        }
      }

      showLoadingOverlay();
    });
  });
});

function showLoadingOverlay() {
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
}

function setSeats(value) {
  const input = document.getElementById("seats");
  input.value = value;
}

function changeSeats(delta) {
  const input = document.getElementById("seats");
  let current = parseInt(input.value || 0, 10);

  current += delta;

  if (current < 1) current = 1;
  if (current > 999) current = 999;

  input.value = current;
}

// =====================
// TYPING EFFECT (CONFIRM LINK)
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
// CANCEL CONFIRM FLOW
// =====================

document.addEventListener("DOMContentLoaded", () => {
  const btn = document.getElementById("cancel-btn");
  const msg = document.getElementById("confirm-msg");
  const actions = document.getElementById("confirm-actions");

  const yes = document.getElementById("confirm-yes");
  const no = document.getElementById("confirm-no");

  if (!btn) return;

  // pega token do HTML (correto)
  const token = btn.dataset.token;

  btn.addEventListener("click", () => {
    msg.style.display = "block";
    actions.style.display = "block";
    btn.style.display = "none";
  });

  yes.addEventListener("click", () => {
    window.location.href = "/cancel?token=" + token;
  });

  no.addEventListener("click", () => {
    msg.style.display = "none";
    actions.style.display = "none";
    btn.style.display = "inline-block";
  });
});

// =====================
// COPY LINK
// =====================

document.addEventListener("DOMContentLoaded", () => {
  const inputs = document.querySelectorAll(".copy-input");

  inputs.forEach((input) => {
    input.addEventListener("click", async () => {
      if (input.dataset.ready !== "true") {
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

const steps = document.querySelectorAll(".hero-flow .step");

let i = 0;

function activateStep(index) {
  steps.forEach((s) => s.classList.remove("active"));
  steps[index].classList.add("active");
}

// delay inicial (6s)
setTimeout(() => {
  activateStep(0);

  setInterval(() => {
    i = (i + 1) % steps.length;
    activateStep(i);
  }, 3000);
}, 6000);
