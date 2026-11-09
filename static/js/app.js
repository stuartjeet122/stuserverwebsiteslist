(() => {
  "use strict";

  const root = document.documentElement;
  root.classList.remove("no-js");
  root.classList.add("js");

  const themeToggle = document.querySelector(".theme-toggle");
  let storedTheme = null;
  try {
    storedTheme = localStorage.getItem("stuserver-theme");
  } catch {
    // Theme persistence can be unavailable in private browsing contexts.
  }
  const systemPrefersLight = window.matchMedia("(prefers-color-scheme: light)").matches;
  const initialTheme = storedTheme === "light" || storedTheme === "dark"
    ? storedTheme
    : (systemPrefersLight ? "light" : "dark");
  root.dataset.theme = initialTheme;

  function updateThemeButton() {
    themeToggle?.setAttribute("aria-label", root.dataset.theme === "dark" ? "Switch to light theme" : "Switch to dark theme");
  }
  updateThemeButton();
  themeToggle?.addEventListener("click", () => {
    root.dataset.theme = root.dataset.theme === "dark" ? "light" : "dark";
    try {
      localStorage.setItem("stuserver-theme", root.dataset.theme);
    } catch {
      // The selected theme still applies for this page view.
    }
    updateThemeButton();
  });

  const menuToggle = document.querySelector(".mobile-menu-toggle");
  const navigation = document.querySelector("#primary-nav");
  menuToggle?.addEventListener("click", () => {
    const open = menuToggle.getAttribute("aria-expanded") !== "true";
    menuToggle.setAttribute("aria-expanded", String(open));
    menuToggle.setAttribute("aria-label", open ? "Close navigation" : "Open navigation");
    navigation?.classList.toggle("is-open", open);
  });
  navigation?.querySelectorAll("a").forEach((link) => {
    link.addEventListener("click", () => {
      navigation.classList.remove("is-open");
      menuToggle?.setAttribute("aria-expanded", "false");
      menuToggle?.setAttribute("aria-label", "Open navigation");
    });
  });

  const search = document.querySelector("#service-search");
  const clearSearch = document.querySelector(".clear-search");
  const categorySelect = document.querySelector("#category-filter");
  const featuredToggle = document.querySelector("[data-featured-filter]");
  const cards = [...document.querySelectorAll("[data-service-card]")];
  const noResults = document.querySelector("#no-results");
  const typeButtons = [...document.querySelectorAll("[data-type-filter]")];
  let activeType = "all";
  let featuredOnly = false;
  let activeCategory = "";

  function applyFilters() {
    const query = (search?.value || "").trim().toLocaleLowerCase();
    let visibleTotal = 0;
    const visibleByType = { website: 0, api: 0 };
    cards.forEach((card) => {
      const matches = (!query || (card.dataset.search || "").toLocaleLowerCase().includes(query))
        && (activeType === "all" || card.dataset.type === activeType)
        && (!activeCategory || card.dataset.category === activeCategory)
        && (!featuredOnly || card.dataset.featured === "true");
      card.hidden = !matches;
      if (matches && !card.closest(".featured-section")) {
        visibleTotal += 1;
        visibleByType[card.dataset.type] += 1;
      }
    });
    document.querySelectorAll("[data-visible-count]").forEach((count) => {
      count.textContent = String(visibleByType[count.dataset.visibleCount] || 0);
    });
    const featuredCount = document.querySelector("[data-featured-count]");
    if (featuredCount) {
      featuredCount.textContent = String(document.querySelectorAll(".featured-section [data-service-card]:not([hidden])").length);
    }
    document.querySelectorAll(".directory-section[id='websites'], .directory-section[id='apis']").forEach((section) => {
      const type = section.id === "websites" ? "website" : "api";
      section.hidden = activeType !== "all" && activeType !== type;
    });
    const featuredSection = document.querySelector(".featured-section");
    if (featuredSection) {
      featuredSection.hidden = !featuredSection.querySelector("[data-service-card]:not([hidden])");
    }
    if (noResults) noResults.hidden = visibleTotal !== 0;
    if (clearSearch) clearSearch.hidden = !search?.value;
  }

  search?.addEventListener("input", applyFilters);
  clearSearch?.addEventListener("click", () => {
    if (search) search.value = "";
    search?.focus();
    applyFilters();
  });
  categorySelect?.addEventListener("change", () => {
    activeCategory = categorySelect.value;
    document.querySelectorAll(".category-chip[data-category]").forEach((chip) => {
      chip.classList.toggle("is-selected", chip.dataset.category === activeCategory);
    });
    applyFilters();
  });
  document.querySelectorAll("button[data-category]").forEach((chip) => {
    chip.addEventListener("click", () => {
      activeCategory = activeCategory === chip.dataset.category ? "" : chip.dataset.category;
      if (categorySelect) categorySelect.value = activeCategory;
      document.querySelectorAll(".category-chip[data-category]").forEach((item) => {
        item.classList.toggle("is-selected", item.dataset.category === activeCategory);
      });
      applyFilters();
    });
  });
  typeButtons.forEach((button) => {
    button.addEventListener("click", () => {
      activeType = button.dataset.typeFilter || "all";
      typeButtons.forEach((item) => {
        const selected = item === button;
        item.classList.toggle("is-active", selected);
        item.setAttribute("aria-pressed", String(selected));
      });
      applyFilters();
    });
  });
  featuredToggle?.addEventListener("click", () => {
    featuredOnly = !featuredOnly;
    featuredToggle.setAttribute("aria-pressed", String(featuredOnly));
    applyFilters();
  });
  document.querySelectorAll("[data-reset-filters]").forEach((button) => {
    button.addEventListener("click", () => {
      if (search) search.value = "";
      if (categorySelect) categorySelect.value = "";
      activeCategory = "";
      activeType = "all";
      featuredOnly = false;
      featuredToggle?.setAttribute("aria-pressed", "false");
      typeButtons.forEach((item) => {
        const selected = item.dataset.typeFilter === "all";
        item.classList.toggle("is-active", selected);
        item.setAttribute("aria-pressed", String(selected));
      });
      document.querySelectorAll(".category-chip[data-category]").forEach((item) => item.classList.remove("is-selected"));
      applyFilters();
    });
  });
  document.addEventListener("keydown", (event) => {
    if (event.key === "/" && !["INPUT", "TEXTAREA", "SELECT"].includes(document.activeElement?.tagName || "")) {
      event.preventDefault();
      search?.focus();
    }
    if (event.key === "Escape" && search && document.activeElement === search) {
      search.value = "";
      applyFilters();
      search.blur();
    }
  });

  let toastTimer;
  function showToast(message) {
    document.querySelector(".toast")?.remove();
    const toast = document.createElement("div");
    toast.className = "toast";
    toast.setAttribute("role", "status");
    toast.textContent = message;
    document.body.append(toast);
    window.clearTimeout(toastTimer);
    toastTimer = window.setTimeout(() => toast.remove(), 2200);
  }
  document.querySelectorAll("[data-copy]").forEach((button) => {
    button.addEventListener("click", async () => {
      try {
        await navigator.clipboard.writeText(button.dataset.copy || "");
        showToast("API base URL copied");
      } catch {
        showToast("Could not copy URL — select and copy it manually");
      }
    });
  });

  const dialog = document.querySelector("#preview-dialog");
  const dialogImage = dialog?.querySelector(".dialog-image");
  const dialogTitle = dialog?.querySelector("#preview-title");
  const dialogDescription = dialog?.querySelector(".dialog-description");
  const dialogLink = dialog?.querySelector(".dialog-open");
  document.querySelectorAll(".preview-modal-trigger").forEach((trigger) => {
    const card = trigger.closest("[data-service-card]");
    const image = card?.querySelector(".preview-image");
    if (!card || !image || !dialog) return;
    trigger.addEventListener("click", () => {
      dialogImage.src = image.src;
      dialogImage.alt = image.alt;
      dialogTitle.textContent = card.querySelector("h3")?.textContent || "";
      dialogDescription.textContent = card.querySelector(".card-description")?.textContent || "";
      dialogLink.href = card.dataset.openUrl || card.querySelector(".preview-link")?.href || "#";
      dialog.showModal();
    });
  });
  dialog?.querySelector(".dialog-close")?.addEventListener("click", () => dialog.close());
  dialog?.addEventListener("click", (event) => {
    if (event.target === dialog) dialog.close();
  });
  dialog?.addEventListener("close", () => {
    if (dialogImage) dialogImage.src = "";
  });
  applyFilters();
})();
