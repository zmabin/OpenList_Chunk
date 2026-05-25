package handles

import (
	"github.com/gin-gonic/gin"
)

func LoginSchedulePage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, loginScheduleHTML)
}

const loginScheduleHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Storage Login Schedule</title>
<style>
  :root {
    --primary: #5b6abf;
    --primary-hover: #4a59ae;
    --bg: #f5f6fa;
    --card: #ffffff;
    --border: #e2e4ea;
    --text: #1a1c23;
    --text-secondary: #6b7085;
    --danger: #e74c3c;
    --success: #27ae60;
    --warning: #f39c12;
    --radius: 8px;
  }
  * { margin: 0; padding: 0; box-sizing: border-box; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    background: var(--bg);
    color: var(--text);
    min-height: 100vh;
  }
  .container {
    max-width: 960px;
    margin: 0 auto;
    padding: 24px 16px;
  }
  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 24px;
  }
  .header h1 {
    font-size: 22px;
    font-weight: 600;
  }
  .header a {
    color: var(--primary);
    text-decoration: none;
    font-size: 14px;
  }
  .header a:hover { text-decoration: underline; }
  .card {
    background: var(--card);
    border-radius: var(--radius);
    border: 1px solid var(--border);
    overflow: hidden;
  }
  .table-wrap { overflow-x: auto; }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
  }
  th, td {
    padding: 12px 16px;
    text-align: left;
    border-bottom: 1px solid var(--border);
  }
  th {
    background: #fafbfc;
    font-weight: 600;
    color: var(--text-secondary);
    font-size: 12px;
    text-transform: uppercase;
    letter-spacing: 0.5px;
  }
  tr:last-child td { border-bottom: none; }
  tr:hover td { background: #f8f9fd; }
  .status {
    display: inline-block;
    padding: 2px 10px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
  }
  .status-work { background: #e8f8ef; color: var(--success); }
  .status-err { background: #fdecea; color: var(--danger); }
  .status-off { background: #f0f0f0; color: #888; }
  input[type="number"] {
    width: 90px;
    padding: 6px 10px;
    border: 1px solid var(--border);
    border-radius: 6px;
    font-size: 14px;
    text-align: center;
    outline: none;
    transition: border-color .2s;
  }
  input[type="number"]:focus { border-color: var(--primary); }
  .btn {
    padding: 6px 16px;
    border: none;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    cursor: pointer;
    transition: background .2s;
  }
  .btn-primary { background: var(--primary); color: #fff; }
  .btn-primary:hover { background: var(--primary-hover); }
  .btn-primary:disabled { opacity: .5; cursor: not-allowed; }
  .msg {
    padding: 4px 0;
    font-size: 12px;
    min-height: 18px;
  }
  .msg-ok { color: var(--success); }
  .msg-err { color: var(--danger); }
  .login-hint {
    text-align: center;
    padding: 60px 20px;
    color: var(--text-secondary);
  }
  .login-hint a { color: var(--primary); text-decoration: none; }
  .login-hint a:hover { text-decoration: underline; }
  .loading { text-align: center; padding: 40px; color: var(--text-secondary); }
  .mount-path { font-family: monospace; font-size: 13px; }
  .driver-name { font-weight: 500; }
  .hint { color: var(--text-secondary); font-size: 12px; margin-top: 4px; }
  .back-link {
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }
</style>
</head>
<body>
<div class="container">
  <div class="header">
    <h1>Storage Login Schedule</h1>
    <a id="back-link" href="/@manage" class="back-link">&larr; Back to Admin</a>
  </div>
  <div id="content">
    <div class="loading">Loading...</div>
  </div>
</div>
<script>
(function() {
  var TOKEN_KEY = "openlist-token";
  var apiBase = window.location.pathname.replace(/\/@manage\/login-schedule.*/, "");

  function getToken() {
    var keys = [TOKEN_KEY, "token", "alist-token", "open_list_token"];
    for (var i = 0; i < keys.length; i++) {
      var v = localStorage.getItem(keys[i]);
      if (v) return v;
    }
    return "";
  }

  function api(method, path, body) {
    var opts = {
      method: method,
      headers: { "Authorization": getToken() }
    };
    if (body) {
      opts.headers["Content-Type"] = "application/json";
      opts.body = JSON.stringify(body);
    }
    return fetch(apiBase + "/api/admin" + path, opts).then(function(r) { return r.json(); });
  }

  function esc(s) {
    var d = document.createElement("div");
    d.textContent = s;
    return d.innerHTML;
  }

  function statusClass(s) {
    if (s === "work") return "status-work";
    if (s && s !== "disabled") return "status-err";
    return "status-off";
  }
  function statusText(s) {
    if (s === "work") return "OK";
    if (s === "disabled") return "Disabled";
    if (s) return s.substring(0, 30);
    return "-";
  }

  function render(storages) {
    if (!storages || storages.length === 0) {
      return '<div class="login-hint"><p>No storages found.</p></div>';
    }
    var h = '<div class="card"><div class="table-wrap"><table>';
    h += '<thead><tr><th>Mount Path</th><th>Driver</th><th>Status</th><th>Login Interval (min)</th><th>Action</th></tr></thead><tbody>';
    storages.forEach(function(s) {
      var val = s.login_interval || 0;
      h += '<tr>';
      h += '<td class="mount-path">' + esc(s.mount_path) + '</td>';
      h += '<td class="driver-name">' + esc(s.driver) + '</td>';
      h += '<td><span class="status ' + statusClass(s.status) + '">' + esc(statusText(s.status)) + '</span></td>';
      h += '<td><input type="number" min="0" value="' + val + '" data-id="' + s.id + '" id="interval-' + s.id + '"/>';
      h += '<div class="hint">0 = disabled</div></td>';
      h += '<td><button class="btn btn-primary" data-id="' + s.id + '" onclick="saveInterval(' + s.id + ')">Save</button>';
      h += '<div class="msg" id="msg-' + s.id + '"></div></td>';
      h += '</tr>';
    });
    h += '</tbody></table></div></div>';
    return h;
  }

  function showMsg(id, text, ok) {
    var el = document.getElementById("msg-" + id);
    if (el) {
      el.textContent = text;
      el.className = "msg " + (ok ? "msg-ok" : "msg-err");
      if (text) setTimeout(function() { el.textContent = ""; }, 3000);
    }
  }

  window.saveInterval = function(id) {
    var input = document.getElementById("interval-" + id);
    if (!input) return;
    var val = parseInt(input.value) || 0;
    showMsg(id, "Saving...", false);

    api("GET", "/storage/get?id=" + id).then(function(res) {
      if (res.code !== 200) {
        showMsg(id, res.message || "Failed to get storage", false);
        return;
      }
      var storage = res.data;
      storage.login_interval = val;
      return api("POST", "/storage/update", storage);
    }).then(function(res) {
      if (!res) return;
      if (res.code === 200) {
        showMsg(id, "Saved", true);
      } else {
        showMsg(id, res.message || "Failed to save", false);
      }
    }).catch(function(e) {
      showMsg(id, "Error: " + e.message, false);
    });
  };

  function init() {
    // Fix back link for base path
    document.getElementById("back-link").href = apiBase + "/@manage";

    var token = getToken();
    if (!token) {
      document.getElementById("content").innerHTML =
        '<div class="login-hint"><p>Please <a href="' + apiBase + '/@manage">log in</a> first, then return to this page.</p></div>';
      return;
    }

    api("GET", "/storage/list?page=1&per_page=100").then(function(res) {
      if (res.code === 200) {
        document.getElementById("content").innerHTML = render(res.data.content);
      } else if (res.code === 401 || res.code === 403) {
        document.getElementById("content").innerHTML =
          '<div class="login-hint"><p>Session expired. Please <a href="' + apiBase + '/@manage">log in</a> again.</p></div>';
      } else {
        document.getElementById("content").innerHTML =
          '<div class="login-hint"><p>Error: ' + esc(res.message || "unknown") + '</p></div>';
      }
    }).catch(function(e) {
      document.getElementById("content").innerHTML =
        '<div class="login-hint"><p>Network error: ' + esc(e.message) + '</p></div>';
    });
  }

  init();
})();
</script>
</body>
</html>`
