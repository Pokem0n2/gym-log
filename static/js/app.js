const API = '/api/v1';
let currentWorkoutId = null;
let exercises = [];
let prChart = null;
let volChart = null;

// 工具
function $(s) { return document.querySelector(s); }
function toast(msg) {
  const t = $('#toast');
  t.textContent = msg;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 2000);
}
async function req(path, opt) {
  const res = await fetch(API + path, opt);
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(txt);
  }
  if (res.status === 204) return null;
  return res.json();
}

// 标签切换
document.querySelectorAll('.tab-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
    document.querySelectorAll('.tab-panel').forEach(p => p.classList.remove('active'));
    btn.classList.add('active');
    $('#tab-' + btn.dataset.tab).classList.add('active');
    if (btn.dataset.tab === 'train') loadWorkouts();
    if (btn.dataset.tab === 'exercises') loadExercises();
  });
});

// 初始化日期
$('#train-date').valueAsDate = new Date();
$('#stat-end').valueAsDate = new Date();
const d30 = new Date(); d30.setDate(d30.getDate() - 30);
$('#stat-start').valueAsDate = d30;

// ========== 动作库 ==========
async function loadExercises() {
  exercises = await req('/exercises');
  const sel = $('#sel-exercise');
  const statSel = $('#stat-exercise');
  sel.innerHTML = '<option value="">请选择</option>';
  statSel.innerHTML = '<option value="">请选择</option>';
  exercises.forEach(e => {
    sel.add(new Option(e.name, e.id));
    statSel.add(new Option(e.name, e.id));
  });

  const list = $('#ex-list');
  list.innerHTML = '';
  exercises.forEach(e => {
    const div = document.createElement('div');
    div.className = 'item';
    div.innerHTML = `
      <div class="item-info">
        <div class="item-title">${e.name}</div>
        <div class="item-meta">${e.category || '未分类'}</div>
      </div>
      <button class="btn small danger" data-id="${e.id}">删除</button>
    `;
    div.querySelector('button').addEventListener('click', async () => {
      await req(`/exercises/${e.id}`, { method: 'DELETE' });
      loadExercises();
      toast('已删除');
    });
    list.appendChild(div);
  });
}

$('#btn-add-ex').addEventListener('click', async () => {
  const name = $('#ex-name').value.trim();
  if (!name) return toast('请输入动作名称');
  await req('/exercises', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, category: $('#ex-category').value.trim() })
  });
  $('#ex-name').value = '';
  $('#ex-category').value = '';
  loadExercises();
  toast('动作已添加');
});

// ========== 训练 ==========
$('#btn-new-workout').addEventListener('click', async () => {
  const date = $('#train-date').value;
  const w = await req('/workouts', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ date, notes: '' })
  });
  currentWorkoutId = w.id;
  $('#workout-area').classList.remove('hidden');
  loadSets();
  toast('训练已创建，开始记录');
});

$('#btn-add-set').addEventListener('click', async () => {
  if (!currentWorkoutId) return toast('请先创建训练');
  const exercise_id = +$('#sel-exercise').value;
  const weight = +$('#set-weight').value;
  const reps = +$('#set-reps').value;
  const rpe = $('#set-rpe').value ? +$('#set-rpe').value : null;
  if (!exercise_id) return toast('请选择动作');
  if (!weight || !reps) return toast('请填写重量和次数');

  await req(`/workouts/${currentWorkoutId}/sets`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ exercise_id, weight, reps, rpe })
  });
  $('#set-weight').value = '';
  $('#set-reps').value = '';
  $('#set-rpe').value = '';
  loadSets();
});

async function loadSets() {
  if (!currentWorkoutId) return;
  const w = await req(`/workouts/${currentWorkoutId}`);
  const list = $('#set-list');
  list.innerHTML = '';
  if (!w.sets || !w.sets.length) {
    list.innerHTML = '<div class="item-meta">暂无记录</div>';
    return;
  }
  const exMap = {};
  exercises.forEach(e => exMap[e.id] = e.name);

  // 按动作分组
  const groups = {};
  w.sets.forEach(s => {
    const name = exMap[s.exercise_id] || '未知';
    if (!groups[name]) groups[name] = [];
    groups[name].push(s);
  });

  for (const [name, sets] of Object.entries(groups)) {
    const h = document.createElement('div');
    h.style.cssText = 'font-weight:600;margin:10px 0 6px;color:var(--accent);';
    h.textContent = name;
    list.appendChild(h);
    sets.forEach(s => {
      const div = document.createElement('div');
      div.className = 'set-item';
      div.innerHTML = `
        <span class="set-data">
          <strong>${s.weight}kg</strong> × ${s.reps}
          ${s.rpe ? `<span style="color:var(--text-secondary)">@RPE${s.rpe}</span>` : ''}
        </span>
        <button class="btn small danger" data-id="${s.id}">删</button>
      `;
      div.querySelector('button').addEventListener('click', async () => {
        await req(`/sets/${s.id}`, { method: 'DELETE' });
        loadSets();
      });
      list.appendChild(div);
    });
  }
}

async function loadWorkouts() {
  const list = await req('/workouts');
  const el = $('#workout-list');
  el.innerHTML = '';
  if (!list.length) {
    el.innerHTML = '<div class="item-meta">暂无训练记录</div>';
    return;
  }
  list.forEach(w => {
    const div = document.createElement('div');
    div.className = 'item';
    div.innerHTML = `
      <div class="item-info">
        <div class="item-title">${w.date}</div>
        <div class="item-meta">${w.notes || '无备注'}</div>
      </div>
      <button class="btn small danger" data-id="${w.id}">删除</button>
    `;
    div.querySelector('button').addEventListener('click', async () => {
      await req(`/workouts/${w.id}`, { method: 'DELETE' });
      loadWorkouts();
    });
    el.appendChild(div);
  });
}

// ========== 统计 ==========
$('#btn-load-pr').addEventListener('click', async () => {
  const id = +$('#stat-exercise').value;
  if (!id) return toast('请选择动作');
  const data = await req(`/stats/exercise/${id}`);
  const ctx = $('#pr-chart').getContext('2d');
  if (prChart) prChart.destroy();

  const labels = data.map((_, i) => `#${data.length - i}`).reverse();
  const weights = data.map(s => s.weight).reverse();
  const reps = data.map(s => s.reps).reverse();

  prChart = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [
        { label: '重量(kg)', data: weights, borderColor: '#4ade80', tension: 0.3, fill: false },
        { label: '次数', data: reps, borderColor: '#60a5fa', tension: 0.3, fill: false }
      ]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { labels: { color: '#e0e0e0' } } },
      scales: {
        x: { ticks: { color: '#888' }, grid: { color: '#2a2a2a' } },
        y: { ticks: { color: '#888' }, grid: { color: '#2a2a2a' } }
      }
    }
  });
});

$('#btn-load-vol').addEventListener('click', async () => {
  const start = $('#stat-start').value;
  const end = $('#stat-end').value;
  if (!start || !end) return toast('请选择日期范围');
  const data = await req(`/stats/volume?start=${start}&end=${end}`);
  const ctx = $('#vol-chart').getContext('2d');
  if (volChart) volChart.destroy();

  const dates = Object.keys(data).sort();
  const volumes = dates.map(d => data[d]);

  volChart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: dates,
      datasets: [{ label: '训练量 (kg×reps)', data: volumes, backgroundColor: '#4ade80', borderRadius: 4 }]
    },
    options: {
      responsive: true,
      maintainAspectRatio: false,
      plugins: { legend: { labels: { color: '#e0e0e0' } } },
      scales: {
        x: { ticks: { color: '#888' }, grid: { color: '#2a2a2a' } },
        y: { ticks: { color: '#888' }, grid: { color: '#2a2a2a' } }
      }
    }
  });
});

// 初始加载
loadExercises();
loadWorkouts();
