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
    if (password.length < 6) {
      errEl.textContent = "密码至少 6 位";
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
