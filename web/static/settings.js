async function api(path, opts = {}) {
  const r = await fetch(path, {
    credentials: "same-origin",
    headers: { "Content-Type": "application/json; charset=utf-8", ...opts.headers },
    ...opts,
  });
  if (r.status === 401) {
    window.location.href = "/login.html";
    throw new Error("unauthorized");
  }
  const text = await r.text();
  if (!r.ok) throw new Error(text || r.statusText);
  return text ? JSON.parse(text) : null;
}

let logoDataUrl = "";

function showLogoPreview(dataUrl) {
  const box = document.getElementById("set-logo-preview");
  const clr = document.getElementById("set-logo-clear");
  if (!dataUrl) {
    box.hidden = true;
    box.innerHTML = "";
    clr.hidden = true;
    return;
  }
  box.hidden = false;
  clr.hidden = false;
  box.innerHTML = `<img src="${dataUrl}" alt="Logo" />`;
}

document.getElementById("set-logo-file")?.addEventListener("change", (e) => {
  const f = e.target.files?.[0];
  if (!f) return;
  const reader = new FileReader();
  reader.onload = () => {
    logoDataUrl = String(reader.result || "");
    if (logoDataUrl.length > 500000) {
      alert("图片过大，请压缩后小于约 350KB。");
      logoDataUrl = "";
      e.target.value = "";
      showLogoPreview("");
      return;
    }
    showLogoPreview(logoDataUrl);
  };
  reader.readAsDataURL(f);
});

document.getElementById("set-logo-clear")?.addEventListener("click", () => {
  logoDataUrl = "";
  const input = document.getElementById("set-logo-file");
  if (input) input.value = "";
  showLogoPreview("");
});

async function loadSettings() {
  const s = await api("/api/settings");
  document.getElementById("set-company").value = s.companyName || "";
  document.getElementById("set-baseurl").value = s.baseUrl || "";
  const bc = document.getElementById("set-base-currency");
  if (bc) bc.value = (s.baseCurrency || "CAD").toUpperCase();
  const ph = document.getElementById("set-company-phone");
  if (ph) ph.value = s.companyPhone || "";
  const fx = document.getElementById("set-company-fax");
  if (fx) fx.value = s.companyFax || "";
  const addr = document.getElementById("set-company-address");
  if (addr) addr.value = s.companyAddress || "";
  const em = document.getElementById("set-company-email");
  if (em) em.value = s.companyEmail || "";
  document.getElementById("set-smtp-host").value = s.smtpHost || "";
  document.getElementById("set-smtp-port").value = s.smtpPort || 587;
  document.getElementById("set-smtp-user").value = s.smtpUser || "";
  document.getElementById("set-smtp-pass").value = "";
  document.getElementById("set-smtp-pass").placeholder = s.smtpPassSet ? "已设置，留空不修改" : "未设置";
  document.getElementById("set-smtp-from").value = s.smtpFrom || "";
  document.getElementById("set-smtp-starttls").checked = s.smtpStartTls !== false;
  document.getElementById("set-smtp-tls").checked = !!s.smtpImplicitTls;
  const hint = document.getElementById("set-env-hint");
  if (hint) {
    const eh = s.envSmtpHost || "(未设置)";
    const eb = s.envBaseUrl || "(未设置)";
    hint.textContent = `当前服务器环境：SMTP_HOST=${eh}，BASE_URL=${eb}`;
  }
  logoDataUrl = s.logoDataUrl || "";
  showLogoPreview(logoDataUrl);
}

function collectSettingsBody() {
  return {
    companyName: document.getElementById("set-company").value.trim(),
    baseUrl: document.getElementById("set-baseurl").value.trim(),
    baseCurrency: (document.getElementById("set-base-currency")?.value || "CAD").trim().toUpperCase() || "CAD",
    companyPhone: document.getElementById("set-company-phone")?.value?.trim() || "",
    companyFax: document.getElementById("set-company-fax")?.value?.trim() || "",
    companyAddress: document.getElementById("set-company-address")?.value?.trim() || "",
    companyEmail: document.getElementById("set-company-email")?.value?.trim() || "",
    logoDataUrl: logoDataUrl,
    smtpHost: document.getElementById("set-smtp-host").value.trim(),
    smtpPort: parseInt(document.getElementById("set-smtp-port").value, 10) || 587,
    smtpUser: document.getElementById("set-smtp-user").value.trim(),
    smtpPass: document.getElementById("set-smtp-pass").value,
    smtpFrom: document.getElementById("set-smtp-from").value.trim(),
    smtpStartTls: document.getElementById("set-smtp-starttls").checked,
    smtpImplicitTls: document.getElementById("set-smtp-tls").checked,
  };
}

async function submitSettings(msgElId) {
  const msg = document.getElementById(msgElId || "save-msg");
  if (msg) msg.textContent = "";
  const body = collectSettingsBody();
  try {
    await api("/api/settings", { method: "PUT", body: JSON.stringify(body) });
    if (msg) msg.textContent = "已保存。";
    await loadSettings();
  } catch (e) {
    alert("保存失败: " + e.message);
  }
}

document.getElementById("btn-save-settings")?.addEventListener("click", () => submitSettings("save-msg"));
document.getElementById("btn-save-smtp")?.addEventListener("click", () => submitSettings("save-msg-smtp"));

function showSettingsView(view) {
  const ids = ["company", "password", "expense-code", "exchange-currencies", "smtp", "users"];
  for (const v of ids) {
    const el = document.getElementById(`settings-view-${v}`);
    if (el) el.hidden = v !== view;
  }
  document.querySelectorAll("[data-settings-view]").forEach((btn) => {
    btn.classList.toggle("active", btn.dataset.settingsView === view);
  });
  if (view === "users") {
    loadUsers();
  }
  if (view === "expense-code") {
    loadExpenseCodes();
  }
  if (view === "exchange-currencies") {
    loadExchangeWatchlist();
  }
}

document.querySelectorAll("[data-settings-view]").forEach((btn) => {
  btn.addEventListener("click", () => showSettingsView(btn.dataset.settingsView));
});

async function loadExchangeWatchlist() {
  const tbody = document.getElementById("erc-body");
  if (!tbody) return;
  tbody.innerHTML = `<tr><td colspan="2" class="hint">加载中…</td></tr>`;
  try {
    const list = asArray(await api("/api/exchange-rate-codes"));
    tbody.innerHTML = "";
    if (list.length === 0) {
      tbody.innerHTML = `<tr><td colspan="2" class="hint">暂无货币。添加如 USD、CNY、EUR（三位字母）。</td></tr>`;
      return;
    }
    for (const row of list) {
      const code = row.code || "";
      const tr = document.createElement("tr");
      tr.innerHTML = `<td>${escapeHtml(code)}</td><td class="row-actions"><button type="button" class="ghost danger" data-act="del">删除</button></td>`;
      tr.querySelector('[data-act="del"]').addEventListener("click", async () => {
        if (!confirm(`从列表移除 ${code}？`)) return;
        try {
          await api(`/api/exchange-rate-codes/${encodeURIComponent(code)}`, { method: "DELETE" });
          await loadExchangeWatchlist();
        } catch (e) {
          alert(e.message || "删除失败");
        }
      });
      tbody.appendChild(tr);
    }
  } catch (e) {
    tbody.innerHTML = `<tr><td colspan="2" class="hint">加载失败</td></tr>`;
    alert(e.message || "加载失败");
  }
}

document.getElementById("btn-erc-add")?.addEventListener("click", async () => {
  const inp = document.getElementById("erc-new-code");
  const raw = (inp?.value || "").trim().toUpperCase();
  if (!/^[A-Z]{3}$/.test(raw)) {
    alert("请输入三位字母货币代码（如 USD）。");
    return;
  }
  try {
    await api("/api/exchange-rate-codes", { method: "POST", body: JSON.stringify({ code: raw }) });
    if (inp) inp.value = "";
    await loadExchangeWatchlist();
  } catch (e) {
    alert(e.message || "添加失败");
  }
});

function asArray(v) {
  return Array.isArray(v) ? v : [];
}

function escapeHtml(s) {
  const d = document.createElement("div");
  d.textContent = s == null ? "" : String(s);
  return d.innerHTML;
}

function fmtEcMoney(n) {
  if (n == null || Number.isNaN(Number(n))) return "0.00";
  return Number(n).toLocaleString("en-CA", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function ensureEcCodeSelectOptions() {
  const sel = document.getElementById("ec-code-new");
  if (!sel || sel.dataset.filled === "1") return;
  sel.dataset.filled = "1";
  const ph = document.createElement("option");
  ph.value = "";
  ph.textContent = "— Select code —";
  sel.appendChild(ph);
  for (let c = 5000; c <= 5999; c++) {
    const o = document.createElement("option");
    o.value = String(c);
    o.textContent = String(c);
    sel.appendChild(o);
  }
}

let ecDlgMode = "new";

function openEcDialog(mode, row) {
  ensureEcCodeSelectOptions();
  ecDlgMode = mode;
  const title = document.getElementById("ec-dlg-title");
  const rowNew = document.getElementById("ec-row-new");
  const rowEdit = document.getElementById("ec-row-edit");
  const sel = document.getElementById("ec-code-new");
  const nameInp = document.getElementById("ec-name");
  const dlg = document.getElementById("dlg-expense-code");
  if (!dlg || !nameInp) return;
  if (title) title.textContent = mode === "edit" ? "Edit expense code" : "New expense code";
  if (mode === "edit" && row) {
    if (rowNew) rowNew.hidden = true;
    if (rowEdit) rowEdit.hidden = false;
    const disp = document.getElementById("ec-code-display");
    const hid = document.getElementById("ec-code-edit");
    if (disp) disp.textContent = row.code || "";
    if (hid) hid.value = row.code || "";
    nameInp.value = row.name || "";
    if (sel) sel.removeAttribute("required");
  } else {
    if (rowNew) rowNew.hidden = false;
    if (rowEdit) rowEdit.hidden = true;
    if (sel) {
      sel.value = "";
      sel.setAttribute("required", "");
    }
    nameInp.value = "";
  }
  dlg.showModal();
}

async function loadExpenseCodes() {
  const tbody = document.getElementById("ec-body");
  const hint = document.getElementById("ec-year-hint");
  if (!tbody) return;
  const year = new Date().getFullYear();
  if (hint) {
    hint.textContent = `Balance (YTD) = ${year} 自然年累计支出（amount 直接相加）。`;
  }
  tbody.innerHTML = `<tr><td colspan="4" class="hint">加载中…</td></tr>`;
  try {
    const list = asArray(await api(`/api/expense-codes?year=${year}`));
    tbody.innerHTML = "";
    if (list.length === 0) {
      tbody.innerHTML = `<tr><td colspan="4" class="hint">暂无科目。可点击 New expense code 添加名称，或在主页 Expense 录入后自动出现已用代码。</td></tr>`;
      return;
    }
    for (const row of list) {
      const tr = document.createElement("tr");
      tr.innerHTML = `
        <td>${escapeHtml(row.code)}</td>
        <td>${escapeHtml(row.name || "—")}</td>
        <td>${escapeHtml(fmtEcMoney(row.balanceYtd))}</td>
        <td class="row-actions"><button type="button" class="ghost" data-act="edit">Edit</button></td>`;
      tr.querySelector('[data-act="edit"]').addEventListener("click", () => openEcDialog("edit", row));
      tbody.appendChild(tr);
    }
  } catch (e) {
    tbody.innerHTML = `<tr><td colspan="4" class="hint">加载失败：${escapeHtml(e.message)}</td></tr>`;
  }
}

document.getElementById("btn-ec-new")?.addEventListener("click", () => openEcDialog("new", null));

document.getElementById("ec-cancel")?.addEventListener("click", () => {
  document.getElementById("dlg-expense-code")?.close();
});

document.getElementById("form-expense-code")?.addEventListener("submit", async (e) => {
  e.preventDefault();
  const name = document.getElementById("ec-name")?.value?.trim() || "";
  if (!name) {
    alert("请填写 Expense name。");
    return;
  }
  try {
    if (ecDlgMode === "edit") {
      const code = document.getElementById("ec-code-edit")?.value?.trim() || "";
      if (!/^5\d{3}$/.test(code)) {
        alert("无效 code。");
        return;
      }
      await api(`/api/expense-codes/${encodeURIComponent(code)}`, {
        method: "PUT",
        body: JSON.stringify({ name }),
      });
    } else {
      const code = document.getElementById("ec-code-new")?.value?.trim() || "";
      if (!/^5\d{3}$/.test(code)) {
        alert("请选择 Code（5XXX）。");
        return;
      }
      await api("/api/expense-codes", {
        method: "POST",
        body: JSON.stringify({ code, name }),
      });
    }
    document.getElementById("dlg-expense-code")?.close();
    await loadExpenseCodes();
  } catch (err) {
    alert("保存失败: " + err.message);
  }
});

document.getElementById("btn-save-password")?.addEventListener("click", async () => {
  const oldPassword = document.getElementById("pwd-old").value;
  const newPassword = document.getElementById("pwd-new").value;
  const new2 = document.getElementById("pwd-new2").value;
  const msg = document.getElementById("pwd-msg");
  msg.hidden = true;
  if (newPassword !== new2) {
    msg.textContent = "两次新密码不一致";
    msg.hidden = false;
    return;
  }
  try {
    await api("/api/auth/password", {
      method: "POST",
      body: JSON.stringify({ oldPassword, newPassword }),
    });
    document.getElementById("pwd-old").value = "";
    document.getElementById("pwd-new").value = "";
    document.getElementById("pwd-new2").value = "";
    msg.textContent = "密码已更新，请牢记新密码。";
    msg.hidden = false;
    msg.style.color = "var(--muted, #8b9cb3)";
  } catch (e) {
    msg.textContent = e.message || "失败";
    msg.hidden = false;
  }
});

function escHtml(s) {
  const d = document.createElement("div");
  d.textContent = String(s || "");
  return d.innerHTML;
}

const ROLE_LABELS = { admin: "Admin", user2: "User Level 2", user1: "User Level 1（仅 Payroll）" };

async function loadUsers() {
  const tbody = document.getElementById("users-body");
  if (!tbody) return;
  tbody.innerHTML = `<tr><td colspan="4" class="hint">加载中…</td></tr>`;
  try {
    const users = await api("/api/users");
    tbody.innerHTML = users.map((u) => `
      <tr>
        <td>${escHtml(u.username)}</td>
        <td>
          ${u.role === "admin"
            ? `<span class="hint">${ROLE_LABELS.admin}</span>`
            : `<select class="user-role-sel" data-username="${escHtml(u.username)}">
                 <option value="user2" ${u.role === "user2" ? "selected" : ""}>User Level 2</option>
                 <option value="user1" ${u.role === "user1" ? "selected" : ""}>User Level 1（仅 Payroll）</option>
               </select>`
          }
        </td>
        <td>
          <button type="button" class="ghost small btn-reset-pwd" data-username="${escHtml(u.username)}">重置密码</button>
        </td>
        <td>
          ${u.role !== "admin"
            ? `<button type="button" class="ghost small btn-del-user" data-username="${escHtml(u.username)}">删除</button>`
            : ""}
        </td>
      </tr>
    `).join("");

    tbody.querySelectorAll(".user-role-sel").forEach((sel) => {
      sel.addEventListener("change", async () => {
        try {
          await api("/api/users", { method: "PUT", body: JSON.stringify({ username: sel.dataset.username, role: sel.value }) });
        } catch (e) {
          alert(e.message || "修改失败");
          loadUsers();
        }
      });
    });

    tbody.querySelectorAll(".btn-reset-pwd").forEach((btn) => {
      btn.addEventListener("click", () => {
        document.getElementById("reset-pwd-username").textContent = btn.dataset.username;
        document.getElementById("reset-pwd-new").value = "";
        document.getElementById("reset-pwd-new").dataset.username = btn.dataset.username;
        document.getElementById("dlg-reset-pwd")?.showModal();
      });
    });

    tbody.querySelectorAll(".btn-del-user").forEach((btn) => {
      btn.addEventListener("click", async () => {
        if (!confirm(`确认删除用户「${btn.dataset.username}」？`)) return;
        try {
          await api("/api/users", { method: "DELETE", body: JSON.stringify({ username: btn.dataset.username }) });
          loadUsers();
        } catch (e) {
          alert(e.message || "删除失败");
        }
      });
    });
  } catch (e) {
    if (tbody) tbody.innerHTML = `<tr><td colspan="4" class="hint">加载失败: ${escHtml(e.message)}</td></tr>`;
  }
}

function initUserManagement() {
  const dlg = document.getElementById("dlg-user");
  const form = document.getElementById("form-user");

  document.getElementById("btn-user-new")?.addEventListener("click", () => {
    document.getElementById("user-name").value = "";
    document.getElementById("user-pwd").value = "";
    document.getElementById("user-role").value = "user2";
    dlg?.showModal();
  });
  document.getElementById("user-cancel")?.addEventListener("click", () => dlg?.close());

  form?.addEventListener("submit", async (e) => {
    e.preventDefault();
    try {
      await api("/api/users", {
        method: "POST",
        body: JSON.stringify({
          username: document.getElementById("user-name").value,
          password: document.getElementById("user-pwd").value,
          role: document.getElementById("user-role").value,
        }),
      });
      dlg?.close();
      loadUsers();
    } catch (e) {
      alert(e.message || "创建失败");
    }
  });

  const resetDlg = document.getElementById("dlg-reset-pwd");
  const resetForm = document.getElementById("form-reset-pwd");

  document.getElementById("reset-pwd-cancel")?.addEventListener("click", () => resetDlg?.close());

  resetForm?.addEventListener("submit", async (e) => {
    e.preventDefault();
    const pwdInput = document.getElementById("reset-pwd-new");
    try {
      await api("/api/users", {
        method: "PUT",
        body: JSON.stringify({ username: pwdInput.dataset.username, newPassword: pwdInput.value }),
      });
      resetDlg?.close();
      alert("密码已重置。");
    } catch (e) {
      alert(e.message || "重置失败");
    }
  });
}

(async function init() {
  try {
    const me = await fetch("/api/me", { credentials: "same-origin" }).then((r) => r.json());
    const btn = document.getElementById("btn-logout");
    if (btn) {
      btn.hidden = !me.authEnabled;
      btn.addEventListener("click", async () => {
        await fetch("/api/logout", { method: "POST", credentials: "same-origin" });
        window.location.href = "/login.html";
      });
    }
    if (me.authEnabled && !me.authenticated) {
      window.location.href = "/login.html";
      return;
    }
    // 仅 admin 显示用户管理
    if (me.role === "admin") {
      const navBtn = document.getElementById("nav-btn-users");
      if (navBtn) navBtn.hidden = false;
      initUserManagement();
    }
    await loadSettings();
    showSettingsView("company");
  } catch {
    window.location.href = "/login.html";
  }
})();
