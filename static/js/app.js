(() => {
  "use strict";

  document.documentElement.classList.remove("no-js");
  document.documentElement.classList.add("js");

  const search = document.querySelector("#service-search");
  const clearSearch = document.querySelector(".clear-search");
  const categoryButtons = [...document.querySelectorAll("[data-category]")];
  const featuredToggle = document.querySelector("[data-featured-filter]");
  const cards = [...document.querySelectorAll("[data-service-card]")];
  const sections = [...document.querySelectorAll(".service-section")];
  const noResults = document.querySelector("#no-results");
  const typeTabs = [...document.querySelectorAll("[data-type-filter]")];
  let activeType = "website";
  let activeCategory = "";
  let featuredOnly = false;

  function applyFilters() {
    const query = (search?.value || "").trim().toLocaleLowerCase();
    let visibleCount = 0;
    const categoryCounts = new Map();
    const countsByType = { website: 0, api: 0 };

    cards.forEach((card) => {
      const text = (card.dataset.search || "").toLocaleLowerCase();
      const matchesQuery = !query || text.includes(query);
      const matchesCategory = !activeCategory || card.dataset.category === activeCategory;
      const matchesFeatured = !featuredOnly || card.dataset.featured === "true";
      const matches = card.dataset.type === activeType && matchesQuery && matchesCategory && matchesFeatured;
      card.hidden = !matches;

      if (card.dataset.type === activeType && matchesQuery && matchesFeatured) {
        const category = card.dataset.category;
        if (category) categoryCounts.set(category, (categoryCounts.get(category) || 0) + 1);
      }
      if (matches) {
        visibleCount += 1;
        countsByType[activeType] += 1;
      }
    });

    sections.forEach((section) => {
      section.hidden = section.id !== (activeType === "website" ? "websites" : "apis");
    });
    document.querySelectorAll("[data-visible-count]").forEach((count) => {
      count.textContent = String(countsByType[count.closest(".service-section")?.id === "websites" ? "website" : "api"] || 0);
    });
    categoryButtons.forEach((button) => {
      const category = button.dataset.category || "";
      button.hidden = Boolean(category) && !categoryCounts.has(category);
      button.classList.toggle("is-selected", category === activeCategory);
    });
    if (noResults) noResults.hidden = visibleCount > 0 || cards.every((card) => card.dataset.type !== activeType);
    if (clearSearch) clearSearch.hidden = !search?.value;
  }

  search?.addEventListener("input", applyFilters);
  clearSearch?.addEventListener("click", () => {
    if (search) search.value = "";
    search?.focus();
    applyFilters();
  });
  categoryButtons.forEach((button) => {
    button.addEventListener("click", () => {
      activeCategory = button.dataset.category || "";
      applyFilters();
    });
  });
  typeTabs.forEach((tab) => {
    tab.addEventListener("click", (event) => {
      event.preventDefault();
      activeType = tab.dataset.typeFilter || "website";
      typeTabs.forEach((item) => {
        const selected = item === tab;
        item.classList.toggle("is-active", selected);
        if (selected) item.setAttribute("aria-current", "page");
        else item.removeAttribute("aria-current");
      });
      activeCategory = "";
      featuredOnly = false;
      featuredToggle?.setAttribute("aria-pressed", "false");
      history.replaceState(null, "", `#${activeType === "website" ? "websites" : "apis"}`);
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
      activeCategory = "";
      featuredOnly = false;
      featuredToggle?.setAttribute("aria-pressed", "false");
      applyFilters();
      search?.focus();
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

  const initialType = window.location.hash === "#apis" ? "api" : "website";
  const initialTab = typeTabs.find((tab) => tab.dataset.typeFilter === initialType);
  if (initialTab) {
    activeType = initialType;
    typeTabs.forEach((tab) => {
      const selected = tab === initialTab;
      tab.classList.toggle("is-active", selected);
      if (selected) tab.setAttribute("aria-current", "page");
      else tab.removeAttribute("aria-current");
    });
  }
  applyFilters();
})();
