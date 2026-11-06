import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import test from 'node:test';
import vm from 'node:vm';

// Execute the script served by adminPageHTML. The DOM fixture only supplies
// browser APIs; admission, selection and retry behavior come from that script.
const page = readFileSync(new URL('../internal/plugin/pages.go', import.meta.url), 'utf8').split('func adminPageHTML')[1];
const script = page.split('<script>')[1].split('</script>')[0];
const fields = ['name', 'url', 'description', 'category', 'iconUrl', 'openMode', 'sortOrder', 'enabled'];

function harness() {
  const nodes = new Map();
  let rows = [];
  function node(id) {
    if (!nodes.has(id)) {
      const value = { value: '', checked: false, disabled: false, textContent: '', dataset: {}, removeAttribute() {}, setAttribute() {}, focus() {} };
      value.click = () => value.disabled ? undefined : value.onclick?.();
      Object.defineProperty(value, 'innerHTML', { set(html) {
        if (id === 'list') rows = [...html.matchAll(/data-id='([^']+)'/g)].map(match => ({ dataset: { id: match[1] }, disabled: false }));
      }});
      nodes.set(id, value);
    }
    return nodes.get(id);
  }
  const links = [
    { id: 'one', name: 'One', url: 'https://one.example', enabled: true, openMode: 'new_tab' },
    { id: 'two', name: 'Two', url: 'https://two.example', enabled: true, openMode: 'iframe' },
  ];
  const writes = [];
  const context = vm.createContext({
    location: { pathname: '/app-links/admin' }, confirm: () => true,
    document: { getElementById: node, querySelectorAll: () => rows },
    fetch: async (url, options = {}) => {
      if (options.method !== 'POST') return { ok: true, json: async () => ({ links }) };
      return new Promise(resolve => writes.push({ url, payload: JSON.parse(options.body), resolve }));
    },
  });
  vm.runInContext(script, context);
  return { node, writes, rows: () => rows };
}

for (const action of ['save', 'delete']) {
  test(`${action} serializes mutations and selection; failure allows retry`, async () => {
    const h = harness();
    await new Promise(setImmediate);
    h.rows()[0].onclick();
    const submit = () => h.node('form').onsubmit({ preventDefault() {} });
    const start = () => action === 'save' ? submit() : h.node('delete').onclick();
    const pending = start();
    assert.equal(h.writes.length, 1);
    assert.equal(h.node('submit').disabled, true, 'save remains disabled during either mutation');
    assert.equal(h.node('delete').disabled, true, 'delete remains disabled during either mutation');
    assert.equal(h.node('new').disabled, true, 'new draft cannot replace the pending editor');
    for (const field of fields) assert.equal(h.node(field).disabled, true, `${field} is locked during mutation`);
    assert.ok(h.rows().every(row => row.disabled), 'selection stays locked');
    await submit();
    await h.node('delete').onclick();
    h.node('new').click();
    h.rows()[1].onclick(); // Even stale queued selection events must be ignored.
    assert.equal(h.node('name').value, 'One');
    assert.equal(h.writes.length, 1, 'overlapping actions must not send another write');
    h.writes[0].resolve({ ok: false, json: async () => ({ error: 'Try again' }) });
    await pending;
    assert.equal(h.node('submit').disabled, false);
    assert.equal(h.node('delete').disabled, false);
    assert.equal(h.node('name').disabled, false);
    assert.equal(h.node('name').value, 'One');
    const retry = start();
    assert.equal(h.writes.length, 2);
    assert.equal(h.writes[1].payload.id, 'one');
    h.writes[1].resolve({ ok: true });
    await retry;
    assert.equal(h.node('name').value, '');
    assert.equal(h.node('submit').disabled, false);
    assert.equal(h.node('new').disabled, false);
  });
}
