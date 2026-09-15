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
    tbody.innerHTML = '<tr><td colspan="12" class="empty-row">Chưa có dữ liệu thiết bị ngoại vi.</td></tr>';
    document.getElementById('count-bang4').textContent = '0 mục';
    return;
  }
  document.getElementById('count-bang4').textContent = records.length + ' mục';
  tbody.innerHTML = records.map((r, i) => {
    const sessionCount = r.sessions ? r.sessions.length : 0;
    const sessionCell = sessionCount > 0
      ? `<a href="#" onclick="showSessionHistory(${i}); return false;" class="session-link">${sessionCount} session${sessionCount > 1 ? 's' : ''} →</a>`
      : '<span class="cell-muted">—</span>';
    return `
    <tr>
      ${rowNumber(i)}
      <td>${escapeHtml(r.device_type)}</td>
      <td><strong>${escapeHtml(r.vendor_model)}</strong></td>
      <td class="cell-mono">${escapeHtml(r.hardware_id)}</td>
      <td class="cell-mono">${escapeHtml(r.vid_pid)}</td>
      <td class="cell-mono">${escapeHtml(r.drive_letter || '—')}</td>
      <td class="cell-mono">${escapeHtml(r.first_plug || '—')}</td>
      <td class="cell-mono">${escapeHtml(r.last_plug || '—')}</td>
      <td><strong>${r.plug_count || 0}</strong></td>
      <td>${formatBool(r.badusb_warning)}</td>
      <td>${sessionCell}</td>
      <td><pre class="cell-mono" style="white-space:pre-wrap;max-width:250px;font-size:10px;margin:0;font-family:Consolas,monospace;">${escapeHtml(r.recent_files_summary || '—')}</pre></td>
    </tr>
  `}).join('');
}

// showSessionHistory hiển thị popup lịch sử từng lần kết nối của thiết bị
function showSessionHistory(recordIndex) {
  if (!lastResult || !lastResult.bang4 || !lastResult.bang4[recordIndex]) {
    alert('Không có dữ liệu session');
    return;
  }
  const r = lastResult.bang4[recordIndex];
  const sessions = r.sessions || [];
  
  const titleEl = document.getElementById('stat-detail-title');
  const bodyEl = document.getElementById('stat-detail-body');
  
  titleEl.textContent = `Lịch sử kết nối: ${r.vendor_model || r.device_type} (${sessions.length} lần)`;
  
  if (sessions.length === 0) {
    bodyEl.innerHTML = '<div class="stat-detail-empty">Không có dữ liệu lịch sử kết nối</div>';
  } else {
    bodyEl.innerHTML = '<ul class="stat-detail-list">' + sessions.map((s, i) => {
      const start = s.start_time ? formatDateTime(s.start_time) : '—';
      const end = s.end_time ? formatDateTime(s.end_time) : '<span class="cell-ok">Đang cắm</span>';
      const duration = calculateDuration(s.start_time, s.end_time);
      return `
        <li>
          <span class="num">${i + 1}</span>
          <div class="main">
            <strong>Lần ${i + 1}</strong>
            <span class="badge-sm" style="background:#631017;color:#fff;">${escapeHtml(s.drive_letter || '—')}</span>
            <span class="badge-sm" style="background:#888;color:#fff;">${escapeHtml(s.source || '—')}</span>
            <div class="meta">
              Cắm: ${start} | Rút: ${end}<br>
              Thời lượng: ${duration}
            </div>
          </div>
        </li>
      `;
    }).join('') + '</ul>';
  }
  openModal('modal-stat-detail');
}

// formatDateTime định dạng thời gian dễ đọc
function formatDateTime(iso) {
  if (!iso) return '—';
  try {
    const d = new Date(iso);
    if (isNaN(d)) return iso;
    return d.toLocaleString('vi-VN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' });
  } catch (e) {
    return iso;
  }
}

// calculateDuration tính thời lượng giữa start và end
function calculateDuration(start, end) {
  if (!start) return '—';
  if (!end) return 'Đang cắm';
  try {
    const s = new Date(start);
    const e = new Date(end);
    if (isNaN(s) || isNaN(e)) return '—';
    const diffMs = e - s;
    if (diffMs < 0) return '—';
    const hours = Math.floor(diffMs / 3600000);
    const minutes = Math.floor((diffMs % 3600000) / 60000);
    if (hours > 0) {
      return hours + ' giờ ' + minutes + ' phút';
    }
    return minutes + ' phút';
  } catch (e) {
    return '—';
  }
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
  
  const tileValues = {
    'stat-crack': stats.total_crack || 0,
    'stat-ports': stats.open_ports || 0,
    'stat-cve': stats.critical_cve || 0,
    'stat-adapters': stats.adapters || 0,
    'stat-usb': stats.peripherals || 0,
    'stat-badusb': stats.badusb || 0,
    'stat-keylogger': stats.keyloggers || 0,
    'stat-duration': (stats.duration_sec || 0).toFixed(2) + 's',
  };
  
  Object.entries(tileValues).forEach(([id, value]) => {
    const el = document.getElementById(id);
    if (el) el.textContent = value;
    const tile = el ? el.closest('.stat-tile') : null;
    if (tile) {
      const numValue = parseInt(value, 10);
      if (!isNaN(numValue) && numValue === 0 && id !== 'stat-duration') {
        tile.classList.add('zero');
        tile.classList.remove('alert');
      } else if (!isNaN(numValue) && numValue > 0) {
        tile.classList.remove('zero');
        if (['stat-crack', 'stat-ports', 'stat-cve', 'stat-badusb', 'stat-keylogger'].includes(id)) {
          tile.classList.add('alert');
        }
      }
    }
  });
}

// ============ Stat Detail Modal - popup khi click vào ô thống kê ============

function showStatDetail(statType) {
  if (!lastResult) {
    alert('Chưa có dữ liệu quét. Vui lòng bấm nút "RÀ QUÉT TOÀN BỘ" trước.');
    return;
  }
  
  const titleEl = document.getElementById('stat-detail-title');
  const bodyEl = document.getElementById('stat-detail-body');
  let html = '';
  let title = '';
  
  switch (statType) {
    case 'crack':
      // Tách thành 2 nhóm: bất hợp pháp và hợp pháp
      const illegalSW = (lastResult.bang1 || []).filter(r => r.legal === false);
      const legalSW = (lastResult.bang1 || []).filter(r => r.legal !== false);
      title = `Chi tiết ${lastResult.bang1.length} phần mềm bản quyền phát hiện`;
      if (lastResult.bang1.length > 0) {
        title += ` (${illegalSW.length} bất hợp pháp, ${legalSW.length} hợp pháp)`;
      }
      html = renderCrackDetailList(lastResult.bang1);
      break;
    case 'ports':
      title = `Chi tiết ${countOpenPorts()} cổng mạng nguy hiểm đang mở`;
      html = renderPortsDetailList(lastResult.bang2);
      break;
    case 'cve':
      title = `Chi tiết CVE có CVSS ≥ 7.0`;
      html = renderCVEDetailList(lastResult.bang2);
      break;
    case 'adapters':
      title = `Chi tiết ${lastResult.bang3.length} card mạng phát hiện`;
      html = renderAdaptersDetailList(lastResult.bang3);
      break;
    case 'usb':
      title = `Chi tiết ${lastResult.bang4.length} thiết bị ngoại vi`;
      html = renderUSBDetailList(lastResult.bang4);
      break;
    case 'badusb':
      title = `Chi tiết thiết bị BadUSB phát hiện`;
      html = renderBadUSBDetailList(lastResult.bang4);
      break;
    case 'keylogger':
      title = `Chi tiết Keylogger phát hiện`;
      html = renderKeyloggerDetailList(lastResult.bang5);
      break;
    default:
      title = 'Chi tiết';
      html = '<div class="stat-detail-empty">Không có dữ liệu chi tiết</div>';
  }
  
  titleEl.textContent = title;
  bodyEl.innerHTML = html;
  openModal('modal-stat-detail');
}

function countOpenPorts() {
  if (!lastResult || !lastResult.bang2 || lastResult.bang2.length === 0) return 0;
  const ports = lastResult.bang2[0].open_ports || '';
  if (!ports) return 0;
  return ports.split(',').filter(p => p.trim()).length;
}

function renderCrackDetailList(records) {
  if (!records || records.length === 0) {
    return '<div class="stat-detail-empty">Không phát hiện phần mềm bản quyền</div>';
  }
  // Sắp xếp: bất hợp pháp lên đầu, hợp pháp xuống cuối
  const sorted = [...records].sort((a, b) => {
    if (a.legal === false && b.legal !== false) return -1;
    if (a.legal !== false && b.legal === false) return 1;
    return 0;
  });
  return '<ul class="stat-detail-list">' + sorted.map((r, i) => {
    const risk = assessLicenseRisk(r);
    return `
    <li>
      <span class="num">${i + 1}</span>
      <div class="main">
        <strong>${escapeHtml(r.software_name || '—')}</strong>
        ${r.legal === false ? '<span class="badge-sm" style="background:#dc143c;color:#fff;">BẤT HỢP PHÁP</span>' : '<span class="badge-sm" style="background:#2d7a3e;color:#fff;">HỢP PHÁP</span>'}
        <span class="badge-sm" style="background:${risk.color};color:#fff;">${risk.level}</span>
        <div class="meta">
          Phiên bản: ${escapeHtml(r.version || '—')} | Product ID: ${escapeHtml(r.product_id || '—')}<br>
          Kênh: ${escapeHtml(r.licensing_channel || '—')} | OEM Key: ${escapeHtml(r.bios_oem_key || '—')}<br>
          ${r.crack_tool ? '⚠ Công cụ crack: ' + escapeHtml(r.crack_tool) : ''}
          ${r.crack_path ? '<br>   Path: ' + escapeHtml(r.crack_path) : ''}
        </div>
        <div class="risk-box">
          <strong>⚠ Đánh giá nguy cơ:</strong> ${risk.description}<br>
          <strong>Căn cứ pháp lý:</strong> ${risk.legal}
        </div>
      </div>
    </li>
  `}).join('') + '</ul>';
}

// assessLicenseRisk đánh giá nguy cơ pháp lý cho phần mềm bản quyền
function assessLicenseRisk(r) {
  if (r.legal === false) {
    // Có crack tool hoặc KMS lậu
    if (r.crack_tool && r.crack_tool.toLowerCase().includes('kms')) {
      return {
        level: 'NGUY HIỂM CAO',
        color: '#8B0000',
        description: 'Sử dụng công cụ crack KMS để kích hoạt bản quyền lậu vi phạm Luật Sở hữu trí tuệ, có thể bị phạt hành chính 50-100 triệu đồng.',
        legal: 'Luật Sở hữu trí tuệ 2005 (sửa đổi 2009); Nghị định 22/2018/NĐ-CP; Luật An ninh mạng 2018 Điều 28.'
      };
    }
    if (r.crack_tool && r.crack_tool.toLowerCase().includes('adobe')) {
      return {
        level: 'NGUY HIỂM CAO',
        color: '#8B0000',
        description: 'Sử dụng phần mềm Adobe crack vi phạm quyền sở hữu trí tuệ, có thể bị phạt hành chính hoặc truy cứu trách nhiệm hình sự.',
        legal: 'Luật Sở hữu trí tuệ; Bộ luật Hình sự 2015 (sửa đổi 2017) Điều 225.'
      };
    }
    if (r.crack_tool) {
      return {
        level: 'NGUY HIỂM CAO',
        color: '#8B0000',
        description: 'Phát hiện công cụ crack phần mềm: ' + r.crack_tool + '. Vi phạm bản quyền phần mềm.',
        legal: 'Luật Sở hữu trí tuệ 2005; Nghị định 22/2018/NĐ-CP; BLHS Điều 225.'
      };
    }
    // Có CVE nghiêm trọng (Chrome/Edge/Java/OpenSSL)
    if (r.crack_tool && r.crack_tool.toLowerCase().includes('cve')) {
      return {
        level: 'NGUY HIỂM TRUNG BÌNH',
        color: '#B8860B',
        description: 'Phần mềm có lỗ hổng CVE nghiêm trọng chưa vá. Cần cập nhật ngay.',
        legal: 'Luật An ninh mạng 2018 Điều 29 (bảo đảm an toàn thông tin); Luật Bảo vệ bí mật nhà nước 2018.'
      };
    }
    return {
      level: 'CẦN KIỂM TRA',
      color: '#888',
      description: 'Phần mềm có dấu hiệu bất hợp pháp, cần kiểm tra giấy phép.',
      legal: 'Luật Sở hữu trí tuệ; quy định cấp phép phần mềm của cơ quan.'
    };
  }
  // Hợp pháp
  return {
    level: 'AN TOÀN',
    color: '#2d7a3e',
    description: 'Phần mềm có bản quyền hợp pháp, không phát hiện dấu hiệu vi phạm.',
    legal: 'Tuân thủ Luật Sở hữu trí tuệ 2005.'
  };
}

function renderPortsDetailList(records) {
  if (!records || records.length === 0) {
    return '<div class="stat-detail-empty">Không có dữ liệu mạng</div>';
  }
  const r = records[0];
  const ports = (r.open_ports || '').split(',').filter(p => p.trim());
  if (ports.length === 0) {
    return '<div class="stat-detail-empty">Không phát hiện cổng nguy hiểm nào đang mở ✓</div>';
  }
  const portInfo = {
    '21': { name: 'FTP', risk: 'CAO', desc: 'Giao thức FTP truyền dữ liệu không mã hoá, có thể bị sniff bắt cắp thông tin', legal: 'Luật An ninh mạng 2018 Điều 28 (bảo đảm an toàn thông tin)' },
    '22': { name: 'SSH', risk: 'TRUNG BÌNH', desc: 'SSH mở ra ngoài có thể bị brute-force attack', legal: 'Luật An ninh mạng 2018 Điều 29' },
    '23': { name: 'Telnet', risk: 'RẤT CAO', desc: 'Telnet truyền dữ liệu clear text, dễ bị sniff', legal: 'Luật An ninh mạng 2018 Điều 28; Nghị định 90/2023/NĐ-CP' },
    '25': { name: 'SMTP', risk: 'TRUNG BÌNH', desc: 'SMTP mở có thể bị spam relay', legal: 'Nghị định 14/2018/NĐ-CP về bưu chính' },
    '80': { name: 'HTTP', risk: 'TRUNG BÌNH', desc: 'HTTP truyền clear text', legal: 'Luật An ninh mạng 2018 Điều 28' },
    '135': { name: 'MSRPC', risk: 'CAO', desc: 'MSRPC mở có thể bị khai thác DCOM', legal: 'Luật An ninh mạng 2018 Điều 29' },
    '139': { name: 'NetBIOS-SSN', risk: 'CAO', desc: 'NetBIOS chia sẻ file, dễ bị rò rỉ thông tin', legal: 'Luật An ninh mạng 2018 Điều 28' },
    '445': { name: 'SMB (EternalBlue)', risk: 'NGUY HIỂM CỰC CAO', desc: 'Cổng 445 mở + SMBv1 có thể bị EternalBlue RCE → WannaCry ransomware → mất dữ liệu vĩnh viễn', legal: 'BLHS 2015 Điều 288 (Tội vi phạm quy định về bảo mật thông tin); Luật An ninh mạng 2018 Điều 28, 29' },
    '1433': { name: 'MSSQL', risk: 'RẤT CAO', desc: 'Cơ sở dữ liệu MSSQL mở ra ngoài có thể bị trộm dữ liệu', legal: 'Luật An ninh mạng 2018 Điều 28; Luật Bảo vệ bí mật nhà nước 2018 Điều 8' },
    '3306': { name: 'MySQL', risk: 'RẤT CAO', desc: 'MySQL mở ra ngoài, có thể bị dump database', legal: 'Luật An ninh mạng 2018 Điều 28' },
    '3389': { name: 'RDP (BlueKeep)', risk: 'NGUY HIỂM CỰC CAO', desc: 'RDP mở có thể bị BlueKeep RCE → chiếm quyền điều khiển máy từ xa → mất toàn bộ quyền kiểm soát', legal: 'BLHS 2015 Điều 288, 290 (Tội vi phạm quy định về bảo mật thông tin, Tội phá rối hoạt động máy tính); Luật An ninh mạng 2018 Điều 28' },
    '5900': { name: 'VNC', risk: 'CAO', desc: 'VNC mở có thể bị chiếm quyền điều khiển desktop', legal: 'BLHS 2015 Điều 290; Luật An ninh mạng 2018' },
    '8080': { name: 'HTTP-Alt', risk: 'TRUNG BÌNH', desc: 'Cổng 8080 mở, có thể có webapp cần kiểm tra', legal: 'Luật An ninh mạng 2018 Điều 29' },
  };
  return '<ul class="stat-detail-list">' + ports.map((p, i) => {
    const port = p.trim();
    const info = portInfo[port] || { name: 'Unknown', risk: 'KHÔNG XÁC ĐỊNH', desc: 'Cổng ít phổ biến, cần đánh giá thêm', legal: 'Luật An ninh mạng 2018' };
    const riskColor = info.risk.includes('CỰC') ? '#8B0000' : (info.risk.includes('RẤT') ? '#dc143c' : (info.risk.includes('CAO') ? '#b8860b' : '#888'));
    return `
      <li>
        <span class="num">${i + 1}</span>
        <div class="main">
          <strong>Cổng ${escapeHtml(port)}</strong> — ${escapeHtml(info.name)}
          <span class="badge-sm" style="background:${riskColor};color:#fff;">${info.risk}</span>
          <div class="meta">Trạng thái: ĐANG MỞ trên máy này</div>
          <div class="risk-box">
            <strong>⚠ Đánh giá nguy cơ:</strong> ${info.desc}<br>
            <strong>Căn cứ pháp lý:</strong> ${info.legal}
          </div>
        </div>
      </li>
    `;
  }).join('') + '</ul>';
}

function renderCVEDetailList(records) {
  if (!records || records.length === 0) {
    return '<div class="stat-detail-empty">Không có dữ liệu CVE</div>';
  }
  const r = records[0];
  if (!r.cve_id) {
    return '<div class="stat-detail-empty">Không phát hiện CVE có CVSS ≥ 7.0 ✓</div>';
  }
  return `<ul class="stat-detail-list">
    <li>
      <span class="num">1</span>
      <div class="main">
        <strong>${escapeHtml(r.cve_id)}</strong>
        <span class="badge-sm" style="background:#dc143c;color:#fff;">CVSS ${r.cvss_score ? r.cvss_score.toFixed(1) : '?'}</span>
        <div class="meta">${escapeHtml(r.notes || '—')}</div>
        <div class="meta" style="margin-top:6px;"><strong>Kết quả pentest:</strong> ${escapeHtml(r.exploit_result || '—')}</div>
      </div>
    </li>
  </ul>`;
}

function renderAdaptersDetailList(records) {
  if (!records || records.length === 0) {
    return '<div class="stat-detail-empty">Không phát hiện card mạng</div>';
  }
  return '<ul class="stat-detail-list">' + records.map((r, i) => `
    <li>
      <span class="num">${i + 1}</span>
      <div class="main">
        <strong>${escapeHtml(r.device_name || '—')}</strong>
        <span class="badge-sm" style="background:#631017;color:#fff;">${escapeHtml(r.adapter_type || '—')}</span>
        <div class="meta">
          Loại: ${escapeHtml(r.adapter_type || '—')} | Vị trí: ${escapeHtml(r.connection_pos || '—')}<br>
          MAC: ${escapeHtml(r.mac || '—')} | Driver: ${escapeHtml(r.driver_status || '—')}<br>
          ${r.ssid_list ? 'SSID đã kết nối: ' + escapeHtml(r.ssid_list) : ''}
          ${r.last_connect ? '<br>Lần gần nhất: ' + escapeHtml(r.last_connect) : ''}
        </div>
      </div>
    </li>
  `).join('') + '</ul>';
}

function renderUSBDetailList(records) {
  if (!records || records.length === 0) {
    return '<div class="stat-detail-empty">Không phát hiện thiết bị ngoại vi</div>';
  }
  return '<ul class="stat-detail-list">' + records.map((r, i) => `
    <li>
      <span class="num">${i + 1}</span>
      <div class="main">
        <strong>${escapeHtml(r.vendor_model || r.device_type || '—')}</strong>
        ${r.badusb_warning ? '<span class="badge-sm" style="background:#dc143c;color:#fff;">BADUSB!</span>' : ''}
        <div class="meta">
          Loại: ${escapeHtml(r.device_type || '—')} | VID/PID: ${escapeHtml(r.vid_pid || '—')}<br>
          Hardware ID: ${escapeHtml(r.hardware_id || '—')}<br>
          Ổ đĩa: ${escapeHtml(r.drive_letter || '—')} | Số lần cắm: ${r.plug_count || 0}<br>
          ${r.first_plug ? 'Lần đầu: ' + escapeHtml(r.first_plug) : ''}
          ${r.last_plug ? ' | Lần cuối: ' + escapeHtml(r.last_plug) : ''}
          ${r.recent_files_summary ? '<br><strong>Recent Files:</strong> ' + escapeHtml(r.recent_files_summary) : ''}
        </div>
      </div>
    </li>
  `).join('') + '</ul>';
}

function renderBadUSBDetailList(records) {
  const badUSBs = (records || []).filter(r => r.badusb_warning);
  if (badUSBs.length === 0) {
    return '<div class="stat-detail-empty">Không phát hiện thiết bị BadUSB ✓</div>';
  }
  return '<ul class="stat-detail-list">' + badUSBs.map((r, i) => `
    <li>
      <span class="num">${i + 1}</span>
      <div class="main">
        <strong>${escapeHtml(r.vendor_model || r.device_type || '—')}</strong>
        <span class="badge-sm" style="background:#dc143c;color:#fff;">BADUSB</span>
        <span class="badge-sm" style="background:#8B0000;color:#fff;">NGUY HIỂM CỰC CAO</span>
        <div class="meta">
          VID/PID: ${escapeHtml(r.vid_pid || '—')}<br>
          Hardware ID: ${escapeHtml(r.hardware_id || '—')}<br>
        </div>
        <div class="risk-box">
          <strong>⚠ Đánh giá nguy cơ:</strong> Thiết bị BadUSB có khả năng giả lập bàn phím tấn công tự động (Keystroke Injection). Khi cắm vào máy, nó có thể tự gõ lệnh PowerShell/CMD, cài backdoor, mở cổng reverse shell, download mã độc. Đây là phương pháp tấn công vật lý cực kỳ nguy hiểm.<br>
          <strong>Căn cứ pháp lý:</strong> Bộ luật Hình sự 2015 Điều 290 (Tội phá rối hoạt động máy tính); Luật An ninh mạng 2018 Điều 18, 28; Quy định về quản lý thiết bị ngoại vi trong cơ quan nhà nước.<br>
          <strong>Hình thức xử lý:</strong> TRÁCH NHIỆM HÌNH SỰ: Phạt tù 01-07 năm (Điều 290 BLHS); Thu giữ thiết bị; Cảnh giác nội bộ cơ quan; Báo cáo Ban Cơ yếu.
        </div>
      </div>
    </li>
  `).join('') + '</ul>';
}

function renderKeyloggerDetailList(records) {
  const keyloggers = (records || []).filter(r => r.type === 'Keylogger');
  if (keyloggers.length === 0) {
    return '<div class="stat-detail-empty">Không phát hiện Keylogger ✓</div>';
  }
  return '<ul class="stat-detail-list">' + keyloggers.map((r, i) => {
    const risk = assessMalwareRisk(r);
    return `
    <li>
      <span class="num">${i + 1}</span>
      <div class="main">
        <strong>${escapeHtml(r.process_name || '—')}</strong>
        <span class="badge-sm" style="background:#dc143c;color:#fff;">PID ${r.pid || '?'}</span>
        ${r.danger_level ? '<span class="badge-sm" style="background:' + risk.color + ';color:#fff;">' + escapeHtml(r.danger_level) + '</span>' : ''}
        <div class="meta">
          Đường dẫn: ${escapeHtml(r.file_path || '—')}<br>
          RAM: ${r.running_in_ram ? 'ĐANG CHẠY' : 'Tắt'} | C2: ${escapeHtml(r.c2_server || '—')}<br>
          ${r.log_wipe_evidence ? '⚠ ' + escapeHtml(r.log_wipe_evidence) : ''}
        </div>
        <div class="risk-box">
          <strong>⚠ Đánh giá nguy cơ:</strong> ${risk.description}<br>
          <strong>Căn cứ pháp lý:</strong> ${risk.legal}<br>
          <strong>Hình thức xử lý:</strong> ${risk.liability}
        </div>
      </div>
    </li>
  `}).join('') + '</ul>';
}

// assessMalwareRisk đánh giá nguy cơ pháp lý cho mã độc/keylogger
function assessMalwareRisk(r) {
  // Keylogger có C2 server đang chạy
  if (r.c2_server && r.running_in_ram) {
    return {
      color: '#8B0000',
      description: 'Keylogger ĐANG CHẠY trong RAM và có kết nối tới C2 server ' + r.c2_server + '. Dữ liệu gõ phím đang bị gửi đi real-time. Đây là tấn công gián điệp mạng nghiêm trọng, có thể làm rò rỉ bí mật nhà nước, mật khẩu, văn bản mật.',
      legal: 'Bộ luật Hình sự 2015 (sửa đổi 2017) Điều 288 (Tội vi phạm quy định về bảo mật thông tin), Điều 289 (Tội đánh cắp thông tin), Điều 290 (Tội phá rối hoạt động máy tính); Luật An ninh mạng 2018 Điều 18, 28; Luật Bảo vệ bí mật nhà nước 2018.',
      liability: 'TRÁCH NHIỆM HÌNH SỰ: Phạt tù từ 01 năm đến 07 năm (Điều 288 BLHS); Phạt tù từ 01 năm đến 12 năm nếu gây hậu quả nghiêm trọng (Điều 289 BLHS).'
    };
  }
  // Keylogger chạy trong RAM nhưng chưa có C2
  if (r.running_in_ram) {
    return {
      color: '#dc143c',
      description: 'Keylogger ĐANG CHẠY trong RAM. Có thể đang ghi lại toàn bộ thao tác gõ phím của người dùng, có nguy cơ rò rỉ thông tin.',
      legal: 'BLHS 2015 Điều 288, 289; Luật An ninh mạng 2018 Điều 28.',
      liability: 'TRÁCH NHIỆM HÌNH SỰ: Phạt tù 01-07 năm (Điều 288); có thể bị kỷ luật công chức hoặc sa thải nếu vi phạm trong cơ quan nhà nước.'
    };
  }
  // Keylogger có file nhưng không chạy
  return {
    color: '#b8860b',
    description: 'Phát hiện file keylogger nhưng tiến trình KHÔNG chạy. Cần xoá file ngay và kiểm tra dấu vết hoạt động trước đó.',
    legal: 'BLHS 2015 Điều 288 (chuẩn bị tội phạm); Luật An ninh mạng 2018 Điều 28.',
    liability: 'TRÁCH NHIỆM HÀNH CHÍNH: Phạt tiền 20-50 triệu đồng (Nghị định 15/2020/NĐ-CP); Cảnh báo hoặc kỷ luật nếu trong cơ quan nhà nước.'
  };
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
  
  // Stat detail modal close buttons
  document.getElementById('stat-detail-close').addEventListener('click', () => closeModal('modal-stat-detail'));
  document.getElementById('stat-detail-done').addEventListener('click', () => closeModal('modal-stat-detail'));
  
  // Click vào stat tile để xem chi tiết
  document.querySelectorAll('.stat-tile[data-stat]').forEach(tile => {
    tile.addEventListener('click', () => {
      const statType = tile.getAttribute('data-stat');
      showStatDetail(statType);
    });
  });

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
