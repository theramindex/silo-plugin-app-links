package plugin

import (
	"encoding/json"
	"strings"

	"github.com/theramindex/silo-plugin-app-links/internal/store"
)

func sharedCSS() string {
	return `
    :root { color-scheme:dark; --fallback-bg:#151415; --fallback-surface:#222124; --fallback-surface-2:#2b292e; --fallback-field:#19181b; --fallback-line:#3b373f; --fallback-text:#f4f0f3; --fallback-muted:#b9aeb6; --fallback-accent:#ff2f74; --fallback-accent-ink:#fff6fa; --fallback-danger:#ff6670;
      /* Silo's theme variables win when the host provides them. */
      --bg:var(--silo-bg, var(--silo-background, var(--fallback-bg))); --surface:var(--silo-card, var(--silo-surface, var(--fallback-surface))); --surface-2:var(--silo-surface-2, var(--silo-card, var(--fallback-surface-2))); --field:var(--silo-surface, var(--fallback-field)); --line:var(--silo-border, var(--silo-line, var(--fallback-line))); --text:var(--silo-text, var(--silo-foreground, var(--fallback-text))); --muted:var(--silo-muted, var(--silo-muted-foreground, var(--fallback-muted))); --accent:var(--silo-accent, var(--silo-primary, var(--fallback-accent))); --accent-ink:var(--silo-on-accent, var(--silo-primary-foreground, var(--fallback-accent-ink))); --danger:var(--silo-danger, var(--fallback-danger)); --shadow:0 18px 44px rgba(0,0,0,.28); }
    * { box-sizing:border-box; }
    html { overflow-x:hidden; }
    body { margin:0; min-height:100vh; overflow-x:hidden; background:var(--bg); color:var(--text); font-family:Avenir Next,Segoe UI,ui-sans-serif,sans-serif; }
    h1, h2, h3, strong { font-family:Georgia,Palatino,ui-serif,serif; letter-spacing:0; }
    h1 { margin:0; font-size:1.72rem; line-height:1.08; }
    h2 { margin:0; font-size:1.18rem; line-height:1.16; }
    a { color:inherit; }
    button, input, textarea, select { font:inherit; }
    button, a, input, textarea, select { outline-offset:3px; }
    button:focus-visible, a:focus-visible, input:focus-visible, textarea:focus-visible, select:focus-visible { outline:2px solid var(--accent); }
    .muted { color:var(--muted); }
    .pill { min-height:2.75rem; border:1px solid var(--line); border-radius:8px; background:var(--surface-2); color:var(--text); display:inline-flex; align-items:center; justify-content:center; gap:.45rem; padding:0 .85rem; text-decoration:none; font-weight:800; transition:background-color .15s ease,border-color .15s ease,transform .15s ease; }
    .pill:hover { background:color-mix(in oklch,var(--surface-2),var(--text) 8%); border-color:color-mix(in oklch,var(--line),var(--text) 18%); transform:translateY(-1px); }
    .primary { border-color:color-mix(in oklch,var(--accent),oklch(96% .02 350) 14%); background:var(--accent); color:var(--accent-ink); }
    .danger { border-color:color-mix(in oklch,var(--danger),oklch(96% .02 350) 10%); background:var(--danger); color:var(--accent-ink); }
    svg { width:1rem; height:1rem; display:block; flex:0 0 auto; }
    @media (prefers-reduced-motion: reduce) { *, *::before, *::after { transition:none!important; transform:none!important; } }`
}

func userPageHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>Apps</title>
  <style>` + sharedCSS() + `
    .shell { min-height:100vh; padding:1.25rem; }
    .top { display:flex; align-items:end; justify-content:space-between; gap:1rem; margin-bottom:1rem; }
    .kicker { margin-top:.22rem; color:var(--muted); font-size:.92rem; }
    .search { width:min(34rem,100%); min-height:2.85rem; border:1px solid var(--line); border-radius:8px; background:var(--field); color:var(--text); padding:.68rem .9rem; transition:border-color .15s ease,background-color .15s ease; }
    .search:focus { border-color:var(--accent); }
    .grid { display:grid; grid-template-columns:repeat(auto-fill,minmax(13rem,1fr)); gap:.7rem; }
    .card { min-height:9.25rem; border:1px solid var(--line); border-radius:8px; background:var(--surface); color:var(--text); text-decoration:none; padding:.85rem; display:flex; flex-direction:column; justify-content:space-between; overflow:hidden; box-shadow:0 1px 0 rgba(255,255,255,.03) inset; transition:background-color .15s ease,border-color .15s ease,transform .15s ease; }
    .card:hover { border-color:color-mix(in oklch,var(--accent),oklch(96% .02 350) 20%); transform:translateY(-1px); }
    .card-top { display:flex; justify-content:space-between; gap:.6rem; align-items:start; }
    .app-icon { width:3.05rem; height:3.05rem; border-radius:8px; background:var(--surface-2); display:grid; place-items:center; overflow:hidden; font-weight:900; color:var(--text); border:1px solid var(--line); }
    .app-icon img { width:100%; height:100%; object-fit:contain; }
    .mode-badge { min-height:1.85rem; border:1px solid var(--line); border-radius:999px; display:inline-flex; align-items:center; gap:.32rem; padding:0 .55rem; color:var(--muted); font-size:.78rem; font-weight:850; white-space:nowrap; }
    .mode-badge.external { border-color:var(--accent); color:var(--text); }
    strong, p, small { overflow:hidden; text-overflow:ellipsis; }
    strong { display:block; white-space:nowrap; margin-top:.78rem; font-size:1.05rem; }
    p { margin:.28rem 0 0; color:var(--muted); display:-webkit-box; -webkit-line-clamp:2; -webkit-box-orient:vertical; line-height:1.35; }
    small { display:block; color:var(--muted); white-space:nowrap; margin-top:.74rem; }
    .state { grid-column:1/-1; min-height:12rem; border:1px dashed var(--line); border-radius:8px; display:grid; place-items:center; padding:2rem; color:var(--muted); text-align:center; }
    .state-inner { max-width:26rem; display:grid; gap:.7rem; justify-items:center; }
    .state strong { margin:0; color:var(--text); white-space:normal; }
    @media (max-width: 720px) { .shell { padding:1rem; } .top { display:grid; align-items:start; } .search { width:100%; } .grid { grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <main class="shell">
    <div class="top"><div><h1>Apps</h1><div class="kicker">Launch the services connected to Silo.</div></div><input id="search" class="search" placeholder="Search apps" aria-label="Search apps"></div>
    <div id="apps" class="grid"></div>
  </main>
  <script>
    const base = location.pathname.endsWith("/app-links") ? location.pathname.slice(0, -"/app-links".length) : "";
    const route = path => base + path;
    const esc = value => String(value || "").replace(/[&<>"']/g, ch => ({ "&":"&amp;", "<":"&lt;", ">":"&gt;", "\"":"&quot;", "'":"&#39;" })[ch]);
    const externalIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M7 17 17 7M9 7h8v8"/></svg>';
    const embedIcon = '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M4 5h16v14H4zM8 9h8M8 13h5"/></svg>';
    let links = [];
    function state(title, detail, action) {
      return "<div class='state'><div class='state-inner'><strong>" + esc(title) + "</strong><span>" + esc(detail || "") + "</span>" + (action || "") + "</div></div>";
    }
    function card(link) {
      const isExternal = link.openMode === "new_tab";
      const icon = link.iconUrl ? "<img src='" + esc(link.iconUrl) + "' alt='' loading='lazy' decoding='async'>" : esc((link.name || "A").slice(0, 2).toUpperCase());
      const badge = "<span class='mode-badge " + (isExternal ? "external" : "embedded") + "'>" + (isExternal ? externalIcon + "New tab" : embedIcon + "Embedded") + "</span>";
      const href = route("/app-links/open?id=" + encodeURIComponent(link.id));
      const targetAttr = isExternal ? " target='_blank' rel='noreferrer'" : "";
      return "<a class='card' data-search='" + esc([link.name, link.description, link.category].join(" ").toLowerCase()) + "' href='" + href + "'" + targetAttr + "><div><div class='card-top'><div class='app-icon'>" + icon + "</div>" + badge + "</div><strong>" + esc(link.name) + "</strong><p>" + esc(link.description || link.url) + "</p></div><small>" + esc(link.category || "App link") + "</small></a>";
    }
    function render() {
      const query = document.getElementById("search").value.trim().toLowerCase();
      const visible = links.filter(link => !query || [link.name, link.description, link.category, link.url].join(" ").toLowerCase().includes(query));
      if (visible.length) { document.getElementById("apps").innerHTML = visible.map(card).join(""); return; }
      document.getElementById("apps").innerHTML = query ? state("No matching apps", "Try a different name, category, or URL.", "<button class='pill' type='button' id='clear-search'>Clear search</button>") : state("No app links yet", "Add links from the App Links admin page.");
      const clear = document.getElementById("clear-search");
      if (clear) clear.onclick = () => { document.getElementById("search").value = ""; render(); };
    }
    function load() {
      document.getElementById("apps").innerHTML = state("Loading apps", "Checking the configured links.");
      fetch(route("/app-links/api/links"), { credentials:"include" }).then(r => {
        if (!r.ok) throw new Error("load failed");
        return r.json();
      }).then(data => { links = (data.links || []).filter(link => link.enabled); render(); }).catch(() => {
        document.getElementById("apps").innerHTML = state("Unable to load app links", "Check your connection and try again. If it keeps failing, ask your Silo administrator.", "<button class='pill primary' type='button' id='retry-load'>Retry</button>");
        document.getElementById("retry-load").onclick = load;
      });
    }
    document.getElementById("search").addEventListener("input", render);
    load();
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
  <style>` + sharedCSS() + `
    body { height:100vh; height:100dvh; overflow:hidden; background:var(--bg); display:flex; flex-direction:column; }
    .bar { position:relative; z-index:3; flex:0 0 auto; min-height:3.65rem; border-bottom:1px solid var(--line); background:var(--surface); display:flex; flex-wrap:wrap; align-items:center; gap:.55rem; padding:.55rem .8rem; box-shadow:var(--shadow); }
    .stage { position:relative; flex:1 1 auto; min-height:0; }
    iframe { position:absolute; inset:0; width:100%; height:100%; border:0; background:var(--bg); }
    .bar .pill { min-height:2.45rem; }
    .icon-only { width:2.45rem; padding:0; }
    .title { min-width:0; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; color:var(--muted); margin-left:.25rem; }
    .embed-note { margin-left:auto; color:var(--muted); font-size:.84rem; white-space:nowrap; }
    .fallback { position:absolute; inset:0; z-index:2; display:none; place-items:center; padding:1rem; background:var(--bg); }
    .fallback.visible { display:grid; }
    .fallback-box { width:min(32rem,100%); border:1px solid var(--line); border-radius:8px; background:var(--surface); padding:1.1rem; box-shadow:var(--shadow); display:grid; gap:.78rem; }
    .fallback-box p { margin:0; color:var(--muted); line-height:1.45; }
    .fallback-actions { display:flex; flex-wrap:wrap; gap:.55rem; }
    @media (max-width: 760px) { .embed-note { flex-basis:100%; order:5; margin-left:0; white-space:normal; } }
    @media (max-width: 640px) { .title { flex-basis:100%; order:4; margin-left:0; } }
  </style>
</head>
<body>
  <nav class="bar">
    <a class="pill icon-only" href="../app-links" aria-label="Back to Apps"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" d="M15.75 19.5 8.25 12l7.5-7.5"/></svg></a>
    <a class="pill" id="silo-home" href="/" aria-label="Back to Silo">Silo</a>
    <a class="pill primary" href="` + esc(link.URL) + `" target="_blank" rel="noreferrer">Open externally</a>
    <span class="title">` + esc(link.Name) + `</span>
    <span class="embed-note">Blank screen? Open externally.</span>
  </nav>
  <div class="stage">
  <iframe id="app-frame" src="` + esc(link.URL) + `" title="` + esc(link.Name) + `" allow="fullscreen; autoplay; encrypted-media; clipboard-read; clipboard-write"></iframe>
  <section id="frame-fallback" class="fallback" aria-live="polite">
    <div class="fallback-box">
      <h1>` + esc(link.Name) + `</h1>
      <p>This app cannot be embedded here.</p>
      <p>Some services block iframe loading for account and security reasons. Open it externally to continue.</p>
      <div class="fallback-actions"><a class="pill primary" href="` + esc(link.URL) + `" target="_blank" rel="noreferrer">Open externally</a><a class="pill" href="../app-links">Back to Apps</a></div>
    </div>
  </section>
  </div>
  <script>
    window.APP_LINK = ` + string(data) + `;
    // Silo may be served under a sub-path; keep the Silo button inside it.
    // Plugin pages are mounted at <silo base>/api/v2/plugin-content/plugins/<id>/...,
    // so Silo's home is everything before the API prefix.
    const apiPrefix = location.pathname.indexOf("/api/v2/plugin-content/");
    if (apiPrefix >= 0) document.getElementById("silo-home").href = location.pathname.slice(0, apiPrefix) + "/";
    const frame = document.getElementById("app-frame");
    const fallback = document.getElementById("frame-fallback");
    const frameLoadTimer = window.setTimeout(() => fallback.classList.add("visible"), 3500);
    frame.addEventListener("load", () => {
      window.clearTimeout(frameLoadTimer);
      fallback.classList.remove("visible");
    });
    frame.addEventListener("error", () => fallback.classList.add("visible"));
  </script>
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
  <style>` + sharedCSS() + `
    main { padding:1.25rem; max-width:74rem; }
    .head { display:flex; justify-content:space-between; gap:1rem; align-items:end; margin-bottom:1rem; }
    .storage { margin-top:.22rem; color:var(--muted); overflow:hidden; text-overflow:ellipsis; }
    .layout { display:grid; grid-template-columns:minmax(0,1fr) 24rem; gap:1rem; align-items:start; }
    .panel { border:1px solid var(--line); border-radius:8px; background:var(--surface); padding:.85rem; }
    .list { display:grid; gap:.5rem; }
    .row { min-height:3.75rem; border:1px solid transparent; border-radius:8px; background:var(--surface-2); color:var(--text); display:grid; grid-template-columns:minmax(0,1fr) auto; gap:.7rem; align-items:center; text-align:left; padding:.72rem; cursor:pointer; transition:background-color .15s ease,border-color .15s ease,transform .15s ease; }
    .row:hover { border-color:color-mix(in oklch,var(--line),var(--text) 18%); transform:translateY(-1px); }
    .row strong, .row small { display:block; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; }
    .row small { color:var(--muted); margin-top:.15rem; }
    .tag { border:1px solid var(--line); border-radius:999px; color:var(--muted); padding:.22rem .48rem; font-size:.78rem; font-weight:850; white-space:nowrap; }
    .tag.on { border-color:var(--accent); color:var(--text); }
    label { display:grid; gap:.3rem; margin-bottom:.65rem; color:var(--muted); font-weight:750; }
    input, textarea, select { width:100%; min-height:2.75rem; border:1px solid var(--line); border-radius:8px; background:var(--field); color:var(--text); padding:.58rem .65rem; transition:border-color .15s ease,background-color .15s ease; }
    input:focus, textarea:focus, select:focus { border-color:var(--accent); }
    [aria-invalid="true"] { border-color:var(--danger); }
    .row[aria-current="true"] { border-color:var(--accent); }
    textarea { min-height:5rem; resize:vertical; }
    .actions { display:flex; gap:.5rem; flex-wrap:wrap; }
    button { border:1px solid var(--line); border-radius:8px; background:var(--surface-2); color:var(--text); padding:.58rem .75rem; min-height:2.75rem; font-weight:850; cursor:pointer; transition:background-color .15s ease,border-color .15s ease,transform .15s ease; }
    button:hover { border-color:color-mix(in oklch,var(--line),var(--text) 18%); transform:translateY(-1px); }
    button.primary { border-color:var(--accent); background:var(--accent); color:var(--accent-ink); }
    button.danger { border-color:color-mix(in oklch,var(--danger),oklch(96% .02 350) 10%); background:var(--danger); color:var(--accent-ink); }
    .check { min-height:2.75rem; display:flex; align-items:center; gap:.5rem; color:var(--text); }
    .check input { width:auto; min-height:auto; }
    .state { border:1px dashed var(--line); border-radius:8px; color:var(--muted); padding:1rem; text-align:center; }
    .form-status { min-height:2.4rem; border:1px solid var(--line); border-radius:8px; margin-bottom:.75rem; padding:.55rem .65rem; color:var(--muted); background:var(--field); }
    .form-status.ok { border-color:var(--accent); color:var(--text); }
    .form-status.error { border-color:var(--danger); color:var(--text); }
    button:disabled { cursor:not-allowed; opacity:.55; transform:none; }
    @media (max-width: 860px) { main { padding:1rem; } .layout { grid-template-columns:1fr; } .head { align-items:start; } }
    @media (max-width: 560px) { .head { display:grid; } .row { grid-template-columns:1fr; } }
  </style>
</head>
<body>
  <main>
    <div class="head"><div><h1>App Links</h1><div class="storage">Storage: ` + esc(dataPath) + `</div></div><button id="new">New link</button></div>
    <div class="layout">
      <section class="panel"><div id="list" class="list"></div></section>
      <form id="form" class="panel">
        <div id="form-status" class="form-status" role="status" aria-live="polite">Ready to edit links.</div>
        <label>Name<input id="name" required></label>
        <label>URL<input id="url" required placeholder="https://app.example.com"></label>
        <label>Description<textarea id="description"></textarea></label>
        <label>Category<input id="category" placeholder="Media"></label>
        <label>Icon URL<input id="iconUrl" placeholder="https://app.example.com/icon.png"></label>
        <label>Open mode<select id="openMode"><option value="new_tab" selected>New tab</option><option value="iframe">Inside Silo (embedded)</option></select></label>
        <label>Sort order<input id="sortOrder" type="number" value="0"></label>
        <label class="check"><input id="enabled" type="checkbox" checked> Enabled</label>
        <div class="actions"><button id="submit" class="primary" type="submit">Save link</button><button id="delete" class="danger" type="button">Delete</button></div>
      </form>
    </div>
  </main>
  <script>
    const base = location.pathname.endsWith("/app-links/admin") ? location.pathname.slice(0, -"/app-links/admin".length) : "";
    const route = path => base + path;
    let links = [], current = null, saving = false, deleting = false;
    const ids = ["name","url","description","category","iconUrl","openMode","sortOrder","enabled"];
    const el = id => document.getElementById(id);
    const esc = value => String(value || "").replace(/[&<>"']/g, ch => ({ "&":"&amp;", "<":"&lt;", ">":"&gt;", "\"":"&quot;", "'":"&#39;" })[ch]);
    function clearFieldError() { ids.forEach(id => el(id).removeAttribute("aria-invalid")); }
    function showFieldError(field) { if (field && ids.includes(field)) { el(field).setAttribute("aria-invalid", "true"); el(field).focus(); } }
    function setStatus(message, tone) { const node = el("form-status"); node.textContent = message; node.className = "form-status" + (tone ? " " + tone : ""); }
    function syncActions() { const submit = el("submit"); if (saving) submit.disabled = true; else submit.disabled = false; submit.textContent = saving ? "Saving..." : "Save link"; el("delete").disabled = deleting || !current; }
    function reset(message) { clearFieldError(); current = null; ids.forEach(id => { if (id === "enabled") el(id).checked = true; else if (id === "openMode") el(id).value = "new_tab"; else if (id === "sortOrder") el(id).value = "0"; else el(id).value = ""; }); setStatus(message || "Ready to create a link.", message ? "ok" : ""); syncActions(); }
    function edit(link) { clearFieldError(); current = link; ids.forEach(id => { if (id === "enabled") el(id).checked = !!link.enabled; else el(id).value = link[id] || (id === "openMode" ? "new_tab" : ""); }); setStatus("Editing " + link.name + "."); syncActions(); render(); }
    function statusLabel(link) { if (!link.enabled) return "Disabled"; return link.openMode === "iframe" ? "Embedded" : "New tab"; }
    function render() {
      el("list").innerHTML = links.length ? links.map(link => "<button class='row' type='button' data-id='" + esc(link.id) + "' aria-current='" + (current && current.id === link.id ? "true" : "false") + "'><span><strong>" + esc(link.name) + "</strong><small>" + esc(link.url) + "</small></span><span class='tag " + (link.enabled ? "on" : "") + "'>" + esc(statusLabel(link)) + "</span></button>").join("") : "<div class='state'>No links yet. Create the first one with the form.</div>";
      document.querySelectorAll("[data-id]").forEach(row => row.onclick = () => edit(links.find(link => link.id === row.dataset.id)));
    }
    async function load() {
      el("list").innerHTML = "<div class='state'>Loading links...</div>";
      try {
        const response = await fetch(route("/app-links/admin/api/links"), { credentials:"include" });
        if (!response.ok) throw new Error("load failed");
        const data = await response.json();
        links = data.links || [];
        render();
      } catch (error) {
        el("list").innerHTML = "<div class='state'>Unable to load links. <button class='pill' type='button' id='retry-admin-load'>Retry</button></div>";
        el("retry-admin-load").onclick = load;
      }
    }
    el("new").onclick = reset;
    el("form").onsubmit = async event => {
      event.preventDefault();
      if (saving) return;
      saving = true; syncActions(); clearFieldError(); setStatus("Saving link...", "");
      const payload = { id: current && current.id || "", name: el("name").value, url: el("url").value, description: el("description").value, category: el("category").value, iconUrl: el("iconUrl").value, openMode: el("openMode").value, sortOrder: Number(el("sortOrder").value || 0), enabled: el("enabled").checked };
      try {
        const response = await fetch(route("/app-links/admin/api/links"), { method:"POST", credentials:"include", headers:{ "content-type":"application/json" }, body:JSON.stringify(payload) });
        if (!response.ok) { const data = await response.json().catch(() => ({})); const failure = new Error(data.error || "Save failed. Try again."); failure.field = data.field; throw failure; }
        reset("Link saved.");
        await load();
      } catch (error) {
        setStatus(error.message || "Save failed.", "error");
        showFieldError(error.field);
      } finally {
        saving = false; syncActions();
      }
    };
    el("delete").onclick = async () => {
      if (!current || deleting || !confirm("Delete " + current.name + "?")) return;
      deleting = true; syncActions(); setStatus("Deleting link...", "");
      try {
        const response = await fetch(route("/app-links/admin/api/delete"), { method:"POST", credentials:"include", headers:{ "content-type":"application/json" }, body:JSON.stringify({ id: current.id }) });
        if (!response.ok) { const data = await response.json().catch(() => ({})); throw new Error(data.error || "Delete failed. Try again."); }
        reset("Link deleted.");
        await load();
      } catch (error) {
        setStatus(error.message || "Delete failed.", "error");
      } finally {
        deleting = false; syncActions();
      }
    };
    reset(); load();
  </script>
</body>
</html>`
}

func notFoundPageHTML() string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width,initial-scale=1">
  <title>App not found</title>
  <style>` + sharedCSS() + `
    main { min-height:100vh; display:grid; place-items:center; padding:1rem; }
    .box { width:min(30rem,100%); border:1px solid var(--line); border-radius:8px; background:var(--surface); padding:1.1rem; display:grid; gap:.78rem; }
    .box p { margin:0; color:var(--muted); line-height:1.45; }
  </style>
</head>
<body>
  <main><div class="box"><h1>App not found</h1><p>This app link was removed or turned off by an administrator.</p><div><a class="pill primary" href="../app-links">Back to Apps</a></div></div></main>
</body>
</html>`
}

func compact(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
