(function(){
  "use strict";

  var uname = window.location.pathname.split('/').filter(Boolean).pop();
  var apiUrl = '/api/sub/' + encodeURIComponent(uname);

  var CONFIG_LINKS = [], CONFIG_NAMES = [], CONFIG_PROTOS = [];
  var SUB_URL = '', SUB_ID = '', IS_ACTIVE = false;
  var TOTAL_BYTES = 0, USED_BYTES = 0, DOWNLOAD_BYTES = 0, UPLOAD_BYTES = 0;
  var EXPIRE_TS = null, EXPIRE_DAYS = null;
  var countdownInterval = null;

  var BYTES_IN_GB = 1024*1024*1024;

  var i18n = {
    fa: {
      title:"ایستگاه فضایی اشتراک", statusOnline:"آنلاین", followUs:"ما را دنبال کنید",
      subLabel:"لینک سابسکریپشن", copyLink:"کپی لینک", accountTitle:"اطلاعات اکانت",
      sId:"شناسه اشتراک", dlExact:"دانلود", ulExact:"آپلود", totalLimit:"حجم سرویس",
      usernameLbl:"نام کاربری", connLimit:"اتصال مجاز", connActive:"اتصال فعلی",
      usage:"مصرف شده", remain:"حجم باقی‌مانده", used:"حجم مصرفی", unlimited:"∞ نامحدود",
      noExpire:"بدون تاریخ انقضا", min:"دقیقه", hour:"ساعت", day:"روز",
      yourConfigs:"کانفیگ‌های شما", copyAllShort:"کپی همه کانفیگ‌ها", copySubText:"کپی متن ساب", all:"همه", search:"جستجو در کانفیگ‌ها...",
      qrHint:"اسکن کنید تا کانفیگ اضافه شود", close:"بستن",
      toastSub:"لینک ساب کپی شد", toastAll:"همه کانفیگ‌ها کپی شدند", toastSubText:"متن ساب کپی شد",
      toastCfg:"کانفیگ کپی شد", toastId:"شناسه کپی شد",
      toastErr:"خطا در کپی", connectedActive:"متصل و فعال",
      secureReady:"اتصال امن شما آماده استفاده است", quickConnect:"اتصال سریع",
      quickImport:"اتصال سریع", qcTitle:"اتصال سریع با یک کلیک", themeLight:"روز", themeDark:"شب", usageChart:"نمودار مصرف",
      noConfig:"کانفیگی موجود نیست", gb:"گیگابایت", dailyAvg:"میانگین روزانه",
      navHome:"داشبورد", navConfigs:"کانفیگ‌ها"
    },
    en: {
      title:"Space Subscription Station", statusOnline:"Online", followUs:"Follow us",
      subLabel:"Subscription Link", copyLink:"Copy Link", accountTitle:"Account Info",
      sId:"Subscription ID", dlExact:"Download", ulExact:"Upload", totalLimit:"Service Limit",
      usernameLbl:"Username", connLimit:"Allowed Connections", connActive:"Active Connections",
      usage:"Used", remain:"Remaining", used:"Used", unlimited:"∞ Unlimited",
      noExpire:"No expiration date", min:"Min", hour:"Hour", day:"Day",
      yourConfigs:"Your Configs", copyAllShort:"Copy All Configs", copySubText:"Copy Sub Text", all:"All", search:"Search configs...",
      qrHint:"Scan to add the config", close:"Close",
      toastSub:"Subscription link copied", toastAll:"All configs copied", toastSubText:"Sub text copied",
      toastCfg:"Config copied", toastId:"ID copied",
      toastErr:"Copy failed", connectedActive:"Connected & Active",
      secureReady:"Your secure connection is ready", quickConnect:"Quick Connect",
      quickImport:"Quick Connect", qcTitle:"One-Tap Quick Connect", themeLight:"Light", themeDark:"Dark", usageChart:"Usage Chart",
      noConfig:"No configs", gb:"GB", dailyAvg:"Daily Avg",
      navHome:"Home", navConfigs:"Configs"
    }
  };

  var lang = localStorage.getItem("white-lang") || "fa";
  var root = document.documentElement;

  function tr(k){ return (i18n[lang] && i18n[lang][k]) || k; }
  function isFa(){ return lang === "fa"; }
  function faDigits(s){ return isFa() ? String(s).replace(/\d/g, function(d){ return '۰۱۲۳۴۵۶۷۸۹'[d]; }) : String(s); }
  function fmtBytes(b){
    if (!b || b < 1) return isFa() ? '۰ B' : '0 B';
    var u = ['B','KB','MB','GB','TB'], i = 0;
    while (b >= 1024 && i < u.length-1){ b /= 1024; i++; }
    var num = b.toFixed(i ? 2 : 0);
    return faDigits(num) + ' ' + u[i];
  }
  function esc(s){
    return String(s == null ? '' : s).replace(/[&<>"']/g, function(c){
      return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c];
    });
  }
  function pad2(v){ return String(v).padStart(2,'0'); }
  function gregorianToJalali(gy, gm, gd){
    var gDm = [0,31,59,90,120,151,181,212,243,273,304,334];
    var jy = gy <= 1600 ? 0 : 979; gy -= gy <= 1600 ? 621 : 1600;
    var gy2 = gm > 2 ? gy + 1 : gy;
    var days = 365*gy + Math.floor((gy2+3)/4) - Math.floor((gy2+99)/100) + Math.floor((gy2+399)/400) - 80 + gd + gDm[gm-1];
    jy += 33*Math.floor(days/12053); days %= 12053;
    jy += 4*Math.floor(days/1461); days %= 1461;
    if (days > 365){ jy += Math.floor((days-1)/365); days = (days-1)%365; }
    var jm = days < 186 ? 1 + Math.floor(days/31) : 7 + Math.floor((days-186)/30);
    var jd = 1 + (days < 186 ? days%31 : (days-186)%30);
    return [jy,jm,jd];
  }
  function formatDateOnly(d){
    if (!d || isNaN(d.getTime())) return '';
    if (isFa()){
      var j = gregorianToJalali(d.getFullYear(), d.getMonth()+1, d.getDate());
      return faDigits(j[0] + '/' + pad2(j[1]) + '/' + pad2(j[2]));
    }
    return d.getFullYear() + '/' + pad2(d.getMonth()+1) + '/' + pad2(d.getDate());
  }

  function applyLang(l){
    lang = l;
    localStorage.setItem("white-lang", l);
    var t = i18n[l] || i18n.fa;
    document.documentElement.lang = l;
    document.documentElement.dir = l === "fa" ? "rtl" : "ltr";
    var lab = document.getElementById("langLabel");
    if (lab) lab.textContent = l === "fa" ? "FA" : "EN";
    updateLangMenu();
    updateThemeMenu();
    document.querySelectorAll("[data-i18n]").forEach(function(el){
      var k = el.getAttribute("data-i18n");
      if (t[k]) el.textContent = t[k];
    });
    renderConfigs();
    if (typeof qcRefresh === "function") qcRefresh();
    if (TOTAL_BYTES || USED_BYTES) applyTraffic();
    if (EXPIRE_TS !== null) renderCountdown();
  }

  var savedTheme = localStorage.getItem("white-theme-v2");
  if (savedTheme === "dark"){ if (root.getAttribute("data-theme") !== "dark") root.setAttribute("data-theme","dark"); }
  else if (root.hasAttribute("data-theme")){ root.removeAttribute("data-theme"); }

  var themeSwitchTimer = null;

  function applyThemeChange(){
    if (root.getAttribute("data-theme") === "dark"){
      root.removeAttribute("data-theme");
      localStorage.setItem("white-theme-v2","light");
      var meta = document.querySelector('meta[name="theme-color"]');
      if (meta) meta.content = "#eef4fb";
    } else {
      root.setAttribute("data-theme","dark");
      localStorage.setItem("white-theme-v2","dark");
      var meta2 = document.querySelector('meta[name="theme-color"]');
      if (meta2) meta2.content = "#01091a";
    }
  }

  function toggleTheme(){
    var reduced = window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches;
    if (reduced){
      applyThemeChange();
      return;
    }

    // روش ۱: View Transitions API — crossfade واقعی
    if (typeof document.startViewTransition === "function"){
      try {
        document.startViewTransition(function(){ applyThemeChange(); });
        return;
      } catch(e){ /* ادامه به روش ۲ */ }
    }

    // روش ۲: fallback با کلاس موقت
    root.classList.add("theme-transition");
    applyThemeChange();
    clearTimeout(themeSwitchTimer);
    themeSwitchTimer = setTimeout(function(){
      root.classList.remove("theme-transition");
    }, 460);
  }

  var toastTimer;
  function showToast(msg, err){
    var old = document.querySelector(".toast"); if (old) old.remove();
    var t = document.createElement("div");
    t.className = "toast" + (err ? " error" : "");
    t.textContent = msg;
    document.body.appendChild(t);
    clearTimeout(toastTimer);
    toastTimer = setTimeout(function(){
      t.classList.add("hide");
      setTimeout(function(){ t.remove(); }, 300);
    }, 2200);
  }

  function legacyCopy(text){
    var ok = false, ta = document.createElement("textarea");
    ta.value = text;
    ta.setAttribute("readonly", "");
    ta.style.cssText = "position:fixed;top:0;left:0;width:1px;height:1px;padding:0;border:0;opacity:0;font-size:16px;";
    document.body.appendChild(ta);
    try {
      ta.focus({ preventScroll: true });
      ta.select();
      ta.setSelectionRange(0, text.length);
      ok = document.execCommand("copy");
    } catch(e){ ok = false; }
    document.body.removeChild(ta);
    return ok;
  }
  /* کپی مطمئن: اول روش همزمان (روی http و وب‌ویو هم کار می‌کند)، بعد Clipboard API، آخر هم نمایش لینک برای کپی دستی */
  function copyText(text){
    text = String(text == null ? "" : text);
    if (!text) return Promise.reject(new Error("empty"));
    if (legacyCopy(text)) return Promise.resolve();
    function manual(){
      try { window.prompt(isFa() ? "لینک را کپی کنید:" : "Copy the link:", text); } catch(e){}
      return Promise.resolve();
    }
    if (navigator.clipboard && navigator.clipboard.writeText){
      return navigator.clipboard.writeText(text).catch(manual);
    }
    return manual();
  }
  function getSubUrl(){
    if (SUB_URL) return SUB_URL;
    var cur = window.location.href.split('?')[0].split('#')[0];
    if (cur && (cur.indexOf('/link/') !== -1 || cur.indexOf('/sub/') !== -1)) return cur;
    return window.location.origin + '/link/' + (SUB_ID || uname);
  }

  function protoOf(link){
    var m = String(link||'').match(/^[a-z0-9]+:\/\//i);
    var proto = m ? m[0].replace('://','').toUpperCase() : 'CONFIG';
    var type = (String(link).match(/type=([a-z0-9-]+)/i) || [])[1] || '';
    var sec = (String(link).match(/security=([a-z0-9-]+)/i) || [])[1] || '';
    var out = proto;
    if (sec) out += ' • ' + sec.toUpperCase();
    else if (type) out += ' • ' + type.toUpperCase();
    return out;
  }

  function dataProtoKey(link){
    var l = String(link || '').toLowerCase();
    var tags = [];
    if (/^(vless|vmess|trojan|ss|socks|hysteria2|tuic):\/\//.test(l)) tags.push('proxy');
    if (l.indexOf('wireguard://') === 0 || l.indexOf('wg://') === 0 || (l.indexOf('peer.') === 0 && l.indexOf('amnezia') >= 0)) tags.push('wireguard');
    if (l.indexOf('security=reality') >= 0) tags.push('reality');
    else if (l.indexOf('security=tls') >= 0 || l.indexOf('&tls=1') >= 0) tags.push('tls');
    return tags.join(' ') || 'other';
  }

  var ACTIVE_FILTER = "all";
  function applyFilters(){
    var input = document.getElementById("searchInput");
    var q = (input ? input.value : "").trim().toLowerCase();
    document.querySelectorAll(".config-item").forEach(function(item){
      var index = parseInt(item.getAttribute("data-index"), 10);
      var tags = (item.getAttribute("data-tags") || "").toLowerCase();
      var name = ((item.querySelector(".config-item-name") || {}).textContent || "").toLowerCase();
      var proto = ((item.querySelector(".config-item-proto") || {}).textContent || "").toLowerCase();
      var link = (CONFIG_LINKS[index] || "").toLowerCase();
      var okFilter = ACTIVE_FILTER === "all" || tags.indexOf(ACTIVE_FILTER) >= 0;
      var okSearch = !q || name.indexOf(q) >= 0 || proto.indexOf(q) >= 0 || link.indexOf(q) >= 0;
      item.style.display = (okFilter && okSearch) ? "" : "none";
    });
  }

  var cfgEnterTimer = null;
  function switchTab(target){
    closeLangMenu(false);
    document.querySelectorAll('.nav-item').forEach(function(n){ n.classList.remove('active'); });
    document.querySelectorAll('.tab-content').forEach(function(c){ c.classList.remove('active'); });
    var btn = document.querySelector('.nav-item[data-target="' + target + '"]');
    if (btn) btn.classList.add('active');
    var tab = document.getElementById(target);
    if (tab) tab.classList.add('active');
    document.body.setAttribute('data-tab', target);
    window.scrollTo({ top: 0, behavior: 'smooth' });
    if (target === 'tab-configs' && tab && !(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches)) {
      tab.classList.remove('cfg-enter');
      void tab.offsetWidth;
      tab.classList.add('cfg-enter');
      clearTimeout(cfgEnterTimer);
      cfgEnterTimer = setTimeout(function(){ tab.classList.remove('cfg-enter'); }, 2400);
    }
    if (target === 'tab-home') {
      var subCard = document.querySelector('.merged-block .sub-card');
      if (subCard && !(window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches)) {
        subCard.classList.remove('sub-enter-show');
        subCard.classList.add('sub-enter-ready');
        void subCard.offsetWidth;
        requestAnimationFrame(function(){ subCard.classList.add('sub-enter-show'); });
      }
    }
  }

  function renderConfigs(){
    var list = document.getElementById("configList");
    var cnt = document.getElementById("cfgCount");
    if (!list) return;
    if (!CONFIG_LINKS.length){
      list.innerHTML = '<div class="empty-configs">' + esc(tr('noConfig')) + '</div>';
      if (cnt) cnt.textContent = '0';
      return;
    }
    list.innerHTML = '';
    CONFIG_LINKS.forEach(function(link, i){
      var item = document.createElement('div');
      item.className = 'config-item';
      item.setAttribute('data-index', i);
      item.setAttribute('data-tags', dataProtoKey(link));
      var name = CONFIG_NAMES[i] || ('Config ' + (i+1));
      var proto = CONFIG_PROTOS[i] || protoOf(link);
      item.innerHTML =
        '<div class="config-item-body">' +
          '<span class="config-item-name">' + esc(name) + '</span>' +
          '<span class="config-item-proto">' + esc(proto) + '</span>' +
        '</div>' +
        '<div class="config-item-actions">' +
          '<button class="mini-btn copy-cfg" type="button" data-index="' + i + '" title="Copy">' +
            '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 01-2-2V4a2 2 0 012-2h9a2 2 0 012 2v1"/></svg>' +
          '</button>' +
          '<button class="mini-btn qr-btn" type="button" data-index="' + i + '" title="QR">' +
            '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7"/><rect x="14" y="3" width="7" height="7"/><rect x="3" y="14" width="7" height="7"/><path d="M14 14h3v3h-3zM20 14v.01M14 20h.01M20 20v.01"/></svg>' +
          '</button>' +
        '</div>';
      list.appendChild(item);
    });
    if (cnt) cnt.textContent = faDigits(CONFIG_LINKS.length);
    bindConfigButtons();
    applyFilters();
  }

  function bindConfigButtons(){
    document.querySelectorAll(".copy-cfg").forEach(function(btn){
      if (btn._bound) return; btn._bound = true;
      btn.addEventListener("click", function(){
        var i = parseInt(btn.getAttribute("data-index"), 10);
        var link = CONFIG_LINKS[i];
        if (!link) return;
        copyText(link).then(function(){
          btn.classList.add("copied");
          var name = (CONFIG_NAMES[i] || "").substring(0, 28);
          showToast(tr("toastCfg") + (name ? ": " + name : ""));
          setTimeout(function(){ btn.classList.remove("copied"); }, 1400);
        }).catch(function(){ showToast(tr("toastErr"), true); });
      });
    });
    document.querySelectorAll(".qr-btn").forEach(function(btn){
      if (btn._bound) return; btn._bound = true;
      btn.addEventListener("click", function(){
        var i = parseInt(btn.getAttribute("data-index"), 10);
        showQR(CONFIG_LINKS[i], CONFIG_NAMES[i] || "config");
        btn.classList.add("qr-active");
        setTimeout(function(){ btn.classList.remove("qr-active"); }, 600);
      });
    });
  }

  function initVisibilityPause(){
    if (!("IntersectionObserver" in window)) return;
    var targets = document.querySelectorAll(".merged-block, .hero-bar, .configs-box, .team-name");
    if (!targets.length) return;
    var io = new IntersectionObserver(function(entries){
      for (var i = 0; i < entries.length; i++){
        var en = entries[i];
        if (en.isIntersecting) en.target.classList.remove("perf-paused");
        else en.target.classList.add("perf-paused");
      }
    }, { rootMargin: "150px 0px" });
    targets.forEach(function(el){ io.observe(el); });
  }

  function initPageEnter(){
    if (window.matchMedia && window.matchMedia('(prefers-reduced-motion: reduce)').matches) return;

    var plan = [
      { sel: ".hero-bar",       delay: 40  },
      { sel: ".badge-row",      delay: 140 },
      { sel: ".merged-block",   delay: 220 },
      { sel: ".import-card",    delay: 380 },
      { sel: ".account-card",   delay: 460 },
      { sel: ".expire-card",    delay: 540 },
      { sel: ".social-section", delay: 620 },
      { sel: ".team-brand",     delay: 700 },
      { sel: ".site-foot",      delay: 760 }
    ];

    var maxEnd = 0;
    plan.forEach(function(p){
      var el = document.querySelector(p.sel);
      if (!el) return;
      el.classList.add("page-enter");

      var startAt = p.delay;
      maxEnd = Math.max(maxEnd, startAt + 950);

      setTimeout(function(){
        if (!el.isConnected) return;
        requestAnimationFrame(function(){
          el.classList.add("page-enter-show");
        });
      }, startAt);
    });

    setTimeout(function(){
      document.querySelectorAll(".page-enter").forEach(function(el){
        el.classList.add("page-enter-done");
      });
    }, maxEnd);
  }

  /* آخرین نقطه‌ی لمس/کلیک: QR مودال از همان‌جا رشد می‌کند */
  var lastPtr = null, qrCurrent = null;
  document.addEventListener("pointerdown", function(e){ lastPtr = { x: e.clientX, y: e.clientY }; }, true);
  function setQrOrigin(){
    var modal = document.getElementById("qrModal");
    var box = modal && modal.querySelector(".modal-box");
    if (!box) return;
    var w = box.offsetWidth, h = box.offsetHeight;
    var left = (window.innerWidth - w) / 2, top = (window.innerHeight - h) / 2;
    var ox = w / 2, oy = h / 2;
    if (lastPtr){
      ox = Math.max(-120, Math.min(w + 120, lastPtr.x - left));
      oy = Math.max(-120, Math.min(h + 120, lastPtr.y - top));
    }
    box.style.setProperty("--qr-ox", Math.round(ox) + "px");
    box.style.setProperty("--qr-oy", Math.round(oy) + "px");
  }
  function showQR(data, title){
    if (!data) return showToast(tr("toastErr"), true);
    clearTimeout(qrCloseTimer);
    document.getElementById("qrModal").classList.remove("closing");
    setQrOrigin();
    qrCurrent = { data: data, title: title };
    document.getElementById("qrLinkUrl").textContent = data;
    document.getElementById("qrName").textContent = title || "config";
    var box = document.getElementById("qrContainer");
    box.innerHTML = '<div class="qr-loader"></div>';
    if (typeof QRCode !== "undefined"){
      try {
        QRCode.toCanvas(data, { width: 180, margin: 1 }, function(err, canvas){
          if (!err && canvas){ box.innerHTML = ''; canvas.classList.add('qr-reveal'); box.appendChild(canvas); return; }
          showServerQr(box, data);
        });
      } catch(e){ showServerQr(box, data); }
    } else {
      showServerQr(box, data);
    }
    document.getElementById("qrModal").classList.add("show");
  }
  function showServerQr(box, data){
    box.innerHTML = '';
    var img = document.createElement('img');
    img.alt = 'QR';
    img.className = 'qr-reveal';
    img.style.cssText = 'width:180px;height:180px;background:#fff;border-radius:8px;padding:6px';
    img.src = 'https://api.qrserver.com/v1/create-qr-code/?size=280x280&margin=1&data=' + encodeURIComponent(data);
    img.onerror = function(){ box.innerHTML = '<div style="color:#888;padding:20px;font-size:12px">QR Error</div>'; };
    box.appendChild(img);
  }
  var qrCloseTimer = null;
  function closeQR(){
    var m = document.getElementById("qrModal");
    if (!m || !m.classList.contains("show") || m.classList.contains("closing")) return;
    var reduced = window.matchMedia && window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (reduced){ m.classList.remove("show"); return; }
    /* اول محو شدن نرم، بعد مخفی کردن کامل */
    m.classList.add("closing");
    clearTimeout(qrCloseTimer);
    qrCloseTimer = setTimeout(function(){ m.classList.remove("show", "closing"); }, 880);
  }

  /* ---------- Quick Connect ---------- */
  var QC_NAME = "WHITE";
  function qcPlay(q){ return "https://play.google.com/store/search?c=apps&q=" + encodeURIComponent(q); }
  function qcIos(q){ return "itms-apps://search.itunes.apple.com/WebObjects/MZSearch.woa/wa/search?media=software&term=" + encodeURIComponent(q); }
  var uV2rayng  = function(s,n){ return "v2rayng://install-sub?url=" + encodeURIComponent(s) + "&name=" + encodeURIComponent(n); };
  var uHapp     = function(s){ return "happ://add/" + s; };
  var uHiddify  = function(s,n){ return "hiddify://import/" + s + "#" + encodeURIComponent(n); };
  var uV2box    = function(s,n){ return "v2box://install-sub?url=" + encodeURIComponent(s) + "&name=" + encodeURIComponent(n); };
  var uStreis   = function(s,n){ return "streisand://import/" + s + "#" + encodeURIComponent(n); };
  var uShadow   = function(s){ return "sub://" + btoa(s); };

  /* i: آیکن، g: گرادیان رنگ، u: دیپ‌لینک، dl: لینک دانلود (در صورت نیاز لینک دقیق استور را جایگزین کنید) */
  var QC_APPS = {
    android: [
      { n:"v2rayNG", i:"fa-shield-halved",        g:["#3b82f6","#1d4ed8"], u:uV2rayng, dl:"https://play.google.com/store/apps/details?id=com.v2ray.ang" },
      { n:"Happ",    i:"fa-bolt",                 g:["#a855f7","#6d28d9"], u:uHapp,    dl:"https://play.google.com/store/apps/details?id=com.happproxy" },
      { n:"Hiddify", i:"fa-wand-magic-sparkles",  g:["#f59e0b","#d97706"], u:uHiddify, dl:"https://play.google.com/store/apps/details?id=app.hiddify.com" },
      { n:"V2Box",   i:"fa-cube",                 g:["#14b8a6","#0f766e"], u:uV2box,   dl:"https://play.google.com/store/apps/details?id=dev.hexasoftware.v2box" }
    ],
    ios: [
      { n:"Streisand",    i:"fa-wind",   g:["#38bdf8","#0369a1"], u:uStreis, dl:"https://apps.apple.com/app/id6450534064" },
      { n:"Happ",         i:"fa-bolt",   g:["#a855f7","#6d28d9"], u:uHapp,   dl:"https://apps.apple.com/app/id6504287215" },
      { n:"V2Box",        i:"fa-cube",   g:["#14b8a6","#0f766e"], u:uV2box,  dl:"https://apps.apple.com/app/id6446814690" },
      { n:"Shadowrocket", i:"fa-rocket", g:["#fb7185","#be123c"], u:uShadow, dl:"https://apps.apple.com/app/id932747118" }
    ],
    windows: [
      { n:"Hiddify", i:"fa-wand-magic-sparkles", g:["#f59e0b","#d97706"], u:uHiddify, dl:"https://github.com/hiddify/hiddify-app/releases/latest" },
      { n:"Happ",    i:"fa-bolt",                g:["#a855f7","#6d28d9"], u:uHapp,    dl:"https://www.happ.su/main" },
      { n:"v2rayN",  i:"fa-paper-plane",         g:["#22c55e","#15803d"], u:null,     dl:"https://github.com/2dust/v2rayN/releases/latest" }
    ]
  };

  var QC_T = {
    fa: {
      sub:"برنامه‌ات رو انتخاب کن؛ اشتراک خودکار اضافه می‌شه",
      steps:[["نصب برنامه","اگه داری، رد شو"],["روی برنامه بزن","ساب خودکار اضافه می‌شه"],["وصل شو","یه کانفیگ بزن و اتصال"]],
      rec:"پیشنهادی", copy:"کپی لینک اشتراک", qr:"کد QR", guideTitle:"راهنمای کامل", dlAria:"دانلود برنامه",
      desc:{ v2rayNG:"محبوب‌ترین و پایدار", Happ:"ساده، سریع و خوش‌دست", Hiddify:"هوشمند؛ بهترین کانفیگ رو خودش انتخاب می‌کنه",
             V2Box:"سبک و روان", Streisand:"رایگان و سریع", Shadowrocket:"پولی ولی حرفه‌ای", v2rayN:"نسخه‌ی دسکتاپ؛ افزودن دستی" },
      missT:"برنامه باز نشد؟",
      missP:"احتمالاً {app} نصب نیست. اول نصبش کن و بعد دوباره روی کارتش بزن. لینک اشتراکت هم کپی شده و آماده‌ی استفاده‌ست.",
      manT:"لینک کپی شد، حالا داخل برنامه اضافه کن",
      manP:"{app} افزودن خودکار نداره. برنامه رو باز کن، بخش اشتراک (Subscription) رو بزن و لینک رو Paste کن. مراحل دقیقش پایین توی راهنما هست.",
      hDl:"نصب {app}", hCopy:"کپی دوباره‌ی لینک", hGuide:"روش دستی",
      g:{
        s1:"اتصال در ۴ قدم",
        s1l:["سیستم‌عاملت رو از بالا انتخاب کن (خودکار تشخیص داده می‌شه).",
             "برنامه رو نداری؟ روی آیکن <b>دانلود</b> کنار اسمش بزن و نصبش کن.",
             "روی <b>کارت برنامه</b> بزن. لینک اشتراک کپی می‌شه و برنامه با ساب تو باز می‌شه.",
             "اگه برنامه پرسید «افزودن اشتراک؟» تأیید کن، یه کانفیگ انتخاب کن و دکمه‌ی <b>اتصال</b> رو بزن."],
        s2:"برنامه باز نشد؟",
        s2l:["مطمئن شو برنامه نصب شده و حداقل یک بار بازش کردی.",
             "اگه این صفحه رو داخل تلگرام یا اینستاگرام باز کردی، از منوی بالا <b>باز کردن در مرورگر</b> (Open in Browser) رو بزن و دوباره امتحان کن.",
             "برنامه‌ی دیگه‌ای از لیست رو امتحان کن؛ همه‌شون همین یک ساب رو می‌خونن.",
             "اگه باز هم نشد، از روش دستی پایین استفاده کن."],
        s3:"روش دستی (کپی لینک)",
        s3p:"اول دکمه‌ی «کپی لینک اشتراک» رو بزن، بعد:",
        man:{
          android:["برنامه رو باز کن و روی <b>+</b> بزن.","گزینه‌ی <b>Subscription</b> یا <b>Import from clipboard</b> رو انتخاب کن.","لینک کپی‌شده رو Paste کن و تأیید بزن.","اشتراک رو <b>Update</b> کن و یه کانفیگ رو برای اتصال انتخاب کن."],
          ios:["برنامه رو باز کن و روی <b>+</b> بالای صفحه بزن.","گزینه‌ی <b>Add from clipboard</b> یا <b>Subscribe</b> رو بزن (توی Shadowrocket: نوع رو روی Subscribe بذار و لینک رو Paste کن).","ذخیره کن و اشتراک رو به‌روزرسانی کن.","یه کانفیگ انتخاب کن و سوییچ اتصال رو روشن کن. بار اول iOS اجازه‌ی افزودن VPN می‌خواد؛ <b>Allow</b> رو بزن."],
          windows:["برنامه رو باز کن. توی v2rayN منوی <b>Subscription ← Subscription settings</b>، توی Hiddify هم <b>Add profile</b>.","روی <b>Add</b> بزن و لینک کپی‌شده رو با <code>Ctrl+V</code> بذار.","تأیید کن و <b>Update subscription</b> رو بزن.","یه کانفیگ انتخاب کن و حالت <b>System Proxy</b> یا <b>TUN</b> رو روشن کن."]
        },
        s4:"بعد از اتصال",
        s4l:["اگه کانفیگ‌ها قدیمی شدن، توی برنامه گزینه‌ی <b>Update subscription</b> (به‌روزرسانی ساب) رو بزن.",
             "اگه کند بود یا وصل نشد، یه کانفیگ دیگه انتخاب کن.",
             "قبل از اتصال VPN یا پروکسی‌های دیگه رو خاموش کن.",
             "ساعت و تاریخ دستگاه باید روی حالت خودکار باشه."]
      }
    },
    en: {
      sub:"Pick your app. Your subscription is added automatically",
      steps:[["Install the app","Skip if you have it"],["Tap your app","Sub is added for you"],["Connect","Pick a config & go"]],
      rec:"Recommended", copy:"Copy subscription link", qr:"QR code", guideTitle:"Full guide", dlAria:"Download app",
      desc:{ v2rayNG:"Most popular and stable", Happ:"Clean, simple and fast", Hiddify:"Smart: auto-picks the best config",
             V2Box:"Light and smooth", Streisand:"Free and fast", Shadowrocket:"Paid, advanced", v2rayN:"Desktop client, manual import" },
      missT:"App didn't open?",
      missP:"{app} is probably not installed. Install it first, then tap its card again. Your subscription link is already copied.",
      manT:"Link copied. Now add it inside the app",
      manP:"{app} has no one-tap import. Open it, go to Subscription and paste the link. Exact steps are in the guide below.",
      hDl:"Install {app}", hCopy:"Copy link again", hGuide:"Manual steps",
      g:{
        s1:"Connect in 4 steps",
        s1l:["Choose your platform at the top (detected automatically).",
             "Don't have the app? Tap the <b>download</b> icon next to its name.",
             "Tap the <b>app card</b>. The link is copied and the app opens with your sub.",
             "If the app asks to add a subscription, confirm, pick a config and press <b>Connect</b>."],
        s2:"App didn't open?",
        s2l:["Make sure the app is installed and opened at least once.",
             "If this page is inside Telegram or Instagram, use <b>Open in Browser</b> from the top menu and try again.",
             "Try another app from the list. They all read the same sub.",
             "Still nothing? Use the manual method below."],
        s3:"Manual method (copy link)",
        s3p:"First tap “Copy subscription link”, then:",
        man:{
          android:["Open the app and tap <b>+</b>.","Choose <b>Subscription</b> or <b>Import from clipboard</b>.","Paste the copied link and confirm.","<b>Update</b> the subscription and pick a config to connect."],
          ios:["Open the app and tap <b>+</b> at the top.","Choose <b>Add from clipboard</b> or <b>Subscribe</b> (in Shadowrocket set Type to Subscribe and paste the link).","Save, then update the subscription.","Pick a config and switch the connection on. On first use iOS asks to add a VPN; tap <b>Allow</b>."],
          windows:["Open the app. In v2rayN use <b>Subscription → Subscription settings</b>, in Hiddify use <b>Add profile</b>.","Click <b>Add</b> and paste the link with <code>Ctrl+V</code>.","Confirm and click <b>Update subscription</b>.","Pick a config and enable <b>System Proxy</b> or <b>TUN</b> mode."]
        },
        s4:"After connecting",
        s4l:["If configs look outdated, hit <b>Update subscription</b> in the app.",
             "If it's slow or won't connect, try another config.",
             "Turn off other VPNs or proxies first.",
             "Keep the device date and time on automatic."]
      }
    }
  };
  function qcT(){ return QC_T[lang] || QC_T.fa; }

  var qcPlatform = null, qcHintApp = -1, qcHintTimer = null, qcStepNo = 1, qcGuideTimer = null;
  function qcDetect(){
    var ua = navigator.userAgent || "";
    if (/iPhone|iPad|iPod/i.test(ua)) return "ios";
    if (/Windows/i.test(ua)) return "windows";
    return "android";
  }
  function movePill(instant){
    var pill = document.getElementById("qcPill");
    var on = document.querySelector("#qcSide .qc-tab.on");
    if (!pill || !on) return;
    if (instant) pill.classList.add("nt");
    pill.style.width = on.offsetWidth + "px";
    pill.style.transform = "translateX(" + on.offsetLeft + "px)";
    if (instant){ void pill.offsetWidth; pill.classList.remove("nt"); }
  }
  function qcFillText(){
    var t = qcT();
    document.querySelectorAll("#qcModal [data-qc]").forEach(function(el){
      var k = el.getAttribute("data-qc");
      if (t[k]) el.textContent = t[k];
    });
  }
  function qcRenderSteps(){
    var box = document.getElementById("qcSteps");
    if (!box) return;
    box.innerHTML = qcT().steps.map(function(s, i){
      var n = i + 1, cls = n < qcStepNo ? "done" : (n === qcStepNo ? "on" : "");
      var num = n < qcStepNo ? '<i class="fa-solid fa-check"></i>' : faDigits(n);
      return '<div class="qc-step ' + cls + '"><span class="qc-num">' + num + '</span>' +
             '<span class="qc-step-t">' + esc(s[0]) + '</span><span class="qc-step-d">' + esc(s[1]) + '</span></div>';
    }).join("");
  }
  function qcSetStep(n){ qcStepNo = n; qcRenderSteps(); }
  function qcRenderGuide(p){
    var g = qcT().g;
    function ul(tag, arr){ return "<" + tag + ">" + arr.map(function(x){ return "<li>" + x + "</li>"; }).join("") + "</" + tag + ">"; }
    var secs = [
      { i:"fa-list-ol",             t:g.s1, h:ul("ol", g.s1l), open:true },
      { i:"fa-triangle-exclamation", t:g.s2, h:ul("ul", g.s2l) },
      { i:"fa-hand-pointer",         t:g.s3, h:"<p>" + g.s3p + "</p>" + ul("ol", g.man[p] || g.man.android), id:"man" },
      { i:"fa-circle-check",         t:g.s4, h:ul("ul", g.s4l) }
    ];
    var box = document.getElementById("qcGuide");
    if (!box) return;
    var wasMan = box.querySelector('details[data-id="man"]');
    var keepMan = wasMan && wasMan.open;
    box.innerHTML = secs.map(function(s){
      return '<details class="qc-g"' + (s.open ? " open" : "") + (s.id ? ' data-id="' + s.id + '"' : "") + '>' +
        '<summary><i class="fa-solid ' + s.i + '"></i><span>' + esc(s.t) + '</span><i class="fa-solid fa-chevron-down qc-chev"></i></summary>' +
        '<div class="qc-gb">' + s.h + '</div></details>';
    }).join("");
    if (keepMan){ var m = box.querySelector('details[data-id="man"]'); if (m) m.open = true; }
  }
  function qcHideHint(){
    var h = document.getElementById("qcHint");
    if (h){ h.classList.remove("show"); h.innerHTML = ""; }
    qcHintApp = -1;
    clearTimeout(qcHintTimer);
  }
  function qcShowHint(idx, kind){
    var a = (QC_APPS[qcPlatform] || [])[idx], h = document.getElementById("qcHint");
    if (!a || !h) return;
    var t = qcT(), miss = kind === "miss";
    qcHintApp = idx;
    h.innerHTML =
      '<span class="qc-hint-ico"><i class="fa-solid ' + (miss ? "fa-circle-question" : "fa-hand-pointer") + '"></i></span>' +
      '<div class="qc-hint-b"><div class="qc-hint-t">' + esc(miss ? t.missT : t.manT) + '</div>' +
      '<div class="qc-hint-p">' + esc((miss ? t.missP : t.manP).replace("{app}", a.n)) + '</div>' +
      '<div class="qc-hint-btns">' +
        (a.dl ? '<button type="button" class="qc-hbtn" data-act="hint-dl"><i class="fa-solid fa-download"></i>' + esc(t.hDl.replace("{app}", a.n)) + '</button>' : '') +
        '<button type="button" class="qc-hbtn" data-act="copy"><i class="fa-solid fa-copy"></i>' + esc(t.hCopy) + '</button>' +
        (miss ? '' : '<button type="button" class="qc-hbtn" data-act="goto-man"><i class="fa-solid fa-book-open"></i>' + esc(t.hGuide) + '</button>') +
      '</div></div>';
    h.classList.remove("show"); void h.offsetWidth; h.classList.add("show");
    var body = document.getElementById("qcBody");
    if (body) body.scrollTo({ top:0, behavior:"smooth" });
    if (!miss) qcSetStep(3);
  }
  function renderQC(p, first){
    qcPlatform = p;
    document.querySelectorAll("#qcSide .qc-tab").forEach(function(b){
      var on = b.getAttribute("data-p") === p;
      b.classList.toggle("on", on);
      b.setAttribute("aria-selected", on ? "true" : "false");
    });
    movePill(!!first);
    qcHideHint();
    qcStepNo = 1; qcRenderSteps();
    qcFillText();
    var t = qcT();
    var grid = document.getElementById("qcGrid");
    grid.style.setProperty("--base", first ? "260ms" : "40ms");
    grid.innerHTML = (QC_APPS[p] || []).map(function(a, idx){
      return '<div class="qc-card" role="button" tabindex="0" data-i="' + idx + '" style="--i:' + idx + ';--g1:' + a.g[0] + ';--g2:' + a.g[1] + '">' +
        '<span class="qc-ico"><i class="fa-solid ' + a.i + '"></i></span>' +
        '<span class="qc-info"><span class="qc-name">' + esc(a.n) + (idx === 0 ? '<em class="qc-rec">' + esc(t.rec) + '</em>' : '') + '</span>' +
        '<span class="qc-desc">' + esc(t.desc[a.n] || "") + '</span></span>' +
        (a.dl ? '<button type="button" class="qc-dl" data-i="' + idx + '" aria-label="' + esc(t.dlAria + " " + a.n) + '"><i class="fa-solid fa-download"></i></button>' : '') +
        '<i class="fa-solid fa-chevron-right qc-go"></i></div>';
    }).join("");
    clearTimeout(qcGuideTimer);
    if (first){
      /* راهنمای پایین صفحه بعد از تمام شدن انیمیشن باز شدن ساخته می‌شود */
      var gbox = document.getElementById("qcGuide"); if (gbox) gbox.innerHTML = "";
      qcGuideTimer = setTimeout(function(){ qcRenderGuide(p); }, 450);
    } else {
      qcRenderGuide(p);
    }
    var body = document.getElementById("qcBody");
    if (body && !first) body.scrollTo({ top:0, behavior:"auto" });
  }
  function qcRefresh(){
    var m = document.getElementById("qcModal");
    if (m && m.classList.contains("show") && qcPlatform){ renderQC(qcPlatform, false); setTimeout(function(){ movePill(true); }, 30); }
  }
  function openQC(){
    var modal = document.getElementById("qcModal");
    var box = modal.querySelector(".qc-box");
    var btn = document.getElementById("fluffyQuickBtn");
    /* انیمیشن از مرکز دکمه شروع می‌شود (در موبایل شیت از پایین بالا می‌آید) */
    if (btn && box){
      var r = btn.getBoundingClientRect();
      var bw = box.offsetWidth, bh = box.offsetHeight;
      var left = (window.innerWidth - bw) / 2, top = (window.innerHeight - bh) / 2;
      box.style.transformOrigin = Math.round(r.left + r.width/2 - left) + "px " + Math.round(r.top + r.height/2 - top) + "px";
      btn.classList.remove("qc-hit"); void btn.offsetWidth; btn.classList.add("qc-hit");
      setTimeout(function(){ btn.classList.remove("qc-hit"); }, 650);
    }
    renderQC(qcPlatform || qcDetect(), true);
    document.documentElement.classList.add("qc-open");
    /* ساخت DOM تمام شد؛ انیمیشن را فریم بعد شروع می‌کنیم تا شروعش گیر نکند */
    requestAnimationFrame(function(){
      requestAnimationFrame(function(){
        modal.classList.add("show");
        movePill(true);
      });
    });
  }
  function closeQC(){
    document.getElementById("qcModal").classList.remove("show");
    qcHideHint();
    clearTimeout(qcGuideTimer);
    /* انیمیشن‌های پس‌زمینه بعد از بسته شدن شیت دوباره ادامه پیدا می‌کنند */
    setTimeout(function(){
      var m = document.getElementById("qcModal");
      if (m && !m.classList.contains("show")) document.documentElement.classList.remove("qc-open");
    }, 320);
  }
  function qcWave(card, e){
    var r = card.getBoundingClientRect();
    var w = document.createElement("i");
    w.className = "qc-wave";
    w.style.left = ((e && e.clientX ? e.clientX : r.left + r.width/2) - r.left) + "px";
    w.style.top  = ((e && e.clientY ? e.clientY : r.top + r.height/2) - r.top) + "px";
    card.appendChild(w);
    setTimeout(function(){ w.remove(); }, 750);
  }
  function qcDone(card){
    var ic = card.querySelector(".qc-ico i");
    if (!ic || card._busy) return;
    card._busy = true;
    var old = ic.className;
    card.classList.remove("done"); void card.offsetWidth;
    ic.className = "fa-solid fa-check";
    card.classList.add("done");
    setTimeout(function(){
      ic.className = old; card.classList.remove("done"); card._busy = false;
    }, 1500);
  }
  function qcOpenDl(idx){
    var a = (QC_APPS[qcPlatform] || [])[idx];
    if (!a || !a.dl) return;
    var link = document.createElement("a");
    link.href = a.dl;
    link.target = "_blank";
    link.rel = "noopener noreferrer";
    link.style.display = "none";
    document.body.appendChild(link);
    link.click();
    setTimeout(function(){ if (link.parentNode) link.parentNode.removeChild(link); }, 0);
  }
  /* اگر بعد از چند ثانیه صفحه هنوز جلوی چشم بود یعنی برنامه باز نشده */
  function qcArmHint(idx){
    clearTimeout(qcHintTimer);
    var left = false;
    function mark(){ left = true; }
    function cleanup(){
      document.removeEventListener("visibilitychange", mark);
      window.removeEventListener("pagehide", mark);
      window.removeEventListener("blur", mark);
    }
    document.addEventListener("visibilitychange", mark);
    window.addEventListener("pagehide", mark);
    window.addEventListener("blur", mark);
    qcHintTimer = setTimeout(function(){
      cleanup();
      if (!left && !document.hidden) qcShowHint(idx, "miss");
    }, 2000);
  }
  function qcOpenApp(a, idx){
    var url = getSubUrl();
    copyText(url).then(function(){ showToast(tr("toastSub")); }).catch(function(){});
    qcHideHint();
    if (a.u){
      qcSetStep(3);
      qcArmHint(idx);
      try { window.location.href = a.u(url, QC_NAME); } catch(e){}
    } else {
      qcShowHint(idx, "manual");
    }
  }
  function qcAction(act){
    if (act === "copy"){
      copyText(getSubUrl()).then(function(){ showToast(tr("toastSub")); }).catch(function(){ showToast(tr("toastErr"), true); });
    } else if (act === "qr"){
      closeQC();
      setTimeout(function(){ showQR(getSubUrl(), "Subscription"); }, 260);
    } else if (act === "hint-dl"){
      qcOpenDl(qcHintApp);
    } else if (act === "goto-man"){
      var m = document.querySelector('#qcGuide details[data-id="man"]');
      if (m){ m.open = true; setTimeout(function(){ m.scrollIntoView({ behavior:"smooth", block:"center" }); }, 60); }
    }
  }

  function renderCountdown(){
    if (countdownInterval) clearInterval(countdownInterval);
    if (EXPIRE_TS === null){
      ['expDay','expHour','expMin'].forEach(function(id){
        document.getElementById(id).textContent = '∞';
      });
      return;
    }
    function tick(){
      var diff = Math.max(0, EXPIRE_TS - Date.now());
      var d = Math.floor(diff / 86400000);
      var h = Math.floor((diff % 86400000) / 3600000);
      var m = Math.floor((diff % 3600000) / 60000);
      document.getElementById("expDay").textContent = faDigits(d);
      document.getElementById("expHour").textContent = faDigits(h);
      document.getElementById("expMin").textContent = faDigits(m);
      if (diff <= 0) clearInterval(countdownInterval);
    }
    tick();
    countdownInterval = setInterval(tick, 1000);
  }

  function applyTraffic(){
    var remain = TOTAL_BYTES > 0 ? Math.max(0, TOTAL_BYTES - USED_BYTES) : 0;
    var pct = TOTAL_BYTES > 0 ? Math.min(100, USED_BYTES / TOTAL_BYTES * 100) : 0;
    var unlimited = TOTAL_BYTES <= 0;

    document.getElementById("statUsed").textContent = fmtBytes(USED_BYTES);
    document.getElementById("statTotal").textContent = unlimited ? '∞' : fmtBytes(TOTAL_BYTES);
    document.getElementById("statRemain").textContent = unlimited ? '∞' : fmtBytes(remain);
    document.getElementById("dlVal").textContent = fmtBytes(DOWNLOAD_BYTES);
    document.getElementById("ulVal").textContent = fmtBytes(UPLOAD_BYTES);

    var dayCount = EXPIRE_DAYS && EXPIRE_DAYS > 0 ? EXPIRE_DAYS : 30;
    var avg = USED_BYTES / dayCount;
    document.getElementById("avgVal").textContent = fmtBytes(avg);

    document.getElementById("remainNum").textContent = unlimited ? '∞' : fmtBytes(remain);
    document.getElementById("ringPct").textContent = unlimited ? '∞' : faDigits(Math.round(pct)) + '٪';
    document.getElementById("barFill").style.width = (unlimited ? 100 : pct).toFixed(1) + '%';
    document.getElementById("barPct").textContent = unlimited ? '∞' : faDigits(Math.round(pct)) + '٪';

    var ringFill = document.querySelector(".fluffy-ring-fill");
    if (ringFill) ringFill.setAttribute("stroke-dashoffset", String(263.9 * (1 - (unlimited ? 1 : pct / 100))));
  }

  /* ---------- کارت انتخاب زبان ---------- */
  var langPop = document.getElementById("langPop");
  var langMenuOpen = false;

  function updateLangMenu(){
    if (!langPop) return;
    langPop.setAttribute("aria-label", lang === "fa" ? "زبان" : "Language");
    langPop.querySelectorAll(".lang-opt").forEach(function(o){
      o.setAttribute("aria-checked", o.getAttribute("data-lang") === lang ? "true" : "false");
    });
  }

  function positionPop(pop, btn){
    if (!btn || !pop) return;
    var r = btn.getBoundingClientRect();
    var vw = document.documentElement.clientWidth || window.innerWidth;
    var w = pop.offsetWidth, edge = 12, gap = 10;
    var cx = r.left + r.width / 2;
    /* لبه‌ی کارت با لبه‌ی بیرونی دکمه هم‌تراز می‌شود (چپ یا راست، بسته به جای دکمه) */
    var left = cx < vw / 2 ? r.left : r.right - w;
    left = Math.max(edge, Math.min(left, vw - w - edge));
    pop.style.left = (left + (window.pageXOffset || 0)) + "px";
    pop.style.top = (r.bottom + (window.pageYOffset || 0) + gap) + "px";
    /* کارت از وسط دکمه «رشد» می‌کند */
    pop.style.transformOrigin = Math.round(cx - left) + "px 0px";
  }
  function positionLangMenu(){ positionPop(langPop, document.getElementById("langBtn")); }

  function openLangMenu(viaKeyboard){
    if (!langPop || langMenuOpen) return;
    var btn = document.getElementById("langBtn");
    closeThemeMenu(false);
    positionLangMenu();
    langMenuOpen = true;
    langPop.classList.add("open");
    if (btn) btn.setAttribute("aria-expanded", "true");
    if (viaKeyboard){
      requestAnimationFrame(function(){
        var cur = langPop.querySelector('.lang-opt[aria-checked="true"]') || langPop.querySelector(".lang-opt");
        if (cur) cur.focus({ preventScroll: true });
      });
    }
  }

  function closeLangMenu(refocus){
    if (!langPop || !langMenuOpen) return;
    langMenuOpen = false;
    langPop.classList.remove("open");
    var btn = document.getElementById("langBtn");
    if (btn){
      btn.setAttribute("aria-expanded", "false");
      if (refocus) btn.focus({ preventScroll: true });
    }
  }

  /* ---------- کارت انتخاب تم (روز / شب) ---------- */
  var themePop = document.getElementById("themePop");
  var themeMenuOpen = false;

  function currentTheme(){ return root.getAttribute("data-theme") === "dark" ? "dark" : "light"; }

  function updateThemeMenu(){
    if (!themePop) return;
    var cur = currentTheme();
    themePop.setAttribute("aria-label", lang === "fa" ? "تم" : "Theme");
    themePop.querySelectorAll(".lang-opt").forEach(function(o){
      o.setAttribute("aria-checked", o.getAttribute("data-theme-opt") === cur ? "true" : "false");
    });
  }

  function openThemeMenu(viaKeyboard){
    if (!themePop || themeMenuOpen) return;
    var btn = document.getElementById("themeBtn");
    closeLangMenu(false);
    updateThemeMenu();
    positionPop(themePop, btn);
    themeMenuOpen = true;
    themePop.classList.add("open");
    if (btn) btn.setAttribute("aria-expanded", "true");
    if (viaKeyboard){
      requestAnimationFrame(function(){
        var cur = themePop.querySelector('.lang-opt[aria-checked="true"]') || themePop.querySelector(".lang-opt");
        if (cur) cur.focus({ preventScroll: true });
      });
    }
  }

  function closeThemeMenu(refocus){
    if (!themePop || !themeMenuOpen) return;
    themeMenuOpen = false;
    themePop.classList.remove("open");
    var btn = document.getElementById("themeBtn");
    if (btn){
      btn.setAttribute("aria-expanded", "false");
      if (refocus) btn.focus({ preventScroll: true });
    }
  }

  function pickTheme(next){
    var btn = document.getElementById("themeBtn");
    closeThemeMenu(themePop && themePop.contains(document.activeElement));
    if (!next || next === currentTheme()) return;
    /* اول کارت جمع می‌شود، بعد تم نرم عوض می‌شود (مثل انتخاب زبان) */
    setTimeout(function(){
      if (btn){
        btn.classList.add("theme-pulse");
        setTimeout(function(){ btn.classList.remove("theme-pulse"); }, 450);
      }
      toggleTheme();
      updateThemeMenu();
    }, 180);
  }

  function pickLang(next){
    var btn = document.getElementById("langBtn");
    closeLangMenu(langPop && langPop.contains(document.activeElement));
    if (!next || next === lang) return;
    if (btn) btn.classList.add("switching");
    setTimeout(function(){
      applyLang(next);
      if (btn) btn.classList.remove("switching");
    }, 180);
  }

  function bind(){
    document.querySelectorAll('.nav-item').forEach(function(btn){
      if (btn._bound) return; btn._bound = true;
      btn.addEventListener('click', function(){
        switchTab(btn.getAttribute('data-target'));
      });
    });

    var themeBtn = document.getElementById("themeBtn");
    if (themeBtn && !themeBtn._bound){
      themeBtn._bound = true;
      themeBtn.addEventListener("click", function(e){
        e.preventDefault();
        if (themeMenuOpen) closeThemeMenu(false);
        else openThemeMenu(e.detail === 0);
      });
    }
    if (themePop && !themePop._bound){
      themePop._bound = true;
      themePop.addEventListener("click", function(e){
        var opt = e.target.closest ? e.target.closest(".lang-opt") : null;
        if (opt) pickTheme(opt.getAttribute("data-theme-opt"));
      });
      themePop.addEventListener("keydown", function(e){
        var opts = Array.prototype.slice.call(themePop.querySelectorAll(".lang-opt"));
        var i = opts.indexOf(document.activeElement);
        var k = e.key;
        if (k === "ArrowDown" || k === "ArrowUp"){
          e.preventDefault();
          var n = k === "ArrowDown" ? (i + 1) % opts.length : (i <= 0 ? opts.length - 1 : i - 1);
          opts[n].focus();
        } else if (k === "Home"){
          e.preventDefault(); opts[0].focus();
        } else if (k === "End"){
          e.preventDefault(); opts[opts.length - 1].focus();
        } else if (k === "Tab"){
          closeThemeMenu(false);
        }
      });
      document.addEventListener("keydown", function(e){
        if (e.key === "Escape" && themeMenuOpen) closeThemeMenu(true);
      });
      document.addEventListener("pointerdown", function(e){
        if (!themeMenuOpen) return;
        var t = e.target;
        if (themePop.contains(t) || (themeBtn && themeBtn.contains(t))) return;
        closeThemeMenu(false);
      });
      window.addEventListener("resize", function(){ closeThemeMenu(false); });
    }
    var langBtn = document.getElementById("langBtn");
    if (langBtn && !langBtn._bound){
      langBtn._bound = true;
      langBtn.addEventListener("click", function(e){
        e.preventDefault();
        if (langMenuOpen) closeLangMenu(false);
        else openLangMenu(e.detail === 0);
      });
    }
    if (langPop && !langPop._bound){
      langPop._bound = true;
      langPop.addEventListener("click", function(e){
        var opt = e.target.closest ? e.target.closest(".lang-opt") : null;
        if (opt) pickLang(opt.getAttribute("data-lang"));
      });
      langPop.addEventListener("keydown", function(e){
        var opts = Array.prototype.slice.call(langPop.querySelectorAll(".lang-opt"));
        var i = opts.indexOf(document.activeElement);
        var k = e.key;
        if (k === "ArrowDown" || k === "ArrowUp"){
          e.preventDefault();
          var n = k === "ArrowDown" ? (i + 1) % opts.length : (i <= 0 ? opts.length - 1 : i - 1);
          opts[n].focus();
        } else if (k === "Home"){
          e.preventDefault(); opts[0].focus();
        } else if (k === "End"){
          e.preventDefault(); opts[opts.length - 1].focus();
        } else if (k === "Tab"){
          closeLangMenu(false);
        }
      });
      document.addEventListener("keydown", function(e){
        if (e.key === "Escape" && langMenuOpen) closeLangMenu(true);
      });
      document.addEventListener("pointerdown", function(e){
        if (!langMenuOpen) return;
        var t = e.target;
        if (langPop.contains(t) || (langBtn && langBtn.contains(t))) return;
        closeLangMenu(false);
      });
      window.addEventListener("resize", function(){ closeLangMenu(false); });
    }
    var fluffyQuick = document.getElementById("fluffyQuickBtn");
    if (fluffyQuick && !fluffyQuick._bound){
      fluffyQuick._bound = true;
      fluffyQuick.addEventListener("click", openQC);
    }
    var qcModal = document.getElementById("qcModal");
    if (qcModal && !qcModal._bound){
      qcModal._bound = true;
      qcModal.addEventListener("click", function(e){
        if (e.target === qcModal) return closeQC();
        var t = e.target;
        var tab = t.closest(".qc-tab");
        if (tab) return renderQC(tab.getAttribute("data-p"), false);
        var act = t.closest("[data-act]");
        if (act) return qcAction(act.getAttribute("data-act"));
        var dl = t.closest(".qc-dl");
        if (dl){ e.preventDefault(); e.stopPropagation(); return qcOpenDl(+dl.getAttribute("data-i")); }
        var card = t.closest(".qc-card");
        if (card){
          var idx = +card.getAttribute("data-i");
          var a = (QC_APPS[qcPlatform] || [])[idx];
          if (a){ qcWave(card, e); qcDone(card); qcOpenApp(a, idx); }
        }
      });
      qcModal.addEventListener("keydown", function(e){
        if (e.key !== "Enter" && e.key !== " ") return;
        var card = e.target.closest && e.target.closest(".qc-card");
        if (!card || e.target.closest(".qc-dl")) return;
        e.preventDefault();
        var idx = +card.getAttribute("data-i");
        var a = (QC_APPS[qcPlatform] || [])[idx];
        if (a){ qcDone(card); qcOpenApp(a, idx); }
      });
      document.getElementById("qcClose").addEventListener("click", closeQC);
      window.addEventListener("resize", function(){ movePill(true); });
      document.addEventListener("keydown", function(e){ if (e.key === "Escape") closeQC(); });
    }
    var copySub = document.getElementById("copySubBtn");
    if (copySub && !copySub._bound){
      copySub._bound = true;
      copySub.addEventListener("click", function(){
        copyText(getSubUrl()).then(function(){
          copySub.classList.add("success");
          showToast(tr("toastSub"));
          clearTimeout(copySub._t);
          copySub._t = setTimeout(function(){ copySub.classList.remove("success"); }, 1800);
        }).catch(function(){ showToast(tr("toastErr"), true); });
      });
    }
    var qrSub = document.getElementById("qrSubBtn");
    if (qrSub && !qrSub._bound){
      qrSub._bound = true;
      qrSub.addEventListener("click", function(){
        showQR(getSubUrl(), "Subscription");
      });
    }
    var btnV2ray = document.getElementById("btnImportV2ray");
    if (btnV2ray && !btnV2ray._bound){
      btnV2ray._bound = true;
      btnV2ray.addEventListener("click", function(){
        copyText(getSubUrl()).then(function(){ showToast(tr("toastSub")); })
          .catch(function(){ showToast(tr("toastErr"), true); });
      });
    }
    var btnNpv = document.getElementById("btnImportNpv");
    if (btnNpv && !btnNpv._bound){
      btnNpv._bound = true;
      btnNpv.addEventListener("click", function(){
        copyText(getSubUrl()).then(function(){ showToast(tr("toastSub")); })
          .catch(function(){ showToast(tr("toastErr"), true); });
      });
    }
    var copyAllBtn = document.getElementById("copyAllConfigsBtn");
    if (copyAllBtn && !copyAllBtn._bound){
      copyAllBtn._bound = true;
      copyAllBtn.addEventListener("click", function(){
        if (!CONFIG_LINKS.length) return showToast(tr("noConfig"), true);
        copyText(CONFIG_LINKS.join("\n")).then(function(){
          copyAllBtn.classList.add("success-state");
          showToast(tr("toastAll") + " (" + CONFIG_LINKS.length + ")");
          setTimeout(function(){ copyAllBtn.classList.remove("success-state"); }, 1800);
        }).catch(function(){ showToast(tr("toastErr"), true); });
      });
    }
    var copySubText = document.getElementById("copySubTextBtn");
    if (copySubText && !copySubText._bound){
      copySubText._bound = true;
      copySubText.addEventListener("click", function(){
        if (!CONFIG_LINKS.length) return showToast(tr("noConfig"), true);
        copyText(CONFIG_LINKS.join("\n")).then(function(){
          copySubText.classList.add("success-state");
          showToast(tr("toastSubText"));
          setTimeout(function(){ copySubText.classList.remove("success-state"); }, 1600);
        }).catch(function(){ showToast(tr("toastErr"), true); });
      });
    }

    var searchInput = document.getElementById("searchInput");
    if (searchInput && !searchInput._bound){
      searchInput._bound = true;
      searchInput.addEventListener("input", applyFilters);
    }

    document.querySelectorAll(".config-chip").forEach(function(chip){
      if (chip._bound) return;
      chip._bound = true;
      chip.addEventListener("click", function(){
        document.querySelectorAll(".config-chip").forEach(function(c){ c.classList.remove("active"); });
        chip.classList.add("active");
        ACTIVE_FILTER = chip.getAttribute("data-filter") || "all";
        applyFilters();
      });
    });

    var closeBtn = document.getElementById("closeModal");
    if (closeBtn && !closeBtn._bound){
      closeBtn._bound = true;
      closeBtn.addEventListener("click", closeQR);
    }
    var modal = document.getElementById("qrModal");
    if (modal && !modal._bound){
      modal._bound = true;
      modal.addEventListener("click", function(e){ if (e.target === modal) closeQR(); });
      document.addEventListener("keydown", function(e){ if (e.key === "Escape") closeQR(); });
      var qrCopyBtn = document.getElementById("qrLinkCopy");
      if (qrCopyBtn) qrCopyBtn.addEventListener("click", function(){
        if (!qrCurrent) return;
        copyText(qrCurrent.data).then(function(){
          showToast(tr(qrCurrent.title === "Subscription" ? "toastSub" : "toastCfg"));
          qrCopyBtn.classList.add("ok");
          setTimeout(function(){ qrCopyBtn.classList.remove("ok"); }, 1400);
        }).catch(function(){ showToast(tr("toastErr"), true); });
      });
    }
    document.querySelectorAll(".copy-id").forEach(function(btn){
      if (btn._bound) return; btn._bound = true;
      btn.addEventListener("click", function(){
        copyText(SUB_ID).then(function(){ showToast(tr("toastId")); });
      });
    });
  }

  async function load(){
    try {
      var r = await fetch(apiUrl, { cache: 'no-store' });
      if (!r.ok) throw new Error(r.status === 404 ? (isFa() ? 'کاربر پیدا نشد' : 'User not found') : (isFa() ? 'خطا در سرور' : 'Server error'));
      var d = await r.json();
      applyData(d);
    } catch (e){
      var list = document.getElementById('configList');
      if (list) list.innerHTML = '<div class="empty-configs">' + esc(e.message) + '</div>';
      document.getElementById("statusTitle").textContent = isFa() ? 'خطا در دریافت اطلاعات' : 'Failed to load';
      console.warn('load failed:', e);
    }
  }

  function setText(id, val){
    var el = document.getElementById(id);
    if (el) el.textContent = val;
  }

  function applyData(d){
    var cfgs = [];
    if (Array.isArray(d.configs) && d.configs.length){
      cfgs = d.configs.filter(function(c){ return c && c.indexOf('%F0%9F%93%8A') === -1; });
    }
    if (!cfgs.length && Array.isArray(d.worker_configs) && d.worker_configs.length) cfgs = d.worker_configs.slice();
    if (!cfgs.length && d.config) cfgs = [d.config];
    CONFIG_LINKS = cfgs;
    CONFIG_NAMES = cfgs.map(function(c){
      var h = c.split('#');
      try { return decodeURIComponent((h[1] || '').replace(/\+/g, ' ')) || 'Config'; }
      catch(e){ return 'Config'; }
    });
    CONFIG_PROTOS = cfgs.map(protoOf);

    SUB_ID = d.config_uuid || uname;
    var curClean = window.location.href.split('?')[0].split('#')[0];
    if (curClean && (curClean.indexOf('/link/') !== -1 || curClean.indexOf('/sub/') !== -1)) {
      SUB_URL = curClean;
    } else if (d.subscription_url || d.sub_url) {
      SUB_URL = d.subscription_url || d.sub_url;
    } else {
      SUB_URL = window.location.origin + '/link/' + SUB_ID;
    }
    setText('subUrl', SUB_URL);
    setText('subId', SUB_ID);

    IS_ACTIVE = d.is_active !== false && d.status === 'active';
    var pill = document.getElementById('statusPill');
    var dot = pill ? pill.querySelector('.status-dot') : null;
    setText('statusPillText', IS_ACTIVE ? tr('statusOnline') : (isFa() ? 'غیرفعال' : 'Inactive'));
    if (pill) pill.style.opacity = IS_ACTIVE ? '' : '.55';
    if (dot) dot.style.background = IS_ACTIVE ? '#22c55e' : '#fb7185';
    setText('statusTitle', IS_ACTIVE ? tr('connectedActive') : (isFa() ? 'اتصال غیرفعال' : 'Connection inactive'));

    USED_BYTES = d.traffic_used_bytes || 0;
    TOTAL_BYTES = d.traffic_limit_bytes || 0;
    DOWNLOAD_BYTES = d.download_bytes || (USED_BYTES * 0.63);
    UPLOAD_BYTES = d.upload_bytes || (USED_BYTES * 0.37);

    var days = d.expire_days;
    EXPIRE_DAYS = (days !== null && days !== undefined) ? days : null;
    applyTraffic();

    setText('usernameVal', d.username || uname);
    var maxConn = parseInt(d.max_ip_per_user, 10);
    if (isNaN(maxConn) || maxConn < 1) maxConn = 0;
    setText('connLimitVal', maxConn > 0 ? faDigits(maxConn) : '∞');
    var usedIps = parseInt(d.used_ips, 10);
    if (isNaN(usedIps)) usedIps = 0;
    var connEl = document.getElementById('connActiveVal');
    if (connEl){
      connEl.textContent = faDigits(usedIps);
      connEl.style.color = (maxConn > 0 && usedIps >= maxConn) ? '#fbbf24' : '#22c55e';
    }

    if (EXPIRE_DAYS !== null){
      EXPIRE_TS = d.expire_at_ts ? d.expire_at_ts * 1000 : Date.now() + EXPIRE_DAYS * 86400000;
      setText('expireMain', faDigits(EXPIRE_DAYS) + ' ' + tr('day'));
      setText('expireSub', new Date(EXPIRE_TS).toLocaleDateString(isFa() ? 'fa-IR' : 'en-US'));
      setText('daysRowText', faDigits(EXPIRE_DAYS) + ' ' + tr('day') + ' ' + (isFa() ? 'باقی‌مانده' : 'left'));
    } else {
      EXPIRE_TS = null;
      setText('expireMain', tr('unlimited'));
      setText('expireSub', tr('noExpire'));
      setText('daysRowText', tr('unlimited'));
    }
    renderCountdown();

    document.title = 'Subscription | ' + (d.username || uname);

    renderConfigs();
  }

  document.addEventListener("visibilitychange", function(){
    if (document.hidden) root.setAttribute("data-paused","");
    else root.removeAttribute("data-paused");
  });

  applyLang(lang);
  bind();
  switchTab('tab-home');
  initVisibilityPause();
  initPageEnter();
  var bootDone = true;

  if (document.readyState === 'loading'){
    document.addEventListener('DOMContentLoaded', function(){
      if (bootDone) return; /* اجرای اول انجام شده؛ تکرار باعث ری‌استارت انیمیشن ورود می‌شد */
      bind(); switchTab('tab-home'); initVisibilityPause(); initPageEnter();
    });
  }

  load();

})();
