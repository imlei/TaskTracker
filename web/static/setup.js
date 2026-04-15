(async function boot() {
  try {
    const me = await fetch("/api/me", { credentials: "same-origin" }).then((r) => r.json());
    if (!me.authEnabled) {
      window.location.replace("/");
      return;
    }
    if (!me.needsSetup) {
      window.location.replace(me.authenticated ? "/" : "/login.html");
      return;
    }
  } catch {
    // 继续显示创建页
  }

  const form = document.getElementById("form-setup");
  const errEl = document.getElementById("setup-err");

  form.addEventListener("submit", async (e) => {
    e.preventDefault();
    errEl.hidden = true;
    const username = document.getElementById("setup-user").value.trim();
    const password = document.getElementById("setup-pass").value;
    const pass2 = document.getElementById("setup-pass2").value;
    if (password.length < 8) {
      errEl.textContent = "密码至少 8 位";
      errEl.hidden = false;
      return;
    }
    if (!/[A-Z]/.test(password)) {
      errEl.textContent = "密码须包含至少一个大写字母";
      errEl.hidden = false;
      return;
    }
    if (!/[a-z]/.test(password)) {
      errEl.textContent = "密码须包含至少一个小写字母";
      errEl.hidden = false;
      return;
    }
    if (!/[0-9]/.test(password)) {
      errEl.textContent = "密码须包含至少一个数字";
      errEl.hidden = false;
      return;
    }
    if (!/[^A-Za-z0-9]/.test(password)) {
      errEl.textContent = "密码须包含至少一个特殊字符";
      errEl.hidden = false;
      return;
    }
    if (password !== pass2) {
      errEl.textContent = "两次输入的密码不一致";
      errEl.hidden = false;
      return;
    }
    try {
      const r = await fetch("/api/setup", {
        method: "POST",
        credentials: "same-origin",
        headers: { "Content-Type": "application/json; charset=utf-8" },
        body: JSON.stringify({ username, password }),
      });
      let data = {};
      try {
        data = await r.json();
      } catch {
        /* ignore */
      }
      if (!r.ok) {
        errEl.textContent = data.error || "创建失败";
        errEl.hidden = false;
        return;
      }
      window.location.href = "/";
    } catch (err) {
      errEl.textContent = err.message || "网络错误";
      errEl.hidden = false;
    }
  });
})();
