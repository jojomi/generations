(function () {
  "use strict";

  function normalize(value) {
    return value
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .replace(/ß/g, "ss");
  }

  function renderResults(container, query, entries) {
    container.innerHTML = "";

    if (query === "") {
      container.classList.add("hidden");
      return;
    }

    var normalizedQuery = normalize(query);
    var matches = [];
    var seenURLs = {};

    for (var i = 0; i < entries.length; i++) {
      var entry = entries[i];
      if (seenURLs[entry.u]) {
        continue;
      }
      for (var j = 0; j < entry.f.length; j++) {
        var field = entry.f[j];
        if (normalize(field.v).indexOf(normalizedQuery) !== -1) {
          matches.push({ entry: entry, field: field });
          seenURLs[entry.u] = true;
          break;
        }
      }
      if (matches.length >= 20) {
        break;
      }
    }

    if (matches.length === 0) {
      var empty = document.createElement("div");
      empty.className = "px-3 py-2 text-gray-500";
      empty.textContent = "Keine Treffer für „" + query + "“";
      container.appendChild(empty);
      container.classList.remove("hidden");
      return;
    }

    matches.forEach(function (match) {
      var link = document.createElement("a");
      link.href = match.entry.u;
      link.className = "block px-3 py-2 hover:bg-gray-200 border-b border-gray-200";
      link.textContent = match.entry.n + " — " + match.field.v + " (" + match.field.l + ")";
      container.appendChild(link);
    });

    container.classList.remove("hidden");
  }

  document.addEventListener("DOMContentLoaded", function () {
    var input = document.getElementById("search-input");
    var results = document.getElementById("search-results");
    if (!input || !results) {
      return;
    }
    var entries = window.SEARCH_INDEX || [];

    input.addEventListener("input", function () {
      renderResults(results, input.value.trim(), entries);
    });

    input.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        input.value = "";
        results.classList.add("hidden");
        results.innerHTML = "";
      }
    });

    document.addEventListener("click", function (event) {
      if (event.target !== input && !results.contains(event.target)) {
        results.classList.add("hidden");
      }
    });
  });
})();
