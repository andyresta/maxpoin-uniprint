// Package ui menyimpan HTML single-page UI PrintBridge sebagai string
// constant Go murni. Pendekatan ini sengaja dipakai (alih-alih
// //go:embed file terpisah) agar binary final benar-benar self-contained
// tanpa membutuhkan file pendamping ketika di-distribusikan.
package ui

// IndexHTML berisi seluruh markup HTML + CSS + JavaScript yang
// digunakan untuk halaman utama PrintBridge. Tidak ada CDN eksternal;
// semua aset di-inline.
const IndexHTML = `<!DOCTYPE html>
<html lang="id">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PrintBridge - Service Bridge Printer</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  :root {
    --bg: #0f1419;
    --bg-elev: #1a1f2e;
    --bg-card: #232a3d;
    --border: #2d3548;
    --text: #e6e9ef;
    --text-dim: #8a93a6;
    --accent: #4f8cff;
    --accent-hover: #6ba0ff;
    --success: #3fb950;
    --warning: #d29922;
    --danger: #f85149;
    --shadow: 0 4px 12px rgba(0,0,0,0.35);
  }
  html, body {
    background: var(--bg);
    color: var(--text);
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    min-height: 100vh;
    font-size: 14px;
    line-height: 1.5;
  }
  a { color: var(--accent); text-decoration: none; }
  .container {
    max-width: 1100px;
    margin: 0 auto;
    padding: 24px 16px 80px 16px;
  }
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 0 24px 0;
    border-bottom: 1px solid var(--border);
    margin-bottom: 24px;
  }
  header h1 {
    font-size: 20px;
    font-weight: 600;
    letter-spacing: 0.3px;
    display: flex;
    align-items: center;
    gap: 10px;
  }
  header h1 .dot {
    width: 10px; height: 10px; border-radius: 50%;
    background: var(--success);
    box-shadow: 0 0 8px var(--success);
  }
  header .ver { color: var(--text-dim); font-size: 12px; }
  .tabs {
    display: flex;
    gap: 4px;
    margin-bottom: 20px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    padding: 4px;
    border-radius: 10px;
    overflow-x: auto;
  }
  .tab {
    flex: 1;
    min-width: 130px;
    padding: 10px 14px;
    text-align: center;
    cursor: pointer;
    border-radius: 7px;
    font-weight: 500;
    color: var(--text-dim);
    background: transparent;
    border: none;
    font-size: 14px;
    transition: all 0.15s ease;
  }
  .tab:hover { color: var(--text); background: rgba(255,255,255,0.04); }
  .tab.active {
    color: #fff;
    background: var(--accent);
    box-shadow: var(--shadow);
  }
  .panel { display: none; }
  .panel.active { display: block; }
  .card {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-radius: 10px;
    padding: 20px;
    box-shadow: var(--shadow);
    margin-bottom: 16px;
  }
  .card h2 {
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 16px;
    color: var(--text);
    display: flex;
    align-items: center;
    gap: 10px;
  }
  .card h2 .badge { font-size: 11px; padding: 2px 8px; border-radius: 10px; background: var(--accent); color: #fff; font-weight: 500; }
  .row { display: flex; flex-wrap: wrap; gap: 12px; align-items: center; margin-bottom: 14px; }
  .row > label { color: var(--text-dim); min-width: 130px; font-size: 13px; }
  .row > .grow { flex: 1; min-width: 220px; }
  input[type="text"], input[type="number"], textarea, select {
    background: var(--bg-elev);
    color: var(--text);
    border: 1px solid var(--border);
    padding: 9px 12px;
    border-radius: 7px;
    font-size: 13px;
    font-family: inherit;
    width: 100%;
    outline: none;
    transition: border-color 0.15s ease;
  }
  input:focus, textarea:focus, select:focus { border-color: var(--accent); }
  textarea { min-height: 110px; resize: vertical; font-family: ui-monospace, "Cascadia Code", "Consolas", monospace; }
  input[type="file"] { color: var(--text-dim); font-size: 12px; }
  input[type="checkbox"], input[type="radio"] { accent-color: var(--accent); width: 16px; height: 16px; cursor: pointer; }
  .checkbox-row, .radio-row { display: flex; align-items: center; gap: 8px; cursor: pointer; }
  .radio-group { display: flex; gap: 16px; }
  button {
    background: var(--accent);
    color: #fff;
    border: none;
    padding: 9px 18px;
    border-radius: 7px;
    cursor: pointer;
    font-size: 13px;
    font-weight: 500;
    transition: background 0.15s ease;
  }
  button:hover:not(:disabled) { background: var(--accent-hover); }
  button:disabled { opacity: 0.5; cursor: not-allowed; }
  button.secondary { background: var(--bg-elev); border: 1px solid var(--border); color: var(--text); }
  button.secondary:hover:not(:disabled) { background: var(--border); }
  button.danger { background: var(--danger); }
  .btn-group { display: flex; gap: 10px; flex-wrap: wrap; }
  .status-pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    padding: 4px 10px;
    border-radius: 12px;
    font-size: 12px;
    font-weight: 500;
    background: var(--bg-elev);
    border: 1px solid var(--border);
  }
  .status-pill.success { color: var(--success); border-color: rgba(63,185,80,0.4); }
  .status-pill.failed  { color: var(--danger);  border-color: rgba(248,81,73,0.4); }
  .status-pill.pending { color: var(--warning); border-color: rgba(210,153,34,0.4); }
  .active-printer-banner {
    display: flex; align-items: center; gap: 12px;
    padding: 12px 14px;
    background: linear-gradient(180deg, rgba(79,140,255,0.15), rgba(79,140,255,0.05));
    border: 1px solid rgba(79,140,255,0.35);
    border-radius: 8px;
    margin-bottom: 16px;
    font-size: 13px;
  }
  .active-printer-banner .label { color: var(--text-dim); font-size: 12px; text-transform: uppercase; letter-spacing: 0.5px; }
  .active-printer-banner .value { font-weight: 600; }
  table { width: 100%; border-collapse: collapse; font-size: 12.5px; }
  th, td { padding: 10px 8px; text-align: left; border-bottom: 1px solid var(--border); }
  th { color: var(--text-dim); font-weight: 500; text-transform: uppercase; font-size: 11px; letter-spacing: 0.5px; background: var(--bg-elev); position: sticky; top: 0; }
  tbody tr:hover { background: rgba(255,255,255,0.02); }
  .table-wrap { overflow-x: auto; max-height: 480px; overflow-y: auto; border: 1px solid var(--border); border-radius: 8px; }
  td.success { color: var(--success); font-weight: 500; }
  td.failed  { color: var(--danger);  font-weight: 500; }
  td.pending { color: var(--warning); font-weight: 500; }
  td .mono { font-family: ui-monospace, "Cascadia Code", "Consolas", monospace; font-size: 11.5px; color: var(--text-dim); }
  .response-box {
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 7px;
    padding: 12px;
    font-family: ui-monospace, "Cascadia Code", "Consolas", monospace;
    font-size: 12px;
    margin-top: 12px;
    white-space: pre-wrap;
    word-break: break-word;
    max-height: 260px; overflow: auto;
  }
  .preview-box {
    background: var(--bg-elev);
    border: 1px dashed var(--border);
    border-radius: 7px;
    padding: 12px;
    margin-top: 8px;
    text-align: center;
  }
  .preview-box img { max-width: 200px; max-height: 200px; border-radius: 4px; }
  .toast-stack { position: fixed; top: 20px; right: 20px; display: flex; flex-direction: column; gap: 10px; z-index: 999; }
  .toast {
    background: var(--bg-card);
    border: 1px solid var(--border);
    border-left: 4px solid var(--accent);
    padding: 12px 16px;
    border-radius: 7px;
    box-shadow: var(--shadow);
    min-width: 280px;
    max-width: 380px;
    font-size: 13px;
    animation: slideIn 0.25s ease;
  }
  .toast.success { border-left-color: var(--success); }
  .toast.error   { border-left-color: var(--danger); }
  .toast.warning { border-left-color: var(--warning); }
  @keyframes slideIn { from { transform: translateX(20px); opacity: 0; } to { transform: translateX(0); opacity: 1; } }
  .empty { padding: 30px; text-align: center; color: var(--text-dim); }
  .footer-note { color: var(--text-dim); font-size: 11.5px; margin-top: 6px; }
  /* ===== Mode toggle (Thermal / Dot Matrix) ===== */
  .mode-toggle {
    display: inline-flex;
    gap: 4px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    padding: 3px;
    border-radius: 8px;
    margin-bottom: 16px;
  }
  .mode-btn {
    padding: 7px 18px;
    border-radius: 6px;
    font-size: 13px;
    font-weight: 500;
    color: var(--text-dim);
    background: transparent;
    border: none;
    cursor: pointer;
    transition: all 0.15s ease;
    white-space: nowrap;
  }
  .mode-btn:hover:not(:disabled) { color: var(--text); background: rgba(255,255,255,0.04); }
  .mode-btn.active {
    background: var(--bg-card);
    color: var(--text);
    box-shadow: 0 1px 5px rgba(0,0,0,0.35);
  }
  .mode-btn.active.thermal  { color: var(--accent); }
  .mode-btn.active.dotmatrix { color: var(--success); }
  .mode-notice {
    display: flex;
    align-items: flex-start;
    gap: 10px;
    padding: 11px 13px;
    background: rgba(210,153,34,0.08);
    border: 1px solid rgba(210,153,34,0.3);
    border-radius: 7px;
    font-size: 12.5px;
    color: var(--warning);
    line-height: 1.5;
    margin-bottom: 14px;
  }
  .printer-list-status {
    display: none;
    align-items: center;
    gap: 10px;
    padding: 10px 12px;
    margin-bottom: 14px;
    background: var(--bg-elev);
    border: 1px solid var(--border);
    border-radius: 7px;
    font-size: 13px;
    color: var(--text-dim);
  }
  .printer-list-status.visible { display: flex; }
  .printer-list-status .spinner {
    flex-shrink: 0;
    width: 18px; height: 18px;
    border: 2px solid var(--border);
    border-top-color: var(--accent);
    border-radius: 50%;
    animation: pb-spin 0.65s linear infinite;
  }
  @keyframes pb-spin { to { transform: rotate(360deg); } }
  @media (max-width: 720px) {
    .row > label { min-width: 100%; }
    .tabs { flex-wrap: wrap; }
    .tab { flex: 1 1 48%; }
  }
</style>
</head>
<body>

<div class="container">
  <header>
    <h1><span class="dot"></span> PrintBridge <span class="ver">v1.0.0</span></h1>
    <div class="status-pill" id="serverStatus">terhubung</div>
  </header>

  <div class="tabs" role="tablist">
    <button class="tab active" data-tab="settings">Pengaturan</button>
    <button class="tab" data-tab="logs">Log Cetak</button>
    <button class="tab" data-tab="testText">Test Cetak Teks</button>
    <button class="tab" data-tab="testImage">Test Cetak Gambar</button>
  </div>

  <!-- ============ TAB: SETTINGS ============ -->
  <section class="panel active" id="panel-settings">
    <div class="active-printer-banner" id="activeBanner">
      <div>
        <div class="label">Printer Aktif</div>
        <div class="value" id="activePrinterText">Belum dikonfigurasi</div>
      </div>
    </div>

    <div class="card">
      <h2>Pengaturan Printer <span class="badge">Persisten</span></h2>
      <div class="row">
        <label>Pilih Printer</label>
        <div class="grow">
          <select id="printerSelect"><option value="">— Memuat printer —</option></select>
        </div>
      </div>
      <div id="printerListStatus" class="printer-list-status" role="status" aria-live="polite" aria-atomic="true">
        <span class="spinner" aria-hidden="true"></span>
        <span id="printerListStatusText">Memuat daftar printer…</span>
      </div>
      <div class="row">
        <label>Mode Printer</label>
        <div class="grow radio-group">
          <label class="radio-row">
            <input type="radio" name="printerMode" value="thermal" checked>
            <span>Thermal <span style="color:var(--text-dim);font-size:12px;">(ESC/POS — bitmap, logo, cut)</span></span>
          </label>
          <label class="radio-row">
            <input type="radio" name="printerMode" value="dotmatrix">
            <span>Dot Matrix <span style="color:var(--text-dim);font-size:12px;">(plain text — tanpa ESC/POS)</span></span>
          </label>
        </div>
      </div>
      <div class="row">
        <label>Lebar Kertas</label>
        <div class="grow radio-group">
          <label class="radio-row"><input type="radio" name="paperWidth" value="58"> 58 mm</label>
          <label class="radio-row"><input type="radio" name="paperWidth" value="80" checked> 80 mm</label>
        </div>
      </div>
      <div class="btn-group" style="margin-top: 8px;">
        <button id="btnSaveSettings">Simpan Pengaturan</button>
        <button id="btnRefreshPrinters" class="secondary">Refresh Daftar Printer</button>
      </div>
      <div class="footer-note">Tipe transport (Spooler / Bluetooth) otomatis mengikuti jenis printer yang dipilih. Tersimpan ke <code>config.json</code>.</div>
    </div>
  </section>

  <!-- ============ TAB: LOGS ============ -->
  <section class="panel" id="panel-logs">
    <div class="card">
      <h2>Log Cetak <span class="badge" id="logCount">0</span></h2>
      <div class="btn-group" style="margin-bottom: 12px;">
        <button id="btnRefreshLogs" class="secondary">Refresh Sekarang</button>
        <button id="btnClearDisplay" class="secondary">Bersihkan Tampilan</button>
        <span class="footer-note" style="align-self:center;">Auto-refresh setiap 5 detik</span>
      </div>
      <div class="table-wrap">
        <table id="logsTable">
          <thead>
            <tr>
              <th>Waktu</th><th>Job ID</th><th>Printer</th><th>Tipe</th>
              <th>Status</th><th>Durasi</th><th>Pesan</th>
            </tr>
          </thead>
          <tbody><tr><td colspan="7" class="empty">Belum ada log…</td></tr></tbody>
        </table>
      </div>
    </div>
  </section>

  <!-- ============ TAB: TEST TEXT ============ -->
  <section class="panel" id="panel-testText">
    <div class="card">
      <h2>Test Cetak Teks</h2>

      <!-- Mode toggle -->
      <div class="mode-toggle" id="textModeToggle" role="group" aria-label="Mode cetak teks">
        <button class="mode-btn thermal active" data-mode="thermal">Thermal (ESC/POS)</button>
        <button class="mode-btn dotmatrix" data-mode="dotmatrix">Dot Matrix / Plain Text</button>
      </div>

      <!-- Konten teks (selalu tampil) -->
      <div class="row">
        <label>Konten Teks</label>
        <div class="grow">
          <textarea id="textContent" placeholder="Enter text to print here..."></textarea>
        </div>
      </div>

      <!-- Opsi khusus Thermal -->
      <div id="textThermalOpts">
        <div class="row">
          <label>Logo (opsional)</label>
          <div class="grow">
            <input type="file" id="textLogoFile" accept="image/png,image/jpeg,image/bmp">
            <div class="preview-box" id="textLogoPreview" style="display:none;"><img id="textLogoImg"></div>
          </div>
        </div>
        <div class="row">
          <label>Posisi Logo</label>
          <div class="grow">
            <select id="textLogoPos">
              <option value="left">Kiri</option>
              <option value="center" selected>Tengah</option>
              <option value="right">Kanan</option>
            </select>
          </div>
        </div>
        <div class="row">
          <label>Cut Kertas</label>
          <div class="grow"><label class="checkbox-row"><input type="checkbox" id="textCut" checked> Potong kertas setelah cetak</label></div>
        </div>
      </div>

      <!-- Keterangan mode dot matrix -->
      <div class="mode-notice" id="textDotmatrixNotice" style="display:none;">
        Mode <strong>Dot Matrix / Plain Text</strong>: teks dikirim langsung sebagai byte polos tanpa ESC/POS.
        Logo, alignment, dan perintah cut paper tidak tersedia pada mode ini.
      </div>

      <div class="row">
        <label>Copies</label>
        <div class="grow"><input type="number" id="textCopies" value="1" min="1" max="10"></div>
      </div>
      <div class="btn-group">
        <button id="btnPrintText">Kirim Test Print</button>
      </div>
      <div class="response-box" id="textResponse" style="display:none;"></div>
      <div class="footer-note">Printer yang dipakai: printer aktif pada tab Pengaturan. Mode default diambil dari Pengaturan.</div>
    </div>
  </section>

  <!-- ============ TAB: TEST IMAGE ============ -->
  <section class="panel" id="panel-testImage">
    <div class="card">
      <h2>Test Cetak Gambar</h2>

      <!-- Mode toggle -->
      <div class="mode-toggle" id="imgModeToggle" role="group" aria-label="Mode cetak gambar">
        <button class="mode-btn thermal active" data-mode="thermal">Thermal (ESC/POS)</button>
        <button class="mode-btn dotmatrix" data-mode="dotmatrix">Dot Matrix / Plain Text</button>
      </div>

      <!-- Opsi Thermal -->
      <div id="imgThermalOpts">
        <div class="row">
          <label>File Gambar</label>
          <div class="grow">
            <input type="file" id="imgFile" accept="image/png,image/jpeg,image/bmp">
            <div class="preview-box" id="imgPreview" style="display:none;"><img id="imgPreviewImg"></div>
          </div>
        </div>
        <div class="row">
          <label>Dithering</label>
          <div class="grow">
            <select id="imgDither">
              <option value="none">None (Threshold)</option>
              <option value="floyd-steinberg" selected>Floyd-Steinberg</option>
              <option value="atkinson">Atkinson</option>
            </select>
          </div>
        </div>
        <div class="row">
          <label>Cut Kertas</label>
          <div class="grow"><label class="checkbox-row"><input type="checkbox" id="imgCut" checked> Potong kertas setelah cetak</label></div>
        </div>
        <div class="row">
          <label>Copies</label>
          <div class="grow"><input type="number" id="imgCopies" value="1" min="1" max="10"></div>
        </div>
      </div>

      <!-- Pesan mode dot matrix (gambar tidak didukung) -->
      <div class="mode-notice" id="imgDotmatrixNotice" style="display:none;">
        Mode <strong>Dot Matrix / Plain Text</strong> tidak mendukung cetak gambar raster (ESC/POS bitmap).
        Untuk mencetak gambar, pilih mode <strong>Thermal (ESC/POS)</strong>.
      </div>

      <div class="btn-group">
        <button id="btnPrintImage">Kirim Test Print</button>
      </div>
      <div class="response-box" id="imgResponse" style="display:none;"></div>
      <div class="footer-note">Printer & lebar kertas mengikuti konfigurasi pada tab Pengaturan. Mode default diambil dari Pengaturan.</div>
    </div>
  </section>

</div>

<div class="toast-stack" id="toastStack"></div>

<script>
'use strict';

const API_BASE = '';
let currentConfig = null;
let knownPrinters = [];
let textLogoBase64 = '';
let imageBase64 = '';
let logsTimer = null;
// Mode aktif masing-masing panel test: 'thermal' atau 'dotmatrix'.
let textPrintMode = 'thermal';
let imgPrintMode  = 'thermal';

// ============================
//  Utilitas umum
// ============================
function $(sel) { return document.querySelector(sel); }
function $$(sel) { return Array.from(document.querySelectorAll(sel)); }

function showToast(msg, kind) {
  const t = document.createElement('div');
  t.className = 'toast ' + (kind || '');
  t.textContent = msg;
  $('#toastStack').appendChild(t);
  setTimeout(() => { t.style.opacity = '0'; t.style.transition = 'opacity .3s'; }, 3000);
  setTimeout(() => t.remove(), 3400);
}

// Konvensi proyek: SEMUA call API/AJAX menggunakan method POST.
// Server PrintBridge sudah disiapkan untuk menerima POST sebagai alias
// dari GET/PUT pada endpoint yang relevan. Untuk endpoint /api/config:
//   - POST tanpa body  → read (sama seperti GET)
//   - POST dengan body → update (sama seperti PUT)
async function apiCall(path, body) {
  const opts = {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
  };
  if (body !== undefined && body !== null) opts.body = body;
  const res = await fetch(API_BASE + path, opts);
  let data;
  try { data = await res.json(); } catch(e) { data = { _raw: await res.text() }; }
  return { ok: res.ok, status: res.status, data };
}

function fileToBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const result = reader.result;
      const idx = result.indexOf(',');
      resolve(idx >= 0 ? result.substring(idx + 1) : result);
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}

function fmtTime(iso) {
  try {
    const d = new Date(iso);
    return d.toLocaleString();
  } catch(e) { return iso; }
}
function shortId(id) {
  if (!id) return '-';
  return String(id).replace(/-/g,'').substring(0,8);
}

// ============================
//  Tab switching
// ============================
$$('.tab').forEach(btn => {
  btn.addEventListener('click', () => {
    const target = btn.dataset.tab;
    $$('.tab').forEach(t => t.classList.toggle('active', t === btn));
    $$('.panel').forEach(p => p.classList.toggle('active', p.id === 'panel-' + target));
    if (target === 'logs') { refreshLogs(); }
  });
});

// ============================
//  Settings: load/save
// ============================
async function loadConfig() {
  const r = await apiCall('/api/config');
  if (!r.ok) { showToast('Gagal memuat config: ' + (r.data.message || r.status), 'error'); return; }
  currentConfig = r.data;
  applyConfigToForm();
  updateActiveBanner();
}

// applyModeToggle menetapkan tombol aktif pada sebuah .mode-toggle container
// dan menampilkan/menyembunyikan panel opsi yang sesuai.
function applyModeToggle(toggleId, mode) {
  const toggle = $('#' + toggleId);
  if (!toggle) return;
  toggle.querySelectorAll('.mode-btn').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.mode === mode);
  });
  if (toggleId === 'textModeToggle') {
    textPrintMode = mode;
    $('#textThermalOpts').style.display     = mode === 'thermal'   ? '' : 'none';
    $('#textDotmatrixNotice').style.display = mode === 'dotmatrix' ? '' : 'none';
  } else if (toggleId === 'imgModeToggle') {
    imgPrintMode = mode;
    $('#imgThermalOpts').style.display      = mode === 'thermal'   ? '' : 'none';
    $('#imgDotmatrixNotice').style.display  = mode === 'dotmatrix' ? '' : 'none';
    $('#btnPrintImage').disabled            = mode === 'dotmatrix';
  }
}

// Pasang event listener untuk semua mode-toggle di halaman.
$$('.mode-toggle').forEach(toggle => {
  toggle.addEventListener('click', e => {
    const btn = e.target.closest('.mode-btn');
    if (!btn) return;
    applyModeToggle(toggle.id, btn.dataset.mode);
  });
});

function applyConfigToForm() {
  if (!currentConfig) return;
  $$('input[name="paperWidth"]').forEach(r => { r.checked = (parseInt(r.value, 10) === currentConfig.default_paper_width_mm); });
  // Set radio printer mode.
  const cfgMode = currentConfig.printer_mode || 'thermal';
  $$('input[name="printerMode"]').forEach(r => { r.checked = (r.value === cfgMode); });
  // Sinkronkan mode toggle di tab test agar default = mode dari config.
  applyModeToggle('textModeToggle', cfgMode);
  applyModeToggle('imgModeToggle', cfgMode);
  // Pilih printer di dropdown jika ada.
  const sel = $('#printerSelect');
  if (sel && currentConfig.default_printer) {
    let found = false;
    Array.from(sel.options).forEach(o => { if (o.value === currentConfig.default_printer) found = true; });
    if (!found && currentConfig.default_printer) {
      const opt = document.createElement('option');
      opt.value = currentConfig.default_printer;
      opt.textContent = currentConfig.default_printer + ' (tidak terdeteksi)';
      sel.appendChild(opt);
    }
    sel.value = currentConfig.default_printer;
  }
}

// getPrinterTypeForName mengembalikan printer_type yang akan disimpan ke
// config: diambil dari hasil deteksi (knownPrinters) agar selalu selaras
// dengan pilihan di dropdown. Jika printer tidak ada di daftar (mis. tidak
// terdeteksi), gunakan nilai config saat ini bila namanya sama, jika tidak
// default spooler.
function getPrinterTypeForName(name) {
  if (!name) return 'spooler';
  const p = knownPrinters.find(function(x) { return x.name === name; });
  if (p && p.type === 'bluetooth') return 'bluetooth';
  if (p && p.type === 'spooler') return 'spooler';
  if (currentConfig && currentConfig.default_printer === name && currentConfig.printer_type === 'bluetooth') {
    return 'bluetooth';
  }
  return 'spooler';
}

function updateActiveBanner() {
  const txt = $('#activePrinterText');
  if (!currentConfig || !currentConfig.default_printer) {
    txt.innerHTML = 'Belum dikonfigurasi <span class="status-pill failed" style="margin-left:8px;">N/A</span>';
    return;
  }
  const info = knownPrinters.find(p => p.name === currentConfig.default_printer);
  const type = info ? info.type : currentConfig.printer_type;
  const status = info ? info.status : 'unknown';
  const statusClass = status === 'ready' ? 'success' : (status === 'offline' ? 'failed' : 'pending');
  const pmode = currentConfig.printer_mode || 'thermal';
  const pmodeLabel = pmode === 'dotmatrix' ? 'Dot Matrix' : 'Thermal';
  const pmodeColor = pmode === 'dotmatrix' ? 'color:var(--success)' : 'color:var(--accent)';
  txt.innerHTML = currentConfig.default_printer
    + ' <span class="status-pill" style="margin-left:8px;">' + (type||'-') + '</span>'
    + ' <span class="status-pill" style="margin-left:6px;' + pmodeColor + '">' + pmodeLabel + '</span>'
    + ' <span class="status-pill" style="margin-left:6px;">' + currentConfig.default_paper_width_mm + 'mm</span>'
    + ' <span class="status-pill ' + statusClass + '" style="margin-left:6px;">' + status + '</span>';
}

// isInitialLoad dipakai untuk membedakan teks loading saat startup
// vs saat klik tombol Refresh manual.
let isInitialLoad = true;

async function loadPrinters(refresh) {
  const showLoading = !!refresh;
  const statusEl = $('#printerListStatus');
  const statusText = $('#printerListStatusText');
  const btnRef = $('#btnRefreshPrinters');
  const sel = $('#printerSelect');

  // Simpan nilai yang harus dipilih setelah select di-enable kembali.
  // Harus dideklarasikan DI LUAR try/finally agar bisa diakses keduanya.
  // Latar belakang: browser Chromium me-reset pilihan select ke option
  // pertama saat elemen di-enable dari kondisi disabled. Oleh karena itu
  // sel.value harus di-set SETELAH sel.disabled = false.
  let pendingSelection = '';

  if (showLoading) {
    statusEl.classList.add('visible');
    statusText.textContent = isInitialLoad
      ? 'Mendeteksi printer saat startup (spooler & Bluetooth)…'
      : 'Mendeteksi ulang printer spooler (OS) dan Bluetooth… mohon tunggu.';
    isInitialLoad = false;
    btnRef.disabled = true;
    sel.disabled = true;
    sel.setAttribute('aria-busy', 'true');
  }

  try {
    const path = refresh ? '/api/printers/refresh' : '/api/printers';
    const r = await apiCall(path);
    if (!r.ok) {
      showToast('Gagal memuat printer: ' + (r.data.message || r.status), 'error');
      return;
    }
    knownPrinters = r.data.printers || [];

    // Tangkap nilai yang sedang terpilih SEBELUM list dibangun ulang.
    // Prioritas: nilai saat ini di sel → config → kosong.
    const prev = sel.value || (currentConfig && currentConfig.default_printer) || '';

    sel.innerHTML = '';
    if (knownPrinters.length === 0) {
      const opt = document.createElement('option');
      opt.value = ''; opt.textContent = '— Tidak ada printer terdeteksi —';
      sel.appendChild(opt);
    } else {
      const placeholder = document.createElement('option');
      placeholder.value = ''; placeholder.textContent = '— Pilih printer —';
      sel.appendChild(placeholder);
      knownPrinters.forEach(p => {
        const opt = document.createElement('option');
        opt.value = p.name;
        opt.textContent = p.name + ' [' + p.type + '] · ' + p.status;
        sel.appendChild(opt);
      });
    }

    // Pastikan opsi untuk prev tersedia di list (tambahkan bila perlu).
    if (prev) {
      const found = Array.from(sel.options).some(o => o.value === prev);
      if (!found) {
        const opt = document.createElement('option');
        opt.value = prev;
        opt.textContent = prev + ' (tidak terdeteksi)';
        sel.appendChild(opt);
      }
      // Simpan dulu; sel.value di-assign di finally SETELAH disabled=false.
      pendingSelection = prev;
    }

    if (refresh) {
      showToast('Daftar printer di-refresh', 'success');
      if (r.data.warning) showToast(String(r.data.warning), 'warning');
    }
  } finally {
    if (showLoading) {
      statusEl.classList.remove('visible');
      btnRef.disabled = false;
      sel.disabled = false;            // enable dulu
      sel.removeAttribute('aria-busy');
    }
    // Set value SETELAH enabled supaya Chromium tidak reset ke option pertama.
    if (pendingSelection) {
      sel.value = pendingSelection;
    }
    // Update banner dengan pilihan yang sudah final.
    updateActiveBanner();
  }
}

$('#printerSelect').addEventListener('change', function() { updateActiveBanner(); });

$('#btnSaveSettings').addEventListener('click', async () => {
  const printer = $('#printerSelect').value;
  const widthRadio = $$('input[name="paperWidth"]').find(r => r.checked);
  const modeRadio  = $$('input[name="printerMode"]').find(r => r.checked);
  const width = widthRadio ? parseInt(widthRadio.value, 10) : 80;
  const pmode = modeRadio ? modeRadio.value : 'thermal';
  const ptype = getPrinterTypeForName(printer);

  if (!printer) { showToast('Silakan pilih printer terlebih dahulu', 'warning'); return; }

  const payload = {
    default_printer: printer,
    default_paper_width_mm: width,
    printer_type: ptype,
    printer_mode: pmode,
  };
  const r = await apiCall('/api/config', JSON.stringify(payload));
  if (r.ok && r.data.success !== false) {
    currentConfig = r.data.config || payload;
    showToast('Settings tersimpan ke config.json', 'success');
    updateActiveBanner();
  } else {
    showToast('Gagal menyimpan: ' + (r.data.message || r.status), 'error');
  }
});

$('#btnRefreshPrinters').addEventListener('click', () => loadPrinters(true));

// ============================
//  Logs
// ============================
async function refreshLogs() {
  const r = await apiCall('/api/logs');
  if (!r.ok) return;
  renderLogs(r.data.logs || []);
}
function renderLogs(logs) {
  $('#logCount').textContent = logs.length;
  const tb = $('#logsTable tbody');
  if (!logs.length) {
    tb.innerHTML = '<tr><td colspan="7" class="empty">Belum ada log…</td></tr>';
    return;
  }
  const rows = logs.map(l => {
    const cls = l.status || 'pending';
    const dur = (l.duration_ms || 0) + ' ms';
    return '<tr>'
      + '<td><span class="mono">' + fmtTime(l.timestamp) + '</span></td>'
      + '<td><span class="mono">' + shortId(l.job_id) + '</span></td>'
      + '<td>' + escapeHTML(l.printer || '-') + '</td>'
      + '<td>' + escapeHTML(l.type || '-') + '</td>'
      + '<td class="' + cls + '">' + (l.status || '-').toUpperCase() + '</td>'
      + '<td><span class="mono">' + dur + '</span></td>'
      + '<td>' + escapeHTML(l.message || '') + '</td>'
      + '</tr>';
  });
  tb.innerHTML = rows.join('');
}
function escapeHTML(s) {
  return String(s == null ? '' : s)
    .replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;')
    .replace(/"/g,'&quot;').replace(/'/g,'&#39;');
}

$('#btnRefreshLogs').addEventListener('click', refreshLogs);
$('#btnClearDisplay').addEventListener('click', () => {
  $('#logsTable tbody').innerHTML = '<tr><td colspan="7" class="empty">Tampilan dibersihkan (server tetap menyimpan log).</td></tr>';
  $('#logCount').textContent = '0';
});

function startLogsAutoRefresh() {
  if (logsTimer) clearInterval(logsTimer);
  logsTimer = setInterval(() => {
    if ($('#panel-logs').classList.contains('active')) refreshLogs();
  }, 5000);
}

// ============================
//  Test Text
// ============================
$('#textLogoFile').addEventListener('change', async (e) => {
  const f = e.target.files[0];
  if (!f) { textLogoBase64 = ''; $('#textLogoPreview').style.display = 'none'; return; }
  textLogoBase64 = await fileToBase64(f);
  $('#textLogoImg').src = 'data:' + (f.type || 'image/png') + ';base64,' + textLogoBase64;
  $('#textLogoPreview').style.display = 'block';
});

$('#btnPrintText').addEventListener('click', async () => {
  const txt = $('#textContent').value;
  if (!txt && !(textLogoBase64 && textPrintMode === 'thermal')) {
    showToast('Isi konten teks terlebih dahulu', 'warning'); return;
  }
  const payload = {
    text: txt,
    logo_base64:   textPrintMode === 'thermal' ? (textLogoBase64 || '') : '',
    logo_position: textPrintMode === 'thermal' ? $('#textLogoPos').value : 'center',
    copies:   parseInt($('#textCopies').value, 10) || 1,
    cut_paper: textPrintMode === 'thermal' && $('#textCut').checked,
    encoding: 'utf-8',
    mode: textPrintMode,
  };
  const btn = $('#btnPrintText'); btn.disabled = true;
  try {
    const r = await apiCall('/api/print/text', JSON.stringify(payload));
    const box = $('#textResponse'); box.style.display = 'block';
    box.textContent = JSON.stringify(r.data, null, 2);
    if (r.ok && r.data.success) showToast('Job teks terkirim', 'success');
    else showToast('Gagal: ' + (r.data.message || r.status), 'error');
  } finally { btn.disabled = false; }
});

// ============================
//  Test Image
// ============================
$('#imgFile').addEventListener('change', async (e) => {
  const f = e.target.files[0];
  if (!f) { imageBase64 = ''; $('#imgPreview').style.display = 'none'; return; }
  imageBase64 = await fileToBase64(f);
  $('#imgPreviewImg').src = 'data:' + (f.type || 'image/png') + ';base64,' + imageBase64;
  $('#imgPreview').style.display = 'block';
});

$('#btnPrintImage').addEventListener('click', async () => {
  if (imgPrintMode === 'dotmatrix') {
    showToast('Cetak gambar tidak didukung pada mode Dot Matrix', 'warning'); return;
  }
  if (!imageBase64) { showToast('Pilih file gambar terlebih dahulu', 'warning'); return; }
  const payload = {
    image_base64: imageBase64,
    width_mm: currentConfig ? currentConfig.default_paper_width_mm : 80,
    dithering: $('#imgDither').value,
    copies: parseInt($('#imgCopies').value, 10) || 1,
    cut_paper: $('#imgCut').checked,
    mode: imgPrintMode,
  };
  const btn = $('#btnPrintImage'); btn.disabled = true;
  try {
    const r = await apiCall('/api/print/image', JSON.stringify(payload));
    const box = $('#imgResponse'); box.style.display = 'block';
    box.textContent = JSON.stringify(r.data, null, 2);
    if (r.ok && r.data.success) showToast('Job gambar terkirim', 'success');
    else showToast('Gagal: ' + (r.data.message || r.status), 'error');
  } finally { btn.disabled = false; }
});

// ============================
//  Init
// ============================
(async function init() {
  try {
    // 1. Muat config lebih dulu agar loadPrinters mengetahui printer mana
    //    yang harus di-pre-select setelah deteksi selesai.
    //    applyConfigToForm() di dalam loadConfig akan mengisi paper width,
    //    mode printer, dan menyiapkan currentConfig.default_printer sebagai
    //    target seleksi dropdown.
    await loadConfig();

    // 2. Jalankan refresh penuh (POST /api/printers/refresh) dengan
    //    loading indicator. Karena currentConfig sudah tersedia, variabel
    //    'prev' di dalam loadPrinters akan menangkap nama printer dari
    //    config dan otomatis memilihnya setelah daftar selesai dimuat.
    await loadPrinters(true);

    refreshLogs();
    startLogsAutoRefresh();
  } catch(e) {
    showToast('Gagal inisialisasi: ' + e.message, 'error');
  }
})();
</script>

</body>
</html>
`
