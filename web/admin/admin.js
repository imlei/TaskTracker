/* Admin module — shared JS for login.html and index.html */

// ─── Utilities ────────────────────────────────────────────────────────────────

async function apiFetch(path, opts = {}) {
  const r = await fetch(path, {
    credentials: "same-origin",
    headers: { "Content-Type": "application/json; charset=utf-8", ...opts.headers },
    ...opts,
  });
  const text = await r.text();
  if (!r.ok) {
    let msg = text;
    try { msg = JSON.parse(text)?.error || text; } catch { /* ignore */ }
    throw new Error(msg || r.statusText);
  }
  return text ? JSON.parse(text) : null;
}

function escHtml(s) {
  const d = document.createElement("div");
  d.textContent = String(s ?? "");
  return d.innerHTML;
}

function showToast(message, type = "info") {
  const wrap = document.createElement("div");
  wrap.className = "toast toast-top toast-end z-[9999]";
  wrap.innerHTML = `<div class="alert alert-${type} text-sm py-2 shadow-lg">${escHtml(message)}</div>`;
  document.body.appendChild(wrap);
  setTimeout(() => wrap.remove(), 3500);
}

// ─── Login Page ───────────────────────────────────────────────────────────────

function initLoginPage() {
  // Already signed in as admin? Skip login.
  fetch("/api/me", { credentials: "same-origin" })
    .then((r) => r.json())
    .then((me) => { if (me.authenticated && me.role === "admin") window.location.href = "/admin/"; })
    .catch(() => { /* ignore — stay on login page */ });

  const form = document.getElementById("login-form");
  const alertEl = document.getElementById("error-alert");
  const alertText = document.getElementById("error-text");
  const btn = document.getElementById("btn-submit");

  function showError(msg) {
    alertText.textContent = msg;
    alertEl.classList.remove("hidden");
  }

  form?.addEventListener("submit", async (e) => {
    e.preventDefault();
    alertEl.classList.add("hidden");
    btn.classList.add("loading", "loading-spinner");
    btn.disabled = true;

    try {
      const data = await apiFetch("/api/login", {
        method: "POST",
        body: JSON.stringify({
          username: document.getElementById("username").value.trim(),
          password: document.getElementById("password").value,
        }),
      });

      if (data?.role !== "admin") {
        // Authenticated but not admin — immediately invalidate the session.
        await fetch("/api/logout", { method: "POST", credentials: "same-origin" }).catch(() => {});
        showError("Access denied. This portal requires an admin account.");
        return;
      }

      window.location.href = "/admin/";
    } catch (err) {
      showError(err.message || "Sign-in failed. Please try again.");
    } finally {
      btn.classList.remove("loading", "loading-spinner");
      btn.disabled = false;
    }
  });
}

// ─── Dashboard Page ───────────────────────────────────────────────────────────

let cachedUsers = [];

function updateStats(users) {
  document.getElementById("stat-total").textContent  = users.length;
  document.getElementById("stat-user2").textContent  = users.filter((u) => u.role === "user2").length;
  document.getElementById("stat-user1").textContent  = users.filter((u) => u.role === "user1").length;
}

function renderUsers(users) {
  const tbody = document.getElementById("users-body");
  if (!tbody) return;

  if (users.length === 0) {
    tbody.innerHTML = `
      <tr><td colspan="4" class="text-center py-12 text-base-content/30 text-sm">
        No users yet. Click <strong>New User</strong> to create one.
      </td></tr>`;
    return;
  }

  tbody.innerHTML = users.map((u) => {
    const initials = u.username.slice(0, 2).toUpperCase();
    const isAdmin  = u.role === "admin";

    const roleBadge = isAdmin
      ? `<span class="badge badge-primary badge-sm font-medium">Admin</span>`
      : `<select class="select select-bordered select-xs role-sel min-w-[10rem]" data-username="${escHtml(u.username)}" data-current="${escHtml(u.role)}">
           <option value="user2" ${u.role === "user2" ? "selected" : ""}>Level 2 — Full Access</option>
           <option value="user1" ${u.role === "user1" ? "selected" : ""}>Level 1 — Payroll Only</option>
         </select>`;

    const deleteBtn = isAdmin ? "" : `
      <button class="btn btn-ghost btn-xs text-error gap-1 btn-del" data-username="${escHtml(u.username)}" title="Delete user">
        <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
            d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
        </svg>
        Delete
      </button>`;

    return `
      <tr>
        <td class="w-12 pl-4">
          <div class="avatar placeholder">
            <div class="bg-base-200 text-base-content/60 rounded-full w-9">
              <span class="text-xs font-bold">${initials}</span>
            </div>
          </div>
        </td>
        <td>
          <span class="font-medium text-sm text-base-content">${escHtml(u.username)}</span>
        </td>
        <td>${roleBadge}</td>
        <td class="text-right pr-4">
          <div class="flex items-center justify-end gap-1.5">
            <button class="btn btn-ghost btn-xs gap-1 btn-reset" data-username="${escHtml(u.username)}" title="Reset password">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-3.5 w-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>
              </svg>
              Reset pwd
            </button>
            ${deleteBtn}
          </div>
        </td>
      </tr>`;
  }).join("");

  // ── Role change ──────────────────────────────────────────────────
  tbody.querySelectorAll(".role-sel").forEach((sel) => {
    sel.addEventListener("change", async () => {
      const { username, current } = sel.dataset;
      try {
        await apiFetch("/api/users", {
          method: "PUT",
          body: JSON.stringify({ username, role: sel.value }),
        });
        sel.dataset.current = sel.value;
        showToast(`Role updated for ${username}`, "success");
      } catch (err) {
        sel.value = current; // revert
        showToast(err.message || "Failed to update role", "error");
      }
    });
  });

  // ── Reset password ───────────────────────────────────────────────
  tbody.querySelectorAll(".btn-reset").forEach((btn) => {
    btn.addEventListener("click", () => {
      document.getElementById("reset-target-name").textContent   = btn.dataset.username;
      document.getElementById("reset-new-pwd").dataset.username  = btn.dataset.username;
      document.getElementById("reset-new-pwd").value             = "";
      document.getElementById("reset-error").classList.add("hidden");
      document.getElementById("modal-reset").showModal();
    });
  });

  // ── Delete user ──────────────────────────────────────────────────
  tbody.querySelectorAll(".btn-del").forEach((btn) => {
    btn.addEventListener("click", async () => {
      const { username } = btn.dataset;
      if (!confirm(`Delete user "${username}"?\n\nThis action cannot be undone.`)) return;
      try {
        await apiFetch("/api/users", {
          method: "DELETE",
          body: JSON.stringify({ username }),
        });
        showToast(`User "${username}" deleted`, "success");
        loadUsers();
      } catch (err) {
        showToast(err.message || "Failed to delete user", "error");
      }
    });
  });
}

async function loadUsers() {
  const tbody = document.getElementById("users-body");
  if (tbody) {
    tbody.innerHTML = `
      <tr><td colspan="4" class="text-center py-10 text-base-content/30">
        <span class="loading loading-spinner loading-md"></span>
      </td></tr>`;
  }
  try {
    cachedUsers = await apiFetch("/api/users") ?? [];
    renderUsers(cachedUsers);
    updateStats(cachedUsers);
  } catch (err) {
    if (tbody) {
      tbody.innerHTML = `
        <tr><td colspan="4" class="text-center py-10 text-error text-sm">
          Failed to load users: ${escHtml(err.message)}
        </td></tr>`;
    }
  }
}

function setAdminInfo(username) {
  const initials = (username || "AD").slice(0, 2).toUpperCase();
  const els = {
    "nav-avatar":       initials,
    "nav-username":     username || "",
    "sidebar-avatar":   initials,
    "sidebar-username": username || "Admin",
  };
  for (const [id, val] of Object.entries(els)) {
    const el = document.getElementById(id);
    if (el) el.textContent = val;
  }
}

function initDashboard() {
  // ── Auth guard ───────────────────────────────────────────────────
  fetch("/api/me", { credentials: "same-origin" })
    .then((r) => r.json())
    .then((me) => {
      if (!me.authenticated || me.role !== "admin") {
        window.location.href = "/admin/login.html";
        return;
      }
      setAdminInfo(me.user);
      loadUsers();
    })
    .catch(() => { window.location.href = "/admin/login.html"; });

  // ── Logout ───────────────────────────────────────────────────────
  document.getElementById("btn-logout")?.addEventListener("click", async () => {
    await fetch("/api/logout", { method: "POST", credentials: "same-origin" }).catch(() => {});
    window.location.href = "/admin/login.html";
  });

  // ── Create user modal ────────────────────────────────────────────
  const createModal = document.getElementById("modal-create");

  document.getElementById("btn-create-user")?.addEventListener("click", () => {
    document.getElementById("form-create").reset();
    document.getElementById("create-error").classList.add("hidden");
    createModal?.showModal();
  });

  document.getElementById("btn-create-cancel")?.addEventListener("click", () => createModal?.close());

  document.getElementById("form-create")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl = document.getElementById("create-error");
    errEl.classList.add("hidden");

    const payload = {
      username: document.getElementById("new-username").value.trim(),
      password: document.getElementById("new-password").value,
      role:     document.getElementById("new-role").value,
    };

    try {
      await apiFetch("/api/users", { method: "POST", body: JSON.stringify(payload) });
      createModal?.close();
      showToast(`User "${payload.username}" created`, "success");
      loadUsers();
    } catch (err) {
      errEl.textContent = err.message || "Failed to create user";
      errEl.classList.remove("hidden");
    }
  });

  // ── Reset password modal ─────────────────────────────────────────
  const resetModal = document.getElementById("modal-reset");

  document.getElementById("btn-reset-cancel")?.addEventListener("click", () => resetModal?.close());

  document.getElementById("form-reset")?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const errEl   = document.getElementById("reset-error");
    const pwdInput = document.getElementById("reset-new-pwd");
    errEl.classList.add("hidden");

    try {
      await apiFetch("/api/users", {
        method: "PUT",
        body: JSON.stringify({
          username:    pwdInput.dataset.username,
          newPassword: pwdInput.value,
        }),
      });
      resetModal?.close();
      showToast("Password reset successfully", "success");
    } catch (err) {
      errEl.textContent = err.message || "Failed to reset password";
      errEl.classList.remove("hidden");
    }
  });
}

// ─── Router: detect page and boot the right init ─────────────────────────────
if (window.location.pathname.includes("login")) {
  initLoginPage();
} else {
  initDashboard();
}
