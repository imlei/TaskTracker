(function () {
  function initials(name) {
    if (!name || !name.trim()) return "?";
    var parts = name.split(/[\s,]+/).filter(Boolean);
    if (parts.length >= 2) {
      return (parts[0][0] + parts[parts.length - 1][0]).toUpperCase();
    }
    return name.slice(0, 2).toUpperCase();
  }

  function typeDisplay(emp) {
    var st = Number(emp.salaryType) === 0 ? "Salaried" : "Time-Based";
    var cat = Number(emp.category) === 0 ? "PERMANENT" : "CONTRACT";
    return st + " " + cat;
  }

  function row(emp) {
    var tr = document.createElement("tr");
    tr.innerHTML =
      '<td><div class="pe-name-cell"><span class="pe-avatar-sm">' +
      initials(emp.legalName) +
      '</span><div class="pe-name-block"><span class="name">' +
      escapeHtml(emp.legalName) +
      '</span><span class="id">' +
      escapeHtml(emp.id) +
      "</span></div></div></td>" +
      "<td>" +
      escapeHtml(emp.position || "") +
      "</td><td>" +
      escapeHtml(emp.payFrequency || "—") +
      '</td><td><span class="pe-dot-active">Active</span></td><td class="pe-type-cell">' +
      escapeHtml(typeDisplay(emp)) +
      '</td><td><button type="button" class="pe-kebab" aria-label="Actions">⋮</button></td>';
    return tr;
  }

  function escapeHtml(s) {
    var d = document.createElement("div");
    d.textContent = s;
    return d.innerHTML;
  }

  window.addEventListener("DOMContentLoaded", function () {
    var tb = document.getElementById("employee-rows");
    if (!tb) return;

    fetch("/api/employees")
      .then(function (r) {
        if (!r.ok) throw new Error("api");
        return r.json();
      })
      .then(function (list) {
        tb.innerHTML = "";
        list.forEach(function (emp) {
          tb.appendChild(row(emp));
        });
        var pending = document.getElementById("pending-count");
        if (pending) pending.textContent = "0";
      })
      .catch(function () {
        /* static rows remain */
      });
  });
})();
