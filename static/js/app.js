const API = '/api/v1';
let currentWorkoutId = null;
let exercises = [];
let prChart = null;
let volChart = null;

function $(s) { return document.querySelector(s); }
function toast(msg) {
  const t = $('#toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 2000);
}
async function req(path, opt) {
  const res = await fetch(API + path, { ...opt, credentials: 'include' });
  if (res.status === 401) {
    window.location.href = '/login';
    throw new Error('未登录');
  }
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(txt);
  }
  if (res.status === 204) return null;
  return res.json();
}

// 检查登录状态
async function checkAuth() {
  try {
    await req('/auth/me', { method: 'GET' });
  } catch {
    window.location.href = '/login';
  }
}

// 退出登录
$('#btn-logout').addEventListener('click', async () => {
  await fetch('/api/v1/auth/logout', { method: 'POST', credentials: 'include' });
  window.location.href = '/login';
});

// 初始化时先验证身份
checkAuth().then(() => {
  loadExercises();
  loadWorkouts();
});

function fmtLocalDate(d) {
  const yyyy = d.getFullYear();
  const mm = String(d.getMonth() + 1).padStart(2, '0');
  const dd = String(d.getDate()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

// ========== 字段元数据 ==========
const FIELD_META = {
  weight:     { label: '重量', unit: 'kg',     type: 'number', step: '0.5', min: '0' },
  reps:       { label: '次数', unit: '个',     type: 'number', step: '1',   min: '1' },
  sets:       { label: '组数', unit: '组',     type: 'number', step: '1',   min: '1' },
  duration:   { label: '时长', unit: '分钟',   type: 'number', step: '0.5', min: '0' },
  speed:      { label: '速度', unit: 'km/h',   type: 'number', step: '0.1', min: '0' },
  incline:    { label: '坡度', unit: '%',      type: 'number', step: '1',   min: '0' },
  resistance: { label: '阻尼', unit: '级',     type: 'number', step: '1',   min: '1' },
  distance:   { label: '距离', unit: '米',     type: 'number', step: '1',   min: '0' },
};

// 根据 fields 数组渲染输入表单
function renderFieldInputs(containerId, fields, prefix) {
  const container = $(containerId);
  container.innerHTML = '';
  for (const key of fields) {
    const meta = FIELD_META[key];
    if (!meta) continue;
    const div = document.createElement('div');
    div.className = 'field';
    const inputId = `${prefix}-${key}`;
    div.innerHTML = `<label>${meta.label} (${meta.unit})</label>
      <input type="${meta.type}" id="${inputId}" step="${meta.step}" min="${meta.min || 0}" ${key === 'sets' && prefix === 'warmup' ? 'value="1"' : ''}>`;
    container.appendChild(div);
  }
}

// 收集字段值
function collectFieldValues(fields, prefix) {
  const result = {};
  for (const key of fields) {
    const el = $(`#${prefix}-${key}`);
    if (el) {
      const v = el.value;
      result[key] = el.type === 'number' ? (v === '' ? 0 : parseFloat(v)) : v;
    }
  }
  return result;
}

// 清空字段输入
function clearFieldValues(fields, prefix) {
  for (const key of fields) {
    const el = $(`#${prefix}-${key}`);
    if (el) el.value = (key === 'sets' && prefix === 'warmup') ? '1' : '';
  }
}

// 获取当前动作的 fields
function getCurrentFields() {
  const id = +$('#sel-exercise').value;
  const ex = exercises.find(e => e.id === id);
  if (!ex || !ex.fields) return ['weight', 'reps', 'sets'];
  try { return JSON.parse(ex.fields); } catch { return ['weight', 'reps', 'sets']; }
}

// ========== 标签切换 ==========
document.querySelectorAll('.tab-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-panel').forEach(p => p.classList.remove('active'));
    btn.classList.add('active');
    $(`#tab-${btn.dataset.tab}`).classList.add('active');
    if (btn.dataset.tab === 'stats') loadStatCategories();
  });
});

// ========== 初始化 ==========
  // 默认训练量统计日期范围为近30天
  const today = new Date();
  const todayStr = fmtLocalDate(today);
  const thirtyDaysAgo = new Date(today);
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30);
  const startStr = fmtLocalDate(thirtyDaysAgo);

  $('#train-date').value = todayStr;
  $('#train-date').max = todayStr;
  $('#stat-start').value = startStr;
  $('#stat-end').value = todayStr;


// ========== 训练 ==========
$('#btn-new-workout').addEventListener('click', async () => {
  const date = $('#train-date').value;
  if (!date) return toast('请选择日期');
  const w = await req('/workouts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ date })
  });
  currentWorkoutId = w.id;
  $('#workout-area').classList.remove('hidden');
  $('#set-list').innerHTML = '';
  toast('训练已创建，开始记录');
});

// 热身组切换
$('#btn-toggle-warmup').addEventListener('click', () => {
  const area = $('#warmup-area');
  const wasHidden = area.classList.contains('hidden');
  area.classList.toggle('hidden');
  const fields = getCurrentFields();
  if (wasHidden) {
    renderFieldInputs('#warmup-fields', fields, 'warmup');
  } else {
    clearFieldValues(fields, 'warmup');
  }
});

// 动作选择变化：动态渲染表单
$('#sel-exercise').addEventListener('change', () => {
  const fields = getCurrentFields();
  renderFieldInputs('#normal-fields', fields, 'normal');
  if (!$('#warmup-area').classList.contains('hidden')) {
    renderFieldInputs('#warmup-fields', fields, 'warmup');
  }
});

// 初始渲染一次表单
renderFieldInputs('#normal-fields', ['weight', 'reps', 'sets'], 'normal');

$('#btn-add-set').addEventListener('click', async () => {
  if (!currentWorkoutId) return toast('请先创建训练');
  const exercise_id = +$('#sel-exercise').value;
  if (!exercise_id) return toast('请选择动作');

  const fields = getCurrentFields();

  // 热身组
  const warmupArea = $('#warmup-area');
  if (!warmupArea.classList.contains('hidden')) {
    const wvals = collectFieldValues(fields, 'warmup');
    const hasVal = fields.some(k => wvals[k] > 0);
    if (hasVal) {
      const payload = buildPayload(exercise_id, fields, wvals, true);
      await req(`/workouts/${currentWorkoutId}/sets`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
    }
  }

  // 正常组
  const vals = collectFieldValues(fields, 'normal');
  const hasVal = fields.some(k => vals[k] > 0);
  if (!hasVal) return toast('请填写至少一项数据');

  const payload = buildPayload(exercise_id, fields, vals, false);
  await req(`/workouts/${currentWorkoutId}/sets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload)
  });

  // 清空输入
  clearFieldValues(fields, 'normal');
  clearFieldValues(fields, 'warmup');

  // 收起热身组
  $('#warmup-area').classList.add('hidden');

  await loadSets();
  loadWorkouts();
  toast('添加成功');
});

// 构建上传 payload
function buildPayload(exercise_id, fields, vals, isWarmup) {
  const payload = { exercise_id, is_warmup: isWarmup, weight: 0, reps: 0, rpe: 1, extra: '{}' };
  const extra = {};
  for (const key of fields) {
    const v = vals[key] || 0;
    if (key === 'weight') payload.weight = v;
    else if (key === 'reps') payload.reps = v;
    else if (key === 'sets') payload.rpe = v || 1;
    else extra[key] = v;
  }
  payload.extra = JSON.stringify(extra);
  return payload;
}

function renderSetItem(s, list) {
  const d = new Date(s.created_at);
  const timeStr = d.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
  const warmupCls = s.is_warmup ? 'set-warmup' : '';

  // 解析 extra
  let extra = {};
  try { if (s.extra) extra = JSON.parse(s.extra); } catch {}

  // 获取动作 fields
  const ex = exercises.find(e => e.id === s.exercise_id);
  let fields = ['weight', 'reps', 'sets'];
  if (ex && ex.fields) {
    try { fields = JSON.parse(ex.fields); } catch {}
  }

// 根据 fields 构建简洁显示文本（带颜色）
  const vals = [];
  const hasWeight = fields.includes('weight');
  for (const key of fields) {
    const meta = FIELD_META[key];
    if (!meta) continue;
    let v;
    if (key === 'weight') v = s.weight;
    else if (key === 'reps') v = s.reps;
    else if (key === 'sets') v = s.rpe;
    else v = extra[key];
    if (v !== undefined && v !== null && v !== 0) {
      if (hasWeight) {
        vals.push(`${v}${meta.unit}`);
      } else {
        if (key === 'duration') vals.push(`${v}min`);
        else if (key === 'distance') vals.push(`${v}m`);
        else vals.push(`${meta.label}${v}`);
      }
    }
  }
  let textHtml = '';
  if (vals.length === 1) {
    textHtml = `<span style="color:var(--accent)">${vals[0]}</span>`;
  } else if (vals.length >= 2) {
    textHtml = `<span style="color:var(--accent)">${vals[0]}</span>`;
    for (let i = 1; i < vals.length - 1; i++) {
      const sep = hasWeight ? ' × ' : ' / ';
      textHtml += sep + vals[i];
    }
    const lastIdx = vals.length - 1;
    const lastSep = hasWeight ? (lastIdx === 1 ? ' × ' : ' @ ') : ' / ';
    textHtml += lastSep + `<span style="color:#0ea5e9">${vals[lastIdx]}</span>`;
  }

  const div = document.createElement('div');
  div.className = `set-item ${warmupCls}`;
  div.innerHTML = `
    <div class="set-data">
      <span>${timeStr}</span>
      <span>${textHtml}</span>
      ${s.is_warmup ? '<span style="color:var(--secondary);font-size:12px;">[热身]</span>' : ''}
    </div>
    <button class="btn small danger" data-id="${s.id}">删除</button>
  `;
  div.querySelector('.btn.small.danger').addEventListener('click', async () => {
    await req(`/sets/${s.id}`, { method: 'DELETE' });
    await loadSets();
    loadWorkouts();
  });
  list.appendChild(div);
}

async function loadSets() {
  if (!currentWorkoutId) return;
  const w = await req(`/workouts/${currentWorkoutId}`);
  const list = $('#set-list');
  list.innerHTML = '';
  if (!w.sets || w.sets.length === 0) {
    list.innerHTML = '<div style="color:var(--secondary);text-align:center;padding:20px;">暂无记录</div>';
    return;
  }

  const groups = {};
  for (const s of w.sets) {
    const name = s.exercise_name || '未知动作';
    if (!groups[name]) groups[name] = { warmups: [], normals: [] };
    if (s.is_warmup) groups[name].warmups.push(s);
    else groups[name].normals.push(s);
  }

  for (const [name, sets] of Object.entries(groups)) {
    const h = document.createElement('div');
    h.style.cssText = 'font-weight:600;margin:10px 0 6px;color:var(--text);';
    h.textContent = name;
    list.appendChild(h);

    if (sets.warmups.length > 0) {
      const wt = document.createElement('div');
      wt.className = 'warmup-title';
      wt.textContent = '热身组';
      list.appendChild(wt);
      sets.warmups.forEach(s => renderSetItem(s, list));
    }
    sets.normals.forEach(s => renderSetItem(s, list));
  }
}

async function loadWorkouts() {
  const list = await req('/workouts');
  const container = $('#workout-list');
  container.innerHTML = '';
  if (list.length === 0) {
    container.innerHTML = '<div style="color:var(--secondary);text-align:center;padding:20px;">暂无训练</div>';
    return;
  }

  // 按日期分组
  const byDate = {};
  for (const w of list) {
    if (!byDate[w.date]) byDate[w.date] = [];
    byDate[w.date].push(w);
  }

  for (const [date, workouts] of Object.entries(byDate)) {
    const validWorkouts = workouts.filter(w => w.sets && w.sets.length > 0);
    if (validWorkouts.length === 0) continue;

    const dateH = document.createElement('div');
    dateH.style.cssText = 'font-weight:600;margin:10px 0 6px;color:var(--accent);';
    dateH.textContent = date;
    container.appendChild(dateH);

    for (const w of validWorkouts) {
      const hasSets = w.sets && w.sets.length > 0;
      const timeInfo = w.time_ranges && w.time_ranges.length > 0 ? w.time_ranges.join(', ') : '';

      const div = document.createElement('div');
      div.className = 'workout-item';
      div.style.cssText = 'display:flex;justify-content:space-between;align-items:center;padding:10px 12px;border-bottom:1px solid var(--border);cursor:pointer;';
      div.innerHTML = `<span>${timeInfo || '未开始训练'}</span>
        <span style="color:var(--secondary);font-size:13px;">${hasSets ? w.sets.length + '次数据录入' : '空'}</span>`;
      div.addEventListener('click', () => {
        currentWorkoutId = w.id;
        $('#workout-area').classList.remove('hidden');
        $('#train-date').value = w.date;
        loadSets();
      });
      container.appendChild(div);

      if (hasSets) {
        const setContainer = document.createElement('div');
        setContainer.style.cssText = 'padding:0 12px 8px;';
        const exNames = [...new Set(w.sets.map(s => s.exercise_name).filter(Boolean))];
        setContainer.innerHTML = exNames.map(n => `<span style="color:var(--secondary);font-size:12px;">${n}</span>`).join(' ');
        container.appendChild(setContainer);
      }
    }
  }
}

// ========== 动作库 ==========
async function loadExercises() {
  exercises = await req('/exercises');
  populateCategorySelect();
  renderExerciseLibrary();
}

function populateCategorySelect() {
  const sel = $('#sel-category');
  sel.innerHTML = '<option value="">请选择</option>';
  const cats = [...new Set(exercises.map(e => e.category).filter(Boolean))];
  cats.forEach(c => {
    const opt = document.createElement('option');
    opt.value = c;
    opt.textContent = c;
    sel.appendChild(opt);
  });
}

$('#sel-category').addEventListener('change', () => {
  const cat = $('#sel-category').value;
  const sel = $('#sel-exercise');
  sel.innerHTML = '<option value="">请选择动作</option>';
  exercises.filter(e => e.category === cat).forEach(e => {
    const opt = document.createElement('option');
    opt.value = e.id;
    opt.textContent = e.name;
    sel.appendChild(opt);
  });
  // 重新渲染表单
  const fields = getCurrentFields();
  renderFieldInputs('#normal-fields', fields, 'normal');
});

function renderExerciseLibrary() {
  const list = $('#ex-list');
  list.innerHTML = '';
  const groups = {};
  exercises.forEach(e => {
    const cat = e.category || '未分类';
    if (!groups[cat]) groups[cat] = [];
    groups[cat].push(e);
  });

  for (const [cat, items] of Object.entries(groups)) {
    const group = document.createElement('div');
    group.className = 'ex-group collapsed';
    group.innerHTML = `
      <div class="ex-group-header">${cat}</div>
      <div class="ex-group-body"></div>
    `;
    const body = group.querySelector('.ex-group-body');
    items.forEach(e => {
      const div = document.createElement('div');
      div.className = 'item';
      div.innerHTML = `
        <div class="item-info">
          <div class="item-title">${e.name}</div>
        </div>
        <button class="btn small danger" data-id="${e.id}">删除</button>
      `;
      div.querySelector('button').addEventListener('click', async () => {
        if (!confirm('确定删除？')) return;
        await req(`/exercises/${e.id}`, { method: 'DELETE' });
        await loadExercises();
        toast('已删除');
      });
      body.appendChild(div);
    });
    group.querySelector('.ex-group-header').addEventListener('click', () => {
      group.classList.toggle('collapsed');
    });
    list.appendChild(group);
  }
}

function getSelectedFields() {
  const checked = $('#ex-fields-wrap').querySelectorAll('input[type="checkbox"]:checked');
  return Array.from(checked).map(cb => cb.value);
}

$('#btn-add-ex').addEventListener('click', async () => {
  const name = $('#ex-name').value.trim();
  const category = $('#ex-category').value.trim();
  if (!name || !category) return toast('请填写动作名称和分类');
  const fields = getSelectedFields();
  if (fields.length === 0) return toast('请至少选择一个字段');

  await req('/exercises', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, category, fields: JSON.stringify(fields) })
  });
  $('#ex-name').value = '';
  $('#ex-category').value = '';
  // 重置复选框
  $('#ex-fields-wrap').querySelectorAll('input').forEach(cb => {
    cb.checked = ['weight', 'reps', 'sets'].includes(cb.value);
  });
  await loadExercises();
  toast('动作已添加');
});

// ========== 统计 ==========
function loadStatCategories() {
  const sel = $('#stat-category');
  sel.innerHTML = '<option value="">请选择</option>';
  const cats = [...new Set(exercises.map(e => e.category).filter(Boolean))];
  cats.forEach(c => {
    const opt = document.createElement('option');
    opt.value = c;
    opt.textContent = c;
    sel.appendChild(opt);
  });
  $('#stat-exercise').innerHTML = '<option value="">请先选择分类</option>';
}

$('#stat-category').addEventListener('change', () => {
  const cat = $('#stat-category').value;
  const sel = $('#stat-exercise');
  sel.innerHTML = '<option value="">请选择动作</option>';
  exercises.filter(e => e.category === cat).forEach(e => {
    const opt = document.createElement('option');
    opt.value = e.id;
    opt.textContent = e.name;
    sel.appendChild(opt);
  });
});

$('#btn-load-pr').addEventListener('click', async () => {
  const id = +$('#stat-exercise').value;
  if (!id) return toast('请选择动作');
  const sets = await req(`/stats/exercise/${id}`);
  if (sets.length === 0) return toast('暂无数据');

  // 检查是否是力量型动作（有 weight 字段）
  const ex = exercises.find(e => e.id === id);
  let isStrength = true;
  if (ex && ex.fields) {
    try {
      const f = JSON.parse(ex.fields);
      isStrength = f.includes('weight');
    } catch {}
  }

  const labels = sets.map(s => s.workout_date || new Date(s.created_at).toLocaleDateString('zh-CN')).reverse();

  if (isStrength) {
    const data = sets.map(s => s.weight).reverse();
    renderChart('pr-chart', prChart, '重量 (kg)', labels, data, 'rgb(74,222,128)');
  } else {
    // 非力量型：显示时长或次数趋势
    const data = sets.map(s => {
      let extra = {};
      try { if (s.extra) extra = JSON.parse(s.extra); } catch {}
      if (extra.duration) return extra.duration;
      if (extra.distance) return extra.distance;
      return s.reps || 0;
    }).reverse();
    const label = ex && ex.fields && ex.fields.includes('duration') ? '时长' :
                  ex && ex.fields && ex.fields.includes('distance') ? '距离' : '次数';
    renderChart('pr-chart', prChart, label, labels, data, 'rgb(74,222,128)');
  }
});

function renderChart(canvasId, chartVar, label, labels, data, color) {
  const ctx = $(`#${canvasId}`).getContext('2d');
  if (chartVar) chartVar.destroy();
  chartVar = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label,
        data,
        borderColor: color,
        backgroundColor: color.replace('rgb', 'rgba').replace(')', ',0.1)'),
        fill: true,
        tension: 0.3,
        pointRadius: 3
      }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { display: false } },
      scales: {
        x: { ticks: { color: '#888', maxTicksLimit: 8 } },
        y: { ticks: { color: '#888' } }
      }
    }
  });
  if (canvasId === 'pr-chart') prChart = chartVar;
  else volChart = chartVar;
}

$('#btn-load-vol').addEventListener('click', async () => {
  const start = $('#stat-start').value;
  const end = $('#stat-end').value;
  if (!start || !end) return toast('请选择日期范围');
  const data = await req(`/stats/volume?start=${start}&end=${end}`);
  const labels = Object.keys(data).sort();
  const values = labels.map(d => data[d]);
  if (labels.length === 0) return toast('暂无数据');
  renderChart('vol-chart', volChart, '训练量 (kg×次)', labels, values, 'rgb(96,165,250)');
});
