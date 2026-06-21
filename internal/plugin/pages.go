package plugin

import (
	"encoding/json"
	"strings"

	"github.com/theramindex/silo-plugin-app-links/internal/store"
)

func userPageHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Apps</title>
  <style>
    :root { color-scheme: dark; --bg:#171717; --panel:#242426; --panel2:#2f2f32; --line:#38383c; --text:#f6f6f6; --muted:#a2a2aa; --accent:#ff2f74; }
    * { box-sizing: border-box; }
    body { margin:0; min-height:100vh; background:var(--bg); color:var(--text); font-family:ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; }
    .shell { min-height:100vh; padding:1.25rem; }
    .top { display:flex; align-items:center; justify-content:space-between; gap:1rem; margin-bottom:1rem; }
    h1 { margin:0; font-size:1.55rem; letter-spacing:0; }
    .search { width:min(34rem,100%); border:1px solid var(--line); border-radius:999px; background:var(--panel); color:var(--text); padding:.68rem .9rem; font:inherit; }
    .grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(13rem,1fr)); gap:.7rem; }
    .card { min-height:8.5rem; border:1px solid rgba(255,255,255,.06); border-radius:.75rem; background:var(--panel); color:var(--text); text-decoration:none; padding:.85rem; display:flex; flex-direction:column; justify-content:space-between; overflow:hidden; }
    .card:hover { background:var(--panel2); border-color:rgba(255,255,255,.14); }
    .icon { width:3.1rem; height:3.1rem; border-radius:.7rem; background:#101011; display:grid; place-items:center; overflow:hidden; font-weight:900; }
    .icon img { width:100%; height:100%; object-fit:contain; }
    strong, p, small { overflow:hidden; text-overflow:ellipsis; }
    strong { display:block; white-space:nowrap; margin-top:.8rem; }
    p { margin:.25rem 0 0; color:var(--muted); display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; }
    small { display:block; color:var(--muted); white-space:nowrap; margin-top:.7rem; }
    .empty { color:var(--muted); padding:2rem 0; }
  </style>
</head>
<body>
  <main class="shell">
    <div class="top"><h1>Apps</h1><input id="search" class="search" placeholder="Search apps"></div>
    <div id="apps" class="grid"></div>
  </main>
  <script>
    const base = location.pathname.endsWith("/app-links") ? location.pathname.slice(0, -"/app-links".length) : "";
    const route = path => base + path;
    const esc = value => String(value || "").replace(/[&<>"']/g, ch => ({ "&":"&amp;", "<":"&lt;", ">":"&gt;", "\"":"&quot;", "'":"&#39;" })[ch]);
    let links = [];
    function card(link) {
      const icon = link.iconUrl ? "<img src='" + esc(link.iconUrl) + "' alt=''>" : esc((link.name || "A").slice(0, 2).toUpperCase());
      return "<a class='card' data-search='" + esc([link.name, link.description, link.category].join(" ").toLowerCase()) + "' href='" + route("/app-links/open?id=" + encodeURIComponent(link.id)) + "'><div><div class='icon'>" + icon + "</div><strong>" + esc(link.name) + "</strong><p>" + esc(link.description || link.url) + "</p></div><small>" + esc(link.category || (link.openMode === "new_tab" ? "External" : "Embedded")) + "</small></a>";
    }
    function render() {
      const query = document.getElementById("search").value.toLowerCase();
      const visible = links.filter(link => !query || [link.name, link.description, link.category].join(" ").toLowerCase().includes(query));
      document.getElementById("apps").innerHTML = visible.length ? visible.map(card).join("") : "<div class='empty'>No app links yet.</div>";
    }
    fetch(route("/app-links/api/links"), { credentials:"include" }).then(r => r.json()).then(data => { links = (data.links || []).filter(link => link.enabled); render(); }).catch(() => { document.getElementById("apps").innerHTML = "<div class='empty'>Unable to load app links.</div>"; });
    document.getElementById("search").addEventListener("input", render);
  </script>
</body>
</html>`
}

func iframePageHTML(link store.Link) string {
	data, _ := json.Marshal(link)
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>` + esc(link.Name) + `</title>
  <style>
    :root { color-scheme:dark; --bg:#050505; --chrome:rgba(16,16,17,.82); --text:#fff; --muted:rgba(255,255,255,.68); --line:rgba(255,255,255,.12); }
    body { margin:0; height:100vh; overflow:hidden; background:var(--bg); color:var(--text); font-family:ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; }
    iframe { position:absolute; inset:0; width:100%; height:100%; border:0; background:#050505; }
    .bar { position:absolute; inset:1rem 1rem auto; z-index:2; display:flex; align-items:center; gap:.55rem; pointer-events:none; }
    .bar a, .bar button { pointer-events:auto; min-height:2.35rem; border:1px solid var(--line); border-radius:999px; background:var(--chrome); color:var(--text); text-decoration:none; display:inline-flex; align-items:center; gap:.4rem; padding:0 .82rem; font:inherit; font-weight:800; backdrop-filter:blur(18px); }
    .icon { width:2.35rem!important; padding:0!important; justify-content:center; }
    .title { max-width:min(32rem,45vw); overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:var(--muted); }
    svg { width:1rem; height:1rem; display:block; }
  </style>
</head>
<body>
  <iframe src="` + esc(link.URL) + `" title="` + esc(link.Name) + `" allow="fullscreen; autoplay; encrypted-media; clipboard-read; clipboard-write"></iframe>
  <nav class="bar">
    <a class="icon" href="../app-links" aria-label="Back to Apps"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5"/></svg></a>
    <a href="/" aria-label="Back to Silo">Silo</a>
    <a href="` + esc(link.URL) + `" target="_blank" rel="noreferrer">Open externally</a>
    <span class="title">` + esc(link.Name) + `</span>
  </nav>
  <script>window.APP_LINK = ` + string(data) + `;</script>
</body>
</html>`
}

func adminPageHTML(dataPath string) string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>App Links Admin</title>
  <style>
    :root { color-scheme:dark; --bg:#141414; --panel:#222224; --panel2:#2d2d30; --line:#3a3a3f; --text:#f7f7f7; --muted:#a6a6ae; --accent:#ff2f74; --danger:#ff5a66; }
    * { box-sizing:border-box; }
    body { margin:0; min-height:100vh; background:var(--bg); color:var(--text); font-family:ui-sans-serif,system-ui,-apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif; }
    main { padding:1.25rem; max-width:74rem; }
    .head { display:flex; justify-content:space-between; gap:1rem; align-items:center; margin-bottom:1rem; }
    h1 { margin:0; font-size:1.55rem; }
    .muted { color:var(--muted); }
    .layout { display:grid; grid-template-columns:minmax(0,1fr) 24rem; gap:1rem; align-items:start; }
    .panel { border:1px solid var(--line); border-radius:.8rem; background:var(--panel); padding:.85rem; }
    .list { display:grid; gap:.5rem; }
    .row { border:0; border-radius:.65rem; background:var(--panel2); color:var(--text); display:grid; grid-template-columns:minmax(0,1fr) auto; gap:.7rem; align-items:center; text-align:left; padding:.72rem; }
    .row strong, .row small { display:block; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
    .row small { color:var(--muted); margin-top:.15rem; }
    label { display:grid; gap:.3rem; margin-bottom:.65rem; color:var(--muted); font-weight:750; }
    input, textarea, select { width:100%; border:1px solid var(--line); border-radius:.55rem; background:#171719; color:var(--text); padding:.58rem .65rem; font:inherit; }
    textarea { min-height:5rem; resize:vertical; }
    .actions { display:flex; gap:.5rem; flex-wrap:wrap; }
    button { border:0; border-radius:.55rem; background:var(--panel2); color:var(--text); padding:.58rem .75rem; font:inherit; font-weight:850; cursor:pointer; }
    button.primary { background:var(--accent); }
    button.danger { color:white; background:var(--danger); }
    .check { display:flex; align-items:center; gap:.5rem; color:var(--text); }
    .check input { width:auto; }
    @media (max-width: 860px) { .layout { grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <main>
    <div class="head"><div><h1>App Links</h1><div class="muted">Storage: ` + esc(dataPath) + `</div></div><button id="new">New link</button></div>
    <div class="layout">
      <section class="panel"><div id="list" class="list"></div></section>
      <form id="form" class="panel">
        <label>Name<input id="name" required></label>
        <label>URL<input id="url" required placeholder="https://app.example.com"></label>
        <label>Description<textarea id="description"></textarea></label>
        <label>Category<input id="category" placeholder="Media, Admin, Requests"></label>
        <label>Icon URL<input id="iconUrl" placeholder="https://app.example.com/icon.png"></label>
        <label>Open mode<select id="openMode"><option value="iframe">Fullscreen iframe</option><option value="new_tab">New tab</option></select></label>
        <label>Sort order<input id="sortOrder" type="number" value="0"></label>
        <label class="check"><input id="enabled" type="checkbox" checked> Enabled</label>
        <div class="actions"><button class="primary" type="submit">Save link</button><button id="delete" class="danger" type="button">Delete</button></div>
      </form>
    </div>
  </main>
  <script>
    const base = location.pathname.endsWith("/app-links/admin") ? location.pathname.slice(0, -"/app-links/admin".length) : "";
    const route = path => base + path;
    let links = [], current = null;
    const ids = ["name","url","description","category","iconUrl","openMode","sortOrder","enabled"];
    const el = id => document.getElementById(id);
    const esc = value => String(value || "").replace(/[&<>"']/g, ch => ({ "&":"&amp;", "<":"&lt;", ">":"&gt;", "\"":"&quot;", "'":"&#39;" })[ch]);
    function reset() { current = null; ids.forEach(id => { if (id === "enabled") el(id).checked = true; else if (id === "openMode") el(id).value = "iframe"; else if (id === "sortOrder") el(id).value = "0"; else el(id).value = ""; }); }
    function edit(link) { current = link; ids.forEach(id => { if (id === "enabled") el(id).checked = !!link.enabled; else el(id).value = link[id] || (id === "openMode" ? "iframe" : ""); }); }
    function render() { el("list").innerHTML = links.length ? links.map(link => "<button class='row' data-id='" + esc(link.id) + "'><span><strong>" + esc(link.name) + "</strong><small>" + esc(link.url) + "</small></span><small>" + esc(link.enabled ? link.openMode : "disabled") + "</small></button>").join("") : "<div class='muted'>No links yet.</div>"; document.querySelectorAll("[data-id]").forEach(row => row.onclick = () => edit(links.find(link => link.id === row.dataset.id))); }
    async function load() { const response = await fetch(route("/app-links/admin/api/links"), { credentials:"include" }); const data = await response.json(); links = data.links || []; render(); }
    el("new").onclick = reset;
    el("form").onsubmit = async event => { event.preventDefault(); const payload = { id: current && current.id || "", name: el("name").value, url: el("url").value, description: el("description").value, category: el("category").value, iconUrl: el("iconUrl").value, openMode: el("openMode").value, sortOrder: Number(el("sortOrder").value || 0), enabled: el("enabled").checked }; const response = await fetch(route("/app-links/admin/api/links"), { method:"POST", credentials:"include", headers:{ "content-type":"application/json" }, body:JSON.stringify(payload) }); if (!response.ok) { const data = await response.json().catch(() => ({})); alert(data.error || "Save failed"); return; } reset(); await load(); };
    el("delete").onclick = async () => { if (!current || !confirm("Delete " + current.name + "?")) return; await fetch(route("/app-links/admin/api/delete"), { method:"POST", credentials:"include", headers:{ "content-type":"application/json" }, body:JSON.stringify({ id: current.id }) }); reset(); await load(); };
    reset(); load();
  </script>
</body>
</html>`
}

func compact(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
