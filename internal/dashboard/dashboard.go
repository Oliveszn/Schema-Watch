package dashboard

import (
	"net/http"

	"github.com/Oliveszn/Schema-Watch/internal/store"
	"github.com/gin-gonic/gin"
)

type Dashboard struct {
	store *store.Store
}

func New(st *store.Store) *Dashboard {
	return &Dashboard{store: st}
}

func (d *Dashboard) RegisterRoutes(r *gin.Engine) {
	r.GET("/__schema-watch/dashboard", d.serveHTML)
	r.GET("/__schema-watch/api/endpoints", d.listEndpoints)
	r.GET("/__schema-watch/api/history", d.getHistory)
}

func (d *Dashboard) serveHTML(c *gin.Context) {
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(dashboardHTML))
}

func (d *Dashboard) listEndpoints(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"endpoints": d.store.Endpoints()})
}

func (d *Dashboard) getHistory(c *gin.Context) {
	endpoint := c.Query("endpoint")
	if endpoint == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endpoint query param is required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"history": d.store.History(endpoint)})
}

const dashboardHTML = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>schema-watch</title>
<style>
  :root { color-scheme: dark; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
    background: #0f1115;
    color: #e6e6e6;
    margin: 0;
    padding: 2rem;
  }
  h1 { font-size: 1.4rem; margin-bottom: 0.25rem; }
  .subtitle { color: #888; font-size: 0.85rem; margin-bottom: 2rem; }
  .endpoint {
    background: #171a21;
    border: 1px solid #262a33;
    border-radius: 8px;
    padding: 1rem 1.25rem;
    margin-bottom: 1rem;
  }
  .endpoint h2 {
    font-size: 0.95rem;
    font-family: ui-monospace, monospace;
    margin: 0 0 0.5rem 0;
    color: #9cd1ff;
  }
  .clean { color: #5a5; font-size: 0.85rem; margin: 0; }
  .diff {
    border-left: 3px solid #d4a72c;
    padding: 0.5rem 0.75rem;
    margin: 0.5rem 0;
    background: #1c1f26;
    border-radius: 4px;
  }
  .diff.breaking { border-left-color: #e5484d; }
  .diff strong { font-size: 0.75rem; letter-spacing: 0.05em; opacity: 0.8; }
  .diff ul { margin: 0.4rem 0 0 0; padding-left: 1.1rem; font-family: ui-monospace, monospace; font-size: 0.85rem; }
  .diff li.added { color: #5ec26a; }
  .diff li.removed { color: #e5484d; }
  .diff li.typechanged { color: #d4a72c; }
  .empty-state { color: #666; }
</style>
</head>
<body>
  <h1>schema-watch</h1>
  <p class="subtitle">watching endpoints for API contract changes — refreshes every 3s</p>
  <div id="endpoints"><p class="empty-state">Loading…</p></div>

<script>
async function refresh() {
  const container = document.getElementById('endpoints');
  try {
    const res = await fetch('/__schema-watch/api/endpoints');
    const data = await res.json();
    const endpoints = data.endpoints || [];

    if (endpoints.length === 0) {
      container.innerHTML = '<p class="empty-state">No traffic observed yet — hit an endpoint through the proxy.</p>';
      return;
    }

    const sections = await Promise.all(endpoints.sort().map(renderEndpoint));
    container.innerHTML = sections.join('');
  } catch (e) {
    container.innerHTML = '<p class="empty-state">Failed to reach schema-watch API.</p>';
  }
}

async function renderEndpoint(ep) {
  const res = await fetch('/__schema-watch/api/history?endpoint=' + encodeURIComponent(ep));
  const data = await res.json();
  const history = data.history || [];

  const body = history.length
    ? history.slice().reverse().map(renderDiff).join('')
    : '<p class="clean">no changes detected</p>';

  return '<div class="endpoint"><h2>' + escapeHTML(ep) + '</h2>' + body + '</div>';
}

function renderDiff(diff) {
  const cls = diff.breaking ? 'breaking' : 'change';
  const label = diff.breaking ? 'BREAKING' : 'CHANGE';
  const changes = (diff.changes || []).map(renderChange).join('');
  return '<div class="diff ' + cls + '"><strong>' + label + '</strong><ul>' + changes + '</ul></div>';
}

function renderChange(c) {
  if (c.type === 'added') {
    return '<li class="added">+ ' + escapeHTML(c.path) + ' added (' + c.new_type + ')</li>';
  }
  if (c.type === 'removed') {
    return '<li class="removed">- ' + escapeHTML(c.path) + ' removed (was ' + c.old_type + ')</li>';
  }
  return '<li class="typechanged">~ ' + escapeHTML(c.path) + ' changed: ' + c.old_type + ' -> ' + c.new_type + '</li>';
}

function escapeHTML(s) {
  const d = document.createElement('div');
  d.innerText = s;
  return d.innerHTML;
}

refresh();
setInterval(refresh, 3000);
</script>
</body>
</html>`
