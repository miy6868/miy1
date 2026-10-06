/* 한국사2 단기 완성 — 앱 로직 (라이브러리 없이 순수 JS) */
(function () {
  'use strict';

  var SECTIONS = window.SECTIONS, UNITS = window.UNITS, TIMELINE = window.TIMELINE,
      TERMS = window.TERMS, QUIZ = window.QUIZ, COMPARES = window.COMPARES;
  var SEC_IDS = Object.keys(SECTIONS);
  var view = document.getElementById('view');
  var LETTERS = ['(가)', '(나)', '(다)', '(라)'];
  var NUMS = ['①', '②', '③', '④', '⑤'];

  /* ───── 저장 ───── */
  var KEY = 'hk2-dana-v1';
  var S = (function () {
    var d = {};
    try { d = JSON.parse(localStorage.getItem(KEY)) || {}; } catch (e) { d = {}; }
    var base = { exam: '', read: {}, stats: {}, wrong: {}, box: {}, blank: false, cover: false, hideYear: false, solved: 0 };
    for (var k in base) if (!(k in d)) d[k] = base[k];
    return d;
  })();
  function save() { try { localStorage.setItem(KEY, JSON.stringify(S)); } catch (e) { /* 저장 불가 환경: 무시 */ } }

  /* ───── 유틸 ───── */
  function esc(s) { return String(s).replace(/[&<>"]/g, function (c) { return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }[c]; }); }
  function fmt(s) { return esc(s).replace(/==(.+?)==/g, '<mark class="k">$1</mark>').replace(/\*\*(.+?)\*\*/g, '<b>$1</b>'); }
  function plain(s) { return String(s).replace(/==|\*\*/g, ''); }
  function shuffle(a) { a = a.slice(); for (var i = a.length - 1; i > 0; i--) { var j = Math.floor(Math.random() * (i + 1)); var t = a[i]; a[i] = a[j]; a[j] = t; } return a; }
  function pick(a) { return a[Math.floor(Math.random() * a.length)]; }
  function secLabel(id) { return SECTIONS[id].no; }
  function ym(e) { return e.y + (e.m ? '.' + e.m : ''); }
  function less(a, b) { return a.y < b.y || (a.y === b.y && !!a.m && !!b.m && a.m < b.m); }
  function comparable(a, b) { return less(a, b) || less(b, a); }
  function todayStr() { var d = new Date(); return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0'); }
  function daysLeft() {
    if (!S.exam) return null;
    var t = new Date(todayStr() + 'T00:00:00'), e = new Date(S.exam + 'T00:00:00');
    return Math.round((e - t) / 86400000);
  }
  function dLabel(n) { return n === null ? '' : (n > 0 ? 'D-' + n : n === 0 ? 'D-DAY' : 'D+' + (-n)); }
  function acc(sid) { var st = S.stats[sid]; return st && st.n ? Math.round(st.c / st.n * 100) : null; }
  function termKey(t) { return t.s + ':' + t.t; }

  function updateTopD() {
    var el = document.getElementById('topD'), n = daysLeft();
    el.hidden = n === null; el.textContent = dLabel(n);
  }

  /* ───── 라우팅 ───── */
  var quizState = null;
  function setTab(tab) {
    document.querySelectorAll('.tab').forEach(function (a) {
      if (a.dataset.tab === tab) a.setAttribute('aria-current', 'page'); else a.removeAttribute('aria-current');
    });
  }
  function route() {
    var h = (location.hash || '#home').slice(1);
    if (SECTIONS[h]) { setTab('learn'); renderSection(h); }
    else if (h === 'learn' || h === 'compare' || h === 'cram') { setTab('learn'); renderLearn(h); }
    else if (h === 'timeline') { setTab('timeline'); renderTimeline(); }
    else if (h === 'quiz') { setTab('quiz'); renderQuizSetup(); }
    else if (h === 'cards') { setTab('cards'); renderCards(); }
    else { setTab('home'); renderHome(); }
    updateTopD();
    window.scrollTo(0, 0);
  }
  window.addEventListener('hashchange', route);
  // 이미 열린 탭을 다시 누르면 처음 화면으로
  document.querySelectorAll('.tab').forEach(function (a) {
    a.addEventListener('click', function (e) {
      if (a.getAttribute('href') === (location.hash || '#home')) { e.preventDefault(); route(); }
    });
  });
  function toQuizHash() {
    if (location.hash === '#quiz') return;
    try { history.replaceState(null, '', '#quiz'); } catch (e) { /* 샌드박스에서 막히면 해시 없이 진행 */ }
  }

  /* ───── 홈 ───── */
  function buildPlan(days) {
    var secs = SEC_IDS.slice();
    var plan = [];
    if (days <= 1) {
      plan.push({ b: '벼락치기 몰아보기', t: '핵심 한 줄 전체 → 연표 → 섞어서 30문제 → 오답 다시', go: '#cram' });
      return plan;
    }
    var studyDays = Math.min(days - 1, secs.length);
    var per = Math.ceil(secs.length / studyDays);
    for (var i = 0; i < secs.length; i += per) {
      var chunk = secs.slice(i, i + per);
      plan.push({
        b: chunk.map(function (s) { return SECTIONS[s].no; }).join(' + ') + ' 개념',
        t: chunk.map(function (s) { return SECTIONS[s].title; }).join(' / ') + ' → 단원 문제 10개',
        secs: chunk, go: '#' + chunk[0]
      });
    }
    var rest = days - plan.length;
    if (rest >= 3) plan.push({ b: '비교표 + 연표 순서 문제', t: '헷갈리는 짝 정리, 순서 배열·시기 찾기 20문제', go: '#compare' });
    if (rest >= 2) plan.push({ b: '전 범위 섞어서 30문제', t: '틀린 문제는 오답 노트로 자동 저장', go: '#quiz' });
    plan.push({ b: '시험 전날: 벼락치기 + 오답 다시', t: '핵심 한 줄 가리고 외우기 → 오답 노트 전부 다시 풀기', go: '#cram' });
    return plan;
  }

  function renderHome() {
    var n = daysLeft();
    var readN = SEC_IDS.filter(function (s) { return S.read[s]; }).length;
    var tc = 0, tn = 0; SEC_IDS.forEach(function (s) { var st = S.stats[s]; if (st) { tc += st.c; tn += st.n; } });
    var mastered = TERMS.filter(function (t) { return (S.box[termKey(t)] || 0) >= 2; }).length;
    var wrongN = Object.keys(S.wrong).length;
    var planDays = (n !== null && n >= 0) ? Math.max(n, 1) : 7;
    var plan = buildPlan(planDays);

    var base = new Date(todayStr() + 'T00:00:00');
    var planHtml = plan.map(function (p, i) {
      var d = new Date(base.getTime() + i * 86400000);
      var dl = (i === 0 ? '오늘' : (d.getMonth() + 1) + '/' + d.getDate());
      var done = p.secs && p.secs.every(function (s) { return S.read[s]; });
      return '<li class="' + (i === 0 ? 'today' : '') + '"><span class="d">' + dl + '</span>' +
        '<a class="what" href="' + p.go + '" style="text-decoration:none;color:inherit">' +
        '<b>' + esc(p.b) + (done ? ' <span class="pill done">완료</span>' : '') + '</b><span>' + esc(p.t) + '</span></a></li>';
    }).join('');

    var secRows = SEC_IDS.map(function (s) {
      var a = acc(s), sec = SECTIONS[s];
      var pill = S.read[s] ? '<span class="pill done">읽음</span>' : '<span class="pill todo">안 읽음</span>';
      if (a !== null && a < 60) pill = '<span class="pill weak">약점 ' + a + '%</span>';
      return '<a class="secrow" href="#' + s + '"><span class="tag">' + sec.no + '</span>' +
        '<span class="t"><b>' + esc(sec.title) + '</b><span class="bar"><i style="width:' + (a === null ? 0 : a) + '%"></i></span></span>' +
        '<span class="row" style="flex-direction:column;align-items:flex-end;gap:2px">' + pill +
        '<span class="acc">' + (a === null ? '문제 전' : '정답 ' + a + '%') + '</span></span></a>';
    }).join('');

    view.innerHTML =
      '<section class="card hero">' +
        '<div class="dday-wrap">' +
          '<div><div class="dday-cap">' + (n === null ? '시험 날짜를 넣으면 남은 날짜에 맞춰 계획을 짜 줘요' : '중간고사까지') + '</div>' +
          '<div class="dday-num">' + (n === null ? 'D-?' : dLabel(n)) + '</div></div>' +
          '<label class="date-field" for="exam-date">시험 날짜<input type="date" id="exam-date" value="' + esc(S.exam) + '"></label>' +
        '</div>' +
        '<div class="scope-line">범위: 한국사2 <b>Ⅰ. 일제 식민지 지배와 민족 운동의 전개</b> 전체 + <b>Ⅱ-1. 8·15 광복과 대한민국 정부의 수립</b></div>' +
      '</section>' +
      '<section class="stat3">' +
        '<div class="stat"><span>개념 읽음</span><b>' + readN + '/' + SEC_IDS.length + '</b><div class="bar"><i style="width:' + (readN / SEC_IDS.length * 100) + '%"></i></div></div>' +
        '<div class="stat"><span>정답률</span><b>' + (tn ? Math.round(tc / tn * 100) + '%' : '–') + '</b><div class="bar"><i style="width:' + (tn ? tc / tn * 100 : 0) + '%"></i></div></div>' +
        '<div class="stat"><span>카드 외움</span><b>' + mastered + '/' + TERMS.length + '</b><div class="bar"><i style="width:' + (mastered / TERMS.length * 100) + '%"></i></div></div>' +
      '</section>' +
      '<section class="card stack">' +
        '<div><div class="eyebrow">남은 ' + planDays + '일 학습 계획' + (n === null ? ' · 예시(7일 기준)' : '') + '</div></div>' +
        '<ol class="plan">' + planHtml + '</ol>' +
      '</section>' +
      '<section class="quick">' +
        '<button class="btn" data-q="mix20">섞어서 20문제<small>객관식·OX·순서·시기</small></button>' +
        '<button class="btn" data-q="order10">순서 배열 10문제<small>연표 감각 잡기</small></button>' +
        '<button class="btn" data-q="period10">시기 찾기 10문제<small>(가)~(라) 구간 고르기</small></button>' +
        '<button class="btn" data-q="wrong"' + (wrongN ? '' : ' disabled') + '>오답 다시 풀기<small>' + wrongN + '문제 저장됨</small></button>' +
        '<a class="btn" href="#cram">벼락치기 핵심 한 줄<small>가리고 외우기</small></a>' +
        '<a class="btn" href="#compare">헷갈리는 것 비교표<small>시험 단골 7개</small></a>' +
      '</section>' +
      '<section class="card"><div class="eyebrow" style="margin-bottom:4px">단원별 현황</div>' + secRows + '</section>' +
      '<p class="footnote">기록은 이 기기의 브라우저에만 저장돼요.</p>';

    document.getElementById('exam-date').addEventListener('change', function (e) {
      S.exam = e.target.value; save(); renderHome(); updateTopD();
    });
    view.querySelectorAll('[data-q]').forEach(function (b) {
      b.addEventListener('click', function () {
        var k = b.dataset.q;
        setTab('quiz');
        if (k === 'mix20') startQuiz({ scope: SEC_IDS, type: 'mix', count: 20 });
        if (k === 'order10') startQuiz({ scope: SEC_IDS, type: 'order', count: 10 });
        if (k === 'period10') startQuiz({ scope: SEC_IDS, type: 'period', count: 10 });
        if (k === 'wrong') startWrong();
      });
    });
  }

  /* ───── 개념 목록 / 비교표 / 벼락치기 ───── */
  function learnSeg(active) {
    return '<div class="seg" role="group" aria-label="개념 보기 방식">' +
      [['learn', '단원 정리'], ['compare', '비교표'], ['cram', '벼락치기']].map(function (x) {
        return '<button data-go="' + x[0] + '" aria-pressed="' + (active === x[0]) + '">' + x[1] + '</button>';
      }).join('') + '</div>';
  }
  function bindSeg() {
    view.querySelectorAll('[data-go]').forEach(function (b) { b.addEventListener('click', function () { location.hash = b.dataset.go; }); });
  }

  function renderLearn(mode) {
    if (mode === 'compare') return renderCompare();
    if (mode === 'cram') return renderCram();
    var html = learnSeg('learn');
    UNITS.forEach(function (u) {
      html += '<div class="unit-h"><span class="no">' + u.no + '</span><h2>' + esc(u.title) + '</h2></div>';
      u.sections.forEach(function (s) {
        var sec = SECTIONS[s], a = acc(s);
        html += '<a class="card seclink" href="#' + s + '">' +
          '<div class="top-line"><span class="row"><span class="tag">' + sec.no + '</span><span class="acc">' + sec.era + '</span></span>' +
          (S.read[s] ? '<span class="pill done">읽음' + (a !== null ? ' · ' + a + '%' : '') + '</span>' : '<span class="pill todo">안 읽음</span>') + '</div>' +
          '<h3>' + esc(sec.title) + '</h3><p>' + esc(sec.oneLine) + '</p></a>';
      });
    });
    html += '<p class="footnote">Ⅱ단원은 1번 중단원(정부 수립)까지만 시험 범위예요.</p>';
    view.innerHTML = html;
    bindSeg();
  }

  function renderSection(id) {
    var sec = SECTIONS[id], idx = SEC_IDS.indexOf(id);
    var prev = SEC_IDS[idx - 1], next = SEC_IDS[idx + 1];
    var html = '<a class="btn ghost" href="#learn" style="justify-self:start;min-height:34px;padding:4px 10px">← 단원 목록</a>' +
      '<header class="sec-head"><div class="row"><span class="tag">' + sec.no + '</span><span class="acc">' + sec.era + '</span></div>' +
      '<h1>' + esc(sec.title) + '</h1><div class="oneline">' + esc(sec.oneLine) + '</div></header>' +
      '<div class="toolbar"><label class="switch" for="blank-sw"><input type="checkbox" id="blank-sw"' + (S.blank ? ' checked' : '') + '>빈칸 모드 (핵심어 가리기)</label>' +
      '<button class="btn ghost" id="reveal" style="min-height:34px;padding:4px 10px"' + (S.blank ? '' : ' hidden') + '>모두 보기</button></div>' +
      '<div id="sec-body" class="stack' + (S.blank ? ' blank' : '') + '">';
    sec.topics.forEach(function (t, i) {
      html += '<article class="card topic"><h3><span class="n">' + String(i + 1).padStart(2, '0') + '</span>' + esc(t.h) + '</h3><ul>' +
        t.points.map(function (p) { return '<li>' + fmt(p) + '</li>'; }).join('') + '</ul>' +
        (t.tip ? '<div class="tip">' + fmt(t.tip) + '</div>' : '') + '</article>';
    });
    html += '</div>' +
      '<section class="card"><div class="eyebrow" style="margin-bottom:6px">헷갈림 주의</div><div class="confuse">' +
      sec.confuse.map(function (c) { return '<div><b>' + esc(c[0]) + '</b><span>' + esc(c[1]) + '</span></div>'; }).join('') + '</div></section>' +
      '<section class="card"><div class="eyebrow" style="margin-bottom:10px">핵심 한 줄 (시험 직전 이것만)</div><ol class="keys">' +
      sec.keys.map(function (k) { return '<li><span>' + esc(k) + '</span></li>'; }).join('') + '</ol></section>' +
      '<section class="stack">' +
        '<button class="btn ' + (S.read[id] ? '' : 'primary') + ' block" id="mark-read">' + (S.read[id] ? '✓ 읽음 완료 (다시 누르면 취소)' : '이 단원 다 읽었어요') + '</button>' +
        '<div class="quick"><button class="btn" id="sec-quiz">이 단원 문제 풀기<small>객관식·OX·용어 10문제</small></button>' +
        '<button class="btn" id="sec-cards">이 단원 암기카드<small>' + TERMS.filter(function (t) { return t.s === id; }).length + '장</small></button></div>' +
      '</section>' +
      '<nav class="secnav">' +
        (prev ? '<a class="btn" href="#' + prev + '"><small>← 이전</small>' + SECTIONS[prev].no + ' ' + esc(SECTIONS[prev].title) + '</a>' : '<span></span>') +
        (next ? '<a class="btn next" href="#' + next + '"><small>다음 →</small>' + SECTIONS[next].no + ' ' + esc(SECTIONS[next].title) + '</a>' : '<span></span>') +
      '</nav>';
    view.innerHTML = html;

    var body = document.getElementById('sec-body');
    document.getElementById('blank-sw').addEventListener('change', function (e) {
      S.blank = e.target.checked; save();
      body.classList.toggle('blank', S.blank);
      body.querySelectorAll('mark.k.open').forEach(function (m) { m.classList.remove('open'); });
      document.getElementById('reveal').hidden = !S.blank;
    });
    document.getElementById('reveal').addEventListener('click', function () {
      body.querySelectorAll('mark.k').forEach(function (m) { m.classList.add('open'); });
    });
    body.querySelectorAll('mark.k').forEach(function (m) {
      m.setAttribute('tabindex', '0');
      function tog() { if (body.classList.contains('blank')) m.classList.toggle('open'); }
      m.addEventListener('click', tog);
      m.addEventListener('keydown', function (e) { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); tog(); } });
    });
    document.getElementById('mark-read').addEventListener('click', function () {
      S.read[id] = !S.read[id]; save(); renderSection(id);
    });
    document.getElementById('sec-quiz').addEventListener('click', function () { setTab('quiz'); startQuiz({ scope: [id], type: 'mix', count: 10 }); });
    document.getElementById('sec-cards').addEventListener('click', function () { cardScope = [id]; location.hash = 'cards'; });
  }

  function renderCompare() {
    var html = learnSeg('compare') +
      '<p class="muted" style="margin:0;font-size:13.5px">시험에서 보기 바꿔치기로 자주 나오는 짝들이에요. 표는 옆으로 밀어서 볼 수 있어요.</p>';
    COMPARES.forEach(function (c) {
      html += '<section class="stack" style="gap:8px"><div class="row" style="justify-content:space-between"><h2 class="h2" style="font-size:18px">' + esc(c.title) + '</h2>' +
        '<a class="tag" href="#' + c.sec + '" style="text-decoration:none">' + secLabel(c.sec) + ' 개념 →</a></div>' +
        '<div class="tbl-wrap"><table><thead><tr>' + c.head.map(function (h) { return '<th>' + esc(h) + '</th>'; }).join('') + '</tr></thead><tbody>' +
        c.rows.map(function (r) { return '<tr>' + r.map(function (x) { return '<td>' + esc(x) + '</td>'; }).join('') + '</tr>'; }).join('') +
        '</tbody></table></div></section>';
    });
    view.innerHTML = html;
    bindSeg();
  }

  function renderCram() {
    var html = learnSeg('cram') +
      '<div class="toolbar" style="top:calc(env(safe-area-inset-top, 0px) + 53px)"><label class="switch" for="cover-sw"><input type="checkbox" id="cover-sw"' + (S.cover ? ' checked' : '') + '>가리고 외우기 (눌러서 확인)</label>' +
      '<span class="acc">' + SEC_IDS.reduce(function (n, s) { return n + SECTIONS[s].keys.length; }, 0) + '줄</span></div>' +
      '<div id="cram-body" class="stack' + (S.cover ? ' cover' : '') + '">';
    SEC_IDS.forEach(function (s) {
      var sec = SECTIONS[s];
      html += '<section class="card"><div class="row" style="justify-content:space-between;margin-bottom:10px"><span class="row"><span class="tag">' + sec.no + '</span><b>' + esc(sec.title) + '</b></span></div>' +
        '<ol class="keys">' + sec.keys.map(function (k) { return '<li><span tabindex="0">' + esc(k) + '</span></li>'; }).join('') + '</ol></section>';
    });
    html += '</div>';
    view.innerHTML = html;
    bindSeg();
    var body = document.getElementById('cram-body');
    document.getElementById('cover-sw').addEventListener('change', function (e) {
      S.cover = e.target.checked; save();
      body.classList.toggle('cover', S.cover);
      body.querySelectorAll('li.open').forEach(function (li) { li.classList.remove('open'); });
    });
    body.querySelectorAll('.keys li > span').forEach(function (sp) {
      function tog() { if (body.classList.contains('cover')) sp.parentNode.classList.toggle('open'); }
      sp.addEventListener('click', tog);
      sp.addEventListener('keydown', function (e) { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); tog(); } });
    });
  }

  /* ───── 연표 ───── */
  var tlFilter = 'all';
  function renderTimeline() {
    var chips = '<button class="chip" data-f="all" aria-pressed="' + (tlFilter === 'all') + '">전체</button>' +
      SEC_IDS.map(function (s) { return '<button class="chip" data-f="' + s + '" aria-pressed="' + (tlFilter === s) + '">' + SECTIONS[s].no + '</button>'; }).join('');
    var evs = TIMELINE.filter(function (e) { return tlFilter === 'all' || e.s === tlFilter; });
    var byYear = [];
    evs.forEach(function (e) {
      var last = byYear[byYear.length - 1];
      if (last && last.y === e.y) last.list.push(e); else byYear.push({ y: e.y, list: [e] });
    });
    var rows = byYear.map(function (g) {
      var major = g.list.some(function (e) { return e.a; });
      return '<div class="tl-year"><div class="tl-y' + (major ? ' big' : '') + '" tabindex="0">' + g.y + '</div><div class="tl-evs">' +
        g.list.map(function (e) {
          return '<div class="tl-ev' + (e.a ? ' major' : '') + '"><span class="tl-m">' + (e.m ? e.m + '월' : '') + '</span><span>' + esc(e.e) + '</span>' +
            '<a class="tag" href="#' + e.s + '" style="text-decoration:none">' + secLabel(e.s) + '</a></div>';
        }).join('') + '</div></div>';
    }).join('');
    view.innerHTML =
      '<header class="stack" style="gap:6px"><h1 class="h2">한눈에 보는 연표</h1><p class="muted" style="margin:0;font-size:13.5px">굵은 글씨는 시기를 나누는 기준 사건이에요. 순서 배열·시기 찾기 문제가 여기서 나와요.</p></header>' +
      '<div class="chips" role="group" aria-label="단원 거르기">' + chips + '</div>' +
      '<div class="toolbar" style="top:calc(env(safe-area-inset-top, 0px) + 53px)"><label class="switch" for="hy-sw"><input type="checkbox" id="hy-sw"' + (S.hideYear ? ' checked' : '') + '>연도 가리기 (눌러서 확인)</label>' +
      '<button class="btn" id="tl-quiz" style="min-height:34px;padding:4px 12px">순서 문제 풀기</button></div>' +
      '<section class="card tl' + (S.hideYear ? ' hideyear' : '') + '" id="tl">' + rows + '</section>';

    view.querySelectorAll('[data-f]').forEach(function (b) { b.addEventListener('click', function () { tlFilter = b.dataset.f; renderTimeline(); }); });
    var tl = document.getElementById('tl');
    document.getElementById('hy-sw').addEventListener('change', function (e) {
      S.hideYear = e.target.checked; save(); tl.classList.toggle('hideyear', S.hideYear);
      tl.querySelectorAll('.open').forEach(function (x) { x.classList.remove('open'); });
    });
    tl.querySelectorAll('.tl-y').forEach(function (y) {
      function tog() { if (tl.classList.contains('hideyear')) { y.classList.toggle('open'); y.parentNode.classList.toggle('open'); } }
      y.addEventListener('click', tog);
      y.addEventListener('keydown', function (e) { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); tog(); } });
    });
    document.getElementById('tl-quiz').addEventListener('click', function () {
      setTab('quiz');
      startQuiz({ scope: tlFilter === 'all' ? SEC_IDS : [tlFilter], type: 'order', count: 10 });
    });
  }

  /* ───── 문제 생성기 ───── */
  function bankQ(q) {
    return { id: q.id, s: q.s, kind: q.t, q: q.q, src: q.src || '', o: q.o || null, a: q.a, ex: q.ex, label: q.t === 'ox' ? 'OX' : '기출형' };
  }
  function termQ(t) {
    var same = shuffle(TERMS.filter(function (x) { return x.s === t.s && x.t !== t.t; }));
    var others = shuffle(TERMS.filter(function (x) { return x.s !== t.s; }));
    var d = same.concat(others).slice(0, 3).map(function (x) { return x.t; });
    var o = shuffle(d.concat([t.t]));
    return { id: 'gt:' + termKey(t), s: t.s, kind: 'mc', label: '용어', q: '다음 설명에 해당하는 것은?', src: t.d, o: o, a: o.indexOf(t.t), ex: t.t + ' — ' + t.d };
  }
  function orderQ(scope) {
    var pool = shuffle(TIMELINE.filter(function (e) { return scope.indexOf(e.s) >= 0 && !e.noq; }));
    for (var size = 4; size >= 3; size--) {
      for (var tries = 0; tries < 30; tries++) {
        var chosen = [];
        // 실제 시험처럼 가까운 시기의 사건끼리 묶기: 기준 사건 ±span년 안에서 고름
        var seed = pick(pool), span = tries < 10 ? 6 : tries < 20 ? 12 : 40;
        var cand = shuffle(pool.filter(function (e) { return Math.abs(e.y - seed.y) <= span; }));
        for (var i = 0; i < cand.length && chosen.length < size; i++) {
          var ok = chosen.every(function (c) { return comparable(c, cand[i]); });
          if (ok) chosen.push(cand[i]);
        }
        if (chosen.length === size) {
          var sorted = chosen.slice().sort(function (a, b) { return less(a, b) ? -1 : 1; });
          var shown = shuffle(chosen);
          var mainS = sorted[0].s;
          var mixed = sorted.some(function (e) { return e.s !== mainS; });
          return {
            id: 'go:' + sorted.map(ym).join('|') + ':' + sorted.map(function (e) { return e.e.length; }).join(''),
            s: mainS, kind: 'order', label: '순서 배열', tagText: mixed ? '연표 종합' : '',
            q: '다음 사건을 일어난 순서대로 눌러 보세요.',
            items: shown.map(function (e) { return { e: e.e, y: ym(e), s: e.s }; }),
            a: sorted.map(function (e) { return shown.indexOf(e); }),
            ex: sorted.map(function (e) { return ym(e) + ' ' + e.e; }).join(' → ')
          };
        }
        if (pool.length < size) break;
      }
    }
    return null;
  }
  var ANCHORS = TIMELINE.filter(function (e) { return e.a; }).sort(function (a, b) { return less(a, b) ? -1 : 1; });
  function periodQ(scope) {
    var targets = shuffle(TIMELINE.filter(function (e) { return scope.indexOf(e.s) >= 0 && !e.noq && !e.a; }));
    for (var k = 0; k < targets.length; k++) {
      var t = targets[k];
      var idx = -1;
      for (var i = 0; i < ANCHORS.length - 1; i++) {
        if (less(ANCHORS[i], t) && less(t, ANCHORS[i + 1])) { idx = i; break; }
      }
      if (idx < 0 || ANCHORS.length < 5) continue;
      var lo = Math.max(0, idx - 3), hi = Math.min(idx, ANCHORS.length - 5);
      var j = lo + Math.floor(Math.random() * (hi - lo + 1));
      var win = ANCHORS.slice(j, j + 5);
      var ans = idx - j;
      return {
        id: 'gp:' + ym(t) + ':' + t.e.slice(0, 12), s: t.s, kind: 'mc', label: '시기 찾기',
        q: '다음 사건이 일어난 시기로 옳은 것은?', src: t.e, strip: win.map(function (w) { return { y: ym(w), e: w.e }; }),
        o: LETTERS.slice(), a: ans,
        ex: t.e + '(' + ym(t) + ')는 「' + win[ans].e + '(' + ym(win[ans]) + ')」와 「' + win[ans + 1].e + '(' + ym(win[ans + 1]) + ')」 사이예요.'
      };
    }
    return null;
  }

  function buildQuiz(opt) {
    var scope = opt.scope, n = opt.count, out = [], seen = {};
    function add(q) { if (q && !seen[q.id]) { seen[q.id] = 1; out.push(q); return true; } return false; }
    function gen(fn, k) { var guard = 0; var start = out.length; while (out.length - start < k && guard++ < k * 8) add(fn()); }
    var bank = shuffle(QUIZ.filter(function (q) { return scope.indexOf(q.s) >= 0; }));
    var terms = TERMS.filter(function (t) { return scope.indexOf(t.s) >= 0; });
    var termGen = function () { return terms.length ? termQ(pick(terms)) : null; };
    var orderGen = function () { return orderQ(scope); };
    var periodGen = function () { return periodQ(scope); };

    if (opt.type === 'mix') {
      var nb = Math.round(n * 0.5), nt = Math.round(n * 0.2), no = Math.round(n * 0.15);
      bank.slice(0, nb).forEach(function (q) { add(bankQ(q)); });
      gen(termGen, nt); gen(orderGen, no); gen(periodGen, n - out.length);
      gen(termGen, n - out.length);
      out = shuffle(out);
    } else if (opt.type === 'bank') {
      bank.filter(function (q) { return q.t === 'mc'; }).slice(0, n).forEach(function (q) { add(bankQ(q)); });
    } else if (opt.type === 'ox') {
      bank.filter(function (q) { return q.t === 'ox'; }).slice(0, n).forEach(function (q) { add(bankQ(q)); });
    } else if (opt.type === 'order') { gen(orderGen, n); }
    else if (opt.type === 'period') { gen(periodGen, n); }
    else if (opt.type === 'term') { gen(termGen, n); }
    return out.slice(0, n);
  }

  /* ───── 문제 설정 화면 ───── */
  var quizOpt = { scope: SEC_IDS.slice(), type: 'mix', count: 20 };
  var TYPES = [['mix', '섞어서'], ['bank', '기출형'], ['ox', 'OX'], ['order', '순서 배열'], ['period', '시기 찾기'], ['term', '용어']];

  function renderQuizSetup() {
    quizState = null;
    var wrong = Object.keys(S.wrong).map(function (k) { return S.wrong[k]; });
    var bankN = QUIZ.filter(function (q) { return quizOpt.scope.indexOf(q.s) >= 0; }).length;
    view.innerHTML =
      '<header class="stack" style="gap:6px"><h1 class="h2">문제 풀기</h1><p class="muted" style="margin:0;font-size:13.5px">직접 만든 기출형 ' + QUIZ.length + '문제 + 연표·용어로 만드는 무한 문제. 틀리면 오답 노트에 자동 저장돼요.</p></header>' +
      '<section class="card stack">' +
        '<div class="eyebrow">범위</div>' +
        '<div class="row" id="scope-chips"><button class="chip" data-all="1" aria-pressed="' + (quizOpt.scope.length === SEC_IDS.length) + '">전체</button>' +
          SEC_IDS.map(function (s) { return '<button class="chip" data-s="' + s + '" aria-pressed="' + (quizOpt.scope.indexOf(s) >= 0) + '">' + SECTIONS[s].no + '</button>'; }).join('') + '</div>' +
        '<div class="eyebrow">문제 유형</div>' +
        '<div class="row" id="type-chips">' + TYPES.map(function (t) { return '<button class="chip" data-t="' + t[0] + '" aria-pressed="' + (quizOpt.type === t[0]) + '">' + t[1] + '</button>'; }).join('') + '</div>' +
        '<div class="eyebrow">문제 수</div>' +
        '<div class="seg" id="count-seg">' + [10, 20, 30].map(function (c) { return '<button data-c="' + c + '" aria-pressed="' + (quizOpt.count === c) + '">' + c + '문제</button>'; }).join('') + '</div>' +
        '<button class="btn primary block" id="start"' + (quizOpt.scope.length ? '' : ' disabled') + '>시작하기</button>' +
        '<p class="muted" style="margin:0;font-size:12.5px">선택한 범위의 기출형 문제 ' + bankN + '개. 문제가 모자라면 있는 만큼만 나와요.</p>' +
      '</section>' +
      '<section class="card stack"><div class="row" style="justify-content:space-between"><div class="eyebrow">오답 노트 · ' + wrong.length + '문제</div>' +
        (wrong.length ? '<button class="btn primary" id="retry-wrong" style="min-height:36px">오답만 다시 풀기</button>' : '') + '</div>' +
        (wrong.length ? '<div class="wlist">' + wrong.slice(-30).reverse().map(function (q) {
          return '<div class="witem"><span class="row"><span class="tag">' + secLabel(q.s) + '</span><span class="acc">' + esc(q.label || '') + '</span></span>' +
            '<span>' + esc(q.src ? q.src.split('\n')[0] : q.q) + '</span><span class="a">정답: ' + esc(answerText(q)) + '</span></div>';
        }).join('') + '</div>' + (wrong.length > 30 ? '<p class="footnote">최근 30개만 보여요.</p>' : '') +
          '<button class="btn ghost" id="clear-wrong" style="justify-self:start">오답 노트 비우기</button>'
          : '<p class="empty" style="margin:0">아직 틀린 문제가 없어요. 문제를 풀면 틀린 것만 여기에 모여요.</p>') +
      '</section>';

    view.querySelectorAll('#scope-chips .chip').forEach(function (b) {
      b.addEventListener('click', function () {
        if (b.dataset.all) quizOpt.scope = quizOpt.scope.length === SEC_IDS.length ? [] : SEC_IDS.slice();
        else {
          var i = quizOpt.scope.indexOf(b.dataset.s);
          if (i >= 0) quizOpt.scope.splice(i, 1); else quizOpt.scope.push(b.dataset.s);
        }
        renderQuizSetup();
      });
    });
    view.querySelectorAll('#type-chips .chip').forEach(function (b) { b.addEventListener('click', function () { quizOpt.type = b.dataset.t; renderQuizSetup(); }); });
    view.querySelectorAll('#count-seg button').forEach(function (b) { b.addEventListener('click', function () { quizOpt.count = +b.dataset.c; renderQuizSetup(); }); });
    document.getElementById('start').addEventListener('click', function () { startQuiz(quizOpt); });
    var rw = document.getElementById('retry-wrong'); if (rw) rw.addEventListener('click', startWrong);
    var cw = document.getElementById('clear-wrong');
    if (cw) cw.addEventListener('click', function () {
      if (cw.dataset.armed) { S.wrong = {}; save(); renderQuizSetup(); }
      else { cw.dataset.armed = '1'; cw.textContent = '한 번 더 누르면 모두 지워져요'; }
    });
  }

  function answerText(q) {
    if (q.kind === 'ox') return q.a ? 'O' : 'X';
    if (q.kind === 'order') return q.ex;
    if (q.strip) return q.o[q.a] + ' — ' + q.ex;
    return (q.o && q.o[q.a]) || '';
  }

  function startQuiz(opt) {
    var qs = buildQuiz({ scope: opt.scope.slice(), type: opt.type, count: opt.count });
    if (!qs.length) {
      view.innerHTML = '<section class="card empty">이 범위에서는 이 유형의 문제를 만들 수 없어요. 범위를 넓히거나 다른 유형을 골라 주세요.<br><br><a class="btn" href="#quiz">설정으로</a></section>';
      return;
    }
    quizState = { qs: qs, i: 0, score: 0, answered: false, log: [], wrongMode: false };
    toQuizHash();
    renderQ();
  }
  function startWrong() {
    var qs = shuffle(Object.keys(S.wrong).map(function (k) { return S.wrong[k]; }));
    if (!qs.length) return;
    quizState = { qs: qs.slice(0, 30), i: 0, score: 0, answered: false, log: [], wrongMode: true };
    toQuizHash();
    renderQ();
  }

  /* ───── 문제 풀이 화면 ───── */
  function renderQ() {
    var st = quizState, q = st.qs[st.i];
    st.answered = false; st.picked = [];
    var head = '<div class="qhead"><span>' + (st.wrongMode ? '오답 다시 · ' : '') + (st.i + 1) + ' / ' + st.qs.length + '</span>' +
      '<span>맞힌 문제 ' + st.score + '</span></div><div class="progress"><i style="width:' + (st.i / st.qs.length * 100) + '%"></i></div>';
    var body = '<div class="row"><span class="tag">' + (q.tagText || secLabel(q.s)) + '</span><span class="acc">' + esc(q.label || '') + '</span></div>' +
      '<div class="qq">' + esc(q.q) + '</div>';
    if (q.strip) {
      body += '<div class="src" style="font-family:var(--f-body);font-weight:700">' + esc(q.src) + '</div><div class="strip">';
      q.strip.forEach(function (w, i) {
        body += '<span class="an"><b>' + w.y + '</b>' + esc(w.e) + '</span>';
        if (i < 4) body += '<span class="gap">' + LETTERS[i] + '</span>';
      });
      body += '</div>';
    } else if (q.src) body += '<div class="src">' + esc(q.src) + '</div>';

    if (q.kind === 'ox') {
      body += '<div class="ox"><button class="opt" data-v="1" aria-label="맞다">O</button><button class="opt" data-v="0" aria-label="틀리다">X</button></div>';
    } else if (q.kind === 'order') {
      body += '<div class="order-pick">' + q.items.map(function (it, i) {
        return '<button class="opt" data-i="' + i + '"><span class="num"></span><span>' + esc(it.e) + '</span></button>';
      }).join('') + '</div><div class="row"><button class="btn" id="o-reset">다시 고르기</button><button class="btn primary" id="o-check" disabled>확인</button></div>';
    } else {
      body += '<div class="opts">' + q.o.map(function (o, i) {
        return '<button class="opt" data-i="' + i + '"><span class="num">' + (q.strip ? '' : NUMS[i]) + '</span><span>' + esc(o) + '</span></button>';
      }).join('') + '</div>';
    }
    body += '<div id="fb"></div>';
    view.innerHTML = head + '<section class="card qcard">' + body + '</section>' +
      '<div class="row" style="justify-content:space-between"><a class="btn ghost" href="#quiz" id="quit">그만 풀기</a><button class="btn primary" id="next" hidden>다음 문제 →</button></div>';

    document.getElementById('quit').addEventListener('click', function (e) {
      e.preventDefault(); if (st.log.length) renderResult(); else renderQuizSetup();
    });
    document.getElementById('next').addEventListener('click', nextQ);

    if (q.kind === 'ox') {
      view.querySelectorAll('.ox .opt').forEach(function (b) {
        b.addEventListener('click', function () { answer((b.dataset.v === '1') === q.a, b); });
      });
    } else if (q.kind === 'order') {
      var btns = view.querySelectorAll('.order-pick .opt');
      var check = document.getElementById('o-check');
      btns.forEach(function (b) {
        b.addEventListener('click', function () {
          if (st.answered) return;
          var i = +b.dataset.i, at = st.picked.indexOf(i);
          if (at >= 0) st.picked.splice(at, 1); else st.picked.push(i);
          btns.forEach(function (x) {
            var p = st.picked.indexOf(+x.dataset.i);
            x.classList.toggle('picked', p >= 0);
            x.querySelector('.num').textContent = p >= 0 ? p + 1 : '';
          });
          check.disabled = st.picked.length !== q.items.length;
        });
      });
      document.getElementById('o-reset').addEventListener('click', function () {
        if (st.answered) return;
        st.picked = []; btns.forEach(function (x) { x.classList.remove('picked'); x.querySelector('.num').textContent = ''; }); check.disabled = true;
      });
      check.addEventListener('click', function () {
        if (st.answered) return;
        var ok = st.picked.every(function (v, i) { return v === q.a[i]; });
        btns.forEach(function (x) {
          var i = +x.dataset.i; x.disabled = true;
          var correctPos = q.a.indexOf(i);
          x.classList.add(st.picked.indexOf(i) === correctPos ? 'right' : 'wrong');
          x.querySelector('.num').textContent = correctPos + 1;
        });
        document.getElementById('o-reset').disabled = true; check.disabled = true;
        answer(ok, null);
      });
    } else {
      view.querySelectorAll('.opts .opt').forEach(function (b) {
        b.addEventListener('click', function () { answer(+b.dataset.i === q.a, b); });
      });
    }
  }

  function answer(ok, btn) {
    var st = quizState, q = st.qs[st.i];
    if (st.answered) return;
    st.answered = true;
    if (q.kind === 'ox') {
      view.querySelectorAll('.ox .opt').forEach(function (b) {
        b.disabled = true;
        if ((b.dataset.v === '1') === q.a) b.classList.add('right');
        else if (b === btn) b.classList.add('wrong');
      });
    } else if (q.kind !== 'order') {
      view.querySelectorAll('.opts .opt').forEach(function (b) {
        b.disabled = true;
        if (+b.dataset.i === q.a) b.classList.add('right');
        else if (b === btn) b.classList.add('wrong');
      });
    }
    if (ok) st.score++;
    st.log.push({ q: q, ok: ok });
    var stat = S.stats[q.s] || (S.stats[q.s] = { c: 0, n: 0 });
    stat.n++; if (ok) stat.c++;
    S.solved++;
    if (ok) { if (st.wrongMode) delete S.wrong[q.id]; }
    else S.wrong[q.id] = q;
    save();

    var exHtml = q.kind === 'order'
      ? '<ol class="order-ans">' + q.a.map(function (i) { var it = q.items[i]; return '<li><span class="y">' + it.y + '</span>' + esc(it.e) + '</li>'; }).join('') + '</ol>'
      : '<span>' + esc(q.ex || '') + '</span>';
    document.getElementById('fb').innerHTML = '<div class="fb ' + (ok ? 'ok' : 'no') + '"><b>' + (ok ? '정답!' : '아쉬워요 · 정답: ' + esc(q.kind === 'order' ? '아래 순서' : answerText(q).split(' — ')[0])) + '</b>' + exHtml +
      (!ok ? '<a href="#' + q.s + '" style="font-size:13px;color:var(--accent);font-weight:600">' + secLabel(q.s) + ' 개념 다시 보기 →</a>' : '') + '</div>';
    var nx = document.getElementById('next');
    nx.hidden = false; nx.textContent = st.i + 1 < st.qs.length ? '다음 문제 →' : '결과 보기';
    nx.focus({ preventScroll: true });
    document.getElementById('fb').scrollIntoView({ block: 'nearest', behavior: 'smooth' });
  }

  function nextQ() {
    var st = quizState;
    if (st.i + 1 < st.qs.length) { st.i++; renderQ(); window.scrollTo(0, 0); }
    else renderResult();
  }

  function renderResult() {
    var st = quizState, total = st.log.length;
    var bySec = {};
    st.log.forEach(function (l) { var b = bySec[l.q.s] || (bySec[l.q.s] = { c: 0, n: 0 }); b.n++; if (l.ok) b.c++; });
    var pct = total ? Math.round(st.score / total * 100) : 0;
    var msg = pct >= 90 ? '완벽해요. 이 감각 그대로 시험장에!' : pct >= 70 ? '좋아요. 틀린 문제만 한 번 더 보면 돼요.' : pct >= 50 ? '절반은 넘었어요. 약한 단원 개념을 다시 읽어 봐요.' : '개념부터 한 번 더! 빈칸 모드로 읽고 다시 도전해요.';
    var wrongs = st.log.filter(function (l) { return !l.ok; });
    view.innerHTML =
      '<section class="card stack" style="text-align:center;justify-items:center">' +
        '<div class="eyebrow">' + (st.wrongMode ? '오답 다시 풀기 결과' : '결과') + '</div>' +
        '<div class="result-num">' + st.score + '<span style="font-size:24px;color:var(--muted)"> / ' + total + '</span></div>' +
        '<div style="font-weight:600">' + msg + '</div>' +
      '</section>' +
      '<section class="card"><div class="eyebrow" style="margin-bottom:4px">단원별</div>' +
        Object.keys(bySec).sort().map(function (s) {
          var b = bySec[s], p = Math.round(b.c / b.n * 100);
          return '<a class="secrow" href="#' + s + '"><span class="tag">' + secLabel(s) + '</span><span class="t"><b>' + esc(SECTIONS[s].title) + '</b><span class="bar"><i style="width:' + p + '%"></i></span></span><span class="acc">' + b.c + '/' + b.n + '</span></a>';
        }).join('') + '</section>' +
      (wrongs.length ? '<section class="card"><div class="eyebrow" style="margin-bottom:4px">틀린 문제 ' + wrongs.length + '개 (오답 노트에 저장됨)</div><div class="wlist">' +
        wrongs.map(function (l) {
          var q = l.q;
          return '<div class="witem"><span class="row"><span class="tag">' + secLabel(q.s) + '</span><span class="acc">' + esc(q.label || '') + '</span></span><span>' + esc(q.src ? q.src.split('\n')[0] : q.q) + '</span><span class="a">정답: ' + esc(answerText(q)) + '</span></div>';
        }).join('') + '</div></section>' : '') +
      '<div class="quick">' +
        (Object.keys(S.wrong).length ? '<button class="btn primary" id="r-wrong">오답 다시 풀기<small style="color:inherit;opacity:.8">' + Object.keys(S.wrong).length + '문제</small></button>' : '') +
        '<button class="btn" id="r-again">같은 설정으로 새 문제</button>' +
        '<a class="btn" href="#home">홈으로</a>' +
      '</div>';
    var rw = document.getElementById('r-wrong'); if (rw) rw.addEventListener('click', startWrong);
    document.getElementById('r-again').addEventListener('click', function () { if (st.wrongMode) startWrong(); else startQuiz(quizOpt); });
    window.scrollTo(0, 0);
  }

  /* ───── 암기 카드 ───── */
  var cardScope = null, cardMode = 'd2t', cardOnlyWeak = false, cardDeck = [], cardI = 0;
  function makeDeck() {
    var scope = cardScope || SEC_IDS;
    var list = TERMS.filter(function (t) { return scope.indexOf(t.s) >= 0; });
    if (cardOnlyWeak) list = list.filter(function (t) { return (S.box[termKey(t)] || 0) < 2; });
    // 덜 외운 카드가 앞으로 오게
    cardDeck = shuffle(list).sort(function (a, b) { return (S.box[termKey(a)] || 0) - (S.box[termKey(b)] || 0); });
    cardI = 0;
  }
  function renderCards(keep) {
    if (!keep) makeDeck();
    var scope = cardScope || SEC_IDS;
    var all = TERMS.filter(function (t) { return scope.indexOf(t.s) >= 0; });
    var mastered = all.filter(function (t) { return (S.box[termKey(t)] || 0) >= 2; }).length;
    var c = cardDeck[cardI];
    var front = '', back = '';
    if (c) {
      if (cardMode === 'd2t') { front = '<div class="desc">' + esc(c.d) + '</div>'; back = '<div class="big">' + esc(c.t) + '</div>'; }
      else { front = '<div class="big">' + esc(c.t) + '</div>'; back = '<div class="desc">' + esc(c.d) + '</div>'; }
    }
    view.innerHTML =
      '<header class="stack" style="gap:6px"><h1 class="h2">암기 카드</h1><p class="muted" style="margin:0;font-size:13.5px">카드를 눌러 뒤집고, 알면 "알아요". 두 번 연속 알면 외운 카드로 쳐요.</p></header>' +
      '<div class="chips" id="c-scope"><button class="chip" data-all="1" aria-pressed="' + (!cardScope) + '">전체</button>' +
        SEC_IDS.map(function (s) { return '<button class="chip" data-s="' + s + '" aria-pressed="' + (!!cardScope && cardScope.length === 1 && cardScope[0] === s) + '">' + SECTIONS[s].no + '</button>'; }).join('') + '</div>' +
      '<div class="toolbar" style="position:static;border:0;padding:0">' +
        '<div class="seg" style="flex:1;min-width:220px"><button data-m="d2t" aria-pressed="' + (cardMode === 'd2t') + '">설명 → 용어</button><button data-m="t2d" aria-pressed="' + (cardMode === 't2d') + '">용어 → 설명</button></div>' +
        '<label class="switch" for="weak-sw"><input type="checkbox" id="weak-sw"' + (cardOnlyWeak ? ' checked' : '') + '>못 외운 것만</label>' +
      '</div>' +
      '<div class="row" style="justify-content:space-between"><span class="acc">' + (c ? (cardI + 1) + ' / ' + cardDeck.length : '0 / 0') + '</span><span class="acc">외운 카드 ' + mastered + ' / ' + all.length + '</span></div>' +
      (c ?
        '<div class="flash" id="flash"><div class="flash-in" tabindex="0" role="button" aria-label="카드 뒤집기">' +
          '<div class="face front"><div><div class="row" style="justify-content:center;margin-bottom:10px"><span class="tag">' + secLabel(c.s) + '</span></div>' + front + '</div><span class="hint">눌러서 뒤집기</span></div>' +
          '<div class="face back"><div>' + back + '</div><span class="hint">' + (cardMode === 'd2t' ? esc(c.d.slice(0, 40)) + (c.d.length > 40 ? '…' : '') : esc(c.t)) + '</span></div>' +
        '</div></div>' +
        '<div class="flash-btns"><button class="btn dunno" id="c-no">몰라요</button><button class="btn know" id="c-yes">알아요</button></div>'
        : '<section class="card empty">' + (cardOnlyWeak ? '이 범위 카드를 모두 외웠어요! "못 외운 것만"을 끄면 전체를 다시 볼 수 있어요.' : '카드가 없어요.') + '</section>') +
      '<button class="btn ghost" id="c-reset" style="justify-self:start">이 범위 외운 기록 초기화</button>';

    view.querySelectorAll('#c-scope .chip').forEach(function (b) {
      b.addEventListener('click', function () { cardScope = b.dataset.all ? null : [b.dataset.s]; renderCards(); });
    });
    view.querySelectorAll('[data-m]').forEach(function (b) { b.addEventListener('click', function () { cardMode = b.dataset.m; renderCards(true); }); });
    document.getElementById('weak-sw').addEventListener('change', function (e) { cardOnlyWeak = e.target.checked; renderCards(); });
    var reset = document.getElementById('c-reset');
    reset.addEventListener('click', function () {
      if (reset.dataset.armed) { all.forEach(function (t) { delete S.box[termKey(t)]; }); save(); renderCards(); }
      else { reset.dataset.armed = '1'; reset.textContent = '한 번 더 누르면 초기화돼요'; }
    });
    if (!c) return;
    var flash = document.getElementById('flash'), inner = flash.querySelector('.flash-in');
    function flip() { flash.classList.toggle('flip'); }
    inner.addEventListener('click', flip);
    inner.addEventListener('keydown', function (e) { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); flip(); } });
    function grade(know) {
      var k = termKey(c);
      S.box[k] = know ? Math.min((S.box[k] || 0) + 1, 3) : 0;
      save();
      if (!know) {
        // 모르는 카드는 몇 장 뒤에 다시 등장
        var again = cardDeck.splice(cardI, 1)[0];
        cardDeck.splice(Math.min(cardI + 3, cardDeck.length), 0, again);
      } else cardI++;
      if (cardI >= cardDeck.length) { makeDeck(); }
      renderCards(true);
    }
    document.getElementById('c-no').addEventListener('click', function () { grade(false); });
    document.getElementById('c-yes').addEventListener('click', function () { grade(true); });
  }

  /* ───── 키보드 단축키(문제 풀이) ───── */
  document.addEventListener('keydown', function (e) {
    if (!quizState || !view.querySelector('.qcard') || e.target.tagName === 'INPUT') return;
    var q = quizState.qs[quizState.i];
    if (!q) return;
    if (!quizState.answered && q.kind !== 'order') {
      if (q.kind === 'ox') {
        if (e.key === 'o' || e.key === 'O' || e.key === '1') { var o = view.querySelector('.ox .opt[data-v="1"]'); if (o) o.click(); }
        if (e.key === 'x' || e.key === 'X' || e.key === '2') { var x = view.querySelector('.ox .opt[data-v="0"]'); if (x) x.click(); }
      } else if (/^[1-4]$/.test(e.key)) {
        var b = view.querySelector('.opts .opt[data-i="' + (+e.key - 1) + '"]'); if (b) b.click();
      }
    } else if (quizState.answered && e.key === 'Enter' && e.target.id !== 'next') {
      var n = document.getElementById('next'); if (n && !n.hidden) n.click();
    }
  });

  route();
})();
