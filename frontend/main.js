/**
 * BCY-VKS Frontend Controller - Government Theme
 * - Lắng nghe event 'scan_progress' từ backend để hiển thị tiến độ real-time
 * - Render 5 bảng dữ liệu với style chính thức
 * - Xử lý 2 nút lệnh (Khắc phục & Anti-Forensics)
 */

// ============ State ============
let lastResult = null;
let scanInProgress = false;
let scanEstimates = [];

// ============ Estimate Table ============

function renderEstimateTable(estimates) {
  scanEstimates = estimates || [];
  const table = document.getElementById('estimate-table');
  if (!scanEstimates.length) {
    table.innerHTML = '<div style="padding:20px;text-align:center;color:#888;">Không có dữ liệu dự báo</div>';
    return;
  }
  table.innerHTML = scanEstimates.map(e => {
    return `
      <div class="estimate-cell ${e.status}" data-id="${e.id}">
        <div class="estimate-cell-group">${escapeHtml(e.group)}</div>
        <div class="estimate-cell-time">${e.min_sec}-${e.max_sec}s</div>
        <div class="estimate-cell-status">${statusText(e.status)}</div>
      </div>
    `;
  }).join('');

  // Tính tổng dự báo
  const totalMin = scanEstimates.reduce((sum, e) => sum + e.min_sec, 0);
  const totalMax = scanEstimates.reduce((sum, e) => sum + e.max_sec, 0);
  document.getElementById('estimate-total').textContent = `Tổng: ${totalMin}-${totalMax} giây`;
}

function statusText(status) {
  switch (status) {
    case 'pending': return 'Chờ';
    case 'running': return 'Đang chạy...';
    case 'done': return '✓ Hoàn tất';
    case 'timeout': return '⚠ Timeout';
    default: return status;
  }
}

function updateEstimateStatus(groupID, status) {
  // Map groupID từ backend (e.g. "license_done") -> estimate ID (e.g. "bang1")
  const mapping = {
    'license': 'bang1', 'license_done': 'bang1', 'license_timeout': 'bang1',
    'network': 'bang2', 'network_done': 'bang2', 'network_timeout': 'bang2',
    'hardware': 'bang3', 'hardware_done': 'bang3', 'hardware_timeout': 'bang3',
    'peripheral': 'bang4', 'peripheral_done': 'bang4', 'peripheral_timeout': 'bang4',
    'malware': 'bang5', 'malware_done': 'bang5', 'malware_timeout': 'bang5',
  };
  const bangID = mapping[groupID];
  if (!bangID) return;
  let newStatus = 'pending';
  if (groupID.endsWith('_done')) newStatus = 'done';
  else if (groupID.endsWith('_timeout')) newStatus = 'timeout';
  else if (groupID) newStatus = 'running';

  scanEstimates = scanEstimates.map(e => {
    if (e.id === bangID) {
      return { ...e, status: newStatus };
    }
    return e;
  });
  renderEstimateTable(scanEstimates);
}

// ============ Helper functions ============
function escapeHtml(s) {
  if (s == null) return '';
  return String(s)
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

function formatBool(v) {
  if (v === true) return '<span class="badge badge-red">CÓ</span>';
  if (v === false) return '<span class="badge badge-green">KHÔNG</span>';
  return '<span class="cell-muted">—</span>';
}

function formatLegal(legal) {
  return legal
    ? '<span class="badge badge-green">HỢP PHÁP</span>'
    : '<span class="badge badge-red">BẤT HỢP PHÁP</span>';
}

function formatBooleanRunning(v) {
  if (v) return '<span class="badge badge-red">ĐANG CHẠY</span>';
  return '<span class="badge badge-gray">TẮT</span>';
}

function formatCVSS(score, cveId) {
  if (!cveId) return '<span class="cell-muted">—</span>';
  let cls = 'badge-green';
  if (score >= 9.0) cls = 'badge-red';
  else if (score >= 7.0) cls = 'badge-orange';
  else if (score >= 4.0) cls = 'badge-yellow';
  return '<span class="badge ' + cls + '">' + escapeHtml(cveId) + ' (CVSS ' + score.toFixed(1) + ')</span>';
}

function formatExploitResult(s) {
  if (!s) return '<span class="cell-muted">—</span>';
  let cls = 'badge-gray';
  if (s.startsWith('SUCCESS')) cls = 'badge-red';
  else if (s.startsWith('BLOCKED')) cls = 'badge-green';
  else if (s.startsWith('PARTIAL')) cls = 'badge-yellow';
  return '<span class="badge ' + cls + '">' + escapeHtml(s) + '</span>';
}

function rowNumber(i) { return '<td class="col-num">' + (i + 1) + '</td>'; }

// ============ Table renderers ============

function renderBang1(records) {
  const tbody = document.getElementById('bang1');
  if (!records || records.length === 0) {
    tbody.innerHTML = '<tr><td colspan="9" class="empty-row">Chưa có dữ liệu bản quyền. Nhấn nút "RÀ QUÉT TOÀN BỘ" để bắt đầu.</td></tr>';
    document.getElementById('count-bang1').textContent = '0 mục';
    return;
  }
  document.getElementById('count-bang1').textContent = records.length + ' mục';
  tbody.innerHTML = records.map((r, i) => `
    <tr>
      ${rowNumber(i)}
      <td><strong>${escapeHtml(r.software_name)}</strong></td>
      <td class="cell-mono">${escapeHtml(r.version)}</td>
      <td class="cell-mono">${escapeHtml(r.product_id)}</td>
      <td>${escapeHtml(r.licensing_channel)}</td>
      <td class="cell-mono">${escapeHtml(r.bios_oem_key)}</td>
      <td class="${r.crack_tool ? 'cell-danger' : 'cell-muted'}">${escapeHtml(r.crack_tool || '—')}</td>
      <td class="cell-mono">${escapeHtml(r.crack_path || '—')}</td>
      <td>${formatLegal(r.legal)}</td>
    </tr>
  `).join('');
}

function renderBang2(records) {
  const tbody = document.getElementById('bang2');
  if (!records || records.length === 0) {
    tbody.innerHTML = '<tr><td colspan="11" class="empty-row">Chưa có dữ liệu mạng. Nhấn nút "RÀ QUÉT TOÀN BỘ" để bắt đầu.</td></tr>';
    document.getElementById('count-bang2').textContent = '0 mục';
    return;
  }
  document.getElementById('count-bang2').textContent = records.length + ' mục';
  tbody.innerHTML = records.map((r, i) => `
    <tr>
      ${rowNumber(i)}
      <td><span class="badge ${r.internet_status === 'Connected' ? 'badge-green' : 'badge-gray'}">${escapeHtml(r.internet_status)}</span></td>
      <td class="cell-mono">${escapeHtml(r.current_ip)}<br><span class="cell-muted">${escapeHtml(r.current_mac)}</span></td>
      <td>${escapeHtml(r.isp)}</td>
      <td class="cell-mono">${escapeHtml(r.open_ports || '—')}</td>
      <td>${formatCVSS(r.cvss_score, r.cve_id)}</td>
      <td><pre class="cell-mono" style="white-space:pre-wrap;max-width:220px;font-size:10px;margin:0;font-family:Consolas,monospace;">${escapeHtml(r.dns_cache || '—')}</pre></td>
      <td><pre class="cell-mono" style="white-space:pre-wrap;max-width:220px;font-size:10px;margin:0;font-family:Consolas,monospace;">${escapeHtml(r.browser_history || '—')}</pre></td>
      <td><pre class="cell-mono" style="white-space:pre-wrap;max-width:200px;font-size:10px;margin:0;font-family:Consolas,monospace;">${escapeHtml(r.lan_config || '—')}</pre></td>
      <td><pre class="cell-mono" style="white-space:pre-wrap;max-width:200px;font-size:10px;margin:0;font-family:Consolas,monospace;">${escapeHtml(r.document_links || '—')}</pre></td>
      <td>${formatExploitResult(r.exploit_result)}</td>
    </tr>
  `).join('');
}

function renderBang3(records) {
  const tbody = document.getElementById('bang3');
  if (!records || records.length === 0) {
    tbody.innerHTML = '<tr><td colspan="9" class="empty-row">Chưa có dữ liệu card mạng.</td></tr>';
    document.getElementById('count-bang3').textContent = '0 mục';
    return;
  }
  document.getElementById('count-bang3').textContent = records.length + ' mục';
  tbody.innerHTML = records.map((r, i) => `
    <tr>
      ${rowNumber(i)}
      <td>${escapeHtml(r.adapter_type)}</td>
      <td>${escapeHtml(r.connection_pos)}</td>
      <td><strong>${escapeHtml(r.device_name)}</strong></td>
      <td class="cell-mono">${escapeHtml(r.serial)}</td>
      <td class="cell-mono">${escapeHtml(r.mac)}</td>
      <td>${escapeHtml(r.driver_status)}</td>
      <td class="cell-mono">${escapeHtml(r.ssid_list || '—')}</td>
      <td class="cell-mono">${escapeHtml(r.last_connect || '—')}</td>
    </tr>
  `).join('');
}

function renderBang4(records) {
  const tbody = document.getElementById('bang4');
  if (!records || records.length === 0) {
    tbody.innerHTML = '<tr><td colspan="11" class="empty-row">Chưa có dữ liệu thiết bị ngoại vi.</td></tr>';
    document.getElementById('count-bang4').textContent = '0 mục';
    return;
  }
  document.getElementById('count-bang4').textContent = records.length + ' mục';
  tbody.innerHTML = records.map((r, i) => `
    <tr>
      ${rowNumber(i)}
      <td>${escapeHtml(r.device_type)}</td>
      <td><strong>${escapeHtml(r.vendor_model)}</strong></td>
      <td class="cell-mono">${escapeHtml(r.hardware_id)}</td>
      <td class="cell-mono">${escapeHtml(r.vid_pid)}</td>
      <td class="cell-mono">${escapeHtml(r.drive_letter || '—')}</td>
      <td class="cell-mono">${escapeHtml(r.first_plug || '—')}</td>
      <td class="cell-mono">${escapeHtml(r.last_plug || '—')}</td>
      <td>${r.plug_count}</td>
      <td>${formatBool(r.badusb_warning)}</td>
      <td><pre class="cell-mono" style="white-space:pre-wrap;max-width:250px;font-size:10px;margin:0;font-family:Consolas,monospace;">${escapeHtml(r.recent_files_summary || '—')}</pre></td>
    </tr>
  `).join('');
}

function renderBang5(records) {
  const tbody = document.getElementById('bang5');
  if (!records || records.length === 0) {
    tbody.innerHTML = '<tr><td colspan="9" class="empty-row">Chưa phát hiện mã độc / keylogger.</td></tr>';
    document.getElementById('count-bang5').textContent = '0 mục';
    return;
  }
  document.getElementById('count-bang5').textContent = records.length + ' mục';
  tbody.innerHTML = records.map((r, i) => `
    <tr>
      ${rowNumber(i)}
      <td class="cell-danger"><strong>${escapeHtml(r.process_name)}</strong></td>
      <td class="cell-mono">${r.pid}</td>
      <td><span class="badge badge-red">${escapeHtml(r.type)}</span></td>
      <td>${formatDangerLevel(r.danger_level)}</td>
      <td>${formatBooleanRunning(r.running_in_ram)}</td>
      <td class="cell-mono">${escapeHtml(r.file_path)}</td>
      <td class="cell-danger">${escapeHtml(r.c2_server || '—')}</td>
      <td class="cell-warn">${escapeHtml(r.log_wipe_evidence || '—')}</td>
    </tr>
  `).join('');
}

// formatDangerLevel định dạng mức độ nguy hiểm với màu sắc tương ứng
function formatDangerLevel(level) {
  if (!level) return '<span class="cell-muted">—</span>';
  switch (level.toUpperCase()) {
    case 'CRITICAL':
      return '<span class="badge badge-red">🔴 CRITICAL</span>';
    case 'HIGH':
      return '<span class="badge badge-orange">🟠 HIGH</span>';
    case 'MEDIUM':
      return '<span class="badge badge-yellow">🟡 MEDIUM</span>';
    case 'LOW':
      return '<span class="badge badge-green">🟢 LOW</span>';
    default:
      return '<span class="badge badge-gray">' + escapeHtml(level) + '</span>';
  }
}

function renderStats(stats) {
  if (!stats) return;
  document.getElementById('stat-crack').textContent = stats.total_crack || 0;
  document.getElementById('stat-ports').textContent = stats.open_ports || 0;
  document.getElementById('stat-cve').textContent = stats.critical_cve || 0;
  document.getElementById('stat-adapters').textContent = stats.adapters || 0;
  document.getElementById('stat-usb').textContent = stats.peripherals || 0;
  document.getElementById('stat-badusb').textContent = stats.badusb || 0;
  document.getElementById('stat-keylogger').textContent = stats.keyloggers || 0;
  document.getElementById('stat-duration').textContent = (stats.duration_sec || 0).toFixed(2) + 's';
}

function renderResult(result) {
  lastResult = result;
  renderBang1(result.bang1);
  renderBang2(result.bang2);
  renderBang3(result.bang3);
  renderBang4(result.bang4);
  renderBang5(result.bang5);
  renderStats(result.stats);

  // Enable action buttons sau khi có kết quả
  ['btn-remediation-popup', 'btn-remediation-html', 'btn-remediation-docx', 'btn-anti-forensics']
    .forEach(id => document.getElementById(id).disabled = false);
}

// ============ Progress UI ============

function showProgressBar(show) {
  const container = document.getElementById('progress-bar-container');
  if (show) {
    container.classList.remove('hidden');
  } else {
    container.classList.add('hidden');
  }
}

function updateProgress(percent, message) {
  document.getElementById('progress-bar-fill').style.width = percent + '%';
  document.getElementById('progress-percent').textContent = percent + '%';
  document.getElementById('progress-message').textContent = message || '';
}

function setScanStatus(state, text) {
  const status = document.getElementById('scan-status');
  status.className = 'scan-status status-' + state;
  status.querySelector('.status-text').textContent = text;
}

// ============ Backend calls ============

async function callScanAll() {
  // Lock ngay lập tức để chống click liên tục (đặt TRƯỚC mọi await)
  if (scanInProgress) {
    console.log('Scan đã đang chạy, bỏ qua click');
    return;
  }
  scanInProgress = true;

  const btn = document.getElementById('btn-scan');
  // Disable button ngay lập tức (đồng bộ, không await)
  if (btn) {
    btn.disabled = true;
    btn.style.opacity = '0.5';
    btn.style.pointerEvents = 'none';
    btn.textContent = 'ĐANG QUÉT...';
  }
  setScanStatus('running', 'Đang khởi tạo...');
  showProgressBar(true);
  updateProgress(0, 'Đang khởi tạo trình quét...');

  // Đặt trạng thái "running" cho mọi nhóm trong bảng dự báo
  scanEstimates = scanEstimates.map(e => ({ ...e, status: 'running' }));
  renderEstimateTable(scanEstimates);

  let offFn = null;
  try {
    // Kiểm tra Wails bindings đã sẵn sàng chưa
    if (!window.go || !window.go.main || !window.go.main.App) {
      throw new Error('Wails bindings chưa sẵn sàng. Vui lòng khởi động lại ứng dụng.');
    }

    // Lấy bảng dự báo thời gian quét (lần đầu)
    try {
      const estimates = await window.go.main.App.GetScanEstimates();
      scanEstimates = estimates || scanEstimates;
      scanEstimates = scanEstimates.map(e => ({ ...e, status: 'running' }));
      renderEstimateTable(scanEstimates);
    } catch (e) {
      console.warn('GetScanEstimates failed:', e);
    }

    // Lắng nghe event 'scan_progress' từ backend Go
    // Kiểm tra wails object tồn tại trước khi gọi
    if (typeof wails !== 'undefined' && wails.Events && wails.Events.On) {
      offFn = wails.Events.On('scan_progress', (data) => {
        if (data && typeof data === 'object') {
          const percent = data.percent || 0;
          const message = data.message || '';
          const step = data.step || '';
          updateProgress(percent, message);
          setScanStatus('running', message);
          updateEstimateStatus(step, '');
        }
      });
    } else {
      console.warn('wails.Events không tồn tại - progress events sẽ không nhận');
    }

    // Gọi hàm ScanAll (đây là điểm blocking, có thể mất 12-52 giây)
    const result = await window.go.main.App.ScanAll();

    // Render kết quả
    renderResult(result);
    setScanStatus('done', 'Quét hoàn tất · ' + new Date().toLocaleTimeString('vi-VN'));
    updateProgress(100, '✓ Hoàn tất');
    setTimeout(() => showProgressBar(false), 2000);

    // Đánh dấu tất cả nhóm đã hoàn tất
    scanEstimates = scanEstimates.map(e => ({ ...e, status: 'done' }));
    renderEstimateTable(scanEstimates);

  } catch (err) {
    console.error('ScanAll error:', err);
    setScanStatus('error', 'Lỗi: ' + (err.message || err));
    showProgressBar(false);
    alert('Lỗi khi quét: ' + (err.message || err));
  } finally {
    // Luôn luôn re-enable button trong finally để không bị stuck
    if (btn) {
      btn.disabled = false;
      btn.style.opacity = '';
      btn.style.pointerEvents = '';
      btn.textContent = 'RÀ QUÉT TOÀN BỘ';
    }
    scanInProgress = false;
    if (offFn) {
      try { offFn(); } catch(e) {}
    }
  }
}

async function callExportRemediation(format) {
  try {
    const result = await window.go.main.App.ExportRemediationReport(format);
    if (format === 'popup') {
      const body = document.getElementById('modal-body');
      const iframe = document.createElement('iframe');
      iframe.srcdoc = result;
      iframe.style.width = '100%';
      iframe.style.height = '500px';
      iframe.style.border = 'none';
      iframe.style.background = '#FFF';
      body.innerHTML = '';
      body.appendChild(iframe);
      document.getElementById('modal-title').textContent = 'Đề xuất Xử lý Khắc phục (Pop-up nhanh)';
      openModal('modal-popup');
    } else {
      alert('Đã xuất file: ' + result + '\nFile đã được mở trong ứng dụng mặc định.');
    }
  } catch (err) {
    console.error('ExportRemediation error:', err);
    alert('Lỗi khi xuất báo cáo: ' + (err.message || err));
  }
}

async function callAntiForensics() {
  const input = document.getElementById('anti-confirm-input');
  const code = input.value.trim();
  if (code !== 'WIPE-CONFIRM-2026') {
    alert('Mã xác nhận không đúng. Phải nhập: WIPE-CONFIRM-2026');
    return;
  }

  closeModal('modal-anti');
  const body = document.getElementById('progress-body');
  body.innerHTML = '<div style="text-align:center;padding:40px;">' +
    '<div style="font-size:14px;color:#8B0000;margin-bottom:16px;font-family:Cambria,serif;">' +
    '⚠ Đang thực hiện wipe, vui lòng chờ... ⚠</div>' +
    '<div class="scan-status status-running" style="display:inline-flex;padding:8px 16px;">' +
    '<span class="status-dot"></span><span class="status-text">Đang wipe</span></div></div>';
  openModal('modal-progress');

  try {
    const steps = await window.go.main.App.RunAntiForensics(code);
    let html = '<div style="font-size:13px;color:#1A1F2C;">' +
      '<p style="color:#8B0000;margin-bottom:12px;font-family:Cambria,serif;font-size:14px;">' +
      'Đã thực hiện <strong>' + steps.length + '</strong> bước wipe:</p>';
    steps.forEach(s => {
      const resultCls = (s.Result || '').toLowerCase();
      html += `
        <div class="wipe-step ${resultCls}">
          <span class="step-action">[${escapeHtml(s.Category)}]</span>
          ${escapeHtml(s.Action)}
          <span class="step-target">${escapeHtml(s.Target)}</span>
          <span class="step-result">${escapeHtml(s.Result)}</span>
          <div style="font-size:11px;color:#6B6B6B;margin-top:4px;">${escapeHtml(s.Detail)}</div>
        </div>
      `;
    });
    html += '</div>';
    html += '<div style="margin-top:16px;padding:12px;background:#FAEFEF;border-left:4px solid #8B0000;border-radius:2px;font-size:12px;">' +
      '<strong style="color:#8B0000;">Lưu ý:</strong> Việc wipe Event Log Security sẽ để lại Event ID 1102 ' +
      '(Audit log cleared) - đây chính là dấu vết cho thấy đã có wipe. ' +
      'Để triệt tiêu hoàn toàn cần khôi phục lại volume shadow copy hoặc restore system image.</div>';
    body.innerHTML = html;
  } catch (err) {
    console.error('RunAntiForensics error:', err);
    body.innerHTML = '<div class="cell-danger">Lỗi: ' + escapeHtml(err.message || err) + '</div>';
  }
  input.value = '';
}

// ============ Modal helpers ============

function openModal(id) {
  document.getElementById(id).classList.remove('hidden');
}

function closeModal(id) {
  document.getElementById(id).classList.add('hidden');
}

// ============ Init ============

document.addEventListener('DOMContentLoaded', async () => {
  console.log('BCY-VKS: DOM đã sẵn sàng, bắt đầu init...');

  // Tab switching
  document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
      document.querySelectorAll('.tab-pane').forEach(p => p.classList.remove('active'));
      tab.classList.add('active');
      document.getElementById(tab.dataset.tab).classList.add('active');
    });
  });

  // Scan button
  document.getElementById('btn-scan').addEventListener('click', callScanAll);

  // Remediation buttons
  document.getElementById('btn-remediation-popup').addEventListener('click', () => callExportRemediation('popup'));
  document.getElementById('btn-remediation-html').addEventListener('click', () => callExportRemediation('html'));
  document.getElementById('btn-remediation-docx').addEventListener('click', () => callExportRemediation('docx'));

  // Anti-forensics button
  document.getElementById('btn-anti-forensics').addEventListener('click', () => {
    openModal('modal-anti');
    document.getElementById('anti-confirm-input').focus();
  });

  // Modal close buttons
  document.getElementById('modal-close').addEventListener('click', () => closeModal('modal-popup'));
  document.getElementById('modal-close-2').addEventListener('click', () => closeModal('modal-popup'));
  document.getElementById('anti-close').addEventListener('click', () => closeModal('modal-anti'));
  document.getElementById('anti-cancel').addEventListener('click', () => closeModal('modal-anti'));
  document.getElementById('anti-confirm').addEventListener('click', callAntiForensics);
  document.getElementById('progress-close').addEventListener('click', () => closeModal('modal-progress'));
  document.getElementById('progress-done').addEventListener('click', () => closeModal('modal-progress'));

  // Enter key để xác nhận anti-forensics
  document.getElementById('anti-confirm-input').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') callAntiForensics();
  });

  // Close modal khi click outside
  document.querySelectorAll('.modal').forEach(m => {
    m.addEventListener('click', (e) => {
      if (e.target === m) m.classList.add('hidden');
    });
  });

  // ESC để đóng modal
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
      document.querySelectorAll('.modal').forEach(m => m.classList.add('hidden'));
    }
  });

  // Render empty state ban đầu
  renderBang1([]);
  renderBang2([]);
  renderBang3([]);
  renderBang4([]);
  renderBang5([]);

  // Render bảng dự báo mặc định (trước khi gọi backend)
  // để UI hiển thị ngay cả khi backend chưa sẵn sàng
  renderEstimateTable([
    { id: 'bang1', group: 'Bản quyền & Crack Tools', min_sec: 2, max_sec: 15, status: 'pending' },
    { id: 'bang2', group: 'Mạng & Pentest', min_sec: 3, max_sec: 30, status: 'pending' },
    { id: 'bang3', group: 'Card mạng & Wi-Fi', min_sec: 1, max_sec: 5, status: 'pending' },
    { id: 'bang4', group: 'USB & Ngoại vi', min_sec: 1, max_sec: 8, status: 'pending' },
    { id: 'bang5', group: 'Mã độc & Memory', min_sec: 3, max_sec: 20, status: 'pending' },
  ]);

  // Đợi 100ms rồi thử gọi backend (để Wails bindings kịp ready)
  await new Promise(r => setTimeout(r, 100));

  // Load bảng dự báo thời gian quét từ backend (ghi đè default)
  // (để UI hiển thị dự báo trước khi user click nút quét)
  try {
    if (window.go && window.go.main && window.go.main.App) {
      console.log('BCY-VKS: Đang gọi GetScanEstimates...');
      const estimates = await window.go.main.App.GetScanEstimates();
      renderEstimateTable(estimates);
      console.log('BCY-VKS: Đã load bảng dự báo từ backend:', estimates);
    } else {
      console.warn('BCY-VKS: window.go.main.App chưa sẵn sàng - dùng dữ liệu mặc định');
    }
  } catch (e) {
    console.warn('BCY-VKS: GetScanEstimates failed on init:', e);
  }

  // Ẩn splash screen sau khi init xong
  const splash = document.getElementById('splash-screen');
  if (splash) {
    splash.classList.add('hidden');
    setTimeout(() => splash.style.display = 'none', 500);
  }

  console.log('BCY-VKS frontend đã sẵn sàng. Click "RÀ QUÉT TOÀN BỘ" để bắt đầu.');
});
