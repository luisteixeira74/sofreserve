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
