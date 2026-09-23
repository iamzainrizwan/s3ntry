const $ = (sel) => document.querySelector(sel);

function fmtTime(iso) {
  return new Date(iso).toLocaleTimeString();
}

function renderServices(services) {
  const body = $("#services-body");
  body.innerHTML = "";
  const names = Object.keys(services).sort();
  for (const name of names) {
    const s = services[name];
    const tr = document.createElement("tr");
    const cls = s.up ? "up" : "down";
    tr.innerHTML = `
      <td>${s.target}</td>
      <td><span class="pill ${cls}"><span class="dot ${cls}"></span>${s.up ? "up" : "down"}</span></td>
      <td>${s.latency_ms}ms</td>
      <td>${fmtTime(s.checked_at)}</td>
    `;
    body.appendChild(tr);
  }
}

function renderHost(host) {
  const el = $("#host-cards");
  el.innerHTML = `
    <div class="card ${host.connectivity ? "ok" : "bad"}">
      <div class="label">connectivity</div>
      <div class="value">${host.connectivity ? "online" : "offline"}</div>
    </div>
    <div class="card ${host.reboot_required ? "warn" : "ok"}">
      <div class="label">reboot required</div>
      <div class="value">${host.reboot_required ? "yes" : "no"}</div>
    </div>
    <div class="card ${host.updates_available > 0 ? "warn" : "ok"}">
      <div class="label">updates available</div>
      <div class="value">${host.updates_available}</div>
    </div>
  `;
}

async function refresh() {
  try {
    const resp = await fetch("status");
    const data = await resp.json();
    renderServices(data.services || {});
    renderHost(data.host || {});
    $("#last-updated").textContent = "updated " + new Date().toLocaleTimeString();
  } catch (err) {
    $("#last-updated").textContent = "fetch failed: " + err;
  }
}

refresh();
setInterval(refresh, 5000);
