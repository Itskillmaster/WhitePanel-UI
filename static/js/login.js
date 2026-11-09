var passwordInput = document.getElementById('password');
var passwordMask = document.getElementById('password-mask');
function syncPasswordMask() {
    if (passwordInput && passwordMask) passwordMask.textContent = '*'.repeat(passwordInput.value.length);
}
passwordInput.addEventListener('input', syncPasswordMask);
passwordInput.addEventListener('change', syncPasswordMask);
setTimeout(syncPasswordMask, 0);

document.getElementById('login-form').addEventListener('submit', async function(e) {
    e.preventDefault();
    var btn = document.getElementById('loginBtn');
    var btnText = document.getElementById('btn-text');
    var btnSpinner = document.getElementById('btn-spinner');
    var errDiv = document.getElementById('error-msg');
    var errText = document.getElementById('error-text');

    errDiv.classList.remove('show');
    btn.disabled = true;
    btnText.classList.add('hidden');
    btnSpinner.classList.remove('hidden');

    try {
        var r = await fetch('/api/login', {
            method: 'POST',
            credentials: 'same-origin',
            cache: 'no-store',
            headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
            body: JSON.stringify({ password: document.getElementById('password').value })
        });
        var d = await r.json().catch(function() { return {}; });
        if (!r.ok || !d.ok) {
            throw new Error(d.detail || 'Login failed');
        }

        // Confirm the browser stored and is sending the session cookie before navigating.
        var check = await fetch('/api/me', {
            method: 'GET',
            credentials: 'same-origin',
            cache: 'no-store',
            headers: { 'Accept': 'application/json' }
        });
        var session = await check.json().catch(function() { return {}; });
        if (!check.ok || session.authenticated !== true) {
            throw new Error('ورود انجام شد اما نشست ورود ذخیره نشد. اتصال HTTPS و تنظیمات Cookie سرور را بررسی کنید.');
        }
        window.location.replace('/white');
    } catch(ex) {
        errText.textContent = ex.message;
        errDiv.classList.add('show');
        btn.disabled = false;
        btnText.classList.remove('hidden');
        btnSpinner.classList.add('hidden');
    }
});

document.addEventListener('wheel', function(e) { if(e.ctrlKey || e.metaKey) e.preventDefault(); }, {passive:false});
document.addEventListener('keydown', function(e) {
    if ((e.ctrlKey || e.metaKey) && ['+', '-', '=', '0'].indexOf(e.key) >= 0) e.preventDefault();
    if ((e.ctrlKey || e.metaKey) && (e.key || '').toLowerCase() === 'c' && !/^(INPUT|TEXTAREA|SELECT)$/.test((e.target || {}).tagName || '')) e.preventDefault();
});
['copy', 'cut', 'contextmenu', 'selectstart', 'dragstart'].forEach(function(x) {
    document.addEventListener(x, function(e) {
        if (!/^(INPUT|TEXTAREA|SELECT)$/.test((e.target || {}).tagName || '')) e.preventDefault();
    });
});
['touchstart', 'touchmove'].forEach(function(x) {
    document.addEventListener(x, function(e) {
        if (e.touches.length > 1) {
            e.preventDefault();
            e.stopPropagation();
        }
    }, {passive:false});
});
document.addEventListener('gesturestart', function(e) { e.preventDefault(); });
document.addEventListener('gesturechange', function(e) { e.preventDefault(); });
document.addEventListener('gestureend', function(e) { e.preventDefault(); });
