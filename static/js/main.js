// ─── State ───
var $ = qs => document.querySelector(qs);
var $id = id => document.getElementById(id);
var curCfg = '';
var ALL_IB = [];
var ALL_USERS = [];
var selectedAccent = localStorage.getItem('white-accent') || 'red';
var selectedTheme = localStorage.getItem('white-theme') || 'dark';

// ─── I18N (FA / EN) — abbreviated, full dict preserved ───
var LANG = localStorage.getItem('white-lang') || 'fa';
var I18N = { fa: {}, en: {} };

Object.assign(I18N.fa, {
  'dash': 'داشبورد', 'navGroup': 'مدیریت', 'systemGroup': 'سیستم', 'advanced': 'پیشرفته', 'users': 'کاربران', 'inbounds': 'اینباندها', 'scanner': 'اسکنر IP', 'worker': 'ورکر (Worker)', 'settings': 'تنظیمات', 'contact': 'تماس با من', 'logout': 'خروج',
  'tab.dash': 'داشبورد', 'tab.users': 'مدیریت کاربران', 'tab.inbounds': 'مدیریت اینباندها', 'tab.scanner': 'اسکنر آی‌پی', 'tab.settings': 'تنظیمات', 'tab.advanced': 'پیشرفته', 'tab.help': 'راهنما',
  'newUser': 'کاربر جدید', 'newInbound': 'اینباند جدید', 'domain': 'دامنه',
  'loading': 'در حال بارگذاری...', 'noUsers': 'هیچ کاربری وجود ندارد', 'noInbounds': 'هیچ اینباندی وجود ندارد', 'errLoadUsers': 'خطا در بارگذاری کاربران', 'errLoadInbounds': 'خطا در بارگذاری اینباندها', 'errLoadStats': 'خطا در بارگذاری آمار',
  'liveBadge': 'زنده', 'cloudflareTitle': 'Cloudflare', 'railwayTitle': 'Railway', 'tcpTitle': 'TCP', 'sniTitle': 'SNI',
  'scnTip': 'اسکن زنده از رنج‌های Cloudflare. IPهای پاسخ‌دهنده مرتب و ۱۰ تای اول به صورت خودکار ذخیره می‌شوند.',
  'ipCount': 'تعداد IP', 'targetPort': 'پورت هدف', 'startScan': 'شروع اسکن', 'stop': 'توقف', 'scanning': 'در حال اسکن...', 'scanStopped': 'اسکن متوقف شد',
  'found': 'IP یافت شده', 'tried': 'تست شده', 'bestPing': 'بهترین پینگ', 'saved': 'ذخیره شده', 'savedFirst10': 'ذخیره شده (اولین ۱۰ آی‌پی)', 'clear': 'پاک کردن', 'notSaved': 'هنوز ذخیره‌ای نیست', 'startPrompt': 'برای شروع اسکن، دکمه «شروع اسکن» را بزنید', 'noHit': 'هیچ آی‌پی پاسخ‌دهنده‌ای یافت نشد', 'saveCleared': 'لیست ذخیره‌شده پاک شد',
  'scnRailwayTip': 'اسکن زنده از رنج‌های اعلام‌شده Railway (اروپا).',
  'scnTcpTip': 'دامنه موردنظر را وارد کنید؛ IP دامنه (DNS) نمایش داده می‌شود و پینگ TCP آن گزارش می‌شود.',
  'addDomain': 'افزودن دامنه', 'checkDomains': 'شروع بررسی دامنه‌ها', 'checking': 'در حال بررسی...', 'checkingDomains': 'در حال بررسی دامنه‌ها...',
  'appearance': 'تنظیمات ظاهری', 'themeMode': 'حالت نمایش', 'dark': 'تاریک', 'light': 'روشن', 'lightMode': 'حالت روشن فعال شد', 'darkMode': 'حالت تاریک فعال شد',
  'panelMusic': 'موسیقی پنل', 'bgMusic': 'موسیقی پس‌زمینه', 'toggleMusic': 'روشن/خاموش', 'changePassword': 'تغییر رمز عبور', 'curPass': 'رمز فعلی', 'newPass': 'رمز جدید', 'changePassBtn': 'تغییر رمز',
  'contactSub': 'برای پشتیبانی و همکاری با من در ارتباط باشید',
  'createUser': 'ایجاد کاربر جدید', 'editUser': 'ویرایش کاربر', 'username': 'نام کاربری', 'conn': 'اتصال همزمان', 'speedLimit': 'محدودیت سرعت (Mbps)', 'speedLimitHint': 'سقف مجموع سرعت اتصال‌های رله‌شده WebSocket/XHTTP؛ صفر یعنی نامحدود.', 'volume': 'حجم', 'unit': 'واحد', 'expire': 'اعتبار', 'day': 'روز', 'month': 'ماه',
  'inboundsMulti': 'اینباندها (چندتایی)', 'noInboundOpt': 'اینباندی وجود ندارد', 'inboundPickHint': 'برای هر اینباند، گزینه‌های مخصوصش نمایش داده می‌شود.',
  'workerCountries': 'کشورهای چندمنطقه‌ای', 'addCfIps': 'افزودن ۱۰ کانفیگ با IPهای Cloudflare اسکن‌شده', 'addRailwayIps': 'افزودن ۱۰ کانفیگ با IPهای Railway اسکن‌شده', 'realityNoExtra': 'Reality گزینه اضافه‌ای ندارد — فقط کانفیگ اصلی Reality.',
  'cancel': 'انصراف', 'submitUser': 'ایجاد کاربر', 'creatingUser': 'در حال ایجاد...', 'saving': 'در حال ذخیره...', 'userCreated': 'با موفقیت ایجاد شد',
  'volEdit': 'حجم (GB - ۰=نامحدود)', 'daysEdit': 'روز اعتبار (۰=نامحدود)', 'status': 'وضعیت', 'active': 'فعال', 'inactive': 'غیرفعال', 'resetTraffic': 'بازنشانی ترافیک', 'saveChanges': 'ذخیره تغییرات',
  'inboundName': 'نام اینباند', 'port': 'پورت', 'securityType': 'نوع امنیت', 'realitySettings': 'تنظیمات Reality', 'autoGen': 'تولید خودکار', 'genRealityKey': 'تولید کلید Reality (x25519)',
  'shortIds': 'شناسه‌های کوتاه', 'hex8': 'هگز ۸ رقمی — خالی = تصادفی', 'generate': '+ تولید', 'whitex': 'WhiteX',
  'extDomain': 'دامنه خارجی', 'extPortCfg': 'پورت خارجی', 'extDomainPlace': 'دامنه خارجی برای کانفیگ کلاینت',
  'network': 'شبکه', 'serviceName': 'نام سرویس', 'createInbound': 'ایجاد اینباند', 'editInbound': 'ویرایش اینباند',
  'config': 'کانفیگ', 'copy': 'کپی', 'close': 'بستن', 'copied': 'کانفیگ کپی شد', 'noCfg': 'کانفیگی موجود نیست', 'errCopy': 'خطا در کپی', 'delete': 'حذف', 'edit': 'ویرایش', 'unlimited': 'نامحدود', 'expired': 'منقضی', 'daysUnit': ' روز', 'usersCount': 'کاربران', 'secNetwork': 'شبکه', 'security': 'امنیت', 'expiry': 'انقضا', 'connTh': 'اتصال',
  'confirmDeleteUser': 'آیا از حذف این کاربر اطمینان دارید؟', 'confirmDeleteInbound': 'آیا از حذف این اینباند اطمینان دارید؟', 'keysGenerated': 'کلیدهای Reality تولید شد', 'errKeyGen': 'خطا در تولید کلید', 'shortIdGen': 'Short ID تولید شد', 'shortIdNew': 'Short ID جدید تولید شد',
  'inboundUpdated': 'اینباند بروزرسانی شد', 'inboundCreated': 'اینباند ایجاد شد', 'updateInbound': 'بروزرسانی اینباند', 'inboundDeleted': 'اینباند حذف شد',
  'download': 'دانلود', 'upload': 'آپلود', 'activeUsers': 'کاربران فعال/کل', 'trafficUsed': 'ترافیک مصرفی', 'activeConns': 'اتصالات فعال', 'uptime': 'زمان فعالیت',
  'goodMorning': 'صبح بخیر', 'goodNoon': 'ظهر بخیر', 'goodEvening': 'عصر بخیر', 'admin': 'ادمین',
  'toast.userCreated': 'کاربر ایجاد شد', 'toast.userUpdated': 'کاربر بروزرسانی شد', 'toast.userDeleted': 'کاربر حذف شد', 'toast.userToggled': 'وضعیت کاربر تغییر کرد', 'toast.passChanged': 'رمز تغییر کرد', 'toast.err': 'خطا',
  'backupDownloaded': 'بکاپ دانلود شد', 'restoreDone': 'بازیابی انجام شد', 'restoreFailed': 'بازیابی ناموفق', 'restoreUploading': 'در حال آپلود و بازیابی...',
  'panelUpdate': 'بروزرسانی پنل', 'updatePanel': 'بروزرسانی', 'confirmUpdate': 'پنل از GitHub بروزرسانی شود؟', 'updateStarted': 'بروزرسانی شروع شد...', 'updateDone': 'بروزرسانی کامل شد — پنل را ری‌استارت کنید', 'updateFailed': 'بروزرسانی ناموفق بود', 'updateRunning': 'در حال بروزرسانی...',
  'panelUpdateHint': 'با زدن این دکمه، آخرین نسخه ریپو از GitHub دریافت و روی سرور نصب می‌شود (Itskillmaster). بعد از اتمام، پنل را در Railway ری‌استارت کنید تا نسخه جدید اعمال شود.',
  'backupRestore': 'بکاپ و بازیابی', 'backup': 'بکاپ', 'restore': 'ریستور', 'backupRestoreHint': 'بکاپ کامل شامل: کاربران، لینک‌ها، اینباندها، اسکن‌ها، تنظیمات، رمز، گروه‌ها و تنظیمات Worker.',
  'autoconfig': 'پیکربندی خودکار', 'tunnel': 'تونل', 'qrCode': 'کد QR',
  'acAutoBtn': 'خودکار', 'acHint': 'اسکن کامل خودکار: پینگ دامنه‌ها → انتخاب Railway/Cloudflare → اسکن IP → تست دامنه‌های TCP → اسکن SNI → پینگ نهایی پنل.',
  'acEnterDomain': 'لطفاً دامنه پنل را وارد کنید', 'acPingingDomains': 'در حال گرفتن پینگ از دامنه‌ها...', 'acYourDomain': 'دامنه شما', 'acPickDomain': 'دامنه را انتخاب کنید:',
  'acScanningRailway': 'اسکن IP های Railway و ذخیره ۱۰ تای برتر...', 'acScanningCf': 'اسکن IP های Cloudflare و ذخیره ۱۰ تای برتر...',
  'acTcpTesting': 'تست دامنه‌های TCP...', 'acSniTesting': 'تست SNI ها...', 'acFinalPing': 'پینگ نهایی به دامنه فعلی پنل...', 'acDone': 'اتمام اسکن خودکار ✓',
  'tunnelNotReady': 'ابتدا یک Worker در تب Worker بسازید — بخش Tunnel بعد از ساخت Worker فعال می‌شود.', 'tunnelPath': 'مسیر اتصال', 'workerPing': 'پینگ سرور تا Worker', 'tunnelKv': 'KV اختصاصی Tunnel', 'tunnelInboundPath': 'اینباند (TLS WS)', 'tunnelCreate': 'ایجاد اینباند Tunnel', 'tunnelCreated': 'اینباند Tunnel ساخته شد',
  'reverse': 'معکوس', 'reverseSwitch': 'حالت Reverse', 'reverseHint': 'با روشن کردن این سوییچ، مسیر اتصال برعکس می‌شود: User → Worker → Railway → Site.', 'reverseOn': 'Reverse فعال شد', 'reverseOff': 'Reverse غیرفعال شد',
  'bot': 'ربات', 'nodes': 'نودها', 'addNode': 'افزودن نود', 'editNode': 'ویرایش نود', 'nodeDomain': 'دامنه نود', 'nodeApiKey': 'کلید API', 'nodeStatus': 'وضعیت', 'nodeOnline': 'آنلاین', 'nodeOffline': 'آفلاین', 'nodeSynced': 'همگام‌سازی شد', 'nodeSyncFailed': 'خطا در همگام‌سازی', 'nodeDelete': 'حذف نود', 'confirmDeleteNode': 'آیا از حذف این نود اطمینان دارید؟', 'nodeDeleted': 'نود حذف شد', 'syncAllNodes': 'همگام‌سازی همه', 'noNodes': 'هنوز نودی اضافه نشده', 'nodeAdded': 'نود اضافه شد', 'nodeUpdated': 'نود بروزرسانی شد',
  'serverIp': 'آی‌پی سرور', 'serverInfo': 'اطلاعات سرور', 'apiKey': 'کلید API', 'copyApiKey': 'کپی API Key', 'inboundType': 'نوع اینباند'
});

Object.assign(I18N.fa, {
  'helpTemplate': 'قالب', 'helpServerEdit': 'ویرایش سرور', 'helpConfigOptions': 'تنظیمات کانفیگ', 'helpCopyPrompt': 'کپی متن راهنما', 'helpPromptCopied': 'متن راهنما کپی شد', 'helpPromptCopyFailed': 'کپی متن راهنما ناموفق بود',
  'helpIntro': 'راهنمای صفحه ساب و قالب آن', 'helpPromptHint': 'این Prompt منطق صفحه ساب فعلی را توضیح می‌دهد.',
  'helpServerEditHint': 'آدرس اتصال کانفیگ‌های پیش‌فرض TLS + WS را تغییر دهید.', 'helpChatgptAddress': 'chatgpt.com در Address', 'helpCustomAddress': 'ذخیره آدرس سفارشی', 'helpAddressLabel': 'آدرس سفارشی', 'helpAddressPlaceholder': 'دامنه یا IP (IPv4/IPv6)', 'helpPanelIpv6': 'IPv6 دامنه پنل', 'helpResetAddress': 'دامنه پنل',
  'helpAddressSaved': 'آدرس ذخیره شد', 'helpAddressInvalid': 'دامنه یا IP معتبر وارد کنید', 'helpIpv6NotFound': 'برای دامنه پنل IPv6 پیدا نشد', 'helpSettingsSaved': 'تنظیمات ذخیره شد',
  'helpEchChoose': 'انتخاب ECH', 'helpEchPickerTitle': 'ECH List', 'helpEchPickerHint': 'یک مقدار را انتخاب کنید.', 'helpEchPresetSaved': 'گزینه ECH ذخیره شد',
  'helpFeatureHint': 'این گزینه‌ها فقط به کانفیگ‌های پیش‌فرض TLS + WS افزوده می‌شوند.', 'helpEch': 'افزودن ECH', 'helpEchHint': 'پارامتر ECH را اضافه می‌کند.', 'helpFf': 'افزودن F+F', 'helpFfHint': 'مجموعه cipher و fragment را اضافه می‌کند.',
  'helpSpoof': 'Spoof', 'helpSpoofHint': 'پارامتر snispoofing را اضافه می‌کند.', 'helpSpoofSni': 'SNI جعلی', 'helpSpoofIp': 'IP مقصد', 'helpSpoofMissing': 'برای فعال کردن Spoof، SNI و IP را وارد کنید', 'helpSpoofSaved': 'تنظیمات Spoof ذخیره شد',
  'advancedInboundHint': 'یک اینباند را انتخاب کنید و JSON کامل آن را ویرایش کنید.', 'advancedJson': 'JSON اینباند', 'advancedSave': 'ذخیره JSON', 'advancedSaved': 'JSON اینباند ذخیره شد', 'advancedInvalidJson': 'JSON معتبر نیست', 'advancedNoInbounds': 'اینباند قابل ویرایشی وجود ندارد', 'defaultTlsWsInbound': 'پیش‌فرض TLS + WS',
  'notif.nav': 'اعلان‌ها', 'notif.title': 'اعلان‌ها', 'notif.refresh': 'بروزرسانی', 'notif.date': 'تاریخ و ساعت', 'notif.hint': 'هر اعلان یک آکاردیون است.', 'notif.loading': 'در حال دریافت...', 'notif.empty': 'اعلانی ثبت نشده.', 'notif.error': 'دریافت اعلان‌ها ممکن نشد.',
  'proxy.nav': 'پروکسی', 'proxy.intro': 'آدرس و پورت پروکسی را وارد کنید؛ پنل SOCKS4/SOCKS5/HTTP/HTTPS را آزمایش می‌کند.', 'proxy.inboundHint': 'ترافیک این اینباند از پروکسی عبور می‌کند.',
  'proxy.add': 'افزودن پروکسی', 'proxy.host': 'آدرس (IP یا دامنه)', 'proxy.port': 'پورت', 'proxy.country': 'کشور (اختیاری)',
  'proxy.auth': 'نام کاربری و رمز (اختیاری)', 'proxy.authHint': 'اگر پروکسی شناسه نمی‌خواهد باز نکنید.', 'proxy.username': 'Username', 'proxy.password': 'Password',
  'proxy.test': 'تست اتصال', 'proxy.testing': 'در حال تست', 'proxy.save': 'ذخیره پروکسی', 'proxy.working': 'سالم ✓', 'proxy.notWorking': 'کار نمی‌کند',
  'proxy.ping': 'پینگ', 'proxy.egress': 'IP خروجی', 'proxy.error': 'خطا', 'proxy.fillHost': 'آدرس و پورت را وارد کنید', 'proxy.testOk': 'اتصال موفق — حالا ذخیره کنید',
  'proxy.testFail': 'اتصال ناموفق بود', 'proxy.testFirst': 'اول «تست اتصال» را بزنید', 'proxy.saved': 'پروکسی ذخیره شد', 'proxy.count': 'پروکسی', 'proxy.empty': 'هنوز پروکسی اضافه نشده — «افزودن پروکسی» را بزنید',
  'proxy.ok': 'فعال ✓', 'proxy.untested': 'تست‌نشده', 'proxy.off': 'غیرفعال', 'proxy.users': 'کاربران', 'proxy.lastTest': 'آخرین تست',
  'proxy.pingBtn': 'پینگ', 'proxy.activate': 'فعال', 'proxy.deactivate': 'غیرفعال', 'proxy.delete': 'حذف', 'proxy.activated': 'فعال شد', 'proxy.deactivated': 'غیرفعال شد', 'proxy.deleted': 'حذف شد', 'proxy.confirmDelete': 'این پروکسی حذف شود؟',
  'wk.sectionHint': 'سه بخش مستقل: Tunnel، Reverse و Worker.', 'wk.tunnel': 'Tunnel', 'wk.reverse': 'Reverse', 'wk.worker': 'Worker',
  'wk.tunnelHint': 'User → Railway → Cloudflare Worker → Site.', 'wk.reverseHint': 'User → Cloudflare Worker → Railway → Site.', 'wk.workerHint': 'Worker چند-لوکیشن با لوکیشن‌های خروجی.',
  'wk.createReverse': 'ایجاد اینباند Reverse', 'wk.createInbound': 'ایجاد اینباند Worker', 'wk.reverseCreated': 'اینباند Reverse ایجاد شد', 'wk.inboundCreated': 'اینباند Worker ایجاد شد',
  'wk.refresh': 'بروزرسانی', 'wk.noLogs': 'لاگی ثبت نشده', 'wk.locations': 'لوکیشن‌ها', 'wk.locationsHint': 'هر لوکیشن یک خروجی در کشور مشخص است.', 'wk.code': 'کد', 'wk.addLocation': 'افزودن لوکیشن', 'wk.location': 'لوکیشن', 'wk.noLocations': 'هنوز لوکیشنی ثبت نشده', 'wk.auto': 'خودکار', 'wk.deploy': 'Deploy / Sync',
  'helpLogs': 'لاگ پنل', 'helpLogsHint': 'لاگ کامل پنل با فیلتر و جست‌وجو.', 'helpLogsSearch': 'جست‌وجو در لاگ…', 'helpLogsAll': 'همه دسته‌ها', 'helpLogsAllLevels': 'همه سطح‌ها', 'helpLogsEmpty': 'در حال بارگذاری…', 'helpLogsTotal': 'کل', 'helpLogsConn': 'اتصال', 'helpLogsAuth': 'ورود/خروج', 'helpLogsErr': 'خطا', 'helpLogsNone': 'هیچ لاگی پیدا نشد.',
  'vercelDomain.nav': 'Vercel Domain', 'vercelDomain.title': 'اتصال دامنه Vercel', 'vercelDomain.intro': 'دامنه پنل Railway را به رله WebSocket روی Vercel وصل کن.', 'vercelDomain.token': 'توکن API وِرسل', 'vercelDomain.getToken': 'دریافت توکن', 'vercelDomain.tokenHint': 'توکن فقط روی سرور ذخیره می‌شود.', 'vercelDomain.create': 'ساخت دامنه Vercel', 'vercelDomain.refresh': 'بروزرسانی', 'vercelDomain.panelDomain': 'دامنه اصلی Railway', 'vercelDomain.domain': 'دامنه Vercel', 'vercelDomain.project': 'نام پروژه', 'vercelDomain.ping': 'پینگ از اینترنت من', 'vercelDomain.delete': 'حذف دامنه Vercel', 'vercelDomain.configTitle': 'کانفیگ WebSocket وِرسل', 'vercelDomain.configHint': 'کاربر را انتخاب کن.', 'vercelDomain.user': 'کاربر', 'vercelDomain.noUsers': 'کانفیگ معتبری پیدا نشد.', 'vercelDomain.copyConfig': 'کپی کانفیگ', 'vercelDomain.connectedHint': 'دامنه متصل است.', 'vercelDomain.loadError': 'دریافت وضعیت Vercel انجام نشد.', 'vercelDomain.enterToken': 'ابتدا توکن را وارد کن.', 'vercelDomain.deploying': 'در حال ساخت…', 'vercelDomain.created': 'دامنه ساخته شد.', 'vercelDomain.createError': 'ساخت دامنه انجام نشد.', 'vercelDomain.confirmDelete': 'پروژه Vercel حذف شود؟', 'vercelDomain.deleted': 'پروژه حذف شد.', 'vercelDomain.deleteError': 'حذف انجام نشد.', 'vercelDomain.noConfig': 'کانفیگی برای کپی وجود ندارد.', 'vercelDomain.legacyProtocol': 'این پروژه رله قدیمی XHTTP است.', 'vercelDomain.notConnected': 'ابتدا دامنه را بساز.', 'vercelDomain.pinging': 'در حال پینگ…', 'vercelDomain.pingingHint': 'در حال اندازه‌گیری…', 'vercelDomain.pingFailed': 'پینگ ناموفق بود.', 'vercelDomain.clientPingResult': 'پینگ: میانه {median} ms · کمینه {min} ms · بیشینه {max} ms · نمونه {samples}/4',
  'tm.nav': 'مدیریت تونل (TM)', 'tm.short': 'TM', 'tm.title': 'مدیریت تونل', 'tm.server': 'سرورها', 'tm.configure': 'پیکربندی', 'tm.servers': 'سرورها', 'tm.links': 'اتصال‌ها', 'tm.entry': 'ورودی', 'tm.exit': 'خروجی', 'tm.lastApply': 'آخرین اعمال',
  'tm.addServer': 'افزودن سرور', 'tm.serverHint': 'هر سرور یک Node واقعی همین پنل است. کلید API فقط روی سرور ذخیره می‌شود.', 'tm.editServer': 'ویرایش', 'tm.delServer': 'حذف', 'tm.pingServer': 'تست', 'tm.noServers': 'هنوز سروری اضافه نشده — «افزودن سرور» را بزنید.', 'tm.loadErr': 'دریافت اطلاعات ممکن نشد',
  'tm.addServerTitle': 'افزودن سرور TM', 'tm.editServerTitle': 'ویرایش سرور', 'tm.serverName': 'نام نمایشی', 'tm.serverDomain': 'دامنه پنل سرور', 'tm.spiKey': 'API Key سرور', 'tm.keyHint': 'در حالت ویرایش، کلید را خالی بگذارید تا کلید قبلی حفظ شود.', 'tm.keyNeverShown': 'کلید ذخیره‌شده: {{masked}}', 'tm.confirmDelServer': 'این سرور حذف شود؟', 'tm.serverAdded': 'سرور اضافه شد', 'tm.serverSaved': 'سرور ذخیره شد', 'tm.serverDeleted': 'سرور حذف شد', 'tm.serverSavedOffline': 'ذخیره شد، اما سرور در دسترس نیست', 'tm.remoteOffline': 'سرور ریموت در دسترس نیست',
  'tm.apply': 'اعمال', 'tm.validate': 'اعتبارسنجی', 'tm.chainPing': 'پینگ زنجیره', 'tm.clientPingAll': 'پینگ کلاینت', 'tm.curved': 'منحنی', 'tm.straight': 'مستقیم', 'tm.connect': 'اتصال جدید', 'tm.disconnect': 'قطع',
  'tm.configureHint': 'حرکت نود: نگه داشتن طولانی · کنترل‌ها: ضربه کوتاه · کلیدها دو سرور را وصل می‌کنند.', 'tm.selectSource': 'سرور مبدأ', 'tm.selectDest': 'سرور مقصد', 'tm.hlConnect': 'اتصال: مبدأ «{{src}}» — مقصد را انتخاب کنید', 'tm.duplicate': 'این اتصال از قبل وجود دارد', 'tm.selfLoop': 'سرور نمی‌تواند به خودش وصل شود',
  'tm.edgeDeleted': 'اتصال حذف شد', 'tm.edgeAdded': 'اتصال ساخته شد', 'tm.styleChanged': 'نوع خط عوض شد', 'tm.validateOk': 'توپولوژی معتبر است', 'tm.applyDone': 'تونل اعمال شد', 'tm.applyFailed': 'اعمال ناموفق بود', 'tm.setDone': 'سرور فعال شد', 'tm.setClear': 'حذف انتخاب', 'tm.chainPingResult': 'تأخیر زنجیره', 'tm.clientPingResult': 'تأخیر از این دستگاه', 'tm.offline': 'قطع / Timeout', 'tm.measuring': 'در حال اندازه‌گیری…', 'tm.samples': 'نمونه', 'tm.noChain': 'زنجیره‌ای وجود ندارد', 'tm.noGraph': 'گراف خالی است', 'tm.pinging': 'در حال پینگ…',
  'tm.generated': 'مسیر کانفیگ تولیدی', 'tm.generatedHint': 'این دقیقاً همان چیزی است که کاربر دریافت می‌کند.', 'tm.regenerate': 'بازتولید', 'tm.copy': 'کپی',
  'tm.state.idle': 'غیرفعال', 'tm.state.active': 'فعال', 'tm.state.dirty': 'تغییر اعمال‌نشده', 'tm.state.applied': 'اعمال‌شده', 'tm.role.initial': 'ورودی', 'tm.role.middle': 'میانی', 'tm.role.final': 'خروجی', 'tm.role.initial_final': 'ورودی + خروجی', 'tm.role.idle': 'غیرفعال',
  'tm.clientPingCardHint': 'اندازه‌گیری واقعی از این دستگاه.', 'tm.serverPingCardHint': 'سلامت از بک‌اند پنل.', 'step.validated': 'گراف تأیید شد', 'step.chain': 'زنجیره مشخص شد', 'step.server': 'سرور', 'step.inbound': 'اینباند به‌روزرسانی شد', 'step.users': 'کاربران همگام شدند', 'step.applied': 'پیکربندی اعمال شد', 'step.rollback': 'بازگشت انجام شد',
  'tm.clientPing': 'پینگ کلاینت', 'tm.serverPing': 'پینگ سرور', 'tm.fromThisDevice': 'از این دستگاه', 'tm.fromBackend': 'از بک‌اند', 'tm.clientPingEmpty': 'هنوز اندازه‌گیری نشده', 'tm.serverPingEmpty': 'هنوز اندازه‌گیری نشده', 'tm.cycle': 'حلقه ایجاد می‌شود', 'tm.destChanged': 'مقصد عوض شد', 'tm.errorApiKey': 'کلید API الزامی است', 'tm.fillDomain': 'دامنه الزامی است', 'tm.selectEdgeFirst': 'ابتدا یک خط انتخاب کنید', 'tm.keyFrom': 'این سرور مبدأ شود', 'tm.keyTo': 'این سرور مقصد شود', 'tm.hlFrom': 'مبدأ «{{src}}» — کلید روی باکس مقصد را بزنید', 'tm.hlTo': 'مقصد «{{dst}}» — کلید پشت باکس مبدأ را بزنید', 'tm.invalid': 'توپولوژی نامعتبر است', 'tm.applying': 'در حال اعمال…', 'tm.balancer': 'بالانسر', 'tm.balancerHint': 'چند سرور موازی.', 'tm.auto': 'خودکار', 'tm.autoDone': 'اعمال شد', 'tm.roleLabel': 'نقش',
  'wpNoListen': 'بدون پورت TCP محلی؛ ترافیک روی HTTPS/WSS خود پنل می‌آید.'
});

Object.assign(I18N.en, {
  'dash': 'Dashboard', 'navGroup': 'Management', 'systemGroup': 'System', 'advanced': 'Advanced', 'users': 'Users', 'inbounds': 'Inbounds', 'scanner': 'IP Scanner', 'worker': 'Worker', 'settings': 'Settings', 'contact': 'Contact Me', 'logout': 'Logout',
  'tab.dash': 'Dashboard', 'tab.users': 'Manage Users', 'tab.inbounds': 'Manage Inbounds', 'tab.scanner': 'IP Scanner', 'tab.settings': 'Settings', 'tab.advanced': 'Advanced', 'tab.help': 'Help',
  'newUser': 'New User', 'newInbound': 'New Inbound', 'domain': 'Domain',
  'loading': 'Loading...', 'noUsers': 'No users found', 'noInbounds': 'No inbounds found', 'errLoadUsers': 'Error loading users', 'errLoadInbounds': 'Error loading inbounds', 'errLoadStats': 'Error loading stats',
  'liveBadge': 'LIVE', 'cloudflareTitle': 'Cloudflare', 'railwayTitle': 'Railway', 'tcpTitle': 'TCP', 'sniTitle': 'SNI',
  'scnTip': 'Live scan of Cloudflare ranges.', 'ipCount': 'IP Count', 'targetPort': 'Target Port', 'startScan': 'Start Scan', 'stop': 'Stop', 'scanning': 'Scanning...', 'scanStopped': 'Scan stopped',
  'found': 'IP Found', 'tried': 'Tested', 'bestPing': 'Best Ping', 'saved': 'Saved', 'savedFirst10': 'Saved (first 10 IPs)', 'clear': 'Clear', 'notSaved': 'Nothing saved yet', 'startPrompt': 'Press "Start Scan" to begin', 'noHit': 'No responding IP found', 'saveCleared': 'Saved list cleared',
  'scnRailwayTip': 'Live scan of announced Railway ranges.', 'scnTcpTip': 'Enter a domain; its IP and TCP ping are reported live.', 'addDomain': 'Add Domain', 'checkDomains': 'Check Domains', 'checking': 'Checking...', 'checkingDomains': 'Checking domains...',
  'appearance': 'Appearance', 'themeMode': 'Display Mode', 'dark': 'Dark', 'light': 'Light', 'lightMode': 'Light mode enabled', 'darkMode': 'Dark mode enabled',
  'panelMusic': 'Panel Music', 'bgMusic': 'Background Music', 'toggleMusic': 'On/Off', 'changePassword': 'Change Password', 'curPass': 'Current Password', 'newPass': 'New Password', 'changePassBtn': 'Change Password',
  'contactSub': 'Get in touch with me', 'createUser': 'Create New User', 'editUser': 'Edit User', 'username': 'Username', 'conn': 'Concurrent Connections', 'speedLimit': 'Speed limit (Mbps)', 'speedLimitHint': 'Combined WebSocket/XHTTP relay cap; 0 means unlimited.', 'volume': 'Traffic', 'unit': 'Unit', 'expire': 'Expiry', 'day': 'Day', 'month': 'Month',
  'inboundsMulti': 'Inbounds (multi-select)', 'noInboundOpt': 'No inbounds available', 'inboundPickHint': 'Each inbound shows its own options.', 'workerCountries': 'Worker Countries', 'addCfIps': 'Add 10 configs with scanned Cloudflare IPs', 'addRailwayIps': 'Add 10 configs with scanned Railway IPs', 'realityNoExtra': 'Reality needs no extras.',
  'cancel': 'Cancel', 'submitUser': 'Create User', 'creatingUser': 'Creating...', 'saving': 'Saving...', 'userCreated': 'created successfully',
  'volEdit': 'Traffic (GB - 0=Unlimited)', 'daysEdit': 'Expiry days (0=Unlimited)', 'status': 'Status', 'active': 'Active', 'inactive': 'Inactive', 'resetTraffic': 'Reset traffic', 'saveChanges': 'Save Changes',
  'inboundName': 'Inbound Name', 'port': 'Port', 'securityType': 'Security Type', 'realitySettings': 'Reality Settings', 'autoGen': 'Auto-generated', 'genRealityKey': 'Generate Reality Keys (x25519)', 'shortIds': 'Short IDs', 'hex8': '8-digit hex', 'generate': '+ Generate', 'whitex': 'WhiteX',
  'extDomain': 'External Domain', 'extPortCfg': 'External Port', 'extDomainPlace': 'External domain shown in the client config', 'network': 'Network', 'serviceName': 'Service Name', 'createInbound': 'Create Inbound', 'editInbound': 'Edit Inbound',
  'config': 'Config', 'copy': 'Copy', 'close': 'Close', 'copied': 'Config copied', 'noCfg': 'No config available', 'errCopy': 'Copy failed', 'delete': 'Delete', 'edit': 'Edit', 'unlimited': 'Unlimited', 'expired': 'Expired', 'daysUnit': ' days', 'usersCount': 'Users', 'secNetwork': 'Network', 'security': 'Security', 'expiry': 'Expiry', 'connTh': 'Conn',
  'confirmDeleteUser': 'Are you sure?', 'confirmDeleteInbound': 'Are you sure?', 'keysGenerated': 'Reality keys generated', 'errKeyGen': 'Key generation failed', 'shortIdGen': 'Short ID generated', 'shortIdNew': 'New Short ID generated',
  'inboundUpdated': 'Inbound updated', 'inboundCreated': 'Inbound created', 'updateInbound': 'Update Inbound', 'inboundDeleted': 'Inbound deleted',
  'download': 'Download', 'upload': 'Upload', 'activeUsers': 'Active / Total Users', 'trafficUsed': 'Traffic Used', 'activeConns': 'Active Connections', 'uptime': 'Uptime',
  'goodMorning': 'Good morning', 'goodNoon': 'Good afternoon', 'goodEvening': 'Good evening', 'admin': 'Admin',
  'toast.userCreated': 'User created', 'toast.userUpdated': 'User updated', 'toast.userDeleted': 'User deleted', 'toast.userToggled': 'User status changed', 'toast.passChanged': 'Password changed', 'toast.err': 'Error',
  'backupDownloaded': 'Backup downloaded', 'restoreDone': 'Restore completed', 'restoreFailed': 'Restore failed', 'restoreUploading': 'Uploading...',
  'panelUpdate': 'Panel Update', 'updatePanel': 'Update', 'confirmUpdate': 'Update from GitHub?', 'updateStarted': 'Update started...', 'updateDone': 'Update finished', 'updateFailed': 'Update failed', 'updateRunning': 'Updating...',
  'panelUpdateHint': 'Pulls the latest repo from GitHub and installs it on the server (Itskillmaster). Restart the panel on Railway after completion.',
  'backupRestore': 'Backup & Restore', 'backup': 'Backup', 'restore': 'Restore', 'backupRestoreHint': 'Full backup includes users, links, inbounds, scans, settings, password, groups, and Worker config.',
  'autoconfig': 'Auto Config', 'tunnel': 'Tunnel', 'qrCode': 'QR Code',
  'acAutoBtn': 'Auto', 'acHint': 'Full automatic scan: ping domains → pick Railway/Cloudflare → IP scan → TCP test → SNI scan → final panel ping.', 'acEnterDomain': 'Please enter the panel domain', 'acPingingDomains': 'Pinging domains...', 'acYourDomain': 'Your domain', 'acPickDomain': 'Pick a domain:', 'acScanningRailway': 'Scanning Railway IPs...', 'acScanningCf': 'Scanning Cloudflare IPs...', 'acTcpTesting': 'Testing TCP domains...', 'acSniTesting': 'Testing SNIs...', 'acFinalPing': 'Final ping...', 'acDone': 'Auto scan complete ✓',
  'tunnelNotReady': 'Create a Worker first.', 'tunnelPath': 'Connect path', 'workerPing': 'Server → Worker ping', 'tunnelKv': 'Dedicated Tunnel KV', 'tunnelInboundPath': 'Inbound (TLS WS)', 'tunnelCreate': 'Create Tunnel Inbound', 'tunnelCreated': 'Tunnel inbound created',
  'reverse': 'Reverse', 'reverseSwitch': 'Reverse mode', 'reverseHint': 'Flip the connect path.', 'reverseOn': 'Reverse enabled', 'reverseOff': 'Reverse disabled',
  'bot': 'Bot', 'nodes': 'Nodes', 'addNode': 'Add Node', 'editNode': 'Edit Node', 'nodeDomain': 'Node Domain', 'nodeApiKey': 'API Key', 'nodeStatus': 'Status', 'nodeOnline': 'Online', 'nodeOffline': 'Offline', 'nodeSynced': 'Synced', 'nodeSyncFailed': 'Sync failed', 'nodeDelete': 'Delete Node', 'confirmDeleteNode': 'Are you sure?', 'nodeDeleted': 'Node deleted', 'syncAllNodes': 'Sync All', 'noNodes': 'No nodes added yet', 'nodeAdded': 'Node added', 'nodeUpdated': 'Node updated',
  'serverIp': 'Server IP', 'serverInfo': 'Server Info', 'apiKey': 'API Key', 'copyApiKey': 'Copy API Key', 'inboundType': 'Inbound Type'
});

Object.assign(I18N.en, {
  'helpTemplate': 'Template', 'helpServerEdit': 'Server Edit', 'helpConfigOptions': 'Config Options', 'helpCopyPrompt': 'Copy Prompt', 'helpPromptCopied': 'Prompt copied', 'helpPromptCopyFailed': 'Could not copy the prompt', 'helpIntro': 'Subscription page help and template', 'helpPromptHint': 'This prompt documents the current subscription page.',
  'helpServerEditHint': 'Change the connect address in default TLS + WS configs.', 'helpChatgptAddress': 'Set Address to chatgpt.com', 'helpCustomAddress': 'Save custom address', 'helpAddressLabel': 'Custom Address', 'helpAddressPlaceholder': 'Domain or IP (IPv4/IPv6)', 'helpPanelIpv6': 'Use panel domain IPv6', 'helpResetAddress': 'Use panel domain',
  'helpAddressSaved': 'Config address saved', 'helpAddressInvalid': 'Enter a valid domain or IP', 'helpIpv6NotFound': 'No IPv6 found', 'helpSettingsSaved': 'Settings saved',
  'helpEchChoose': 'Choose ECH', 'helpEchPickerTitle': 'ECH List', 'helpEchPickerHint': 'Choose a value.', 'helpEchPresetSaved': 'ECH preset saved',
  'helpFeatureHint': 'These options are added only to default TLS + WS configs.', 'helpEch': 'Add ECH', 'helpEchHint': 'Adds the ECH parameter.', 'helpFf': 'Add F+F', 'helpFfHint': 'Adds sample cipher and fragment settings.',
  'helpSpoof': 'Spoof', 'helpSpoofHint': 'Adds the snispoofing parameter.', 'helpSpoofSni': 'Fake SNI', 'helpSpoofIp': 'Target IP', 'helpSpoofMissing': 'Enter the SNI and IP to enable Spoof', 'helpSpoofSaved': 'Spoof settings saved',
  'advancedInboundHint': 'Select an inbound and edit its full JSON.', 'advancedJson': 'Inbound JSON', 'advancedSave': 'Save JSON', 'advancedSaved': 'Inbound JSON saved', 'advancedInvalidJson': 'Invalid JSON', 'advancedNoInbounds': 'No editable inbounds', 'defaultTlsWsInbound': 'Default TLS + WS',
  'notif.nav': 'Notifications', 'notif.title': 'Notifications', 'notif.refresh': 'Refresh', 'notif.date': 'Date & time', 'notif.hint': 'Each announcement is an accordion.', 'notif.loading': 'Loading...', 'notif.empty': 'No announcements.', 'notif.error': 'Could not fetch announcements.',
  'proxy.nav': 'Proxies', 'proxy.intro': 'Enter a proxy address and port. The panel tests SOCKS4, SOCKS5, HTTP, and HTTPS.', 'proxy.inboundHint': 'Traffic on this inbound exits through the selected proxy.',
  'proxy.add': 'Add proxy', 'proxy.host': 'Host (IP or domain)', 'proxy.port': 'Port', 'proxy.country': 'Country (optional)', 'proxy.auth': 'Username & password (optional)', 'proxy.authHint': 'Leave collapsed if no credentials.', 'proxy.username': 'Username', 'proxy.password': 'Password',
  'proxy.test': 'Test connection', 'proxy.testing': 'Testing', 'proxy.save': 'Save proxy', 'proxy.working': 'Working ✓', 'proxy.notWorking': 'Not working', 'proxy.ping': 'Ping', 'proxy.egress': 'Egress IP', 'proxy.error': 'Error', 'proxy.fillHost': 'Enter the host and port', 'proxy.testOk': 'Connection OK — save it now', 'proxy.testFail': 'Connection failed', 'proxy.testFirst': 'Run «Test connection» first', 'proxy.saved': 'Proxy saved', 'proxy.count': 'proxies', 'proxy.empty': 'No proxy yet',
  'proxy.ok': 'Active ✓', 'proxy.untested': 'Untested', 'proxy.off': 'Disabled', 'proxy.users': 'Users', 'proxy.lastTest': 'Last test', 'proxy.pingBtn': 'Ping', 'proxy.activate': 'Enable', 'proxy.deactivate': 'Disable', 'proxy.delete': 'Delete', 'proxy.activated': 'Enabled', 'proxy.deactivated': 'Disabled', 'proxy.deleted': 'Deleted', 'proxy.confirmDelete': 'Delete this proxy?',
  'wk.sectionHint': 'Three independent sections.', 'wk.tunnel': 'Tunnel', 'wk.reverse': 'Reverse', 'wk.worker': 'Worker', 'wk.tunnelHint': 'User → Railway → Cloudflare Worker → Site.', 'wk.reverseHint': 'User → Cloudflare Worker → Railway → Site.', 'wk.workerHint': 'Multi-location Worker.', 'wk.createReverse': 'Create Reverse inbound', 'wk.createInbound': 'Create Worker inbound', 'wk.reverseCreated': 'Reverse inbound created', 'wk.inboundCreated': 'Worker inbound created', 'wk.refresh': 'Refresh', 'wk.noLogs': 'No logs yet', 'wk.locations': 'Locations', 'wk.locationsHint': 'Each location is an egress.', 'wk.code': 'Code', 'wk.addLocation': 'Add location', 'wk.location': 'locations', 'wk.noLocations': 'No location yet', 'wk.auto': 'auto', 'wk.deploy': 'Deploy / Sync',
  'helpLogs': 'Panel Logs', 'helpLogsHint': 'Full panel log with filter and search.', 'helpLogsSearch': 'Search the log…', 'helpLogsAll': 'All kinds', 'helpLogsAllLevels': 'All levels', 'helpLogsEmpty': 'Loading…', 'helpLogsTotal': 'Total', 'helpLogsConn': 'Connections', 'helpLogsAuth': 'Auth', 'helpLogsErr': 'Errors', 'helpLogsNone': 'No entries match.',
  'vercelDomain.nav': 'Vercel Domain', 'vercelDomain.title': 'Connect Vercel Domain', 'vercelDomain.intro': 'Connect the Railway panel domain to a Vercel WebSocket relay.', 'vercelDomain.token': 'Vercel API Token', 'vercelDomain.getToken': 'Get token', 'vercelDomain.tokenHint': 'The token is stored on the panel server.', 'vercelDomain.create': 'Create Vercel Domain', 'vercelDomain.refresh': 'Refresh', 'vercelDomain.panelDomain': 'Railway panel domain', 'vercelDomain.domain': 'Vercel domain', 'vercelDomain.project': 'Project name', 'vercelDomain.ping': 'Ping from my internet', 'vercelDomain.delete': 'Delete Vercel Domain', 'vercelDomain.configTitle': 'Vercel WebSocket configuration', 'vercelDomain.configHint': "Choose a user to show that user's link.", 'vercelDomain.user': 'User', 'vercelDomain.noUsers': 'No valid configs.', 'vercelDomain.copyConfig': 'Copy configuration', 'vercelDomain.connectedHint': 'Domain connected.', 'vercelDomain.loadError': 'Could not load Vercel status.', 'vercelDomain.enterToken': 'Enter a Vercel token first.', 'vercelDomain.deploying': 'Creating…', 'vercelDomain.created': 'Vercel domain created.', 'vercelDomain.createError': 'Could not create.', 'vercelDomain.confirmDelete': 'Delete Vercel project?', 'vercelDomain.deleted': 'Project deleted.', 'vercelDomain.deleteError': 'Could not delete.', 'vercelDomain.noConfig': 'No configuration.', 'vercelDomain.legacyProtocol': 'This is an older XHTTP relay.', 'vercelDomain.notConnected': 'Create a domain first.', 'vercelDomain.pinging': 'Pinging…', 'vercelDomain.pingingHint': 'Measuring…', 'vercelDomain.pingFailed': 'Ping failed.', 'vercelDomain.clientPingResult': 'Browser ping: median {median} ms · min {min} ms · max {max} ms · successful {samples}/4',
  'tm.nav': 'Tunnel Management (TM)', 'tm.short': 'TM', 'tm.title': 'Tunnel Management', 'tm.server': 'Server', 'tm.configure': 'Configure', 'tm.servers': 'Servers', 'tm.links': 'Links', 'tm.entry': 'Entry', 'tm.exit': 'Exit', 'tm.lastApply': 'Last Apply', 'tm.addServer': 'Add Server', 'tm.serverHint': 'Every server is a real Node.', 'tm.editServer': 'Edit', 'tm.delServer': 'Delete', 'tm.pingServer': 'Test', 'tm.noServers': 'No server yet.', 'tm.loadErr': 'Could not load.', 'tm.addServerTitle': 'Add TM server', 'tm.editServerTitle': 'Edit server', 'tm.serverName': 'Display name', 'tm.serverDomain': 'Server panel domain', 'tm.spiKey': 'API key (server)', 'tm.keyHint': 'When editing, leave the key empty to keep the stored one.', 'tm.keyNeverShown': 'Stored key: {{masked}}', 'tm.confirmDelServer': 'Delete this server?', 'tm.serverAdded': 'Server added', 'tm.serverSaved': 'Server saved', 'tm.serverDeleted': 'Server deleted', 'tm.serverSavedOffline': 'Saved, but server unreachable', 'tm.remoteOffline': 'Server unreachable', 'tm.apply': 'Apply', 'tm.validate': 'Validate', 'tm.chainPing': 'Chain Ping', 'tm.clientPingAll': 'Client Ping', 'tm.curved': 'Curved', 'tm.straight': 'Straight', 'tm.connect': 'Connect', 'tm.disconnect': 'Disconnect', 'tm.configureHint': 'Long-press to move · tap for controls · keys connect two servers.', 'tm.selectSource': 'Source server', 'tm.selectDest': 'Destination server', 'tm.hlConnect': 'Source «{{src}}» — pick the destination', 'tm.duplicate': 'Connection already exists', 'tm.selfLoop': 'Cannot self-connect', 'tm.edgeDeleted': 'Connection removed', 'tm.edgeAdded': 'Connection created', 'tm.styleChanged': 'Line style changed', 'tm.validateOk': 'Topology valid', 'tm.applyDone': 'Tunnel applied', 'tm.applyFailed': 'Apply failed', 'tm.setDone': 'Active server set', 'tm.setClear': 'Active server cleared', 'tm.chainPingResult': 'Chain latency', 'tm.clientPingResult': 'Latency from this device', 'tm.offline': 'Offline / Timeout', 'tm.measuring': 'Measuring…', 'tm.samples': 'samples', 'tm.noChain': 'No active chain', 'tm.noGraph': 'Graph empty', 'tm.pinging': 'Pinging…', 'tm.generated': 'Generated client route', 'tm.generatedHint': 'Exactly what a user receives.', 'tm.regenerate': 'Regenerate', 'tm.copy': 'Copy',
  'tm.state.idle': 'Inactive', 'tm.state.active': 'Active', 'tm.state.dirty': 'Unapplied changes', 'tm.state.applied': 'Applied', 'tm.role.initial': 'Initial', 'tm.role.middle': 'Middle', 'tm.role.final': 'Final', 'tm.role.initial_final': 'Initial + Final', 'tm.role.idle': 'Idle',
  'tm.clientPingCardHint': 'Real measurement from this device.', 'tm.serverPingCardHint': 'Server-to-server from backend.', 'step.validated': 'Graph validated', 'step.chain': 'Chain resolved', 'step.server': 'Server', 'step.inbound': 'TM inbound updated', 'step.users': 'Users synced', 'step.applied': 'Configuration applied', 'step.rollback': 'Rolled back',
  'tm.clientPing': 'Client Ping', 'tm.serverPing': 'Server Ping', 'tm.fromThisDevice': 'from this device', 'tm.fromBackend': 'from backend', 'tm.clientPingEmpty': 'Not measured', 'tm.serverPingEmpty': 'Not measured', 'tm.cycle': 'Would create a loop', 'tm.destChanged': 'Destination changed', 'tm.errorApiKey': 'API key required', 'tm.fillDomain': 'Domain required', 'tm.selectEdgeFirst': 'Select a line first', 'tm.keyFrom': 'Make this server the source', 'tm.keyTo': 'Make this server the destination', 'tm.hlFrom': 'Source «{{src}}» — tap the key on the destination', 'tm.hlTo': 'Destination «{{dst}}» — tap the key behind the source', 'tm.invalid': 'Invalid topology', 'tm.applying': 'Applying…', 'tm.balancer': 'Balancer', 'tm.balancerHint': 'Multiple parallel servers.', 'tm.auto': 'Auto', 'tm.autoDone': 'Applied', 'tm.roleLabel': 'Role',
  'wpNoListen': 'No local TCP port; traffic arrives over HTTPS/WSS.'
});

function t(key) {
  var dict = I18N[LANG] || I18N.fa;
  return dict[key] != null ? dict[key] : (I18N.fa[key] != null ? I18N.fa[key] : key);
}

function applyLang() {
  document.documentElement.lang = LANG;
  document.documentElement.dir = 'rtl';
  document.querySelectorAll('[data-i18n]').forEach(function(el) {
    var k = el.getAttribute('data-i18n');
    if (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA') el.placeholder = t(k);
    else el.textContent = t(k);
  });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(function(el) { el.setAttribute('placeholder', t(el.getAttribute('data-i18n-placeholder'))); });
  document.querySelectorAll('[data-i18n-title]').forEach(function(el) { el.setAttribute('title', t(el.getAttribute('data-i18n-title'))); });
  var lb = $id('lang-label');
  if (lb) { lb.textContent = LANG === 'en' ? 'FA' : 'EN'; }
  updateGreeting();
  if (typeof ALL_USERS !== 'undefined' && ALL_USERS.length) renderUsers();
  if (typeof ALL_IB !== 'undefined' && ALL_IB.length) renderInbounds();
  if (typeof ALL_ADV_IB !== 'undefined' && ALL_ADV_IB.length) renderAdvancedInboundList();
  if ($id('dashboard-content') && $id('dashboard-content').innerHTML.indexOf('stat') > -1) loadDashboard();
  if (typeof renderNotifs === 'function') renderNotifs();
}
function toggleLang() {
  LANG = LANG === 'fa' ? 'en' : 'fa';
  localStorage.setItem('white-lang', LANG);
  applyLang();
  toast(t('dash'), 'ok');
}

async function api(url, opts) {
  opts = opts || {};
  opts.credentials = 'same-origin';
  opts.headers = opts.headers || {};
  if (opts.body && typeof opts.body === 'object' && !(opts.body instanceof FormData)) {
    opts.headers['Content-Type'] = 'application/json';
    opts.body = JSON.stringify(opts.body);
  }
  var r = await fetch(url, opts);
  var txt = await r.text();
  var data = null;
  var parseError = null;
  if (txt.trim()) {
    try { data = JSON.parse(txt); }
    catch (e) { parseError = e; }
  }
  if (!r.ok) {
    var detail = data && typeof data === 'object' && (data.detail || data.error || data.message);
    if (detail) throw new Error(typeof detail === 'string' ? detail : JSON.stringify(detail));
    if (parseError) throw new Error('پاسخ API معتبر نیست (HTTP ' + r.status + ')');
    throw new Error('خطای سرور: ' + r.status);
  }
  if (!txt.trim()) return {};
  if (parseError) throw new Error('پاسخ API JSON نیست (HTTP ' + r.status + ')');
  return data;
}

var toastTimer;
function toast(msg, type) {
  var el = $id('toast');
  el.textContent = msg;
  el.style.color = type === 'err' ? 'var(--red)' : type === 'ok' ? 'var(--green)' : 'var(--txt)';
  el.classList.add('show');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(function() { el.classList.remove('show'); }, 2800);
}

(function installInteractionLock(){
  var meta=document.querySelector('meta[name="viewport"]');
  if(meta) meta.setAttribute('content','width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no');
  else { meta=document.createElement('meta'); meta.name='viewport'; meta.content='width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no'; document.head.appendChild(meta); }
  document.addEventListener('wheel', function(e){ if(e.ctrlKey || e.metaKey) e.preventDefault(); }, {passive:false});
  document.addEventListener('keydown', function(e){
    var k=(e.key||'').toLowerCase();
    if((e.ctrlKey||e.metaKey) && ['+','-','=','0'].indexOf(e.key)>=0) e.preventDefault();
    if((e.ctrlKey||e.metaKey) && k==='c') { var t=e.target; var formField=t && /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName); if(!formField) e.preventDefault(); }
  });
  ['copy','cut','contextmenu','selectstart','dragstart'].forEach(function(evt){
    document.addEventListener(evt, function(e){ var t=e.target; if(t && /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName)) return; if(evt==='copy' && typeof window.getSelection==='function' && !window.getSelection().toString()) return; e.preventDefault(); });
  });
})();

document.addEventListener('DOMContentLoaded', function() {
  loadAccent();
  loadTheme();
  applyLang();
  checkAuth();
});

function applyThemeUI() {
  var isLight = document.documentElement.getAttribute('data-theme') === 'light';
  var ic = $id('theme-ic'), lb = $id('theme-label');
  if (ic) ic.textContent = isLight ? 'dark_mode' : 'light_mode';
  if (lb) lb.textContent = isLight ? (LANG === 'en' ? 'Dark' : 'تاریک') : (LANG === 'en' ? 'Light' : 'روشن');
}
function loadTheme() {
  document.documentElement.setAttribute('data-theme', selectedTheme || 'dark');
  applyThemeUI();
}
function setTheme(t) {
  selectedTheme = (t === 'light') ? 'light' : 'dark';
  document.documentElement.setAttribute('data-theme', selectedTheme);
  localStorage.setItem('white-theme', selectedTheme);
  applyThemeUI();
}
function toggleThemeMode() {
  setTheme(selectedTheme === 'light' ? 'dark' : 'light');
  toast(t(selectedTheme === 'light' ? 'lightMode' : 'darkMode'), 'ok');
}
function loadAccent() {}
function setAccent() {}

async function checkAuth() {
  try { var r = await api('/api/me'); if (r.authenticated) { showApp(); return; } } catch(e) {}
  window.location.href = '/login';
}
function showLogin() { $id('login-wrap').style.display = 'flex'; $id('app-wrap').classList.remove('active'); }
function showApp() { $id('login-wrap').style.display = 'none'; $id('app-wrap').classList.add('active'); loadAll(); updateGreeting(); }
async function doLogin() { /* handled by login page */ }
async function doLogout() {
  try { await api('/api/logout', {method: 'POST'}); } catch(e) {}
  localStorage.removeItem('white_session');
  window.location.href = '/login';
}
function updateGreeting() {
  var h = new Date().getHours(), g = h < 12 ? t('goodMorning') : h < 17 ? t('goodNoon') : t('goodEvening');
  var sep = LANG === 'en' ? ', <strong>' : '، <strong>';
  $id('grt').innerHTML = g + sep + t('admin') + '</strong>';
}

function toggleNavAccordion(forceOpen) { var wrap = $id('main-nav-accordion'); if (!wrap) return; var open = wrap.classList.contains('open'); if (typeof forceOpen === 'boolean') open = !forceOpen; wrap.classList.toggle('open', !open); var head = wrap.querySelector('.nav-accordion-head'); if (head) head.setAttribute('aria-expanded', String(!open)); }
function toggleSystemNavAccordion(forceOpen) { var wrap = $id('system-nav-accordion'); if (!wrap) return; var open = wrap.classList.contains('open'); if (typeof forceOpen === 'boolean') open = !forceOpen; wrap.classList.toggle('open', !open); var head = wrap.querySelector('.nav-accordion-head'); if (head) head.setAttribute('aria-expanded', String(!open)); }
function spdToggleAcc(id, btn) { var body = $id('acc-' + id); if (!body) return; var acc = body.parentElement; var open = !acc.classList.contains('open'); acc.classList.toggle('open', open); var head = btn || acc.querySelector('.spd-acc-head'); if (head) head.setAttribute('aria-expanded', String(open)); }
function toggleNodeAccordion(btn) { if (!btn) return; var acc = btn.parentElement; if (!acc || !acc.classList.contains('node-item-accordion')) return; var open = !acc.classList.contains('open'); acc.classList.toggle('open', open); btn.setAttribute('aria-expanded', String(open)); var body = acc.querySelector('.spd-acc-body'); if (body) body.setAttribute('aria-hidden', String(!open)); }

function switchTab(tab, el) {
  if (['users','inbounds','scanner','worker','bot','settings','proxy'].indexOf(tab) >= 0) toggleNavAccordion(true);
  if (['advanced','help','notif','vercel-domain'].indexOf(tab) >= 0) toggleSystemNavAccordion(true);
  document.querySelectorAll('.nav-item').forEach(function(n) { n.classList.remove('active'); });
  if (el) el.classList.add('active');
  var content = document.querySelector('.content');
  document.querySelectorAll('.tab-content').forEach(function(t) { t.classList.add('hidden'); });
  var target = $id('tab-' + tab);
  if (!target) return;
  if (content) { content.style.scrollBehavior = 'auto'; content.scrollTop = 0; content.scrollLeft = 0; }
  target.classList.remove('hidden');
  target.style.setProperty('margin-top', '0px', 'important');
  target.style.setProperty('top', '0px', 'important');
  target.style.setProperty('transform', 'none', 'important');
  if (tab === 'worker' || tab === 'contact' || tab === 'settings' || tab === 'nodes' || tab === 'advanced' || tab === 'help' || tab === 'notif' || tab === 'vercel-domain' || tab === 'proxy' || tab === 'suggest' || tab === 'tm') {
    target.style.setProperty('position', 'static', 'important');
    target.style.setProperty('flex', 'none', 'important');
  }
  var resetScroll = function() { if (content) { content.scrollTop = 0; content.scrollLeft = 0; } target.scrollTop = 0; target.scrollLeft = 0; document.documentElement.scrollTop = 0; document.body.scrollTop = 0; };
  resetScroll();
  requestAnimationFrame(resetScroll);
  stopLiveStats();
  if (tab !== 'advanced') stopHelpLogsAuto();
  loadTabData(tab);
  closeSidebar();
}

function openSidebar() { $id('sidebar').classList.add('open'); $id('sidebar-backdrop').classList.add('show'); }
function closeSidebar() { $id('sidebar').classList.remove('open'); $id('sidebar-backdrop').classList.remove('show'); }
function toggleSidebar() { $id('sidebar').classList.contains('open') ? closeSidebar() : openSidebar(); }

async function loadAll() { loadDashboard(); }
var TAB_BUSY = {};
var TAB_LOADERS = {
  dashboard: function () { return loadDashboard(); },
  users: function () { return Promise.all([loadUsers(), ensureInbounds(), ensureWorkerPool()]); },
  inbounds: function () { return loadInbounds(); },
  scanner: function () { return scnInit(); },
  worker: function () { return loadWorker(); },
  'vercel-domain': function () { return loadVercelDomain(); },
  tunnel: function () { return loadTunnel(); },
  nodes: function () { return loadNodes(); },
  tm: function () { return loadTM(); },
  bot: function () { return loadBot(); },
  proxy: function () { return loadProxies(); },
  advanced: function () { startHelpLogsAuto(); return Promise.all([loadAdvancedInbounds(), loadHelpSettings(), loadHelpLogs()]); },
  settings: function () { return loadServerInfo(); },
  notif: function () { return loadNotifs(false); },
  suggest: function () {}, help: function () {}, contact: function () {}
};
function loadTabData(tab) {
  var run = TAB_LOADERS[tab]; if (!run) return Promise.resolve();
  if (TAB_BUSY[tab]) return TAB_BUSY[tab];
  var p = Promise.resolve().then(run).catch(function (e) { console.error('tab load failed:', tab, e); }).then(function (v) { TAB_BUSY[tab] = null; return v; });
  TAB_BUSY[tab] = p;
  return p;
}
var INBOUNDS_PROMISE = null;
function ensureInbounds() { if (INBOUNDS_PROMISE) return INBOUNDS_PROMISE; INBOUNDS_PROMISE = settle(function () { return loadInbounds(); }, function (v) { INBOUNDS_PROMISE = null; return v; }); return INBOUNDS_PROMISE; }
var WORKER_POOL_PROMISE = null;
function ensureWorkerPool() { if (WORKER_POOL_PROMISE) return WORKER_POOL_PROMISE; if (typeof loadWorker !== 'function') return Promise.resolve(); WORKER_POOL_PROMISE = settle(loadWorker, function (v) { WORKER_POOL_PROMISE = null; return v; }); return WORKER_POOL_PROMISE; }
function settle(fn, done) { return Promise.resolve().then(fn).then(done, function (e) { done(undefined); throw e; }); }

var liveInterval = null;
var DASH_HISTORY = {cpu: [], ram: [], swap: [], disk: []};
var DASH_IPS_VISIBLE = false;
var DASH_IP_VALUES = {ipv4: '', ipv6: ''};
var DASH_NET_PREVIOUS = null;
var DASH_NET_HISTORY_LIMIT = 22;
var DASH_NET_HISTORY = {download: [], upload: []};

function startLiveStats() { if (liveInterval) clearInterval(liveInterval); refreshLiveStats(); liveInterval = setInterval(refreshLiveStats, 4000); }
function stopLiveStats() { if (liveInterval) { clearInterval(liveInterval); liveInterval = null; } }

async function loadDashboard() {
  try {
    var s = await api('/stats');
    var html = '<div class="stats">';
    html += statCard(iconUsers, '<span id="dash-users">' + s.active_users + ' / ' + s.total_users + '</span>', t('activeUsers'));
    html += statCard(iconTraffic, '<span id="dash-traffic">' + fmtBytes((s.total_traffic_used_gb || s.traffic_usage_gb || 0) * 1073741824) + '</span>', t('trafficUsed'));
    html += statCard(iconConn, '<span id="dash-conn">' + (s.active_connections || '۰') + '</span>', t('activeConns'));
    html += statCard(iconUptime, '<span id="dash-uptime">...</span>', t('uptime'));
    html += '</div>';
    html += '<section class="panel dash-system-card">';
    html += '<div class="ds-title"><span>SYSTEM STATUS</span></div>';
    html += dashResourceRow('cpu', 'CPU', 'developer_board', '#62c8b2');
    html += dashResourceRow('ram', 'RAM', 'memory', '#62c8b2');
    html += dashResourceRow('swap', 'Swap', 'swap_vert', '#62c8b2');
    html += dashResourceRow('disk', 'Storage', 'storage', '#4489e4');
    html += '<div class="ds-section"><div class="ds-section-title">CONNECTION STATS</div>';
    html += '<div class="ds-conn-grid"><div><div class="ds-conn-label"><span class="material-symbols-outlined">cable</span>TCP</div><b class="ds-conn-count" id="ds-tcp">0</b></div>';
    html += '<div><div class="ds-conn-label udp"><span class="material-symbols-outlined">cable</span>UDP</div><b class="ds-conn-count" id="ds-udp">0</b></div></div>';
    html += '<div class="ds-conn-bar"><span class="ds-conn-tcp" id="ds-tcp-bar"></span><span class="ds-conn-udp" id="ds-udp-bar"></span></div></div>';
    html += '<div class="ds-ips"><div class="ds-ips-heading"><div class="ds-section-title" style="margin:0">IP ADDRESSES</div>';
    html += '<button class="ds-ip-toggle" id="ds-ip-toggle" type="button" onclick="toggleDashIps()"><span class="material-symbols-outlined">visibility_off</span></button></div>';
    html += '<div class="ds-ip-row"><span>IPv4</span><b class="ds-ip-value" id="ds-ipv4">—</b></div>';
    html += '<div class="ds-ip-row"><span>IPv6</span><b class="ds-ip-value" id="ds-ipv6">—</b></div></div>';
    html += '</section>';
    html += '<section class="panel dash-bandwidth-card"><div class="dash-bw-grid">';
    html += dashBandwidthColumn('download', t('download'));
    html += dashBandwidthColumn('upload', t('upload'));
    html += '</div></section>';
    $id('dashboard-content').innerHTML = html;
    DASH_HISTORY = {cpu: [], ram: [], swap: [], disk: []};
    DASH_IPS_VISIBLE = false;
    DASH_NET_PREVIOUS = null;
    DASH_NET_HISTORY = {download: [], upload: []};
    startLiveStats();
  } catch(e) { $id('dashboard-content').innerHTML = '<div class="empty">' + t('errLoadStats') + '</div>'; }
}

function dashBandwidthColumn(key, label) {
  return '<div class="dash-bw-col"><div class="dash-bw-heading"><strong id="ds-bw-' + key + '-rate">0.0</strong><span class="dash-bw-unit">Mb/s</span><span class="dash-bw-label">' + label + '</span><span class="dash-bw-marker" aria-hidden="true"></span></div><div class="dash-bw-bars" id="ds-bw-' + key + '-bars" aria-hidden="true"></div><div class="dash-bw-peak" id="ds-bw-' + key + '-peak"></div></div>';
}
function dashFormatPeak(rate) { var value = (Number(rate) || 0).toFixed(1); return LANG === 'fa' ? 'اوج ' + value + ' Mb/s' : 'Peak ' + value + ' Mb/s'; }
function renderDashBandwidthColumn(key) {
  var history = DASH_NET_HISTORY[key] || [];
  var current = history.length ? history[history.length - 1] : 0;
  var peak = history.length ? Math.max.apply(Math, history) : 0;
  var valueEl = $id('ds-bw-' + key + '-rate'); var peakEl = $id('ds-bw-' + key + '-peak'); var barsEl = $id('ds-bw-' + key + '-bars');
  if (valueEl) valueEl.textContent = current.toFixed(1);
  if (peakEl) peakEl.textContent = dashFormatPeak(peak);
  if (!barsEl) return;
  var max = Math.max(peak, 0.01); var padding = Math.max(0, DASH_NET_HISTORY_LIMIT - history.length); var bars = [];
  for (var i = 0; i < DASH_NET_HISTORY_LIMIT; i++) {
    var sampleIndex = i - padding; var sample = sampleIndex >= 0 ? history[sampleIndex] : 0;
    var height = sample > 0 ? Math.max(9, sample / max * 100) : 5;
    bars.push('<span' + (i === DASH_NET_HISTORY_LIMIT - 1 ? ' class="current"' : '') + ' style="height:' + height.toFixed(1) + '%"></span>');
  }
  barsEl.innerHTML = bars.join('');
}
function updateDashBandwidth(stats) {
  var sent = Number(stats.net_sent_bytes); var received = Number(stats.net_recv_bytes);
  if (!Number.isFinite(sent)) sent = (Number(stats.net_sent_mb) || 0) * 1024 * 1024;
  if (!Number.isFinite(received)) received = (Number(stats.net_recv_mb) || 0) * 1024 * 1024;
  var now = Date.now(); var rates = {download: 0, upload: 0};
  if (DASH_NET_PREVIOUS) {
    var elapsed = (now - DASH_NET_PREVIOUS.at) / 1000;
    if (elapsed > 0 && elapsed <= 60) {
      rates.download = Math.max(0, received - DASH_NET_PREVIOUS.received) * 8 / elapsed / 1000000;
      rates.upload = Math.max(0, sent - DASH_NET_PREVIOUS.sent) * 8 / elapsed / 1000000;
    }
  }
  DASH_NET_PREVIOUS = {at: now, sent: sent, received: received};
  ['download', 'upload'].forEach(function(key) {
    var history = DASH_NET_HISTORY[key]; history.push(rates[key]);
    if (history.length > DASH_NET_HISTORY_LIMIT) history.shift();
    renderDashBandwidthColumn(key);
  });
}
function dashResourceRow(key, label, icon, color) {
  return '<div class="ds-row"><div class="ds-gauge" id="ds-gauge-' + key + '" style="--pct:0;--gauge-color:' + color + '"><span class="material-symbols-outlined">' + icon + '</span></div><div class="ds-main"><div class="ds-line"><strong>' + label + '</strong><b id="ds-' + key + '-pct">0%</b></div><div class="ds-sub" id="ds-' + key + '-sub">—</div><svg class="ds-spark" viewBox="0 0 320 24" preserveAspectRatio="none" aria-hidden="true"><polyline id="ds-' + key + '-spark" points="0,12 320,12"></polyline></svg></div></div>';
}
function dashSetResource(key, percent, detail) {
  var pct = Math.max(0, Math.min(100, Number(percent) || 0));
  var gauge = $id('ds-gauge-' + key); if (gauge) gauge.style.setProperty('--pct', pct);
  var label = $id('ds-' + key + '-pct'); if (label) label.textContent = pct.toFixed(2).replace(/0+$/, '').replace(/\.$/, '') + '%';
  var sub = $id('ds-' + key + '-sub'); if (sub) sub.textContent = detail;
  var history = DASH_HISTORY[key] || (DASH_HISTORY[key] = []); history.push(pct); if (history.length > 32) history.shift();
  var line = $id('ds-' + key + '-spark');
  if (line) {
    var low = Math.min.apply(Math, history), high = Math.max.apply(Math, history), spread = Math.max(high - low, 1);
    var points = history.map(function(v, i) { var x = history.length < 2 ? 320 : i * 320 / (history.length - 1); var y = 20 - ((v - low) / spread) * 15; return x.toFixed(1) + ',' + y.toFixed(1); }).join(' ');
    if (history.length === 1) { var flatY = (20 - ((history[0] - low) / spread) * 15).toFixed(1); points = '0,' + flatY + ' 320,' + flatY; }
    line.setAttribute('points', points);
  }
}
function dashGb(value) { return (Number(value) || 0).toFixed(2); }
function dashMaskIp(value, family) { if (!value) return '—'; return family === 'ipv4' ? '•••.•••.•••.•••' : '••••:••••:••••:••••'; }
function renderDashIps() {
  var v4 = $id('ds-ipv4'), v6 = $id('ds-ipv6'), toggle = $id('ds-ip-toggle');
  if (v4) v4.textContent = DASH_IPS_VISIBLE ? (DASH_IP_VALUES.ipv4 || '—') : dashMaskIp(DASH_IP_VALUES.ipv4, 'ipv4');
  if (v6) v6.textContent = DASH_IPS_VISIBLE ? (DASH_IP_VALUES.ipv6 || '—') : dashMaskIp(DASH_IP_VALUES.ipv6, 'ipv6');
  if (toggle) toggle.innerHTML = '<span class="material-symbols-outlined">' + (DASH_IPS_VISIBLE ? 'visibility' : 'visibility_off') + '</span>';
}
function toggleDashIps() { DASH_IPS_VISIBLE = !DASH_IPS_VISIBLE; renderDashIps(); }
async function refreshLiveStats() {
  try {
    var st = await api('/api/server/stats');
    dashSetResource('cpu', st.cpu_percent, (st.cpu_count || '?') + ' Cores');
    dashSetResource('ram', st.ram_percent, dashGb(st.ram_used_gb) + ' GB / ' + dashGb(st.ram_total_gb) + ' GB');
    dashSetResource('swap', st.swap_percent, dashGb(st.swap_used_gb) + ' GB / ' + dashGb(st.swap_total_gb) + ' GB');
    dashSetResource('disk', st.disk_percent, dashGb(st.disk_used_gb) + ' GB / ' + dashGb(st.disk_total_gb) + ' GB');
    var duptEl = $id('dash-uptime'); if (duptEl) duptEl.textContent = fmtUptime(st.uptime_seconds);
    var connEl = $id('dash-conn'); if (connEl) connEl.textContent = String(st.active_connections || 0);
    var trEl = $id('dash-traffic'); if (trEl) trEl.textContent = (typeof st.total_traffic_used_gb !== 'undefined' ? st.total_traffic_used_gb + ' GB' : fmtBytes((st.total_traffic_used_tb || 0) * 1099511627776));
    var usrEl = $id('dash-users'); if (usrEl) usrEl.textContent = (st.active_users || 0) + ' / ' + (st.total_users || 0);
    var tcp = Number(st.tcp_connections != null ? st.tcp_connections : st.active_connections) || 0;
    var udp = Number(st.udp_connections) || 0;
    var tcpEl = $id('ds-tcp'); if (tcpEl) tcpEl.textContent = String(tcp);
    var udpEl = $id('ds-udp'); if (udpEl) udpEl.textContent = String(udp);
    var totalProto = tcp + udp, tcpPct = totalProto ? tcp * 100 / totalProto : 50;
    var tcpBar = $id('ds-tcp-bar'); if (tcpBar) tcpBar.style.width = tcpPct + '%';
    var udpBar = $id('ds-udp-bar'); if (udpBar) udpBar.style.width = (100 - tcpPct) + '%';
    updateDashBandwidth(st);
    DASH_IP_VALUES.ipv4 = String(st.public_ipv4 || st.public_ip || '');
    DASH_IP_VALUES.ipv6 = String(st.public_ipv6 || '');
    renderDashIps();
  } catch(e) {}
}
function statCard(icon, val, label) { return '<div class="stat"><div class="stat-icon">' + icon + '</div><div class="stat-v">' + val + '</div><div class="stat-l">' + label + '</div></div>'; }

async function loadUsers() {
  try { var d = await api('/api/users'); ALL_USERS = d.users || []; renderUsers(); }
  catch(e) { $id('users-list').innerHTML = '<div class="empty">' + t('errLoadUsers') + '</div>'; }
}
function renderUsers() {
  var el = $id('users-list');
  if (!ALL_USERS.length) { el.innerHTML = '<div class="empty">' + t('noUsers') + '</div>'; return; }
  var html = '<div class="spd-acc-list">';
  for (var i = 0; i < ALL_USERS.length; i++) { html += _userRow(ALL_USERS[i]); }
  html += '</div>';
  el.innerHTML = html;
}
function _userTraffic(u) {
  var used = u.traffic_used_bytes || 0, limit = u.traffic_limit_bytes || 0;
  var pct = limit > 0 ? Math.min((used / limit) * 100, 100) : 0;
  return '<div style="font-size:11px">' + fmtBytes(used) + ' / ' + (limit ? fmtBytes(limit) : t('unlimited')) + '</div><div class="pbar"><div class="pbar-f" style="width:' + pct + '%"></div></div>';
}
function _userExp(u) {
  var exp = u.expire_at ? Math.ceil((new Date(u.expire_at) - new Date()) / 86400000) : null;
  return exp === null ? t('unlimited') : exp <= 0 ? '<span style="color:var(--red)">' + t('expired') + '</span>' : exp + t('daysUnit');
}
function _userStatus(u) { return u.status === 'active' ? '<span class="badge badge-g">' + t('active') + '</span>' : '<span class="badge badge-r">' + t('inactive') + '</span>'; }
function _userAccHead(u, idx) {
  return '<button type="button" class="spd-acc-head" onclick="spdToggleAcc(\'ua-' + u.user_id + '\',this)" aria-expanded="false"><span class="spd-acc-title"><strong>' + esc(u.username) + '</strong></span><span class="spd-acc-badges">' + _userStatus(u) + '</span><span class="material-symbols-outlined spd-acc-arrow">expand_more</span></button>';
}
function _userAccBody(u) {
  var connTxt = u.concurrent_connections ? u.concurrent_connections : t('unlimited');
  var speedVal = parseFloat(u.speed_limit_mbps) || 0;
  var speedTxt = speedVal > 0 ? speedVal + ' Mbps' : t('unlimited');
  return '<div class="spd-acc-body" id="acc-ua-' + u.user_id + '">' +
    '<div class="spd-acc-sec"><div class="spd-acc-sec-t">' + t('volume') + '</div>' + _userTraffic(u) + '</div>' +
    '<div class="spd-acc-kv"><span>' + t('expiry') + '</span><b>' + _userExp(u) + '</b></div>' +
    '<div class="spd-acc-kv"><span>' + t('connTh') + '</span><b>' + connTxt + '</b></div>' +
    '<div class="spd-acc-kv"><span>' + t('speedLimit') + '</span><b>' + speedTxt + '</b></div>' +
    '<div class="spd-acc-btns user-action-btns">' +
      '<button class="btn btn-sm btn-g" onclick="viewCfg(\'' + esc(u.config_uuid || '') + '\')" title="Config">'+iconCfg+'</button>' +
      '<button class="btn btn-sm btn-g" onclick="viewSub(\'' + (u.config_uuid || '') + '\')" title="Sub Link">'+iconLink+'</button>' +
      '<button class="btn btn-sm btn-g" onclick="viewQR(\'' + u.user_id + '\')" title="QR Code">'+iconQr+'</button>' +
      '<button class="btn btn-sm btn-g" onclick="editUser(\'' + u.user_id + '\')" title="Edit">'+iconEdit+'</button>' +
      '<button class="btn btn-sm btn-g" onclick="resetUserTraffic(\'' + u.user_id + '\')" title="Reset Traffic">'+iconReset+'</button>' +
      '<button class="btn btn-sm btn-g" onclick="toggleUser(\'' + u.user_id + '\')" title="Toggle">'+iconToggle+'</button>' +
      '<button class="btn btn-sm btn-g" onclick="deleteUser(\'' + u.user_id + '\')" title="Delete">'+iconTrash+'</button>' +
    '</div></div>';
}
function _userRow(u) { return '<div class="spd-acc">' + _userAccHead(u) + _userAccBody(u) + '</div>'; }
async function resetUserTraffic(uid) { if (!confirm(t('resetTraffic'))) return; try { await api('/api/users/' + uid + '/reset', {method: 'PATCH'}); toast(t('toast.userUpdated'), 'ok'); loadUsers(); } catch(e) { toast(e.message, 'err'); } }
if (window.matchMedia) { window.matchMedia('(max-width: 640px)').addListener(function(m) { if ($id('users-list') && ALL_USERS.length) renderUsers(); }); }

async function openCreateUser() {
  cuInboundState = { picks: {}, countries: {} };
  await ensureInbounds();
  if (ALL_IB.some(function(ib) { return (ib.protocol || '').toLowerCase() === 'worker'; })) { try { await ensureWorkerPool(); } catch(e) {} }
  populateInboundSelects();
  $id('cu-name').value = ''; $id('cu-vol').value = ''; $id('cu-vol-u').value = 'GB'; $id('cu-exp').value = ''; $id('cu-exp-u').value = 'days'; $id('cu-conn').value = ''; $id('cu-speed').value = '0';
  openModal('mod-user');
}
async function submitCreateUser(e) {
  e.preventDefault();
  var btn = $id('cu-submit'); btn.disabled = true; btn.textContent = t('creatingUser');
  var v = parseFloat($id('cu-vol').value) || 0, vu = $id('cu-vol-u').value;
  var ex = parseInt($id('cu-exp').value) || 0, eu = $id('cu-exp-u').value;
  var inbound_ids = selectedInboundIds();
  var cuSw = cuCustomStates(); var cuCountries = cuWorkerCountries();
  var cfIds = inbound_ids.filter(function(i){ return cuSw[i] && cuSw[i].cf; });
  var rwIds = inbound_ids.filter(function(i){ return cuSw[i] && cuSw[i].railway; });
  var allCountries = [];
  Object.keys(cuCountries).forEach(function(ibId) { (cuCountries[ibId] || []).forEach(function(c) { if (allCountries.indexOf(c) === -1) allCountries.push(c); }); });
  var firstCountry = allCountries[0] || '';
  try {
    var r = await api('/api/users', {method: 'POST', body: {
      username: $id('cu-name').value, traffic_limit_gb: vu === 'GB' ? v : v / 1024, expire_days: eu === 'months' ? ex * 30 : ex,
      concurrent_connections: parseInt($id('cu-conn').value) || 0, speed_limit_mbps: parseFloat($id('cu-speed').value) || 0,
      inbound_ids: inbound_ids, proxy_countries: allCountries, proxy_country: firstCountry, proxy_ip_enabled: !!allCountries.length,
      custom_ip_inbounds: { cf: cfIds, railway: rwIds }
    }});
    toast(r.username + ' ' + t('userCreated'), 'ok');
    closeModal('mod-user');
    loadTabData('users');
  } catch(e) { toast(e.message || t('toast.err'), 'err'); }
  btn.disabled = false; btn.textContent = t('submitUser');
}

var EU_INBOUND_PICKS = {};
var EU_COUNTRIES = {};
function renderEditInboundPicker(u) {
  var wrap = $id('eu-inbound'); if (!wrap) return;
  var cur = u.inbound_ids || (u.inbound_id ? [u.inbound_id] : []);
  EU_INBOUND_PICKS = {};
  cur.forEach(function(i) { EU_INBOUND_PICKS[i] = true; });
  if (!Object.keys(EU_COUNTRIES).length) {
    var existingCountries = u.proxy_countries || [];
    ALL_IB.forEach(function(ib) { if ((ib.protocol || '').toLowerCase() === 'worker') { EU_COUNTRIES[ib.inbound_id] = { worker: existingCountries.slice() }; } });
  }
  var html = '';
  ALL_IB.forEach(function(ib) {
    var id = ib.inbound_id; var selected = EU_INBOUND_PICKS[id] === true;
    var type = (ib.protocol || 'vless').toLowerCase();
    var isWebProxy = type === 'telegram-webproxy' || String(ib._managed_kind || '') === 'webproxy';
    var proto = isWebProxy ? (ib.display_protocol || 'Telegram Web Proxy') : type.toUpperCase();
    var meta = isWebProxy ? (proto + ' · TLS + WS · ' + (ib.status || '')) : (proto + ' · ' + (ib.network || 'ws') + ' · ' + (ib.security || 'tls'));
    html += '<div class="ib-pick' + (selected ? ' selected' : '') + '" data-id="' + esc(id) + '" onclick="euToggleIbPick(this)">' +
      '<div class="ib-pick-ic"><span class="material-symbols-outlined">' + (type === 'worker' ? 'bolt' : isWebProxy ? 'language' : type === 'reality' ? 'shield' : 'cloud') + '</span></div>' +
      '<div class="ib-pick-info"><div class="ib-pick-name">' + esc(inboundDisplayName(ib)) + '</div><div class="ib-pick-meta">' + esc(meta) + '</div></div>' +
      '<div class="ib-pick-check"><span class="material-symbols-outlined">check</span></div></div>';
    if (selected && type === 'worker') {
      var sel = EU_COUNTRIES[id] || {}; var chipHtml = '';
      cuWorkerProxies().forEach(function(p) {
        var flag = window.cuCodeToFlag ? cuCodeToFlag(p.code) : '';
        var checked = (sel.worker || []).indexOf(p.code) !== -1 ? ' selected' : '';
        chipHtml += '<span class="country-tick' + checked + '" data-code="' + esc(p.code) + '" onclick="euPickWorkerCountry(\'' + esc(id) + '\',this)"><span class="ck material-symbols-outlined">check</span><span class="country-name">' + flag + ' ' + esc(p.country || p.code) + '</span></span>';
      });
      html += '<div style="flex-basis:100%;padding:8px 12px;background:var(--bg2);border:1px solid var(--border);border-radius:var(--radius);margin-top:4px"><div style="font-size:11px;color:var(--txt2);margin-bottom:6px">' + (t('workerCountries') || 'Worker Countries') + '</div><div style="display:flex;flex-wrap:wrap;gap:6px;direction:ltr;text-align:left">' + chipHtml + '</div></div>';
    }
  });
  wrap.innerHTML = html || '<span style="font-size:11px;color:var(--txt2)">' + t('noInboundOpt') + '</span>';
}
function euPickWorkerCountry(ibId, el) {
  el.classList.toggle('selected');
  var arr = EU_COUNTRIES[ibId] = EU_COUNTRIES[ibId] || {}; arr.worker = arr.worker || [];
  var code = el.getAttribute('data-code'); var idx = arr.worker.indexOf(code);
  if (el.classList.contains('selected')) { if (idx === -1) arr.worker.push(code); } else { if (idx !== -1) arr.worker.splice(idx, 1); }
}
function euToggleIbPick(el) { var id = el.getAttribute('data-id'); if (EU_INBOUND_PICKS[id]) delete EU_INBOUND_PICKS[id]; else EU_INBOUND_PICKS[id] = true; renderEditInboundPicker(ALL_USERS.find(function(x) { return x.user_id === $id('eu-id').value; }) || {}); }
function selectedEditInboundIds() { return Object.keys(EU_INBOUND_PICKS).filter(function(k) { return EU_INBOUND_PICKS[k]; }); }
async function editUser(uid) {
  await ensureInbounds();
  var u = ALL_USERS.find(function(x) { return x.user_id === uid; }); if (!u) return;
  EU_COUNTRIES = {};
  $id('eu-id').value = uid; $id('eu-name').value = u.username || '';
  $id('eu-vol').value = (u.traffic_limit_bytes ? (u.traffic_limit_bytes / 1073741824).toFixed(0) : '0');
  $id('eu-exp').value = u.expire_at ? Math.ceil((new Date(u.expire_at) - new Date()) / 86400000) : 0;
  $id('eu-status').value = u.status || 'active';
  $id('eu-conn').value = (u.concurrent_connections != null ? u.concurrent_connections : 0);
  $id('eu-speed').value = (u.speed_limit_mbps != null ? u.speed_limit_mbps : 0);
  renderEditInboundPicker(u);
  openModal('mod-edit-user');
}
async function submitEditUser(e) {
  e.preventDefault();
  var uid = $id('eu-id').value;
  var allEuCountries = [];
  Object.keys(EU_COUNTRIES).forEach(function(ibId) { var c = EU_COUNTRIES[ibId] || {}; (c.worker || []).forEach(function(code) { if (allEuCountries.indexOf(code) === -1) allEuCountries.push(code); }); });
  try {
    await api('/api/users/' + uid, {method: 'PATCH', body: {
      username: $id('eu-name').value, traffic_limit_gb: $id('eu-vol').value ? parseFloat($id('eu-vol').value) : 0,
      expire_days: parseInt($id('eu-exp').value) || 0, concurrent_connections: parseInt($id('eu-conn').value) || 0,
      speed_limit_mbps: parseFloat($id('eu-speed').value) || 0, status: $id('eu-status').value,
      reset_traffic: $id('eu-reset') ? $id('eu-reset').checked : false,
      inbound_ids: selectedEditInboundIds(), proxy_countries: allEuCountries, proxy_country: allEuCountries[0] || ''
    }});
    toast(t('toast.userUpdated'), 'ok'); closeModal('mod-edit-user'); loadTabData('users');
  } catch(e) { toast(e.message || t('toast.err'), 'err'); }
}
async function toggleUser(uid) { try { await api('/api/users/' + uid + '/toggle', {method: 'PATCH'}); loadUsers(); toast(t('toast.userToggled'), 'ok'); } catch(e) { toast(e.message, 'err'); } }
async function deleteUser(uid) { if (!confirm(t('confirmDeleteUser'))) return; try { await api('/api/users/' + uid, {method: 'DELETE'}); toast(t('toast.userDeleted'), 'ok'); loadTabData('users'); } catch(e) { toast(e.message, 'err'); } }

async function viewCfg(configUuid) {
  try {
    var d = await api('/api/sub/' + encodeURIComponent(configUuid));
    var list = Array.isArray(d.configs) ? d.configs : [];
    list = list.filter(function(c){ return c && c.indexOf('%F0%9F%93%8A') === -1; });
    curCfg = list.join('\n');
    $id('cfg-code').textContent = curCfg || t('noCfg');
    openModal('mod-cfg');
  } catch(e) { toast(e.message, 'err'); }
}
function copyCfg() { navigator.clipboard.writeText(curCfg).then(function() { toast(t('copied'), 'ok'); }).catch(function() { toast(t('errCopy'), 'err'); }); }
async function viewSub(configUuid) {
  try {
    var d = await api('/api/sub/' + encodeURIComponent(configUuid));
    var uuid = d.config_uuid || configUuid;
    var subUrl = window.location.origin + '/link/' + uuid;
    if (navigator.clipboard) { await navigator.clipboard.writeText(subUrl); toast('لینک ساب کپی شد', 'ok'); } else { window.open(subUrl, '_blank'); }
  } catch (e) { console.error(e); }
}
var qrSubscriptionUrl = '';
function viewQR(userId) {
  var user = ALL_USERS.find(function(u) { return String(u.user_id) === String(userId); });
  var configUuid = user && (user.config_uuid || user.subscription_uuid);
  qrSubscriptionUrl = configUuid ? window.location.origin + '/link/' + encodeURIComponent(configUuid) : '';
  $id('qr-img').src = '/api/users/' + encodeURIComponent(userId) + '/qr?t=' + Date.now();
  openModal('mod-qr');
}
function openQrSubscription() { if (!qrSubscriptionUrl) return; window.open(qrSubscriptionUrl, '_blank', 'noopener,noreferrer'); }

async function loadInbounds() {
  try { var d = await api('/api/inbounds'); ALL_IB = d.inbounds || []; renderInbounds(); populateInboundSelects(); }
  catch(e) { var list = $id('inbounds-list'); if (list) list.innerHTML = '<div class="empty">' + esc((t('errLoadInbounds') || 'Error') + ': ' + (e.message || '')) + '</div>'; }
}
function isWebProxyIb(ib) { var p = String((ib && ib.protocol) || '').toLowerCase(); return p === 'telegram-webproxy' || String((ib && ib._managed_kind) || '') === 'webproxy'; }
function _ibDetail(ib) {
  var proto = (ib.protocol || 'vless').toLowerCase();
  var d = '<div style="font-size:11px;color:var(--txt2);line-height:1.8">';
  if (isWebProxyIb(ib)) {
    var st = ib.status || 'not_ready';
    d += t('wpNoListen') + '<br>';
  } else if (proto === 'telegram') {
    var tg = ib.telegram_settings || {};
    d += 'Domain: ' + (tg.external_domain || ib.external_domain || '—') + '<br>';
    d += 'Internal Port: ' + (tg.internal_port || ib.port || '—') + '<br>';
    d += 'External Port: ' + (tg.external_port || ib.external_port || '—');
  } else if (proto === 'worker') {
    d += t('domain') + ': ' + (ib.external_domain || '—') + '<br>' + t('port') + ': 443 (Worker)';
  } else if (ib.protocol === 'reality' || ib.security === 'reality') {
    d += t('domain') + ': ' + (ib.external_domain || '—') + '<br>' + t('port') + ': ' + (ib.port || '—');
  } else {
    d += t('domain') + ': ' + (ib.external_domain || '—') + ' (خودکار)<br>' + t('port') + ': 443 / relay';
  }
  if (proto !== 'telegram' && !isWebProxyIb(ib)) { d += '<br>' + t('secNetwork') + ': ' + (ib.network || 'tcp') + ' | ' + t('security') + ': ' + (ib.security || 'none'); }
  d += '<br>' + t('usersCount') + ': ' + (ib.users_count || 0);
  d += '</div>'; return d;
}
function renderInbounds() {
  var el = $id('inbounds-list');
  if (!ALL_IB.length) { el.innerHTML = '<div class="empty">' + t('noInbounds') + '</div>'; return; }
  var html = '<div class="spd-acc-list">';
  for (var i = 0; i < ALL_IB.length; i++) {
    var ib = ALL_IB[i]; var proto = (ib.protocol || 'vless').toLowerCase();
    var inboundTitle = inboundDisplayName(ib); var protoBadge = proto === 'reality' ? 'badge-p' : 'badge-b';
    html += '<div class="spd-acc"><button type="button" class="spd-acc-head" onclick="spdToggleAcc(\'ia-' + ib.inbound_id + '\',this)" aria-expanded="false"><span class="spd-acc-title"><strong>' + esc(inboundTitle) + '</strong></span><span class="spd-acc-badges"><span class="badge ' + protoBadge + '">' + proto.toUpperCase() + '</span></span><span class="material-symbols-outlined spd-acc-arrow">expand_more</span></button><div class="spd-acc-body" id="acc-ia-' + ib.inbound_id + '">' + _ibDetail(ib) +
      (ib.proxy_managed ? '' : isWebProxyIb(ib) ? '' :
        '<div class="spd-acc-btns"><button class="btn btn-sm btn-g" onclick="editInbound(\'' + ib.inbound_id + '\')">' + t('edit') + '</button>' + ((ib.system || ib.inbound_id === 'Node') ? '' : '<button class="btn btn-sm btn-g" onclick="deleteInbound(\'' + ib.inbound_id + '\')">' + t('delete') + '</button>') + '</div>') +
      '</div></div>';
  }
  html += '</div>'; el.innerHTML = html;
}
function populateInboundSelects() {
  if (!Object.keys(cuInboundState.picks).length) {
    cuInboundState.picks = {};
    ALL_IB.forEach(function(ib){ cuInboundState.picks[ib.inbound_id] = !ib.proxy_managed; });
    cuInboundState.countries = {};
  }
  cuRenderInboundOpts();
}
function toggleIbPick(el) { cuToggleIbPick(el); }
function selectedInboundIds() {
  var ids = []; Object.keys(cuInboundState.picks).forEach(function(id) { if (cuInboundState.picks[id]) ids.push(id); });
  if (!ids.length) { document.querySelectorAll('.ib-pick.selected').forEach(function(c){ ids.push(c.getAttribute('data-id')); }); }
  return ids;
}
function onSecurityChange() {
  var sec = $id('ib-security').value; var isReality = sec === 'reality';
  toggleById('ib-reality-fields', isReality); toggleById('ib-external-address-fields', isReality);
  if (!isReality) { if ($id('ib-extdomain')) $id('ib-extdomain').value = ''; if ($id('ib-extport')) $id('ib-extport').value = ''; }
}
async function loadTelegramRailwayInfo(){ /* ... */ }
function onInboundTypeChange() {
  var type = $id('ib-inbound-type').value; var isTg = type === 'telegram'; var isConfig = type === 'config';
  toggleById('ib-tls-fields', isConfig); toggleById('ib-net-wrap', isConfig); toggleById('ib-internal-port-wrap', isConfig); toggleById('ib-tg-fields', isTg);
  if (isTg) loadTelegramRailwayInfo();
}
function onNetworkChange() { var net = $id('ib-network').value; toggleById('ib-ws-fields', net === 'ws'); toggleById('ib-xhttp-fields', net === 'xhttp'); toggleById('ib-grpc-fields', net === 'grpc'); }
function toggleById(id, show) { var el = $id(id); if (el) { if (show) el.classList.remove('hidden'); else el.classList.add('hidden'); } }
async function generateRealityKeysUI() {
  try { var r = await api('/api/tools/generate-reality-keys', {method: 'POST'}); $id('ib-privkey').value = r.private_key || ''; $id('ib-pubkey').value = r.public_key || ''; toast(t('keysGenerated'), 'ok'); }
  catch(e) { toast(t('errKeyGen'), 'err'); }
}
async function generateShortId() {
  var iid = $id('ib-id').value;
  if (!iid) { var hex = ''; for(var i=0;i<10;i++) hex += '0123456789abcdef'[Math.floor(Math.random()*16)]; $id('ib-shortids').value = hex; toast(t('shortIdGen'), 'ok'); return; }
  try { var r = await api('/api/inbounds/' + iid + '/generate-short-id', {method: 'POST'}); $id('ib-shortids').value = r.short_id || ''; toast(t('shortIdNew'), 'ok'); }
  catch(e) { toast(t('toast.err'), 'err'); }
}
function setInboundNodeOnlyMode(enabled) { /* ... */ }
function ensureNodeSelector(ib) { /* ... */ }
function openCreateInbound() {
  setInboundNodeOnlyMode(false);
  var nodeWrap = $id('ib-node-select'); if (nodeWrap) nodeWrap.classList.add('hidden');
  EDIT_IB = null;
  $id('ib-id').value = ''; $id('ib-title').textContent = t('newInbound');
  $id('ib-name').value = ''; $id('ib-inbound-type').value = 'config'; $id('ib-port').value = '443'; $id('ib-security').value = 'tls'; $id('ib-network').value = 'ws';
  $id('ib-extdomain').value = ''; $id('ib-extport').value = '443'; $id('ib-sni').value = ''; $id('ib-privkey').value = ''; $id('ib-pubkey').value = '';
  $id('ib-shortids').value = ''; $id('ib-whitex').value = '/'; $id('ib-wspath').value = '/'; $id('ib-xhttppath').value = '/'; $id('ib-xhttp-mode').value = 'auto'; $id('ib-grpc-svc').value = 'grpc';
  _ibSet('ib-fingerprint', 'chrome'); _ibSet('ib-nameprefix', ''); _ibSet('ib-host', ''); _ibSet('ib-alpn', ''); _ibSet('ib-flow', ''); _ibSet('ib-packetencoding', ''); _ibSet('ib-advanced', '');
  if ($id('ib-allowinsecure')) $id('ib-allowinsecure').checked = false;
  $id('ib-submit').textContent = t('createInbound');
  onInboundTypeChange(); onSecurityChange(); onNetworkChange();
  openModal('mod-inbound');
}
var EDIT_IB = null;
function _ibSet(id, v) { var el = $id(id); if (el) el.value = (v == null ? '' : v); }
async function editInbound(iid) {
  var ib = null;
  for (var i = 0; i < ALL_IB.length; i++) { if (ALL_IB[i].inbound_id === iid) { ib = ALL_IB[i]; break; } }
  if (!ib) {
    try { var resp = await api('/api/inbounds/' + iid); ib = resp.inbound; } catch(e) { toast(e.message, 'err'); return; }
  }
  if (!ib) { toast(t('toast.err'), 'err'); return; }
  EDIT_IB = ib;
  var proto = (ib.protocol || 'vless').toLowerCase();
  var isTg = proto === 'telegram';
  $id('ib-id').value = ib.inbound_id || '';
  $id('ib-title').textContent = _advancedInboundTitle(ib);
  $id('ib-inbound-type').value = isTg ? 'telegram' : 'config';
  _ibSet('ib-name', ib.name);
  _ibSet('ib-port', ib.port || 443);
  _ibSet('ib-security', ib.security || 'tls');
  _ibSet('ib-network', ib.network || 'ws');
  _ibSet('ib-extdomain', ib.external_domain);
  _ibSet('ib-extport', ib.external_port || 443);
  _ibSet('ib-sni', ib.sni);
  _ibSet('ib-fingerprint', ib.fingerprint || 'chrome');
  _ibSet('ib-nameprefix', ib.name_prefix);
  _ibSet('ib-host', ib.host);
  _ibSet('ib-alpn', ib.alpn);
  _ibSet('ib-flow', ib.flow);
  _ibSet('ib-packetencoding', ib.packet_encoding);
  if ($id('ib-allowinsecure')) $id('ib-allowinsecure').checked = !!ib.allow_insecure;
  _ibSet('ib-advanced', (ib.advanced_settings && Object.keys(ib.advanced_settings).length) ? JSON.stringify(ib.advanced_settings, null, 2) : '');
  var rs = ib.reality_settings || {};
  _ibSet('ib-privkey', rs.private_key);
  _ibSet('ib-pubkey', rs.public_key);
  var sid = rs.short_id; if (Array.isArray(sid)) sid = sid[0] || '';
  _ibSet('ib-shortids', sid);
  _ibSet('ib-whitex', rs.rx_path || '/');
  var ws = ib.ws_settings || {}; _ibSet('ib-wspath', ws.path || '/');
  var xs = ib.xhttp_settings || {}; _ibSet('ib-xhttppath', xs.path || '/'); _ibSet('ib-xhttp-mode', xs.mode || 'auto');
  var gr = ib.grpc_settings || {}; _ibSet('ib-grpc-svc', gr.serviceName || gr.service_name || 'grpc');
  if (isTg) {
    var tg = ib.telegram_settings || {};
    _ibSet('ib-tg-intport', tg.internal_port || ib.port || 44344);
    _ibSet('ib-tg-extport', tg.external_port || ib.external_port || 443);
    _ibSet('ib-tg-domain', tg.external_domain || ib.external_domain);
  }
  $id('ib-submit').textContent = (LANG === 'en' ? 'Save' : 'ذخیره');
  onInboundTypeChange(); onSecurityChange(); onNetworkChange();
  openModal('mod-inbound');
}
async function submitInbound(e) {
  if (e) e.preventDefault();
  var id = $id('ib-id').value;
  var type = $id('ib-inbound-type').value;
  var payload = { name: $id('ib-name').value.trim() };
  if (type === 'telegram') {
    payload.protocol = 'telegram';
    payload.network = 'tcp';
    payload.security = 'tls';
    payload.telegram_settings = {
      internal_port: Number($id('ib-tg-intport').value || 0),
      external_port: Number($id('ib-tg-extport').value || 0),
      external_domain: $id('ib-tg-domain').value.trim()
    };
    payload.port = payload.telegram_settings.internal_port;
    payload.external_port = payload.telegram_settings.external_port;
    payload.external_domain = payload.telegram_settings.external_domain;
  } else {
    var sec = $id('ib-security').value;
    var net = $id('ib-network').value;
    payload.protocol = sec === 'reality' ? 'reality' : 'vless';
    payload.security = sec;
    payload.network = net;
    payload.port = Number($id('ib-port').value || 443);
    payload.external_port = Number($id('ib-extport').value || 443);
    payload.external_domain = $id('ib-extdomain').value.trim();
    payload.sni = $id('ib-sni').value.trim();
    payload.fingerprint = $id('ib-fingerprint').value.trim();
    payload.name_prefix = $id('ib-nameprefix').value.trim();
    payload.host = $id('ib-host').value.trim();
    payload.alpn = $id('ib-alpn').value.trim();
    payload.flow = $id('ib-flow').value.trim();
    payload.packet_encoding = $id('ib-packetencoding').value.trim();
    payload.allow_insecure = !!$id('ib-allowinsecure').checked;
    var advRaw = $id('ib-advanced').value.trim();
    if (advRaw) {
      try { payload.advanced_settings = JSON.parse(advRaw); }
      catch(err) { toast(t('advancedInvalidJson'), 'err'); return; }
    } else { payload.advanced_settings = {}; }
    if (sec === 'reality') {
      var rs = Object.assign({}, (EDIT_IB && EDIT_IB.reality_settings) || {});
      rs.private_key = $id('ib-privkey').value.trim();
      rs.public_key = $id('ib-pubkey').value.trim();
      rs.short_id = $id('ib-shortids').value.trim();
      rs.sni = payload.sni;
      rs.rx_path = $id('ib-whitex').value.trim() || '/';
      payload.reality_settings = rs;
    }
    if (net === 'ws') { var ws = Object.assign({}, (EDIT_IB && EDIT_IB.ws_settings) || {}); ws.path = $id('ib-wspath').value.trim() || '/'; payload.ws_settings = ws; }
    if (net === 'xhttp') { var xs = Object.assign({}, (EDIT_IB && EDIT_IB.xhttp_settings) || {}); xs.path = $id('ib-xhttppath').value.trim() || '/'; xs.mode = $id('ib-xhttp-mode').value; payload.xhttp_settings = xs; }
    if (net === 'grpc') { var gr = Object.assign({}, (EDIT_IB && EDIT_IB.grpc_settings) || {}); gr.serviceName = $id('ib-grpc-svc').value.trim() || 'grpc'; payload.grpc_settings = gr; }
  }
  try {
    if (id) {
      await api('/api/inbounds/' + id, {method: 'PATCH', body: payload});
      toast(LANG === 'en' ? 'Inbound saved' : 'اینباند ذخیره شد', 'ok');
    } else {
      await api('/api/inbounds', {method: 'POST', body: payload});
      toast(LANG === 'en' ? 'Inbound created' : 'اینباند ایجاد شد', 'ok');
    }
    closeModal('mod-inbound');
    EDIT_IB = null;
    loadInbounds();
  } catch(err) { toast(err.message, 'err'); }
}
async function deleteInbound(iid) { if (!confirm(t('confirmDeleteInbound'))) return; try { await api('/api/inbounds/' + iid, {method: 'DELETE'}); toast(t('inboundDeleted'), 'ok'); loadInbounds(); } catch(e) { toast(e.message, 'err'); } }

function loadSettings() {
  try {
    api('/api/settings').then(function(s) {
      var el = $id('set-config-prefix');
      if (el) el.value = (s && s.config_name_prefix) ? s.config_name_prefix : 'White-';
    }).catch(function() {});
  } catch(e) {}
  checkPanelUpdate();
}
async function saveConfigPrefix() {
  var el = $id('set-config-prefix');
  var v = (el && el.value || '').trim() || 'White-';
  try {
    await api('/api/settings', {method: 'POST', body: {config_name_prefix: v}});
    if (el) el.value = v;
    toast(LANG === 'en' ? 'Saved' : 'ذخیره شد', 'ok');
  } catch(e) { toast(e.message, 'err'); }
}
var ALL_ADV_IB = [];
var SELECTED_ADV_IB = '';
function _advancedInboundTitle(ib) { if (ib && ib._managed_kind === 'default_tls_ws' && (!ib.advanced_customized || ib.name === 'پیش‌فرض TLS + WS')) return t('defaultTlsWsInbound'); return (ib && ib.name) || (ib && ib.inbound_id) || t('inboundName'); }
async function loadAdvancedInbounds() {
  var el = $id('advanced-inbound-list');
  try {
    var r = await api('/api/inbounds');
    ALL_ADV_IB = (r && r.inbounds) || [];
    renderAdvancedInboundList();
  } catch(e) {
    if (el) el.innerHTML = '<div class="empty">' + esc(e.message) + '</div>';
  }
}
function renderAdvancedInboundList() {
  var el = $id('advanced-inbound-list');
  if (!el) return;
  if (!ALL_ADV_IB.length) { el.innerHTML = '<div class="empty">' + t('advancedNoInbounds') + '</div>'; return; }
  el.innerHTML = ALL_ADV_IB.map(function(ib) {
    var active = (ib.inbound_id === SELECTED_ADV_IB) ? ' btn-p' : ' btn-g';
    return '<button type="button" class="btn btn-sm' + active + '" style="margin:4px" onclick="selectAdvancedInbound(\'' + ib.inbound_id + '\')">' + esc(_advancedInboundTitle(ib)) + '</button>';
  }).join('');
}
function selectAdvancedInbound(iid) {
  SELECTED_ADV_IB = iid;
  var ib = null;
  for (var i = 0; i < ALL_ADV_IB.length; i++) { if (ALL_ADV_IB[i].inbound_id === iid) { ib = ALL_ADV_IB[i]; break; } }
  var wrap = $id('advanced-editor-wrap'), ta = $id('advanced-inbound-json');
  if (ib && ta) {
    ta.value = JSON.stringify(ib, null, 2);
    if (wrap) wrap.classList.remove('hidden');
  }
  renderAdvancedInboundList();
}
async function saveAdvancedInbound(iid) {
  iid = iid || SELECTED_ADV_IB;
  if (!iid) { toast(t('advancedNoInbounds'), 'err'); return; }
  var ta = $id('advanced-inbound-json');
  if (!ta) return;
  var obj;
  try { obj = JSON.parse(ta.value); }
  catch(e) { toast(t('advancedInvalidJson'), 'err'); return; }
  try {
    await api('/api/inbounds/' + iid, {method: 'PATCH', body: obj});
    toast(t('advancedSaved'), 'ok');
    await loadAdvancedInbounds();
    selectAdvancedInbound(iid);
  } catch(e) { toast(e.message, 'err'); }
}
function _setHelpSwitch(id, enabled) { var el = $id(id); if (!el) return; el.classList.toggle('on', !!enabled); el.setAttribute('aria-checked', String(!!enabled)); }
var ECH_PRESET_LABELS = { 'cf-config': 'CF Config', 'crypto-cloudflare': 'CF DNS', 'ridanbe': 'Ridanbe', 'iranva': 'Iranva' };
var SELECTED_ECH_PRESET = 'cf-config';
function setEchPresetUi(preset) { /* ... */ }
function openEchPresetPicker() { setEchPresetUi(SELECTED_ECH_PRESET); openModal('mod-ech-picker'); }
async function selectEchPreset(preset) { /* ... */ }
async function loadHelpSettings() { /* ... */ }
async function saveHelpAddress(mode) { /* ... */ }
async function toggleHelpFeature(feature, el) { /* ... */ }
async function saveHelpSpoofFields() { /* ... */ }
async function copyHelpPrompt() { /* ... */ }

var HL_AUTO_TIMER = null;
var HL_DEBOUNCE = null;
function helpLogsToggled() { /* ... */ }
function helpLogsDebounced() { /* ... */ }
function startHelpLogsAuto() { /* ... */ }
function stopHelpLogsAuto() { /* ... */ }
function hlKindColor(kind, level) { /* ... */ }
function helpLogsHtml(d) { /* ... */ }
async function loadHelpLogs(quiet) { /* ... */ }
async function changePassword(e) { /* ... */ }
var UPD_POLL = null;
function renderUpdVersion(s) { /* ... */ }
async function checkPanelUpdate() {
  var el = $id('upd-ver');
  if (!el) return;
  try {
    el.innerHTML = '<span style="color:var(--txt2)">در حال بررسی نسخه جدید...</span>';
    var d = await api('/api/update-check');
    if (!d || !d.ok) {
      el.innerHTML = '<span style="color:var(--txt2)">نسخه فعلی: 4.0.0</span>';
      return;
    }
    if (d.update_available) {
      el.innerHTML = '<span style="color:var(--green);font-weight:700">نسخه جدید موجود است: ' + d.latest + '</span>' +
        '<div style="font-size:11px;color:var(--txt2);margin-top:4px">نسخه شما: ' + d.version + '</div>' +
        (d.changelog ? '<div style="font-size:10px;color:var(--neon);margin-top:2px">' + d.changelog + '</div>' : '');
    } else {
      el.innerHTML = '<span style="color:var(--green)">شما از آخرین نسخه پنل (' + d.version + ') استفاده می‌کنید ✔</span>';
    }
  } catch (e) {
    el.innerHTML = '<span style="color:var(--txt2)">نسخه فعلی: 4.0.0</span>';
  }
}

async function runPanelUpdate() {
  await checkPanelUpdate();
  var conf = confirm('آیا می‌خواهید به مخزن گیت‌هاب هدایت شوید؟');
  if (conf) {
    window.open('https://github.com/Itskillmaster/WhitePanel-UI', '_blank');
  }
}
async function downloadBackup() {
  try {
    var r = await fetch('/api/settings/backup', { method: 'GET', credentials: 'same-origin', cache: 'no-store', headers: {'Accept': 'application/json'} });
    if (!r.ok) { var errText = await r.text(); var errMsg = 'بکاپ ناموفق: ' + r.status; try { var errJson = JSON.parse(errText); errMsg = errJson.detail || errJson.error || errMsg; } catch (_) {} throw new Error(errMsg); }
    var blob = await r.blob(); if (!blob.size) throw new Error('فایل بکاپ خالی است');
    var url = URL.createObjectURL(blob); var a = document.createElement('a'); a.href = url;
    var cd = r.headers.get('Content-Disposition') || ''; var match = cd.match(/filename="([^"]+)"/i);
    a.download = match && match[1] ? match[1] : ('white-panel-backup-' + new Date().toISOString().slice(0,19).replace(/[:T]/g,'-') + '.json');
    a.style.display = 'none'; document.body.appendChild(a); a.click();
    setTimeout(function(){ document.body.removeChild(a); URL.revokeObjectURL(url); }, 1000);
    toast(t('backupDownloaded'), 'ok');
  } catch(e) { toast(e.message || t('toast.err'), 'err'); }
}
async function uploadRestore(input) { /* preserved */ }

function openModal(id) { $id(id).classList.add('show'); }
function closeModal(id) { $id(id).classList.remove('show'); }
function wireModalCloseButtons() {
  document.querySelectorAll('.modal').forEach(function(m) {
    if (m.querySelector('.modal-x')) return;
    var x = document.createElement('button'); x.className = 'modal-x'; x.setAttribute('aria-label', 'close'); x.innerHTML = '&#10005;';
    x.onclick = function(ev) { ev.stopPropagation(); closeModal(m.id); };
    m.appendChild(x);
    m.addEventListener('click', function(ev) { if (ev.target === m) closeModal(m.id); });
  });
}
document.addEventListener('keydown', function(ev) { if (ev.key === 'Escape') { document.querySelectorAll('.modal.show').forEach(function(m) { closeModal(m.id); }); } });
document.addEventListener('DOMContentLoaded', wireModalCloseButtons);
wireModalCloseButtons();

function esc(s) { return (s || '').replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;'); }
function escAttr(s) { return esc(s).replace(/'/g,'&#39;').replace(/`/g,'&#96;'); }
function toFa(n) { return String(n).replace(/\d/g, function(d) { return '۰۱۲۳۴۵۶۷۸۹'[d]; }); }
function fmtBytes(b) { if (!b || b <= 0) return '0 B'; var u = ['B','KB','MB','GB','TB'], i = 0; while (b >= 1024 && i < 4) { b /= 1024; i++; } return b.toFixed(i === 0 ? 0 : 2) + ' ' + u[i]; }
function fmtUptime(s) { if (!s) return '۰s'; var d = Math.floor(s / 86400), h = Math.floor((s % 86400) / 3600), m = Math.floor((s % 3600) / 60); return (d ? d + 'd ' : '') + h + 'h ' + m + 'm'; }

var iconUsers = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2m8-10a4 4 0 100-8 4 4 0 000 8z"/><circle cx="17" cy="7" r="4"/><path d="M23 21v-2a4 4 0 00-3-3.87"/></svg>';
var iconTraffic = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>';
var iconConn = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg>';
var iconUptime = '<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>';
var iconCfg = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-4 0v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 010-4h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 012.83-2.83l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 014 0v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 010 4h-.09a1.65 1.65 0 00-1.51 1z"/></svg>';
var iconLink = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg>';
var iconQr = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><line x1="14" y1="14" x2="14" y2="14.01"/><line x1="18" y1="14" x2="18" y2="14.01"/><line x1="21" y1="14" x2="21" y2="14.01"/><line x1="14" y1="17" x2="14" y2="17.01"/><line x1="16" y1="17" x2="16" y2="21"/><line x1="14" y1="21" x2="21" y2="21"/></svg>';
var iconEdit = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>';
var iconToggle = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round"><path d="M12 2v9"/><path d="M6.6 5.7a9 9 0 106.8 0"/></svg>';
var iconTrash = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/></svg>';
var iconReset = '<svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M3 12a9 9 0 109-9c-2.5 0-4.76 1.02-6.4 2.67L3 8"/><path d="M3 3v5h5"/></svg>';

// Audio (preserved behavior)
var MUSIX_TRACKS = [];
function musixPickTrack() { if (!MUSIX_TRACKS.length) return ''; return MUSIX_TRACKS[Math.floor(Math.random() * MUSIX_TRACKS.length)]; }
function musixPlayRandom() { var p = $id('panel-audio-player'); if (!p) return; var url = musixPickTrack(); if (!url) return; if (p.getAttribute('src') !== url) { p.src = url; p.load(); } var lbl = $id('audio-file'); if (lbl) lbl.textContent = decodeURIComponent(url.split('/').pop()); p.play().catch(function(){}); }
function togglePanelAudio() { var p = $id('panel-audio-player'); var sw = $id('audio-autoplay'); sw.classList.toggle('on'); var on = sw.classList.contains('on'); localStorage.setItem('white_audio', on ? '1' : '0'); if (on) musixPlayRandom(); else if (p) p.pause(); }
(function initPanelAudio(){ var p = $id('panel-audio-player'); if (!p) return; p.volume = 0.3; var sw = $id('audio-autoplay'); if (!sw) return; var saved = localStorage.getItem('white_audio'); var on = saved === null ? true : saved !== '0'; var loading = false; if (saved === null) { try { localStorage.setItem('white_audio', '1'); } catch(e) {} } sw.classList.toggle('on', on); p.addEventListener('ended', function(){ if (sw.classList.contains('on')) musixPlayRandom(); }); function startAfterGesture() { if (!sw.classList.contains('on') || loading) return; loading = true; document.removeEventListener('click', startAfterGesture); document.removeEventListener('touchstart', startAfterGesture); document.removeEventListener('keydown', startAfterGesture); api('/api/musix').then(function(d) { MUSIX_TRACKS = (d.tracks || []).map(function(x){ return x.url; }); if (MUSIX_TRACKS.length && sw.classList.contains('on')) musixPlayRandom(); }).catch(function(){ loading = false; }); } document.addEventListener('click', startAfterGesture); document.addEventListener('touchstart', startAfterGesture); document.addEventListener('keydown', startAfterGesture); })();

function loadProxyCountries(){}


// ─── Notifications ───
// Announcements live as notif/*.txt files in a GitHub repo. Every notif file
// becomes one accordion: the subject is on the head, the body and the date/time
// are revealed when it is clicked. `force` bypasses the server-side cache.
var NOTIF_DATA = null;
var NOTIF_PENDING = null;

function renderNotifs() {
  var list = $id('notif-list');
  if (!list) return;
  var data = NOTIF_DATA;
  var src = $id('notif-source');
  if (src) src.textContent = (data && data.source) ? data.source : '';
  if (!data) return;

  var items = data.items || [];
  if (data.error) {
    list.innerHTML = '<div class="empty">' + esc(t('notif.error')) +
      '<div style="margin-top:6px;font-size:10px;opacity:.75">' + esc(data.error) + '</div></div>';
    return;
  }
  if (!items.length) {
    list.innerHTML = '<div class="empty">' + esc(t('notif.empty')) + '</div>';
    return;
  }

  list.innerHTML = items.map(function(n, i) {
    // spdToggleAcc() looks up `acc-` + id, so the body must carry that prefix.
    var accId = 'notif-' + i;
    var date = String(n.date || '').trim();
    return '<div class="spd-acc notif-item">' +
      '<button type="button" class="spd-acc-head" onclick="spdToggleAcc(\'' + accId + '\',this)" aria-expanded="false">' +
        '<span class="spd-acc-title">' + esc(n.title || n.name || '') + '</span>' +
        (date ? '<span class="badge badge-b spd-acc-badges" style="font-size:9px">' + esc(date) + '</span>' : '') +
        '<span class="material-symbols-outlined spd-acc-arrow">expand_more</span>' +
      '</button>' +
      '<div class="spd-acc-body" id="acc-' + accId + '">' +
        '<div class="notif-text">' + esc(n.text || '') + '</div>' +
        '<div class="notif-date"><span class="material-symbols-outlined">schedule</span>' +
          '<span>' + t('notif.date') + ':</span>' +
          '<span dir="ltr">' + esc(date || '—') + '</span>' +
        '</div>' +
      '</div>' +
    '</div>';
  }).join('');
}

async function loadNotifs(force) {
  if (force) { NOTIF_DATA = null; NOTIF_PENDING = null; }
  if (NOTIF_DATA) { renderNotifs(); return; }

  var tab = $id('tab-notif');
  var list = $id('notif-list');
  if (list && tab && !tab.classList.contains('hidden') && !list.querySelector('.spd-acc')) {
    list.innerHTML = '<div class="empty">' + esc(t('notif.loading')) + '</div>';
  }
  if (NOTIF_PENDING) return NOTIF_PENDING;

  NOTIF_PENDING = (async function() {
    try {
      NOTIF_DATA = await api('/api/notifications' + (force ? '?refresh=1' : ''));
      renderNotifs();
    } catch(e) {
      console.error(e);
      NOTIF_DATA = {items: [], error: e.message || 'error', source: ''};
      renderNotifs();
    } finally {
      NOTIF_PENDING = null;
    }
  })();
  return NOTIF_PENDING;
}

// ═══ IP SCANNER ═══
var SCN_CF_RANGES = [];
var SCN_RW_RANGES = [];
var SCN_SNI_LIST = [];
var SCN_DOM_DEFAULTS = [
  'shuttle.proxy.rlwy.net', 'monorail.proxy.rlwy.net', 'caboose.proxy.rlwy.net',
  'viaduct.proxy.rlwy.net', 'acela.proxy.rlwy.net', 'sakura.proxy.rlwy.net',
  'tokaido.proxy.rlwy.net', 'switchback.proxy.rlwy.net', 'hayabusa.proxy.rlwy.net',
  'reseau.proxy.rlwy.net', 'shortline.proxy.rlwy.net'
];
var SCN_STOP = false;
var scnCtx = { cf: {found: [], tried: 0, gen: 0, seq: 0}, railway: {found: [], tried: 0, gen: 0, seq: 0} };

// Load CF subnets dynamically from server
async function scnLoadRanges() {
  try {
    var r = await api('/api/scanner/cf-subnets');
    if (r.subnets && r.subnets.length) {
      SCN_CF_RANGES = r.subnets.map(function(s) {
        var parts = s.split('/')[0].split('.');
        var mask = parseInt(s.split('/')[1]) || 24;
        return [parseInt(parts[0])||0, parseInt(parts[1])||0, parseInt(parts[2])||0, parseInt(parts[3])||0, mask];
      });
    }
  } catch(e) { console.error('Failed to load CF subnets:', e); }
  // Fallback ranges if API fails
  if (!SCN_CF_RANGES.length) {
    SCN_CF_RANGES = [[104, 16, 0, 0, 12], [172, 64, 0, 0, 13], [162, 159, 0, 0, 16]];
  }
  SCN_RW_RANGES = [[66, 33, 22, 0, 23], [66, 33, 23, 0, 24], [152, 55, 176, 0, 22]];
}
scnLoadRanges();

function scnRandIp(range) {
  var seg = [range[0], range[1], range[2], range[3]];
  if (range[4] === 12) { seg[1] += Math.floor(Math.random() * 16); seg[2] = Math.floor(Math.random() * 256); }
  else if (range[4] === 13) { seg[1] += Math.floor(Math.random() * 8); seg[2] = Math.floor(Math.random() * 256); }
  else if (range[4] === 14) { seg[2] += Math.floor(Math.random() * 4); }
  else if (range[4] === 15) { seg[2] += Math.floor(Math.random() * 2); }
  else if (range[4] === 16) { seg[2] = Math.floor(Math.random() * 256); }
  else if (range[4] === 22) { seg[2] = Math.floor(Math.random() * 64); }
  else if (range[4] === 23) { seg[2] = Math.floor(Math.random() * 128); }
  else if (range[4] === 24) { seg[2] = Math.floor(Math.random() * 256); }
  seg[3] = Math.floor(Math.random() * 253) + 2;
  return seg.join('.');
}

function scnMsClass(ms) {
  if (ms == null) return 'soon';
  return ms < 150 ? 'good' : ms < 400 ? 'ok' : 'bad';
}

function scnMsHtml(ms, failed) {
  if (ms == null && failed) return '<span class="scn-ms bad">✗</span>';
  if (ms == null) return '<span class="scn-ms soon">…</span>';
  return '<span class="scn-ms ' + scnMsClass(ms) + '">' + ms + ' ms</span>';
}

function scnChipClass(ms) {
  if (ms == null) return 'soon';
  return ms < 150 ? 'good' : ms < 400 ? 'ok' : 'bad';
}

// Browser-side ping: loads an image from the target IP:port and times the
// round-trip. This scans from the USER's own network (like the Reza_Gm
// scanner), not from the Railway server. Port 443 uses https; other ports
// fall back to http. Browsers may block direct requests to arbitrary IPs
// (mixed-content / TLS on HTTPS pages); those targets are re-measured by the
// server as a fallback so the scan still completes.
// Browser-side ping using fetch() with no-cors — measures the round-trip from
// the USER's own network (works on mobile + desktop, no server needed). This is
// the simplest reliable way to rank IPs by latency from the user's connection.
function scnBrowserPing(target, timeoutMs) {
  return new Promise(function(resolve) {
    var port = parseInt((target.split(':')[1]) || '443');
    var url = (port === 443 ? 'https' : 'http') + '://' + target + '/';
    var controller = null;
    try { controller = new AbortController(); } catch (e) {}
    var t0 = performance.now();
    fetch(url, { mode: 'no-cors', cache: 'no-store', signal: controller ? controller.signal : undefined })
      .then(function() {
        var ms = performance.now() - t0;
        resolve({ ok: ms < timeoutMs, ms: ms });
      })
      .catch(function() {
        // A fast error (TLS/cert) is still a live result — just not a good one.
        var ms = performance.now() - t0;
        resolve({ ok: ms < timeoutMs && ms >= 150, ms: ms });
      });
    setTimeout(function() {
      if (controller) { try { controller.abort(); } catch (e) {} }
      resolve({ ok: false, ms: null });
    }, timeoutMs);
  });
}

async function scnPingBatch(targets, timeout) {
  var out = {};
  var tmo = timeout || 1.2;
  await Promise.all(targets.map(function(t) {
    return scnBrowserPing(t, tmo * 1000).then(function(res) {
      var parts = t.split(':');
      var port = parseInt(parts[1]) || 443;
      out[t] = { target: t, ip: parts[0], port: port, latency_ms: res.ms, ok: !!res.ok, src: 'browser' };
    });
  }));
  return out;
}

async function scnRunCF() {
  scnRunIPs('cf');
}
async function scnRunRailway() {
  scnRunIPs('railway');
}

async function scnRunIPs(which) {
  if (scnCtx[which].running) return;
  scnCtx[which].running = true;
  scnCtx[which].found = [];
  scnCtx[which].tried = 0;
  // Generation token: bumped by scnClearSaved() so a still-running scan loop
  // can't re-write the file after the user cleared it.
  var runGen = ++scnCtx[which].gen;
  SCN_STOP = false;
  var pfx = which === 'cf' ? 'scn-cf' : 'scn-rw';
  var ranges = which === 'cf' ? SCN_CF_RANGES : SCN_RW_RANGES;
  var count = parseInt($id(pfx + '-count').value) || 30;
  var port = $id(pfx + '-port').value || '443';
  var list = $id(pfx + '-list');
  var bar = $id(pfx + '-bar');
  var startBtn = list.parentElement.querySelector('.btn-p');
  list.innerHTML = '<div class="scn-empty">' + t('scanning') + '</div>';
  $id(pfx + '-found').textContent = '0';
  $id(pfx + '-tried').textContent = '0';
  $id(pfx + '-best').textContent = '--';
  $id(pfx + '-bar').style.width = '0%';
  if (startBtn) { startBtn.disabled = true; startBtn.textContent = t('scanning'); }

  var found = [];
  var tried = 0;
  var live = {};   // ip -> {ms, ok} of last attempt (success and failure both shown)
  var loops = 0;
  var BATCH = 32;
  var pfx = which === 'cf' ? 'scn-cf' : 'scn-rw';
  var scanDelay = parseInt($id(pfx + '-delay').value) || 200;
  var scanTimeout = parseFloat($id(pfx + '-timeout').value) || 1.0;
  while (found.length < count && loops < count * 20) {
    if (SCN_STOP) break;
    loops++;
    var batch = [];
    for (var b = 0; b < BATCH && found.length + batch.length < count * 3; b++) {
      batch.push(scnRandIp(ranges[Math.floor(Math.random() * ranges.length)]) + ':' + port);
    }
    var resMap = await scnPingBatch(batch, scanTimeout);
    // Rate limit: delay between batches to avoid flood behavior
    if (scanDelay > 0 && loops < count * 20) {
      await new Promise(function(resolve) { setTimeout(resolve, scanDelay); });
    }
    Object.keys(resMap).forEach(function(t) {
      tried++;
      var r = resMap[t];
      live[t.split(':')[0]] = { ms: r.latency_ms, ok: r.ok };
      if (r && r.ok) {
        found.push({ip: t.split(':')[0], port: r.port, ms: r.latency_ms});
      }
    });
    found.sort(function(a, b) { return a.ms - b.ms; });
    found = found.slice(0, count);
    scnCtx[which].found = found.slice();
    scnCtx[which].tried = tried;
    // Live-save the current best list (replaces the file, capped at 10 by the
    // backend) so the saved first-10 stays in sync with the scan. Skip if the
    // scan was cleared mid-run (generation changed).
    if (found.length && scnCtx[which].gen === runGen) {
      scnSaveLive(which, found.map(function(f) { return f.ip + ':' + f.port; }), true);
    }

    // render: success (green) and failure (red) both live
    var rows = '';
    Object.keys(live).forEach(function(ip) {
      var r = live[ip];
      rows += '<div class="scn-item"><span class="scn-ip">' + ip + '</span><span class="scn-port">:' + port + '</span>' +
        (r.ok ? scnMsHtml(r.ms, false) : '<span class="scn-ms bad">✗</span>') + '</div>';
    });
    found.forEach(function(f) {
      if (!live[f.ip] || !live[f.ip].ok) {
        rows += '<div class="scn-item"><span class="scn-ip">' + f.ip + '</span><span class="scn-port">:' + f.port + '</span>' + scnMsHtml(f.ms, false) + '</div>';
      }
    });
    list.innerHTML = rows || '<div class="scn-empty">' + t('noHit') + '</div>';
    $id(pfx + '-found').textContent = found.length;
    $id(pfx + '-tried').textContent = tried;
    if (found.length) {
      $id(pfx + '-best').textContent = found[0].ms + ' ms';
      $id(pfx + '-best').className = 'scn-ms ' + scnChipClass(found[0].ms);
    }
    $id(pfx + '-bar').style.width = Math.min(100, (found.length / count) * 100) + '%';
    $id(pfx + '-saved').textContent = found.length + '/' + count;
  }
  scnCtx[which].running = false;
  if (startBtn) { startBtn.disabled = false; startBtn.textContent = t('startScan'); }
  scnLoadSaved(which);
  if (SCN_STOP) toast(t('scanStopped'), 'warn');
}

function scnSaveLive(which, entries, replace) {
  // seq (per-type write token) guards against a clear() being overwritten by
  // an older save that was already in flight — the server drops stale writes.
  api('/api/scanner/save', {method: 'POST', body: {type: which, ips: entries, replace: !!replace, seq: scnCtx[which].seq}}).then(function(d) {
    if (d && typeof d.seq === 'number') scnCtx[which].seq = d.seq;
    if (d && d.stale) return;
    scnRenderSavedTags(which, d.ips || []);
  }).catch(function() {});
}

async function scnLoadSaved(which) {
  try {
    var d = await api('/api/scanner/ips/' + which);
    if (typeof d.seq === 'number') scnCtx[which].seq = d.seq;
    scnRenderSavedTags(which, d.ips || []);
  } catch (e) {}
}

function scnRenderSavedTags(which, ips) {
  var el = $id(which === 'cf' ? 'scn-cf-saved-tags' : 'scn-rw-saved-tags');
  if (!el) return;
  el.innerHTML = ips.length ? ips.map(function(x) { return '<span class="scn-tag">' + esc(x) + '</span>'; }).join('') : '<span class="scn-empty">' + t('notSaved') + '</span>';
}

function scnClearSaved(which) {
  // Stop any running scan so it can't re-write the file after we clear it.
  SCN_STOP = true;
  scnCtx[which].running = false;
  scnCtx[which].found = [];
  scnCtx[which].tried = 0;
  // Bump the generation token: a still-running scan loop will skip its next
  // save, and the seq token makes the server drop any in-flight stale write.
  scnCtx[which].gen++;
  var pfx = which === 'cf' ? 'scn-cf' : 'scn-rw';
  var list = $id(pfx + '-list');
  if (list) list.innerHTML = '<div class="scn-empty">' + t('startPrompt') + '</div>';
  $id(pfx + '-found').textContent = '0';
  $id(pfx + '-tried').textContent = '0';
  $id(pfx + '-best').textContent = '--';
  $id(pfx + '-bar').style.width = '0%';
  var sb = list ? list.parentElement.querySelector('.btn-p') : null;
  if (sb) { sb.disabled = false; sb.textContent = t('startScan'); }
  api('/api/scanner/save', {method: 'POST', body: {type: which, ips: [], replace: true, seq: scnCtx[which].seq}}).then(function(d) {
    if (d && typeof d.seq === 'number') scnCtx[which].seq = d.seq;
    scnRenderSavedTags(which, []);
    toast(t('saveCleared') + ' (' + which + ')', 'ok');
  }).catch(function() {});
}

function scnSwitch(which, btn) {
  if (btn) btn.classList.add('active');
  var body = $id('acc-scn-' + which);
  if (body && !body.parentElement.classList.contains('open')) {
    body.parentElement.classList.add('open');
    var head = body.parentElement.querySelector('.spd-acc-head');
    if (head) head.setAttribute('aria-expanded', 'true');
  }
  if (which === 'cf') scnLoadSaved('cf');
  if (which === 'railway') scnLoadSaved('railway');
  if (which === 'tcp') scnRenderDomains();
  if (which === 'sni') {
    scnLoadSniResults();
  }
}

function scnStop() { SCN_STOP = true; toast(t('scanStopped'), 'warn'); }

// ── SNI Scanner ──
var SCN_SNI_STOP = false;

function scnSwitchSni(which) {
  // Placeholder — all sni sub-tabs use the same pane
}

async function scnRunSingleSni() {
  var sni = ($id('scn-sni-input').value || '').trim().toLowerCase();
  if (!sni) { toast('SNI را وارد کنید', 'err'); return; }
  SCN_SNI_STOP = false;
  var list = $id('scn-sni-list');
  list.innerHTML = '<div class="scn-empty">در حال تست TLS/SNI...</div>';
  $id('scn-sni-tried').textContent = '۰';
  $id('scn-sni-found').textContent = '۰';
  $id('scn-sni-best').textContent = '--';
  $id('scn-sni-bar').style.width = '0%';
  try {
    var r = await api('/api/scanner/sni-check?host=' + encodeURIComponent(sni) + '&port=443');
    var ok = !!(r && r.ok);
    $id('scn-sni-tried').textContent = '۱';
    $id('scn-sni-found').textContent = ok ? '۱' : '۰';
    $id('scn-sni-best').textContent = ok ? Math.round(r.latency_ms) + 'ms' : '--';
    $id('scn-sni-bar').style.width = '100%';
    list.innerHTML = '<div class="scn-item"><span class="scn-ip">' + esc(sni) + '</span>' +
      (ok ? '<span class="scn-ms good">' + Math.round(r.latency_ms) + 'ms</span>' : '<span class="scn-ms bad">✗</span>') + '</div>';
    if (ok) {
      // Immediately merge/save this SNI; no waiting for a batch scan.
      await api('/api/scanner/sni-save-results', {method: 'POST', body: {results: [{sni: sni, latency_ms: r.latency_ms}], top: 3}});
      await scnLoadSniResults();
      toast('SNI ذخیره شد: ' + sni, 'ok');
    }
  } catch(e) { list.innerHTML = '<div class="scn-empty">خطا: ' + esc(e.message || String(e)) + '</div>'; }
}

async function scnRunAutoSni() {
  SCN_SNI_STOP = false;
  var btn = $id('scn-sni-auto-btn');
  var list = $id('scn-sni-list');
  var savedEl = $id('scn-sni-results');
  if (btn) { btn.disabled = true; btn.textContent = 'در حال اسکن...'; }
  if (list) list.innerHTML = '<div class="scn-empty">در حال دریافت لیست SNI...</div>';
  if (savedEl) savedEl.innerHTML = '<span class="scn-empty">در حال اسکن...</span>';
  $id('scn-sni-tried').textContent = '۰';
  $id('scn-sni-found').textContent = '۰';
  $id('scn-sni-best').textContent = '--';
  $id('scn-sni-bar').style.width = '0%';

  try {
    // Clear previous results before the new run, but do not wait for any
    // scanner work. This keeps the UI responsive from the first frame.
    try { await api('/api/scanner/sni-clear-results', {method:'POST'}); } catch (_) {}
    var sr = await api('/api/scanner/sni-scan-list');
    var snis = Array.isArray(sr.snis) ? sr.snis.map(function(x){ return String(x).trim(); }).filter(Boolean) : [];
    if (!snis.length) {
      if (list) list.innerHTML = '<div class="scn-empty">لیست SNI خالی است</div>';
      if (btn) { btn.disabled = false; btn.textContent = 'Auto Scan'; }
      return;
    }

    var total = snis.length, tried = 0, found = 0;
    var live = [];
    var active = 0, next = 0;
    var CONCURRENCY = 4;

    function renderLive() {
      var ordered = live.slice().sort(function(a,b){ return a.latency_ms - b.latency_ms; });
      if ($id('scn-sni-tried')) $id('scn-sni-tried').textContent = toFa(tried);
      if ($id('scn-sni-found')) $id('scn-sni-found').textContent = toFa(found);
      if ($id('scn-sni-best')) $id('scn-sni-best').textContent = ordered.length ? Math.round(ordered[0].latency_ms) + 'ms' : '--';
      if ($id('scn-sni-bar')) $id('scn-sni-bar').style.width = Math.round((tried / total) * 100) + '%';
      if (list) {
        var rows = ordered.slice(0, 20).map(function(x) {
          return '<div class="scn-item"><span class="scn-ip">' + esc(x.sni) + '</span><span class="scn-ms good">' + Math.round(x.latency_ms) + 'ms</span></div>';
        }).join('');
        list.innerHTML = rows || '<div class="scn-empty">در حال اسکن...</div>';
      }
      var top3 = ordered.slice(0,3).map(function(x){ return {sni:x.sni, latency_ms:Math.round(x.latency_ms)}; });
      if (top3.length) {
        // Save on every improvement. The file therefore follows the live UI.
        api('/api/scanner/sni-save-results', {method:'POST', body:{results:top3, top:3}})
          .then(function(){ scnLoadSniResults(); }).catch(function(){});
      }
    }

    return new Promise(function(resolve) {
      function launch() {
        if (SCN_SNI_STOP || next >= total) {
          if (active === 0) resolve();
          return;
        }
        while (!SCN_SNI_STOP && active < CONCURRENCY && next < total) {
          (function(sni) {
            next++; active++;
            api('/api/scanner/sni-check?host=' + encodeURIComponent(sni))
              .then(function(r) {
                tried++;
                if (r && r.ok) {
                  found++;
                  live.push({sni:sni, latency_ms:Number(r.latency_ms || 0)});
                  live.sort(function(a,b){return a.latency_ms-b.latency_ms;});
                  live = live.slice(0, 20);
                }
                renderLive();
              })
              .catch(function(){ tried++; renderLive(); })
              .finally(function(){
                active--;
                if (next >= total && active === 0) resolve(); else launch();
              });
          })(snis[next]);
        }
      }
      launch();
    });

  } catch (e) {
    if (list) list.innerHTML = '<div class="scn-empty">خطا: ' + esc(e.message || String(e)) + '</div>';
  } finally {
    if (btn) { btn.disabled = false; btn.textContent = 'Auto Scan'; }
    scnLoadSniResults();
  }
}
function scnStopSni() { SCN_SNI_STOP = true; toast(t('scanStopped'), 'warn'); }

async function scnLoadSniResults() {
  try {
    var r = await api('/api/scanner/sni-results');
    var results = r.results || [];
    var el = $id('scn-sni-results');
    if (!el) return;
    if (!results.length) { el.innerHTML = '<span class="scn-empty">هنوز ذخیره‌ای نیست</span>'; return; }
    el.innerHTML = results.map(function(x) {
      return '<span class="scn-tag">' + esc(x.sni) + ' <small>(' + x.latency_ms + 'ms)</small></span>';
    }).join('');
  } catch(e) { console.error(e); }
}

async function snPickFastest() {
  try {
    var r = await api('/api/scanner/sni-fastest');
    if (r.ok && r.sni) {
      $id('ib-sni').value = r.sni;
      toast('سریع‌ترین SNI: ' + r.sni + ' (' + r.latency_ms + 'ms)', 'ok');
    } else {
      toast('هنوز SNI ذخیره نشده — ابتدا اسکن کنید', 'warn');
    }
  } catch(e) { toast('خطا در دریافت SNI', 'err'); }
}

// ── Clipboard helper ──
// Uses the async Clipboard API when available and falls back to a temporary
// textarea for mobile/webview environments where navigator.clipboard is absent.
function copyText(value) {
  value = String(value == null ? '' : value);
  if (!value) return Promise.reject(new Error('empty clipboard value'));
  if (navigator.clipboard && typeof navigator.clipboard.writeText === 'function') {
    return navigator.clipboard.writeText(value);
  }
  return new Promise(function(resolve, reject) {
    var ta = document.createElement('textarea');
    ta.value = value;
    ta.setAttribute('readonly', '');
    ta.style.position = 'fixed';
    ta.style.top = '0';
    ta.style.left = '-9999px';
    ta.style.opacity = '0';
    ta.style.pointerEvents = 'none';
    document.body.appendChild(ta);
    ta.focus();
    ta.select();
    ta.setSelectionRange(0, ta.value.length);
    var ok = false;
    try { ok = document.execCommand('copy'); } catch (e) {}
    document.body.removeChild(ta);
    if (ok) resolve(); else reject(new Error('clipboard unavailable'));
  });
}

// ── TCP tab ──
var scnDomains = SCN_DOM_DEFAULTS.slice();

function scnRenderDomains() {
  var el = $id('scn-tcp-list');
  if (!el) return;
  el.innerHTML = scnDomains.map(function(d, i) {
    return '<div class="scn-dom">' + esc(d) + '<button class="rm" onclick="scnRemoveDomain(' + i + ')" title="حذف">✕</button></div>';
  }).join('') || '<span class="scn-empty">دامنه‌ای اضافه نکرده‌اید</span>';
}

function scnAddDomain() {
  var inp = $id('scn-tcp-domain');
  var v = (inp.value || '').trim().toLowerCase();
  if (!v) return;
  if (scnDomains.indexOf(v) === -1) scnDomains.push(v);
  inp.value = '';
  scnRenderDomains();
}

function scnRemoveDomain(i) {
  scnDomains.splice(i, 1);
  scnRenderDomains();
}

async function scnRunTCP() {
  var resEl = $id('scn-tcp-results');
  var btn = document.querySelector('#scn-pane-tcp .btn-p');
  resEl.innerHTML = '<div class="scn-empty">' + t('checkingDomains') + '</div>';
  if (btn) { btn.disabled = true; btn.textContent = t('checking'); }
  var out = [];
  var targets = [];
  for (var i = 0; i < scnDomains.length; i++) targets.push(scnDomains[i] + ':443');
  var resMap = await scnPingBatch(targets, 3);
  for (var i = 0; i < scnDomains.length; i++) {
    var dom = scnDomains[i];
    var ip = null;
    try {
      var d = await api('/api/scanner/resolve?host=' + encodeURIComponent(dom));
      if (d.ips && d.ips.length) ip = d.ips[0];
    } catch (e) {}
    var r = resMap[dom + ':443'];
    var ms = r && r.ok ? r.latency_ms : null;
    var failed = !!(r && !r.ok);
    out.push({dom: dom, ip: ip, ms: ms, failed: failed});
  }
  var html = '';
  out.forEach(function(o) {
    var dom = String(o.dom || '');
    var ip = String(o.ip || '');
    html += '<div class="scn-result">' +
      '<button type="button" class="scn-copy-value scn-dom-name" data-copy="' + esc(dom) + '" title="Copy domain">' + esc(dom) + '</button>' +
      '<button type="button" class="scn-copy-value scn-ip" data-copy="' + esc(ip) + '" title="Copy IP"' + (ip ? '' : ' disabled') + '>' + esc(ip || '—') + '</button>' +
      scnMsHtml(o.ms, o.failed) +
      '</div>';
  });
  resEl.innerHTML = html || '<div class="scn-empty">' + t('noResult') + '</div>';
  resEl.querySelectorAll('.scn-copy-value').forEach(function(el) {
    if (el._copyBound) return;
    el._copyBound = true;
    el.addEventListener('click', function() {
      var value = el.getAttribute('data-copy') || '';
      if (!value) return;
      copyText(value).then(function() {
        toast(value, 'ok');
      }).catch(function() {
        toast(t('toastErr') || 'Copy failed', 'err');
      });
    });
  });
  if (btn) { btn.disabled = false; btn.textContent = t('checkDomains'); }
}

function scnInit() {
  scnRenderDomains();
  scnLoadSaved('cf');
  scnLoadSaved('railway');
}

// ═══ AUTO CONFIG ═══
// Full automatic pipeline, all steps run from the USER'S internet (browser
// ping + backend TCP probe). Flow:
//   1. ping user-domain, railway.app, worker.dev
//   2. user picks the domain (Railway → scan Railway IPs, else Cloudflare IPs)
//   3. scan that provider's IPs, save top 10 (with ping, 1 decimal)
//   4. test TCP domains (own domain + IP both)
//   5. scan SNIs, save the fastest
//   6. final ping to the panel's current domain
var AC_PING = { user: null, railway: null, worker: null };
var AC_CHOSEN = null;

function acLog(msg) {
  var el = $id('ac-log');
  if (!el) return;
  el.style.display = 'block';
  var line = document.createElement('div');
  line.className = 'scn-item';
  line.innerHTML = '<span class="scn-ip">' + esc(msg) + '</span>';
  el.appendChild(line);
}

async function acPingOne(host) {
  var t0 = performance.now();
  try {
    var ctrl = new AbortController();
    var timer = setTimeout(function(){ try{ctrl.abort();}catch(e){} }, 6000);
    await fetch('https://' + host + '/', {mode:'no-cors', cache:'no-store', signal: ctrl.signal});
    clearTimeout(timer);
    return Math.round((performance.now() - t0) * 10) / 10;
  } catch (e) {
    return null;
  }
}

async function acStart() {
  var d = ($id('ac-domain').value || '').trim().toLowerCase();
  if (!d) { toast(t('acEnterDomain'), 'warn'); return; }
  AC_PING = { user: null, railway: null, worker: null };
  var btn = $id('ac-auto-btn');
  if (btn) { btn.disabled = true; }
  $id('ac-log').style.display = 'block';
  $id('ac-log').innerHTML = '';
  $id('ac-results').innerHTML = '';
  AC_PING.user = await acPingOne(d);
  acLog(t('acYourDomain') + ' (' + esc(d) + '): ' + (AC_PING.user == null ? '✗' : AC_PING.user + ' ms'));
  acLog(t('acPingingDomains'));
  var [rw, wk] = await Promise.all([acPingOne('railway.app'), acPingOne('worker.dev')]);
  AC_PING.railway = rw; AC_PING.worker = wk;
  acLog('railway.app: ' + (rw == null ? '✗' : rw + ' ms') + '  |  worker.dev: ' + (wk == null ? '✗' : wk + ' ms'));
  acShowPick();
  if (btn) { btn.disabled = false; }
}

function acShowPick() {
  var results = $id('ac-results');
  if (!results) return;
  var wrap = document.getElementById('ac-domain-wrap');
  if (wrap) wrap.style.display = 'flex';
  results.innerHTML =
    '<div class="px-hint" style="margin-bottom:8px">' + t('acPickDomain') + '</div>' +
    '<div style="display:flex;gap:8px;flex-wrap:wrap">' +
      '<button class="btn btn-sm btn-p" onclick="acPick(\'railway\')">' + esc(AC_PING.user == null ? '—' : AC_PING.user + 'ms') + ' ▸ railway.app (' + (AC_PING.railway==null?'✗':AC_PING.railway+'ms') + ')</button>' +
      '<button class="btn btn-sm btn-g" onclick="acPick(\'cloudflare\')">worker.dev (' + (AC_PING.worker==null?'✗':AC_PING.worker+'ms') + ') ▸ cloudflare</button>' +
    '</div>';
}

async function acPick(which) {
  AC_CHOSEN = which;
  var log = $id('ac-log');
  var results = $id('ac-results');
  if (results) results.innerHTML = '';
  var domWrap = document.getElementById('ac-domain-wrap');
  if (domWrap) domWrap.style.display = 'none';
  acLog('✓ ' + (which === 'railway' ? t('acScanningRailway') : t('acScanningCf')));
  await acScanProvider(which);
  acLog(t('acTcpTesting'));
  await acTcpTest();
  acLog(t('acSniTesting'));
  await acSniTest();
  acLog(t('acFinalPing'));
  var finalPing = await acPingOne(($id('ac-domain').value || '').trim().toLowerCase());
  acLog((($id('ac-domain').value||'').trim()) + ' → ' + (finalPing == null ? '✗' : finalPing + ' ms'));
  acLog('✅ ' + t('acDone'));
}

async function acScanProvider(which) {
  // Use the same IP ranges the manual scanner uses, ping from user's browser.
  var ranges = which === 'railway' ? SCN_RW_RANGES : SCN_CF_RANGES;
  var port = which === 'railway' ? 443 : 443;
  var count = 60, found = [], tried = 0, loops = 0, BATCH = 32;
  while (found.length < 10 && loops < 120 && ranges.length) {
    loops++;
    var batch = [];
    for (var b=0; b<BATCH && batch.length < 40; b++)
      batch.push(scnRandIp(ranges[Math.floor(Math.random()*ranges.length)]) + ':' + port);
    var resMap = await scnPingBatch(batch, 1.2);
    Object.keys(resMap).forEach(function(k){
      tried++;
      var r = resMap[k];
      if (r && r.ok) found.push({ip: k.split(':')[0], port: r.port, ms: r.latency_ms});
    });
    found.sort(function(a,c){return a.ms - c.ms;});
    found = found.slice(0, 10);
    if (loops % 4 === 0) acLog('   یافته: ' + found.length + '/10 (تست ' + tried + ')');
  }
  found.sort(function(a,c){return a.ms - c.ms;});
  found = found.slice(0, 10);
  // Save top 10 (backend caps at 10) with 1-decimal ping precision.
  try {
    var save = await api('/api/scanner/save', {method:'POST', body:{type: which, ips: found.map(function(f){return f.ip+':'+f.port;}), replace: true, seq: (window['scnCtx'] && scnCtx[which] && scnCtx[which].seq) || 0}});
    if (save && typeof save.seq === 'number' && window['scnCtx'] && scnCtx[which]) scnCtx[which].seq = save.seq;
  } catch(e){}
  found.forEach(function(f){ acLog('   💾 ' + f.ip + ':' + f.port + ' (' + (Math.round(f.ms*10)/10) + ' ms)'); });
}

async function acTcpTest() {
  var domains = scnDomains.slice();
  var userDom = ($id('ac-domain').value || '').trim().toLowerCase();
  if (userDom && domains.indexOf(userDom) === -1) domains.unshift(userDom);
  var targets = [];
  domains.forEach(function(d){ targets.push(d + ':443'); });
  var resMap = await scnPingBatch(targets, 3);
  for (var i=0; i<domains.length; i++) {
    var dom = domains[i];
    var ip = null;
    try { var dd = await api('/api/scanner/resolve?host=' + encodeURIComponent(dom)); if (dd.ips && dd.ips.length) ip = dd.ips[0]; } catch(e){}
    var r = resMap[dom + ':443'];
    var domMs = (r && r.ok) ? r.latency_ms : null;
    var ipMs = null;
    if (ip) { var ipr = await scnPingBatch([ip + ':443'], 3); ipMs = (ipr[ip+':443'] && ipr[ip+':443'].ok) ? ipr[ip+':443'].latency_ms : null; }
    acLog('   🔌 ' + esc(dom) + (ip ? ' (' + ip + ')' : '') + ' → dom:' + (domMs==null?'✗':domMs+'ms') + ' ip:' + (ipMs==null?'✗':ipMs+'ms'));
  }
}

async function acSniTest() {
  var list = SPF_SNI_LIST.slice();
  if (!list.length) { try { var sr = await api('/api/scanner/sni-list'); list = sr.sni_list || []; } catch(e){} }
  if (!list.length) { acLog('   (لیست SNI خالی)'); return; }
  var opts = $id('spf-sni-port') ? $id('spf-sni-port').value : '443';
  var port = opts || '443';
  var results = [];
  for (var i=0; i<list.length; i+=8) {
    var batch = list.slice(i, i+8);
    var proms = batch.map(function(s){
      return api('/api/scanner/sni-check?host=' + encodeURIComponent(s) + '&port=' + port)
        .then(function(r){ return r; }).catch(function(){ return {ok:false}; });
    });
    var rs = await Promise.all(proms);
    rs.forEach(function(r, j){
      if (r && r.ok) results.push({sni: batch[j], ms: r.latency_ms});
    });
  }
  results.sort(function(a,c){return a.ms - c.ms;});
  var top = results.slice(0, 5);
  try {
    await api('/api/scanner/sni-save-results', {method:'POST', body:{results: top.map(function(x){return {sni:x.sni, latency_ms:x.ms};}), top: 5}});
  } catch(e){}
  top.forEach(function(x){ acLog('   ⚡ ' + esc(x.sni) + ' (' + (Math.round(x.ms*10)/10) + ' ms)'); });
}

// ═══ TUNNEL ═══
async function loadTunnel() {
  var notready = $id('tunnel-notready');
  var ready = $id('tunnel-ready');
  if (!notready || !ready) return;
  try {
    var s = await api('/api/tunnel/status');
    var showReady = s.worker_connected;
    notready.classList.toggle('hidden', showReady);
    ready.classList.toggle('hidden', !showReady);
    if (!showReady) return;
    var badge = $id('tn-status-badge');
    if (badge) {
      badge.textContent = (s.remote_status === 'online') ? '● online' : '○ offline';
      badge.className = 'badge ' + ((s.remote_status === 'online') ? 'badge-g' : 'badge-r');
    }
    var pingEl = $id('tn-ping');
    if (pingEl) pingEl.textContent = s.ping_ms != null ? (s.ping_ms + ' ms') : 'unreachable';
    var urlEl = $id('tn-url');
    if (urlEl) { urlEl.textContent = s.worker_url || '—'; urlEl.href = s.worker_url || '#'; }
    var kvEl = $id('tn-kv');
    if (kvEl) kvEl.textContent = (s.tunnel_kv_title ? s.tunnel_kv_title + ' · ' : '') + (s.tunnel_kv_id || '—');
    var pathEl = $id('tn-inbound-path');
    if (pathEl) pathEl.textContent = s.inbound_exists ? (s.panel_domain + '/tunnel/{uuid}  ·  TLS+WS') : '—';
    var hbEl = $id('tn-heartbeat');
    if (hbEl) hbEl.textContent = s.last_heartbeat || '—';
    var logsEl = $id('tn-logs');
    if (logsEl) {
      logsEl.innerHTML = (s.logs || []).slice().reverse().map(function(l){
        return '<div class="scn-item"><span class="scn-ip">' + esc(l.time.slice(11)) + '</span><span style="font-size:12px;color:var(--txt2)">' + esc(l.msg) + '</span></div>';
      }).join('') || '<div class="scn-empty">لاگی ثبت نشده</div>';
    }
    var btn = $id('tn-create-btn');
    if (btn) {
      var exists = !!s.inbound_exists;
      btn.disabled = exists;
      btn.textContent = exists ? t('tunnelInboundPath') + ': ✓' : t('tunnelCreate');
      if (exists) { btn.disabled = false; btn.textContent = t('tunnelCreate'); }
    }
    // Reverse section (below tunnel info)
    var sw = $id('tn-reverse-sw');
    if (sw) sw.classList.toggle('on', !!s.reverse_enabled);
    var revInfo = $id('tn-reverse-info');
    if (revInfo) revInfo.classList.toggle('hidden', !s.reverse_enabled);
    if (s.reverse_enabled) {
      var rPing = $id('tn-reverse-ping');
      if (rPing) rPing.textContent = s.reverse_ping_ms != null ? (s.reverse_ping_ms + ' ms') : 'unreachable';
      var rPath = $id('tn-reverse-path');
      if (rPath) rPath.textContent = s.worker_domain + '/reverse/{uuid}';
      var rKv = $id('tn-reverse-kv');
      if (rKv) rKv.textContent = (s.reverse_kv_title ? s.reverse_kv_title + ' · ' : '') + (s.reverse_kv_id || '—');
    }
  } catch(e) {}
}

async function tunnelToggleReverse(sw) {
  if (!sw) return;
  var enabling = !sw.classList.contains('on');
  sw.disabled = true;
  try {
    await api('/api/tunnel/reverse', {method: 'POST', body: {enabled: enabling}});
    toast(enabling ? t('reverseOn') : t('reverseOff'), 'ok');
  } catch(e) {
    toast(e.message || t('toast.err'), 'err');
  }
  sw.disabled = false;
  loadTunnel();
}

async function tunnelCreate() {
  var btn = $id('tn-create-btn');
  if (!btn) return;
  btn.disabled = true;
  btn.textContent = t('updateRunning');
  try {
    await api('/api/tunnel/create', {method: 'POST'});
    toast(t('tunnelCreated'), 'ok');
  } catch(e) { toast(e.message || t('toast.err'), 'err'); }
  btn.disabled = false;
  loadTunnel();
}

// ═══ VERCEL DOMAIN — client WSS -> Vercel WS bridge -> Railway ═══
var VD_STATE = {connected: false, configs: [], domain: '', url: ''};

function vdSetMessage(message, isError) {
  var el = $id('vd-message');
  if (!el) return;
  el.textContent = message || '';
  el.style.color = isError ? 'var(--red)' : 'var(--txt2)';
}

function renderVercelDomain(data) {
  VD_STATE = data || {connected: false, configs: [], domain: '', url: ''};
  VD_STATE.configs = Array.isArray(VD_STATE.configs) ? VD_STATE.configs : [];
  var connected = !!VD_STATE.connected;
  var setup = $id('vd-setup'), managed = $id('vd-managed'), badge = $id('vd-badge');
  if (setup) setup.classList.toggle('hidden', connected);
  if (managed) managed.classList.toggle('hidden', !connected);
  if (badge) {
    badge.textContent = connected ? t('connected') : t('disconnected');
    badge.className = 'badge ' + (connected ? 'badge-g' : 'badge-r');
  }
  var panelDomain = $id('vd-panel-domain'); if (panelDomain) panelDomain.textContent = VD_STATE.target_domain || '—';
  var domain = $id('vd-domain');
  if (domain) {
    domain.textContent = VD_STATE.domain || '—';
    domain.href = VD_STATE.url || (VD_STATE.domain ? 'https://' + VD_STATE.domain : '#');
  }
  var project = $id('vd-project'); if (project) project.textContent = VD_STATE.project_name || '—';
  var count = $id('vd-config-count'); if (count) count.textContent = String(VD_STATE.configs.length);
  var select = $id('vd-user-select');
  var oldValue = select ? select.value : '';
  if (select) {
    select.replaceChildren();
    VD_STATE.configs.forEach(function (item, index) {
      var opt = document.createElement('option');
      opt.value = String(index);
      opt.textContent = (item.username || item.user_id || ('User ' + (index + 1))) + (item.active ? '' : ' · ' + t('inactive'));
      select.appendChild(opt);
    });
    if (oldValue !== '' && Number(oldValue) < VD_STATE.configs.length) select.value = oldValue;
    select.disabled = !VD_STATE.configs.length;
  }
  var empty = $id('vd-config-empty'); if (empty) { empty.classList.toggle('hidden', !connected || VD_STATE.configs.length > 0); empty.textContent = VD_STATE.legacy_protocol ? t('vercelDomain.legacyProtocol') : t('vercelDomain.noUsers'); }
  var output = $id('vd-config-output');
  if (output) output.value = VD_STATE.configs.length ? String(VD_STATE.configs[Number(select && select.value) || 0].config || '') : '';
  var pingResult = $id('vd-ping-result');
  if (pingResult && !connected) pingResult.textContent = '';
  if (VD_STATE.last_error) vdSetMessage(VD_STATE.last_error, true);
  else if (VD_STATE.legacy_protocol) vdSetMessage(t('vercelDomain.legacyProtocol'), true);
  else if (connected) vdSetMessage(t('vercelDomain.connectedHint'), false);
  else vdSetMessage('', false);
}

async function loadVercelDomain() {
  try {
    var data = await api('/api/system/vercel-domain');
    renderVercelDomain(data);
  } catch (e) {
    vdSetMessage(e.message || t('vercelDomain.loadError'), true);
  }
}

async function deployVercelDomain() {
  var input = $id('vd-token'), button = $id('vd-deploy-btn');
  var token = input ? input.value.trim() : '';
  if (!token) { toast(t('vercelDomain.enterToken'), 'err'); if (input) input.focus(); return; }
  if (button) {
    button.disabled = true;
    var label = button.querySelector('[data-i18n="vercelDomain.create"]');
    if (label) label.textContent = t('vercelDomain.deploying');
  }
  vdSetMessage(t('vercelDomain.deploying'), false);
  try {
    var data = await api('/api/system/vercel-domain', {method: 'POST', body: {token: token}});
    if (input) input.value = '';
    renderVercelDomain(data);
    toast(t('vercelDomain.created'), 'ok');
  } catch (e) {
    vdSetMessage(e.message || t('vercelDomain.createError'), true);
    toast(e.message || t('vercelDomain.createError'), 'err');
  } finally {
    if (button) {
      button.disabled = false;
      var label = button.querySelector('[data-i18n="vercelDomain.create"]');
      if (label) label.textContent = t('vercelDomain.create');
    }
  }
}

async function deleteVercelDomain() {
  if (!confirm(t('vercelDomain.confirmDelete'))) return;
  try {
    var data = await api('/api/system/vercel-domain', {method: 'DELETE'});
    renderVercelDomain(data);
    toast(t('vercelDomain.deleted'), 'ok');
  } catch (e) {
    vdSetMessage(e.message || t('vercelDomain.deleteError'), true);
    toast(e.message || t('vercelDomain.deleteError'), 'err');
  }
}

function selectVercelConfig() {
  var select = $id('vd-user-select'), output = $id('vd-config-output');
  if (!select || !output) return;
  var item = VD_STATE.configs[Number(select.value)];
  output.value = item ? String(item.config || '') : '';
}

async function copyVercelConfig() {
  var output = $id('vd-config-output');
  if (!output || !output.value) { toast(t('vercelDomain.noConfig'), 'err'); return; }
  try {
    if (navigator.clipboard && window.isSecureContext) await navigator.clipboard.writeText(output.value);
    else {
      output.focus(); output.select();
      if (!document.execCommand('copy')) throw new Error('copy failed');
    }
    toast(t('copied'), 'ok');
  } catch (e) {
    toast(t('errCopy'), 'err');
  }
}

async function pingVercelDomain() {
  var button = $id('vd-ping-btn'), out = $id('vd-ping-result');
  var base = String(VD_STATE.url || (VD_STATE.domain ? 'https://' + VD_STATE.domain : '')).replace(/\/$/, '');
  if (!base || !VD_STATE.connected) { vdSetMessage(t('vercelDomain.notConnected'), true); return; }
  if (button) {
    button.disabled = true;
    var pingLabel = button.querySelector('[data-i18n="vercelDomain.ping"]');
    if (pingLabel) pingLabel.textContent = t('vercelDomain.pinging');
  }
  if (out) out.textContent = t('vercelDomain.pingingHint');
  var samples = [];
  for (var i = 0; i < 4; i++) {
    var controller = new AbortController();
    var timer = setTimeout(function (c) { return function () { c.abort(); }; }(controller), 8000);
    var start = performance.now();
    try {
      // Measure HTTPS reachability to the production relay from THIS browser.
      // The health handler does not need CORS access because no-cors timing is enough.
      await fetch(base + '/_white-health?cb=' + Date.now() + '-' + i, {
        method: 'GET', mode: 'no-cors', cache: 'no-store', signal: controller.signal
      });
      samples.push(Math.round(performance.now() - start));
    } catch (_) {
      // A failed browser fetch (DNS, TLS, timeout or blocked network) is not a ping result.
    } finally { clearTimeout(timer); }
    if (i < 3) await new Promise(function (resolve) { setTimeout(resolve, 180); });
  }
  if (button) {
    button.disabled = false;
    var pingLabel = button.querySelector('[data-i18n="vercelDomain.ping"]');
    if (pingLabel) pingLabel.textContent = t('vercelDomain.ping');
  }
  if (!samples.length) {
    if (out) out.textContent = t('vercelDomain.pingFailed');
    return;
  }
  var sorted = samples.slice().sort(function (a, b) { return a - b; });
  var median = sorted.length % 2 ? sorted[(sorted.length - 1) / 2] : Math.round((sorted[sorted.length / 2 - 1] + sorted[sorted.length / 2]) / 2);
  if (out) out.textContent = t('vercelDomain.clientPingResult').replace('{median}', median).replace('{min}', sorted[0]).replace('{max}', sorted[sorted.length - 1]).replace('{samples}', samples.length);
}

// ═══ CLOUDFLARE WORKER — Tunnel / Reverse / Worker as three sections ═══
var WK_STATE = {connected: false, worker: {}, tunnel: {}, reverse: {}, locations: [], proxies: []};
var WK_HB_TIMER = null;

function wkFlag(code) {
  var c = String(code || '').trim().toUpperCase();
  if (c.length !== 2 || !/^[A-Z]{2}$/.test(c)) return '🌐';
  return String.fromCodePoint(0x1F1E6 + c.charCodeAt(0) - 65, 0x1F1E6 + c.charCodeAt(1) - 65);
}

function wkSetBadge(el, on, onText, offText) {
  if (!el) return;
  el.textContent = on ? onText : offText;
  el.className = 'badge ' + (on ? 'badge-g' : 'badge-b');
}

function wkFmtTime(iso) {
  if (!iso) return '—';
  return String(iso).replace('T', ' ').slice(0, 19);
}

async function loadWorkerSections() {
  var badge = $id('worker-badge');
  try {
    var d = await api('/api/worker/sections');
    var w = d.worker || {}, tn = d.tunnel || {}, rv = d.reverse || {};
    WK_STATE = {connected: !!w.connected, worker: w, tunnel: tn, reverse: rv, locations: w.locations || []};
    if (typeof WK_PROXY_LIST !== 'undefined') WK_PROXY_LIST = WK_STATE.locations;

    if (badge) {
      badge.textContent = w.connected ? t('connected') : t('disconnected');
      badge.className = 'badge ' + (w.connected ? 'badge-g' : 'badge-r');
    }

    var setup = $id('wk-setup'), managed = $id('wk-managed');
    if (setup) setup.classList.toggle('hidden', !!w.connected);
    if (managed) managed.classList.toggle('hidden', !w.connected);

    // ── 1. TUNNEL ──
    wkSetBadge($id('wk-tun-badge'), !!tn.enabled, t('connected'), t('notCreated'));
    wkSetBadge($id('wk-tun-status'), !!tn.enabled, t('connected'), t('notCreated'));
    var tp = $id('wk-tun-ping'); if (tp) tp.textContent = tn.ping_ms != null ? tn.ping_ms + ' ms' : '—';
    var td = $id('wk-tun-domain'); if (td) td.textContent = tn.domain || '—';
    var tpa = $id('wk-tun-path'); if (tpa) tpa.textContent = tn.path || '—';
    var tkv = $id('wk-tun-kv'); if (tkv) tkv.textContent = (tn.kv_title ? tn.kv_title + ' · ' : '') + (tn.kv_id || '—');
    var thb = $id('wk-tun-hb'); if (thb) thb.textContent = wkFmtTime(tn.last_heartbeat);
    var tib = $id('wk-tun-inbound'); if (tib) tib.textContent = tn.inbound_id ? tn.inbound_id : '—';
    var tlogs = $id('wk-tun-logs');
    if (tlogs) {
      tlogs.innerHTML = (tn.logs || []).slice().reverse().map(function (l) {
        return '<div class="scn-item"><span class="scn-ip">' + esc(String(l.time || '').slice(11)) + '</span>' +
               '<span style="font-size:12px;color:var(--txt2)">' + esc(l.msg) + '</span></div>';
      }).join('') || '<div class="scn-empty">' + esc(t('wk.noLogs')) + '</div>';
    }

    // ── 2. REVERSE ──
    wkSetBadge($id('wk-rev-badge'), !!rv.enabled, t('connected'), t('notCreated'));
    wkSetBadge($id('wk-rev-status'), !!rv.enabled, t('connected'), t('notCreated'));
    var rp = $id('wk-rev-ping'); if (rp) rp.textContent = rv.ping_ms != null ? rv.ping_ms + ' ms' : '—';
    var rd = $id('wk-rev-domain'); if (rd) rd.textContent = rv.domain || '—';
    var rpa = $id('wk-rev-path'); if (rpa) rpa.textContent = rv.path || '—';
    var rkv = $id('wk-rev-kv'); if (rkv) rkv.textContent = (rv.kv_title ? rv.kv_title + ' · ' : '') + (rv.kv_id || '—');
    var rib = $id('wk-rev-inbound'); if (rib) rib.textContent = rv.inbound_id ? rv.inbound_id : '—';

    // ── 3. WORKER ──
    wkSetBadge($id('wk-badge'), !!w.connected, t('connected'), t('disconnected'));
    wkSetBadge($id('wk-status'), !!w.connected, t('connected'), t('disconnected'));
    var rOk = w.remote_status === 'online';
    var rs = $id('wk-remote');
    if (rs) { rs.textContent = w.remote_status || '—'; rs.className = 'badge ' + (rOk ? 'badge-g' : 'badge-r'); }
    var u = $id('wk-url'); if (u) { u.textContent = w.worker_url || ''; u.href = w.worker_url || '#'; }
    var nm = $id('wk-name'); if (nm) nm.textContent = w.worker_name || '—';
    var kv = $id('wk-kv'); if (kv) kv.textContent = (w.kv_title ? w.kv_title + ' · ' : '') + (w.kv_id || '—');
    var wping = $id('wk-ping'); if (wping) wping.textContent = w.ping_ms != null ? w.ping_ms + ' ms' : '—';
    var on = $id('wk-online'); if (on) on.textContent = String(w.online || 0);
    var tr = $id('wk-traffic'); if (tr) tr.textContent = fmtBytes(w.traffic_bytes || 0);
    var sy = $id('wk-sync'); if (sy) sy.textContent = wkFmtTime(w.last_sync);
    var er = $id('wk-error'); if (er) er.textContent = w.last_error || '';
    var wi = $id('wk-inbound');
    if (wi) wi.textContent = w.inbound_id ? w.inbound_id + '  ·  ' + (w.inbound_domain || '') + w.inbound_path : '—';
    var auto = $id('wk-auto'); if (auto) auto.classList.toggle('on', !!w.auto_sync);

    renderWorkerLocations();
    if (w.connected) wkHeartbeat();
  } catch (e) {
    if (badge) { badge.textContent = t('disconnected'); badge.className = 'badge badge-r'; }
  }
}

// Back-compat alias: the tab and a few helpers call loadWorker().
async function loadWorker() { return loadWorkerSections(); }

function renderWorkerLocations() {
  var box = $id('wk-locations');
  if (!box) return;
  var list = WK_STATE.locations || [];
  var cnt = $id('wk-loc-count');
  if (cnt) cnt.textContent = list.length + ' ' + t('wk.location');
  if (!list.length) {
    box.innerHTML = '<div class="px-empty">' + esc(t('wk.noLocations')) + '</div>';
    return;
  }
  box.innerHTML = list.map(function (p) {
    return '<div class="set-row" style="border-bottom:1px solid var(--border)">' +
      '<span class="set-lbl">' + esc(p.country_flag || wkFlag(p.code)) + ' ' + esc(p.country) +
        (p.manual ? '' : ' <span class="badge badge-b" style="font-size:9px">' + esc(t('wk.auto')) + '</span>') + '</span>' +
      '<span style="display:flex;align-items:center;gap:8px">' +
        '<span dir="ltr" style="font-size:11px;font-family:monospace;color:var(--txt2)">' + esc(p.proxy) + ':' + esc(String(p.port)) + '</span>' +
        '<button class="btn btn-sm btn-r" type="button" onclick="wkDelLocation(\'' + escAttr(String(p.code)) + '\')">' + esc(t('delete')) + '</button>' +
      '</span></div>';
  }).join('');
}

async function workerCreateTunnel() {
  try {
    await api('/api/tunnel/create', {method: 'POST'});
    toast(t('tunnelCreated'), 'ok');
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
  await loadWorkerSections();
}

async function workerCreateReverse() {
  try {
    await api('/api/reverse/create', {method: 'POST'});
    toast(t('wk.reverseCreated'), 'ok');
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
  await loadWorkerSections();
}

async function workerCreateInbound() {
  try {
    await api('/api/worker/inbound', {method: 'POST'});
    toast(t('wk.inboundCreated'), 'ok');
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
  await loadWorkerSections();
}

async function workerSetup() {
  var token = ($id('wk-token').value || '').trim();
  if (!token) { toast(t('workerFillAll'), 'err'); return; }
  var btn = document.querySelector('#wk-setup .btn-p');
  if (btn) { btn.disabled = true; btn.textContent = t('connecting'); }
  try {
    var w = await api('/api/worker/setup', {method: 'POST', body: {token: token}});
    $id('wk-token').value = '';
    if (w && w.deferred_sync) toast(t('workerConnected') + ' — ' + (w.deferred_reason || ''), 'warn');
    else toast(t('workerConnected'), 'ok');
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
  if (btn) { btn.disabled = false; btn.textContent = t('createNewWorker'); }
  await loadWorkerSections();
}

function wkHeartbeat() {
  if (WK_HB_TIMER) clearTimeout(WK_HB_TIMER);
  WK_HB_TIMER = setTimeout(async function () {
    if (!WK_STATE.connected) return;
    try {
      var w = await api('/api/worker/heartbeat', {method: 'POST'});
      var rs = $id('wk-remote');
      if (rs) { rs.textContent = w.ok ? 'online' : (w.detail || 'unreachable'); rs.className = 'badge ' + (w.ok ? 'badge-g' : 'badge-r'); }
      if (w.ok) {
        var on = $id('wk-online'); if (on) on.textContent = String(w.worker_users_online || 0);
        var tr = $id('wk-traffic'); if (tr) tr.textContent = fmtBytes(w.worker_traffic_bytes || 0);
        var hb = $id('wk-tun-hb'); if (hb) hb.textContent = wkFmtTime(w.last_heartbeat);
      }
    } catch (e) {}
    wkHeartbeat();
  }, 15000);
}

async function workerSync() {
  toast(t('syncing'), 'warn');
  try {
    await api('/api/worker/sync', {method: 'POST'});
    toast(t('synced'), 'ok');
  } catch (e) { toast(e.message, 'err'); }
  await loadWorkerSections();
}

function workerToggleAuto() {
  var sw = $id('wk-auto');
  if (!sw) return;
  sw.classList.toggle('on');
  api('/api/worker/settings', {method: 'POST', body: {auto_sync: sw.classList.contains('on')}})
    .then(loadWorkerSections)
    .catch(function (e) { toast(e.message, 'err'); sw.classList.toggle('on'); });
}

async function workerSyncSource() {
  var st = $id('wk-source-status');
  if (st) st.textContent = t('updatingFromSource') + '...';
  try {
    var r = await api('/api/worker/sync-source', {method: 'POST'});
    if (st) {
      st.textContent = (r.sync && r.sync.ok)
        ? t('sourceUpdated') + ': ' + r.sync.countries + ' ' + t('countries')
        : t('sourceError');
      st.style.color = (r.sync && r.sync.ok) ? 'var(--green)' : 'var(--red)';
    }
    if (r.sync && r.sync.ok) toast(t('sourceUpdated') + ' ✓', 'ok');
    else toast(t('sourceError'), 'err');
  } catch (e) {
    if (st) { st.textContent = e.message || t('sourceError'); st.style.color = 'var(--red)'; }
    toast(e.message || t('sourceError'), 'err');
  }
  await loadWorkerSections();
}

async function workerDisconnect() {
  if (!confirm(t('confirmRemoveWorker'))) return;
  try {
    await api('/api/worker', {method: 'DELETE'});
    toast(t('workerRemoved'), 'ok');
  } catch (e) { toast(e.message, 'err'); }
  await loadWorkerSections();
}

function wkAutoCode(country, host) {
  var c = String(country || '').trim();
  if (/^[A-Za-z]{2}$/.test(c)) return c.toLowerCase();
  var h = String(host || '').split('.')[0] || '';
  var m = h.match(/^([a-z]{2})[-.]/i);
  if (m) return m[1].toLowerCase();
  return (h.replace(/[^a-z0-9]/gi, '').slice(0, 8) || 'loc').toLowerCase();
}

async function wkAddLocation() {
  var country = ($id('wk-loc-country').value || '').trim();
  var host = ($id('wk-loc-host').value || '').trim();
  var port = parseInt($id('wk-loc-port').value || '443', 10);
  var code = ($id('wk-loc-code').value || '').trim().toLowerCase() || wkAutoCode(country, host);
  if (!country || !host) { toast(t('workerFillAll'), 'err'); return; }
  try {
    await api('/api/worker/proxies', {
      method: 'POST',
      body: {code: code, country: country, proxy: host, port: port}
    });
    $id('wk-loc-country').value = ''; $id('wk-loc-code').value = '';
    $id('wk-loc-host').value = ''; $id('wk-loc-port').value = '443';
    toast(t('proxySaved'), 'ok');
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
  await loadWorkerSections();
}

async function wkDelLocation(code) {
  if (!confirm(t('confirmDelProxy'))) return;
  try {
    await api('/api/worker/proxies/' + encodeURIComponent(code), {method: 'DELETE'});
    toast(t('proxyDeleted'), 'ok');
  } catch (e) { toast(e.message, 'err'); }
  await loadWorkerSections();
}

// ═══ EXTERNAL PROXIES (PROXY tab) ═══
// Flow: fill the form → «تست اتصال» → a ping/working report appears → only then
// does the Save button show up. Each saved proxy renders as an accordion whose
// body holds ping / delete / active / users (a scrollable checklist).
var PX_STATE = {list: [], tested: null};

function pxToggleAdd(btn) {
  var box = $id('px-add');
  if (!box) return;
  var opening = box.classList.contains('hidden');
  box.classList.toggle('hidden', !opening);
  if (btn) btn.setAttribute('aria-expanded', String(opening));
  if (!opening) pxResetAdd();
}

function pxInvalidateTest() {
  PX_STATE.tested = null;
  var out = $id('px-test-result');
  if (out) { out.classList.remove('show'); out.innerHTML = ''; }
  var sv = $id('px-save-btn');
  if (sv) sv.classList.add('hidden');
}

function pxResetAdd() {
  ['px-host', 'px-country', 'px-username', 'px-password'].forEach(function(id) {
    var el = $id(id); if (el) el.value = '';
  });
  var port = $id('px-port'); if (port) port.value = '';
  pxInvalidateTest();
}

function pxTestPayload() {
  return {
    host: (($id('px-host') || {}).value || '').trim(),
    port: parseInt(($id('px-port') || {}).value || '0', 10),
    username: (($id('px-username') || {}).value || '').trim(),
    password: ($id('px-password') || {}).value || '',
    country: (($id('px-country') || {}).value || '').trim()
  };
}

function pxTestHtml(r) {
  var ok = !!r.working;
  return '<b style="color:' + (ok ? 'var(--green)' : 'var(--red)') + '">' +
    (ok ? '✔ ' : '✖ ') + esc(ok ? t('proxy.working') : t('proxy.notWorking')) + '</b><br>' +
    esc(t('proxy.protocol')) + ': <b dir="ltr">' + esc(r.protocol || '—') + '</b><br>' +
    (r.tls_hostname_verified === false ? '<span style="color:#eab308">' + esc(t('proxy.ipTlsNameWarning')) + '</span><br>' : '') +
    esc(t('proxy.ping')) + ': <b dir="ltr">' + (r.ping_ms != null ? esc(String(r.ping_ms)) + ' ms' : '—') + '</b><br>' +
    esc(t('proxy.egress')) + ': <b dir="ltr">' + esc(r.egress_ip || '—') + '</b>' +
    (r.country_name ? '<br>' + esc(t('proxy.detectedCountry')) + ': <b>' + esc((r.country_flag || '') + ' ' + r.country_name) + '</b>' : '') +
    (r.error ? '<br><span style="color:var(--red)">' + esc(r.error) + '</span>' : '');
}

async function pxTestNew() {
  var btn = $id('px-test-btn'), out = $id('px-test-result'), sv = $id('px-save-btn');
  var payload = pxTestPayload();
  if (!payload.host || !payload.port) { toast(t('proxy.fillHost'), 'err'); return; }
  PX_STATE.tested = null;
  if (sv) sv.classList.add('hidden');
  if (out) { out.classList.add('show'); out.innerHTML = '<span style="color:var(--txt2)">⏳ ' + esc(t('proxy.testing')) + '…</span>'; }
  if (btn) btn.disabled = true;
  try {
    var r = await api('/api/proxies/test', {method: 'POST', body: payload});
    if (out) out.innerHTML = pxTestHtml(r);
    if (r.working) {
      PX_STATE.tested = Object.assign({}, payload, {
        protocol: r.protocol || '', working: true, ping_ms: r.ping_ms, egress_ip: r.egress_ip || '',
        country_code: r.country_code || '', country_flag: r.country_flag || '', country_name: r.country_name || ''
      });
      if (sv) sv.classList.remove('hidden');
      toast(t('proxy.testOk'), 'ok');
    } else {
      toast(r.error || t('proxy.testFail'), 'err');
    }
  } catch (e) {
    if (out) out.innerHTML = '<span style="color:var(--red)">✖ ' + esc(e.message || '') + '</span>';
    toast(e.message || t('toast.err'), 'err');
  }
  if (btn) btn.disabled = false;
}

async function pxSaveNew() {
  if (!PX_STATE.tested) { toast(t('proxy.testFirst'), 'warn'); return; }
  var btn = $id('px-save-btn');
  if (btn) btn.disabled = true;
  try {
    await api('/api/proxies', {method: 'POST', body: PX_STATE.tested});
    toast(t('proxy.saved'), 'ok');
    pxResetAdd();
    var add = $id('px-add');
    if (add) add.classList.add('hidden');
    await loadProxies();
    await loadInbounds();
  } catch (e) {
    toast(e.message || t('toast.err'), 'err');
  }
  if (btn) btn.disabled = false;
}

async function loadProxies() {
  var list = $id('proxy-list');
  if (!list) return;
  try {
    var d = await api('/api/proxies');
    PX_STATE.list = d.proxies || [];
    renderProxies();
    var c = $id('px-count');
    if (c) {
      var n = PX_STATE.list.length;
      c.textContent = n + ' ' + t('proxy.count');
      c.className = 'badge ' + (n ? 'badge-g' : 'badge-b');
    }
  } catch (e) {
    list.innerHTML = '<div class="px-empty">' + esc(e.message || t('toast.err')) + '</div>';
  }
}

function pxRow(label, value) {
  return '<div class="set-row"><span class="set-lbl">' + esc(label) + '</span>' +
    '<span dir="ltr" style="font-size:12px;color:var(--txt);word-break:break-all">' + esc(String(value == null ? '—' : value)) + '</span></div>';
}

function pxCardHtml(p) {
  var statusTxt = p.active ? (p.working ? t('proxy.ok') : t('proxy.untested')) : t('proxy.off');
  var statusCls = p.active ? (p.working ? 'badge-g' : 'badge-b') : 'badge-r';
  var accId = 'px-' + String(p.id);
  return '<div class="spd-acc px-card' + (p.active ? '' : ' off') + '" id="px-card-' + escAttr(p.id) + '">' +
    '<button type="button" class="spd-acc-head px-head" onclick="spdToggleAcc(\'' + escAttr(accId) + '\',this)" aria-expanded="false" aria-controls="acc-' + escAttr(accId) + '">' +
      '<span class="px-flag">' + esc(p.country_flag || '🌐') + '</span>' +
      '<span class="px-meta"><b>' + esc(p.country || p.host) + '</b>' +
        '<span>' + esc(p.protocol) + ' · ' + esc(p.host) + ':' + esc(String(p.port)) + (p.username ? ' · ' + esc(p.username) : '') + '</span></span>' +
      '<span class="badge ' + statusCls + '">' + esc(statusTxt) + '</span>' +
      '<span class="material-symbols-outlined spd-acc-arrow">expand_more</span>' +
    '</button>' +
    '<div class="spd-acc-body px-body" id="acc-' + escAttr(accId) + '">' +
      '<div class="px-rows">' +
        pxRow(t('proxy.ping'), p.ping_ms != null ? p.ping_ms + ' ms' : '—') +
        pxRow(t('proxy.egress'), p.egress_ip || '—') +
        pxRow(t('proxy.lastTest'), p.last_test_at ? String(p.last_test_at).replace('T', ' ').slice(0, 19) : '—') +
        (p.error ? pxRow(t('proxy.error'), p.error) : '') +
      '</div>' +
      '<div class="px-actions">' +
        '<button class="btn btn-g btn-sm" type="button" onclick="event.stopPropagation();pxPing(\'' + escAttr(p.id) + '\')">🔄 ' + esc(t('proxy.pingBtn')) + '</button>' +
        '<button class="btn btn-sm ' + (p.active ? 'btn-r' : 'btn-p') + '" type="button" onclick="event.stopPropagation();pxToggleActive(\'' + escAttr(p.id) + '\',' + (p.active ? 'false' : 'true') + ')">' +
          esc(p.active ? t('proxy.deactivate') : t('proxy.activate')) + '</button>' +
        '<button class="btn btn-sm btn-r" type="button" onclick="event.stopPropagation();pxDelete(\'' + escAttr(p.id) + '\')">🗑 ' + esc(t('proxy.delete')) + '</button>' +
      '</div>' +
    '</div>' +
  '</div>';
}

function renderProxies() {
  var list = $id('proxy-list');
  if (!list) return;
  if (!PX_STATE.list.length) {
    list.innerHTML = '<div class="px-empty">' + esc(t('proxy.empty')) + '</div>';
    return;
  }
  list.innerHTML = PX_STATE.list.map(pxCardHtml).join('');
}

async function pxPing(pid) {
  try {
    var r = await api('/api/proxies/' + encodeURIComponent(pid) + '/ping', {method: 'POST'});
    toast(t('proxy.ping') + ': ' + (r.ping_ms != null ? r.ping_ms + ' ms' : '—'), r.working ? 'ok' : 'err');
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
  await loadProxies();
}

async function pxToggleActive(pid, on) {
  try {
    await api('/api/proxies/' + encodeURIComponent(pid), {method: 'PATCH', body: {active: !!on}});
    toast(on ? t('proxy.activated') : t('proxy.deactivated'), 'ok');
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
  await loadProxies();
}

async function pxDelete(pid) {
  if (!confirm(t('proxy.confirmDelete'))) return;
  try {
    await api('/api/proxies/' + encodeURIComponent(pid), {method: 'DELETE'});
    toast(t('proxy.deleted'), 'ok');
    await loadProxies();
    await loadInbounds();
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
}

// ═══ PER-INBOUND CONFIG OPTIONS (create-user modal) ═══
// For each selected inbound:
//   - worker → country multi-select dropdown (from the worker proxy pool)
//   - tls (ws/xhttp) → small iOS switch: adds 10 Railway scanned-IP configs
//   - reality → nothing (Reality can't use scanned IPs / proxy)
//   - worker → the same iOS switch additionally adds 10 Cloudflare scanned-IP configs
var cuInboundState = { picks: {}, countries: {} };
var cuWorkerProxyList = (WK_STATE && WK_STATE.proxies) || [];
function cuWorkerProxies() {
  // `locations` is the rebuilt Worker section's name for the proxy pool; the
  // `proxies` fallback keeps older payloads working.
  if (WK_STATE && (WK_STATE.locations || WK_STATE.proxies)) return WK_STATE.locations || WK_STATE.proxies;
  return cuWorkerProxyList;
}

function cuInboundType(ib) {
  var p = (ib && ib.protocol || '').toLowerCase();
  if (p === 'worker') return 'worker';
  if (p === 'telegram') return 'telegram';
  if (p === 'telegram-webproxy') return 'telegram-webproxy';
  if (p === 'reality' || (ib && ib.security || '').toLowerCase() === 'reality') return 'reality';
  return 'tls';
}

function inboundDisplayName(ib) {
  if (ib && ib._managed_kind === 'default_tls_ws' && (!ib.advanced_customized || ib.name === 'پیش‌فرض TLS + WS')) return t('defaultTlsWsInbound');
  return (ib && ib.name) || '';
}

function cuToggleIbPick(el) {
  el.classList.toggle('selected');
  var id = el.getAttribute('data-id');
  cuInboundState.picks[id] = el.classList.contains('selected');
  cuRenderInboundOpts();
}

function cuToggleIpSwitch(ibId, el) {
  el.classList.toggle('on');
  var on = el.classList.contains('on');
  cuInboundState.countries[ibId] = cuInboundState.countries[ibId] || {};
  var type = el.getAttribute('data-type');
  if (type === 'cf') cuInboundState.countries[ibId].cf = on;
  else if (type === 'railway') cuInboundState.countries[ibId].railway = on;
}
function setWebProxyEnabled(inboundId, enabled) {
  if (!confirm(enabled ? 'اینباند Web Proxy را فعال می‌کنیم' : 'اینباند Web Proxy را غیرفعال می‌کنیم\n\n' +
    'فعال‌سازی: لینک‌های کاربران باز می‌شوند. غیرفعال‌سازی: دسترسی کنسل می‌شود.')) return;
  api('/api/inbounds/' + encodeURIComponent(inboundId) + '/enabled', {
    method: 'POST', body: { enabled: !!enabled }
  }).then(function(j) {
    if (j && j.ok) {
      alert(enabled ? t('wpEnable') + ': OK' : t('wpDisable') + ': OK');
      // Refresh inbounds and the selector cache after the server confirms the change.
      location.reload();
    } else {
      alert(t('errLoadInbounds') + ': ' + ((j && (j.error || j.message)) || 'پاسخ نامعتبر از سرور'));
    }
  }).catch(function(e) { alert('Request failed: ' + (e.message || 'خطای نامشخص')); });
}

function cuPickWorkerCountries(ibId, el) {
  el.classList.toggle('selected');
  var arr = cuInboundState.countries[ibId] = cuInboundState.countries[ibId] || {};
  arr.worker = arr.worker || [];
  var code = el.getAttribute('data-code');
  var idx = arr.worker.indexOf(code);
  if (el.classList.contains('selected')) {
    if (idx === -1) arr.worker.push(code);
  } else {
    if (idx !== -1) arr.worker.splice(idx, 1);
  }
}

function cuCustomStates() {
  var out = {};
  Object.keys(cuInboundState.countries || {}).forEach(function(id) {
    out[id] = out[id] || {};
    var c = cuInboundState.countries[id] || {};
    out[id].cf = !!c.cf;
    out[id].railway = !!c.railway;
  });
  return out;
}

function cuWorkerCountries() {
  var out = {};
  Object.keys(cuInboundState.countries || {}).forEach(function(id) {
    var c = cuInboundState.countries[id] || {};
    if (c.worker && c.worker.length) out[id] = c.worker;
  });
  return out;
}

function cuRenderInboundOpts() {
  var wrap = $id('cu-inbound');
  if (!wrap) return;
  var html = '';
  ALL_IB.forEach(function(ib) {
    var id = ib.inbound_id;
    var selected = cuInboundState.picks[id] === true;
    var type = cuInboundType(ib);
    var proto = (ib.protocol || 'vless').toUpperCase();
    var meta = proto + ' · ' + (ib.network || 'ws').toUpperCase() + ' · ' + (ib.security || 'tls').toUpperCase();
    if (type === 'telegram-webproxy') meta = (ib.display_protocol || 'Telegram Web Proxy') + ' · TLS + WS · ' + (ib.status || '');
    html += '<div class="ib-pick' + (selected ? ' selected' : '') + '" data-id="' + esc(id) + '" onclick="cuToggleIbPick(this)">' +
      '<div class="ib-pick-ic"><span class="material-symbols-outlined">' + (type === 'worker' ? 'bolt' : type === 'reality' ? 'shield' : type === 'telegram-webproxy' ? 'language' : 'cloud') + '</span></div>' +
      '<div class="ib-pick-info">' +
        '<div class="ib-pick-name">' + esc(inboundDisplayName(ib)) + '</div>' +
        '<div class="ib-pick-meta">' + esc(meta) + '</div>' +
      '</div>' +
      '<div class="ib-pick-check"><span class="material-symbols-outlined">check</span></div>' +
    '</div>';
    if (selected) {
      html += '<div class="ib-opts" data-for="' + esc(id) + '" style="flex-basis:100%">';
      if (ib.proxy_managed) {
        html += '<div class="px-hint" style="margin:0">' + t('proxy.inboundHint') + '</div>';
      } else if (type === 'worker') {
        var sel = cuInboundState.countries[id] || {};
        var chipHtml = '';
        cuWorkerProxies().forEach(function(p) {
          var flag = window.cuCodeToFlag ? cuCodeToFlag(p.code) : '';
          var checked = (sel.worker || []).indexOf(p.code) !== -1 ? ' selected' : '';
          chipHtml += '<span class="country-tick' + checked + '" data-code="' + esc(p.code) + '" onclick="cuPickWorkerCountries(\'' + esc(id) + '\',this)">' +
            '<span class="ck material-symbols-outlined">check</span><span class="country-name">' + flag + ' ' + esc(p.country || p.code) + '</span></span>';
        });
        html += '<div class="set-row" style="padding:0;border:none;flex-wrap:wrap">' +
          '<span class="set-lbl" style="width:100%">' + t('workerCountries') + '</span>' +
          '<div style="display:flex;flex-wrap:wrap;gap:6px;direction:ltr;text-align:left">' + chipHtml + '</div>' +
        '</div>' +
        '<div class="set-row" style="padding:0;border:none">' +
          '<span class="set-lbl" style="font-size:11px">' + t('addCfIps') + '</span>' +
          '<button type="button" class="ios-switch' + (sel.cf ? ' on' : '') + '" data-type="cf" onclick="cuToggleIpSwitch(\'' + esc(id) + '\',this)"></button>' +
        '</div>';
      } else if (type === 'tls' && (ib.protocol || '').toLowerCase() === 'vless' && !String(id).toLowerCase().startsWith('default')) {
        var selt = cuInboundState.countries[id] || {};
        html += '<div class="set-row" style="padding:0;border:none"><span class="set-lbl" style="font-size:11px">Cloudflare IP</span>' +
          '<button type="button" class="ios-switch' + (selt.cf ? ' on' : '') + '" data-type="cf" onclick="cuToggleIpSwitch(\'' + esc(id) + '\',this)"></button></div>' +
          '<div class="set-row" style="padding:0;border:none"><span class="set-lbl" style="font-size:11px">Railway IP</span>' +
          '<button type="button" class="ios-switch' + (selt.railway ? ' on' : '') + '" data-type="railway" onclick="cuToggleIpSwitch(\'' + esc(id) + '\',this)"></button></div>';
      } else if (type === 'tls') {
        var selt2 = cuInboundState.countries[id] || {};
        html += '<div class="set-row" style="padding:0;border:none"><span class="set-lbl" style="font-size:11px">' + t('addRailwayIps') + '</span>' +
          '<button type="button" class="ios-switch' + (selt2.railway ? ' on' : '') + '" data-type="railway" onclick="cuToggleIpSwitch(\'' + esc(id) + '\',this)"></button></div>';
      } else if (type === 'telegram') {
        html += '<div class="px-hint" style="margin:0">Telegram Proxy — نیازی به IP اضافی نیست</div>';
      } else if (type === 'telegram-webproxy') {
        html += '<div class="px-hint" style="margin:0">' + esc(t('wpNoListen')) + '</div>';
      } else {
        html += '<div class="px-hint" style="margin:0">' + t('realityNoExtra') + '</div>';
      }
      html += '</div>';
    }
  });
  wrap.innerHTML = html || '<span style="font-size:11px;color:var(--txt2)">' + t('noInboundOpt') + '</span>';
}

function cuCodeToFlag(code) {
  var c = (code || '').toUpperCase();
  if (c.length !== 2 || !/^[A-Z]{2}$/.test(c)) return '';
  return String.fromCodePoint(0x1F1E6 + c.charCodeAt(0) - 65, 0x1F1E6 + c.charCodeAt(1) - 65);
}

// ─── NODES ─────────────────────────────────────────────────────────────────
var ALL_NODES = [];
var EDITING_NODE_ID = null;

function _nodeStatusText(node) {
  var st = String(node.status || node.last_status || 'offline').toLowerCase();
  if (st === 'online') return '● Online';
  if (st === 'checking') return '● Checking...';
  if (st === 'unauthorized') return '! Unauthorized';
  if (st === 'orphan_cleanup_pending') return '! Cleanup pending';
  if (st === 'error' || st === 'remote_error') return '! Error';
  return '○ Offline';
}

function _nodeStatusClass(node) {
  var st = String(node.status || node.last_status || 'offline').toLowerCase();
  if (st === 'online') return 'node-online';
  if (st === 'checking') return 'node-checking';
  if (st === 'unauthorized') return 'node-unauth';
  return 'node-offline';
}

function _nodeFlag(node) { return node.country_flag || node.remote_flag || '🌐'; }
function _nodeIp(node) { return node.public_ip || node.remote_ip || '--'; }

async function loadNodes() {
  try {
    var response = await api('/api/nodes');
    ALL_NODES = response.nodes || response || [];
    renderNodes();
  } catch(e) {
    var el = $id('nodes-list');
    if (el) el.innerHTML = '<div style="color:var(--red);padding:20px">' + esc(e.message || t('toast.err')) + '</div>';
  }
}

function renderNodes() {
  var wrap = $id('nodes-list');
  if (!wrap) return;
  if (!ALL_NODES.length) {
    wrap.innerHTML = '<div style="color:var(--txt2);text-align:center;padding:40px 0;font-size:14px">' + esc(t('noNodes')) + '</div>';
    return;
  }
  var html = '';
  ALL_NODES.forEach(function(node, nodeIndex) {
    var flag = _nodeFlag(node);
    var ip = _nodeIp(node);
    var country = node.country || node.country_code || '';
    var status = _nodeStatusText(node);
    var statusCls = _nodeStatusClass(node);
    var lat = node.latency_ms != null ? node.latency_ms + ' ms' : (node.latency != null ? node.latency + ' ms' : '—');
    var lastSeen = node.last_seen || node.last_checked || '—';
    var lastSync = node.last_sync || '—';
    var lastErr = node.last_error || node.error || '—';
    var nodeAccId = 'node-item-' + nodeIndex;
    html += '<div class="spd-acc node-item-accordion" style="margin-bottom:8px">' +
      '<button type="button" class="spd-acc-head" onclick="toggleNodeAccordion(this)" aria-expanded="false" aria-controls="acc-' + nodeAccId + '\">' +
        '<span class="spd-acc-title"><strong>' + esc(flag + ' ' + (node.name || node.domain || node.node_id)) + '</strong><span style="font-size:10px;color:var(--txt2);display:block;margin-top:3px" dir="ltr">' + esc(country + (country && ip ? ' · ' : '') + ip) + '</span></span>' +
        '<span class="node-status-pill ' + statusCls + '">' + esc(status) + '</span>' +
        '<span class="material-symbols-outlined spd-acc-arrow">expand_more</span>' +
      '</button>' +
      '<div class="spd-acc-body" id="acc-' + nodeAccId + '" aria-hidden="true">' +
        '<div class="set-row"><span class="set-lbl">Domain</span><span dir="ltr" style="font-size:12px">' + esc(node.domain || '—') + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">IP / Country</span><span dir="ltr" style="font-size:12px">' + esc(ip) + ' · ' + esc(flag + ' ' + country) + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">Status</span><span class="node-status-pill ' + statusCls + '">' + esc(status) + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">Latency</span><span dir="ltr" style="font-size:12px">' + esc(lat) + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">Last Seen</span><span dir="ltr" style="font-size:11px">' + esc(lastSeen) + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">Last Sync</span><span dir="ltr" style="font-size:11px">' + esc(lastSync) + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">Inbound Count</span><b>' + esc(String(node.enabled_inbound_count != null ? node.enabled_inbound_count : 0)) + '</b></div>' +
        '<div class="set-row"><span class="set-lbl">Synced Users</span><b>' + esc(String(node.synced_user_count != null ? node.synced_user_count : node.remote_users || 0)) + '</b></div>' +
        '<div class="set-row"><span class="set-lbl">Last Error</span><span style="font-size:11px;color:' + (lastErr === '—' ? 'var(--txt2)' : 'var(--red)') + ';word-break:break-word">' + esc(lastErr) + '</span></div>' +
        '<div style="display:flex;gap:8px;margin-top:14px;flex-wrap:wrap">' +
          '<button class="btn btn-g btn-sm" onclick="syncNode(\'' + esc(node.node_id) + '\')">Refresh / Test</button>' +
          '<button class="btn btn-g btn-sm" onclick="editNode(\'' + esc(node.node_id) + '\')">' + esc(t('edit')) + '</button>' +
          '<button class="btn btn-g btn-sm" onclick="deleteNode(\'' + esc(node.node_id) + '\')">' + esc(t('delete')) + '</button>' +
        '</div>' +
      '</div></div>';
  });
  wrap.innerHTML = html;
}

function openNodeModal(nodeId) {
  EDITING_NODE_ID = nodeId || null;
  var node = nodeId ? ALL_NODES.find(function(n) { return n.node_id === nodeId; }) : null;
  var title = node ? t('editNode') : t('addNode');
  var modal = $id('mod-node');
  if (!modal) {
    modal = document.createElement('div');
    modal.className = 'modal';
    modal.id = 'mod-node';
    var inner = document.createElement('div');
    inner.className = 'modal-c';
    inner.style.maxWidth = '420px';
    modal.appendChild(inner);
    document.body.appendChild(modal);
  }
  var form = modal.querySelector('form');
  if (!form) {
    form = document.createElement('form');
    form.onsubmit = submitNode;
    modal.querySelector('.modal-c').appendChild(form);
    var header = document.createElement('div');
    header.className = 'modal-t';
    header.innerHTML = '<span class="glow"></span><span id="node-modal-title"></span>';
    form.appendChild(header);
    var fields = [
      {id:'node-name', label:'Name', type:'text', placeholder:'Germany'},
      {id:'node-domain', label:'Domain', type:'text', placeholder:'example.com'},
      {id:'node-apikey', label:'API Key', type:'password', placeholder:'spdr_xxxxxxxxx'}
    ];
    fields.forEach(function(f){
      var g = document.createElement('div'); g.className='f-g';
      var lab = document.createElement('label'); lab.textContent=f.label; g.appendChild(lab);
      var inp = document.createElement('input'); inp.type=f.type; inp.id=f.id; inp.dir='ltr'; inp.placeholder=f.placeholder; inp.autocomplete='off';
      if (f.id !== 'node-name') inp.required = (f.id === 'node-domain' || !EDITING_NODE_ID);
      g.appendChild(inp); form.appendChild(g);
    });
    var hint = document.createElement('div');
    hint.className='px-hint';
    hint.textContent='در حالت ویرایش، API Key را خالی بگذارید تا همان کلید قبلی حفظ شود.';
    form.appendChild(hint);
    var btns=document.createElement('div'); btns.className='modal-btns';
    var bc=document.createElement('button'); bc.type='button'; bc.className='btn btn-g'; bc.textContent=t('cancel'); bc.onclick=function(){ closeModal('mod-node'); };
    var bs=document.createElement('button'); bs.type='submit'; bs.className='btn btn-g'; bs.textContent=t('saveChanges'); bs.id='node-submit';
    btns.appendChild(bc); btns.appendChild(bs); form.appendChild(btns);
  }
  var titleEl=$id('node-modal-title'); if(titleEl) titleEl.textContent=title;
  $id('node-name').value=node ? (node.name || '') : '';
  $id('node-domain').value=node ? (node.domain || '').replace(/^https?:\/\//i,'') : '';
  $id('node-apikey').value='';
  $id('node-apikey').required=!EDITING_NODE_ID;
  $id('node-domain').required=true;
  $id('node-apikey').placeholder=node && node.api_key_masked ? node.api_key_masked : 'spdr_xxxxxxxxx';
  modal.classList.add('show');
}

async function submitNode(e) {
  e.preventDefault();
  var name=$id('node-name').value.trim();
  var domain=$id('node-domain').value.trim();
  var apiKey=$id('node-apikey').value.trim();
  if(!domain || (!EDITING_NODE_ID && !apiKey)){ toast('Domain و API Key الزامی است','err'); return; }
  var btn=e.target.querySelector('button[type="submit"]'); if(btn){btn.disabled=true;btn.textContent='...';}
  try{
    var method=EDITING_NODE_ID?'PATCH':'POST';
    var url=EDITING_NODE_ID?'/api/nodes/'+encodeURIComponent(EDITING_NODE_ID):'/api/nodes';
    var body={name:name,domain:domain}; if(apiKey) body.api_key=apiKey;
    var result=await api(url,{method:method,body:body});
    if(!result || !result.ok) throw new Error('پاسخ نامعتبر از سرور');
    toast(EDITING_NODE_ID?t('nodeUpdated'):t('nodeAdded'),'ok'); closeModal('mod-node'); await loadNodes();
    loadInbounds();
  }catch(e){toast(e.message || t('nodeSyncFailed'),'err');}
  if(btn){btn.disabled=false;btn.textContent=t('saveChanges');}
}

function editNode(nodeId){ openNodeModal(nodeId); }

async function deleteNode(nodeId){
  if(!confirm(t('confirmDeleteNode'))) return;
  try{
    var r=await api('/api/nodes/'+encodeURIComponent(nodeId),{method:'DELETE'});
    if(r && r.pending_cleanup){ toast('حذف نود به cleanup روی Node مقصد وابسته است؛ وضعیت pending ذخیره شد.','warn'); }
    else { toast(t('nodeDeleted'),'ok'); }
    await loadNodes(); await loadInbounds();
  }catch(e){toast(e.message || t('toast.err'),'err');}
}

async function syncNode(nodeId){
  try{
    var r=await api('/api/nodes/'+encodeURIComponent(nodeId)+'/refresh',{method:'POST'});
    if(r && r.node && r.node.status==='online') toast(t('nodeSynced'),'ok');
    else toast((r && r.node && r.node.last_error) || t('nodeSyncFailed'),'warn');
    await loadNodes();
  }catch(e){toast(e.message || t('nodeSyncFailed'),'err');}
}

async function syncAllNodes(){
  try{var result=await api('/api/nodes/sync-all',{method:'POST'});toast(t('nodeSynced')+' ('+(result.synced || 0)+')','ok');await loadNodes();}
  catch(e){toast(e.message || t('nodeSyncFailed'),'err');}
}


// ─── TELEGRAM BOT ───────────────────────────────────────────────────────────
function toggleBotAcc(id){
  var el=$id(id);
  if(!el) return;
  var open=el.classList.toggle('open');
  var btn=el.querySelector('.bot-acc-head');
  if(btn) btn.setAttribute('aria-expanded', open ? 'true' : 'false');
}
function switchBotSubtab(which) {
  ['channel','sell'].forEach(function(k){
    var pane=$id('bot-sub-'+k), btn=$id('bot-tab-'+k+'-btn');
    if(pane) pane.classList.toggle('active', k===which);
    if(btn) btn.classList.toggle('active', k===which);
  });
  if(which==='sell'){ loadBotInbounds(); loadSellPlans(); }
}
async function loadBotInbounds() {
  try{
    var d=await api('/api/inbounds'); var items=Array.isArray(d)?d:(d.inbounds||[]);
    ['bot-channel-inbound','bot-plan-inbound'].forEach(function(id){
      var el=$id(id); if(!el) return;
      var first = id==='bot-channel-inbound' ? '<option value="">Default TLS/WS</option>' : '<option value="">انتخاب Inbound</option>';
      // The API field is inbound_id (not id) — reading x.id produced options with
      // an empty value, so every inbound the admin picked was silently dropped.
      el.innerHTML = first + items.filter(function(x){return x.inbound_id!=='Node' && !(x.system)}).map(function(x){ return '<option value="'+escAttr(String(x.inbound_id||''))+'">'+esc(x.name||x.inbound_id||'Inbound')+'</option>'; }).join('');
    });
  }catch(e){console.error(e);}
}
// Select an inbound, re-adding a lost option so a configured inbound_id is
// never silently reset to "" when the list reloads.
function selBotInbound(el, value) {
  if (!el) return;
  value = String(value || '');
  var found = false;
  for (var i = 0; i < el.options.length; i++) { if (el.options[i].value === value) { found = true; break; } }
  if (!found && value) el.insertAdjacentHTML('beforeend', '<option value="' + escAttr(value) + '">' + esc(value) + '</option>');
  el.value = value;
}
function fillBotStatus(d){
  var b=$id('bot-status-badge'); if(b){b.textContent=d.token_configured?'ON':'OFF';b.className='badge '+(d.token_configured?'badge-g':'badge-b');}
  var info=$id('bot-account-info'); if(info&&d.token_masked){ info.innerHTML='<div class="bot-status-card"><span>API Key</span><b dir="ltr">'+esc(d.token_masked)+'</b></div>'; }
  var ch=d.channel||{}, sell=d.sell||{};
  var cs=$id('bot-channel-state'); if(cs){cs.textContent=ch.enabled?'ON':'OFF';cs.className='badge '+(ch.enabled?'badge-g':'badge-b');}
  var ss=$id('bot-sell-state'); if(ss){ss.textContent=sell.enabled?'ON':'OFF';ss.className='badge '+(sell.enabled?'badge-g':'badge-b');}
  var cn=$id('bot-channel-next'); if(cn) cn.innerHTML=
    '<div class="bot-status-card"><span>موفق</span><b>'+esc(String(ch.success_count||0))+'</b></div>'+
    '<div class="bot-status-card"><span>خطا</span><b>'+esc(String(ch.error_count||0))+'</b></div>'+
    '<div class="bot-status-card"><span>آخرین اجرا</span><b dir="ltr" style="font-size:11px">'+esc(ch.last_run_at||'—')+'</b></div>'+
    '<div class="bot-status-card"><span>اجرای بعدی</span><b dir="ltr" style="font-size:11px">'+esc(ch.next_run_at||'—')+'</b></div>'+
    '<div class="bot-status-card"><span>حذف معلق</span><b>'+esc(String((ch.pending_delete_user_ids||[]).length))+'</b></div>'+
    '<div class="bot-status-card"><span>خطای حذف</span><b style="font-size:11px;color:var(--red)">'+esc(ch.last_delete_error||'—')+'</b></div>';
  var so=$id('bot-sell-orders'); if(so) so.innerHTML=
    '<div class="bot-status-card"><span>در انتظار رسید</span><b>'+esc(String((d.orders||{}).pending||0))+'</b></div>'+
    '<div class="bot-status-card"><span>رسیدهای ارسالی</span><b>'+esc(String((d.orders||{}).payment_submitted||0))+'</b></div>'+
    '<div class="bot-status-card"><span>Approved</span><b>'+esc(String((d.orders||{}).approved||0))+'</b></div>'+
    '<div class="bot-status-card"><span>Rejected</span><b>'+esc(String((d.orders||{}).rejected||0))+'</b></div>'+
    '<div class="bot-status-card"><span>Cancelled</span><b>'+esc(String((d.orders||{}).cancelled||0))+'</b></div>';
  renderSellPlans(d.plans || sell.plans || []);
  renderBotJoinRows(sell.required_channels||[]);
}
function renderSellPlans(plans){
  var wrap=$id('bot-sell-plans-list'); if(!wrap) return;
  if(!plans.length){ wrap.innerHTML='<div class="bot-note" style="margin-top:12px">هیچ Plan فعالی/ذخیره‌شده‌ای وجود ندارد.</div>'; return; }
  var html='';
  plans.forEach(function(p){
    var traffic=Number(p.traffic_gb||0)<=0?'نامحدود':(Number(p.traffic_gb)||0)+' GB';
    var days=Number(p.expire_days||0)<=0?'نامحدود':(Number(p.expire_days)||0)+' روز';
    html+='<div class="bot-plan-card">'+
      '<div class="bot-plan-title"><b>'+esc(p.name||'Plan')+'</b><span class="badge '+(p.enabled===false?'badge-b':'badge-g')+'">'+(p.enabled===false?'OFF':'ON')+'</span></div>'+
      '<div class="bot-plan-meta">💰 '+esc(p.price||'—')+' · 💾 '+esc(traffic)+' · ⏳ '+esc(days)+' · 🌐 '+esc(p.inbound_id||'Default')+' · 👤 '+esc(p.username_prefix||'shop')+'</div>'+
      '<div class="bot-plan-actions">'+
      '<button class="btn-glass" onclick="editSellPlan(\''+escAttr(String(p.id||''))+'\')">ویرایش</button>'+
      '<button class="btn-glass" onclick="toggleSellPlan(\''+escAttr(String(p.id||''))+'\','+(p.enabled===false?'true':'false')+')">'+(p.enabled===false?'فعال کردن':'غیرفعال کردن')+'</button>'+
      '<button class="btn-glass" onclick="deleteSellPlan(\''+escAttr(String(p.id||''))+'\')">حذف</button>'+
      '</div></div>';
  });
  wrap.innerHTML=html;
}
function resetSellPlanForm(){
  if($id('bot-plan-id')) $id('bot-plan-id').value='';
  if($id('bot-plan-name')) $id('bot-plan-name').value='';
  if($id('bot-plan-price')) $id('bot-plan-price').value='';
  if($id('bot-plan-traffic')) $id('bot-plan-traffic').value='50';
  if($id('bot-plan-days')) $id('bot-plan-days').value='30';
  if($id('bot-plan-prefix')) $id('bot-plan-prefix').value='shop';
  if($id('bot-plan-inbound')) $id('bot-plan-inbound').value='';
  if($id('bot-plan-save-label')) $id('bot-plan-save-label').textContent='➕ افزودن Plan';
}
async function loadSellPlans(){try{var d=await api('/api/bot/sell/plans');renderSellPlans(d.plans||[]);}catch(e){console.error(e);}}
function sellPlanPayload(){return{ name:($id('bot-plan-name').value||'').trim(), price:($id('bot-plan-price').value||'').trim(), inbound_id:$id('bot-plan-inbound').value||'', traffic_gb:parseFloat($id('bot-plan-traffic').value||0), expire_days:parseInt($id('bot-plan-days').value||0,10), username_prefix:($id('bot-plan-prefix').value||'shop').trim(), enabled:true};}
async function saveSellPlan(){try{var payload=sellPlanPayload();if(!payload.name||!payload.price||!payload.inbound_id)throw new Error('نام Plan، قیمت و Inbound الزامی است');var id=($id('bot-plan-id').value||'').trim();if(id){await api('/api/bot/sell/plans/'+encodeURIComponent(id),{method:'PATCH',body:payload});toast('Plan بروزرسانی شد','ok');}else{await api('/api/bot/sell/plans',{method:'POST',body:payload});toast('Plan اضافه شد','ok');}resetSellPlanForm();await loadSellPlans();await loadBot();}catch(e){toast(e.message||'خطا در ذخیره Plan','err');}}
async function editSellPlan(id){try{var d=await api('/api/bot/sell/plans');var p=(d.plans||[]).find(function(x){return String(x.id)===String(id);});if(!p)throw new Error('Plan پیدا نشد');$id('bot-plan-id').value=p.id||'';$id('bot-plan-name').value=p.name||'';$id('bot-plan-price').value=p.price||'';selBotInbound($id('bot-plan-inbound'),p.inbound_id||'');$id('bot-plan-traffic').value=p.traffic_gb||0;$id('bot-plan-days').value=p.expire_days||0;$id('bot-plan-prefix').value=p.username_prefix||'shop';$id('bot-plan-save-label').textContent='💾 ذخیره تغییرات';toggleBotAcc('botacc-sell-products');window.scrollTo({top:$id('tab-bot').offsetTop,behavior:'smooth'});}catch(e){toast(e.message||'خطا','err');}}
async function toggleSellPlan(id,enabled){try{var d=await api('/api/bot/sell/plans');var p=(d.plans||[]).find(function(x){return String(x.id)===String(id);});if(!p)throw new Error('Plan پیدا نشد');p.enabled=!!enabled;await api('/api/bot/sell/plans/'+encodeURIComponent(id),{method:'PATCH',body:p});await loadSellPlans();await loadBot();toast(enabled?'Plan فعال شد':'Plan غیرفعال شد','ok');}catch(e){toast(e.message||'خطا','err');}}
async function deleteSellPlan(id){if(!confirm('این Plan حذف شود؟'))return;try{await api('/api/bot/sell/plans/'+encodeURIComponent(id),{method:'DELETE'});resetSellPlanForm();await loadSellPlans();await loadBot();toast('Plan حذف شد','ok');}catch(e){toast(e.message||'خطا','err');}}
function addBotJoinRow(item){var wrap=$id('bot-join-list');if(!wrap)return;item=item||{};var row=document.createElement('div');row.className='bot-join-row';row.innerHTML='<div class="f-g" style="margin:0"><label>Channel / Chat ID</label><input class="bot-join-chat" type="text" dir="ltr" placeholder="@channel یا -100123456789" value="'+escAttr(item.chat_id||item.channel||'')+'"></div><div class="f-g" style="margin:0"><label>Join Link</label><input class="bot-join-url" type="url" dir="ltr" placeholder="https://t.me/channel" value="'+escAttr(item.join_url||item.url||'')+'"></div><button class="btn-glass" type="button" onclick="this.parentElement.remove()">حذف</button>';wrap.appendChild(row);}
function renderBotJoinRows(items){var wrap=$id('bot-join-list');if(!wrap)return;wrap.innerHTML='';(items||[]).forEach(addBotJoinRow);if(!items||!items.length)addBotJoinRow();}
function getBotJoinPayload(){return Array.prototype.slice.call(document.querySelectorAll('#bot-join-list .bot-join-row')).map(function(row){return{chat_id:(row.querySelector('.bot-join-chat').value||'').trim(),join_url:(row.querySelector('.bot-join-url').value||'').trim()};}).filter(function(x){return x.chat_id;});}
async function loadBot(){try{var d=await api('/api/bot/status');fillBotStatus(d);if($id('bot-token')&&!d.token_configured)$id('bot-token').value='';var ch=d.channel||{},sell=d.sell||{};$id('bot-channel-enabled').checked=!!ch.enabled;$id('bot-channel').value=ch.channel||'';$id('bot-channel-interval').value=ch.interval_minutes||60;$id('bot-channel-prefix').value=ch.username_prefix||'white';$id('bot-channel-traffic').value=ch.traffic_limit_gb||0;$id('bot-channel-days').value=ch.expire_days||30;$id('bot-channel-replace').checked=ch.replace_previous!==false;await loadBotInbounds();selBotInbound($id('bot-channel-inbound'),ch.inbound_id||'');$id('bot-sell-enabled').checked=!!sell.enabled;$id('bot-sell-admin').value=sell.admin_chat_id||'';$id('bot-sell-support').value=sell.support_username||'';$id('bot-sell-support-text').value=sell.support_text||'💬 برای پشتیبانی با مدیر فروش تماس بگیرید.';$id('bot-sell-payment').value=sell.payment_url||'';$id('bot-sell-payment-details').value=sell.payment_details||'';renderBotJoinRows(sell.required_channels||[]);}catch(e){console.error(e);}}
function botPayload(){return{token:($id('bot-token').value||'').trim(),channel:{enabled:$id('bot-channel-enabled').checked,channel:($id('bot-channel').value||'').trim(),interval_minutes:parseInt($id('bot-channel-interval').value||60,10),username_prefix:($id('bot-channel-prefix').value||'white').trim(),traffic_limit_gb:parseFloat($id('bot-channel-traffic').value||0),expire_days:parseInt($id('bot-channel-days').value||0,10),inbound_id:$id('bot-channel-inbound').value||'',replace_previous:$id('bot-channel-replace').checked},sell:{enabled:$id('bot-sell-enabled').checked,admin_chat_id:($id('bot-sell-admin').value||'').trim(),support_username:($id('bot-sell-support').value||'').trim(),support_text:$id('bot-sell-support-text').value||'',payment_url:($id('bot-sell-payment').value||'').trim(),payment_details:$id('bot-sell-payment-details').value||'',required_channels:getBotJoinPayload(),welcome_text:'سلام 👋\nبرای استفاده از فروشگاه یکی از بخش‌های زیر را انتخاب کنید.'}};}
async function saveBotConfig(){try{var p=botPayload();if(!p.token){var st=await api('/api/bot/status');if(!st.token_configured)throw new Error('Telegram Bot API Key را وارد کنید');}await api('/api/bot/config',{method:'POST',body:p});await loadBot();toast('تنظیمات Bot ذخیره شد','ok');}catch(e){toast(e.message||'خطا در ذخیره تنظیمات','err');}}
async function testBotConfig(){try{var p=botPayload();var d=await api('/api/bot/test',{method:'POST',body:{token:p.token,channel:p.channel.channel}});var msg='Bot: @'+(d.bot&&d.bot.username||'?');if(d.channel){msg+=' · Channel: '+(d.channel.username||d.channel.title||d.channel.id);if(!d.channel.is_admin)msg+=' · Bot admin نیست';}toast(msg,d.channel&&!d.channel.is_admin?'warn':'ok');}catch(e){toast(e.message||'Telegram test failed','err');}}
async function runChannelBotNow(){try{await api('/api/bot/config',{method:'POST',body:botPayload()});var d=await api('/api/bot/channel/run',{method:'POST'});toast('کاربر '+((d.user||{}).username||'')+' ساخته و در Channel ارسال شد','ok');await loadBot();}catch(e){toast(e.message||'خطا در اجرای Channel Bot','err');}}

// ─── SERVER INFO IN SETTINGS ───────────────────────────────────────────────
async function loadServerInfo() {
  try {
    var info = await api('/api/server-info');
    var wrap = $id('server-info-wrap');
    if (!wrap) return;
    var apiKey = ''; try { var kr = await api('/api/panel-api-key'); apiKey = kr.api_key || ''; } catch(e) { apiKey = ''; }
    var maskedKey = apiKey.length > 8 ? apiKey.substring(0, 8) + '••••••••' : (apiKey || '--');
    window._fullApiKey = apiKey;
    var flagHtml = (info.country_flag || '🌐') + ' ' + (info.country || info.country_code || '');
    var html = '<div class="set-row"><span class="set-lbl">' + t('serverIp') + '</span><span dir="ltr" style="font-size:12px">' + (info.public_ip || '--') + ' ' + flagHtml + '</span></div>' +
      '<div class="set-row"><span class="set-lbl">' + t('apiKey') + '</span><span dir="ltr" style="font-size:11px;font-family:monospace" id="api-key-display">' + maskedKey + '</span>' +
      (apiKey ? '<button type="button" class="btn-glass" onclick="copyApiKey()">' + t('copyApiKey') + '</button>' : '') +
      '<button type="button" class="btn-glass" onclick="generateApiKey()">تولید خودکار</button>' +
      '</div>';
    wrap.innerHTML = html;
  } catch(e) { console.error('loadServerInfo:', e); }
}

function copyApiKey() {
  navigator.clipboard.writeText(window._fullApiKey || '').then(function() {
    toast('کپی شد', 'ok');
  });
}

async function generateApiKey() {
  try {
    var r = await api('/api/me/generate-key', {method: 'POST'});
    if (r.ok && r.api_key) {
      window._fullApiKey = r.api_key;
      loadServerInfo();
      toast('API Key جدید تولید شد', 'ok');
    }
  } catch(e) { toast(e.message || 'خطا', 'err'); }
}

// ─── TUNNEL MANAGEMENT (TM) ──────────────────────────────────────────────────
// Everything below talks to /api/tm/* over the panel's normal session cookie.
// No API key ever reaches this file: the backend keeps it server-side and only
// returns masked metadata, so there is nothing secret to leak from here.
//
// Two latency numbers live side by side and are never mixed:
//   * `pings[id].clientMs`  – measured BY THIS BROWSER against the server's own
//                             reachable endpoint (several samples, median).
//   * `srv.server_latency_ms` – measured BY THE PANEL BACKEND (health probe).
// A client measurement is never replaced by a backend one, and a failure is
// always reported as "Offline / Timeout" instead of a guessed number.

var TM = {
  servers: [],
  edges: [],
  activeServer: '',
  status: null,
  appliedEdges: [],
  appliedChain: [],
  pings: {},          // node_id -> {clientMs: number|null, testing: bool, at: number}
  chainPing: null,
  edgePings: {},      // edge id -> ms, measured server-to-server by the backend
  edgePingAt: {},     // edge id -> when it was measured (the chip hides after a while)
  edgePingBusy: '',
  chipTimer: null,    // auto-hide a ping result some time after it arrived
  progressTimer: null,
  nodeTimer: null,    // auto-collapse the expanded node popover
  bar: null,          // a bar being dragged: {edgeId, pointerId, mode, …}
  snapHl: '',         // node currently lit up as the magnet target
  section: 'server',
  zoom: 1,
  panX: 0,
  panY: 0,
  pendingSrc: '',     // a box picked as the source (its key behind the box)
  pendingDst: '',     // a box picked as the destination (its key on the face)
  selectedEdge: '',
  openNode: '',       // node whose popover is expanded
  sizes: {},          // node_id -> {w, h} measured after render
  drag: null,
  pan: null,
  press: null,        // pending long-press before a node drag starts
  pointers: {},       // pointerId -> {x, y} for pinch zoom
  pinch: null,
  saveTimer: null,
  edgeFrame: null,
  defaultEdgeStyle: 'curved'
};

function $t(id) { return document.getElementById(id); }

function tmSleep(ms) { return new Promise(function (r) { setTimeout(r, ms); }); }
function tmServerById(id) { return TM.servers.filter(function (s) { return s.node_id === id; })[0] || null; }
function tmEdgeById(id) { return TM.edges.filter(function (e) { return e.id === id; })[0] || null; }
function tmEdgeFor(from, to) {
  return TM.edges.filter(function (e) { return e.from === from && e.to === to; })[0] || null;
}
function tmNodeName(id) { var s = tmServerById(id); return s ? (s.country_flag + ' ' + s.name) : id; }

// The role of a node is always DERIVED from the graph, never stored per node.
function tmRoleOf(nodeId) {
  var roles = (TM.status && TM.status.roles) || {};
  var role = roles[nodeId];
  if (role) return role;
  var chain = TM.appliedChain || [];
  var idx = chain.indexOf(nodeId);
  if (idx < 0) return null;
  var label = chain.length === 1 ? t('tm.role.initial_final')
    : idx === 0 ? t('tm.role.initial')
    : idx === chain.length - 1 ? t('tm.role.final')
    : t('tm.role.middle');
  return { node_id: nodeId, role_label: label, active: true };
}
function tmRoleLabel(nodeId) {
  var r = tmRoleOf(nodeId);
  if (!r) return '';
  // The backend's role_label is English; the UI's own key follows the locale.
  var label = t('tm.role.' + (r.role || 'idle')) || r.role_label || '';
  // A fan-out hop is both a role position and the balancer for what follows.
  if (r.balancer) label += (label ? ' · ' : '') + t('tm.balancer');
  return label;
}

// ─── Data load ──────────────────────────────────────────────────────────────
async function loadTM() {
  try {
    var data = await api('/api/tm');
    TM.servers = (data.graph && data.graph.servers) || [];
    TM.edges = (data.graph && data.graph.edges) || [];
    TM.activeServer = (data.graph && data.graph.active_server) || '';
    TM.status = data.status || null;
    TM.appliedEdges = ((data.status && data.status.applied_edges) || []);
    TM.appliedChain = ((data.status && data.status.applied_chain) || []);
    tmDropMissingRefs();
    tmRenderServers();
    tmRenderStatus();
    tmRenderGraph();
  } catch (e) {
    var el = $t('tm-server-list');
    if (el) el.innerHTML = '<div style="color:var(--red);padding:16px">' + esc(e.message || t('tm.loadErr')) + '</div>';
  }
}

// A deleted server must not survive anywhere in the rendered graph.
function tmDropMissingRefs() {
  var known = {};
  TM.servers.forEach(function (s) { known[s.node_id] = true; });
  TM.edges = TM.edges.filter(function (e) { return known[e.from] && known[e.to]; });
  if (TM.activeServer && !known[TM.activeServer]) TM.activeServer = '';
  if (TM.pendingSrc && !known[TM.pendingSrc]) TM.pendingSrc = '';
  if (TM.pendingDst && !known[TM.pendingDst]) TM.pendingDst = '';
  TM.servers.forEach(function (s) {
    if (s.x == null || s.y == null || (s.x === 0 && s.y === 0)) tmPlaceServer(s);
  });
}

// Deterministic first-placement for a server that has never been dragged.
function tmPlaceServer(srv) {
  var index = TM.servers.indexOf(srv);
  var cols = Math.max(1, Math.min(4, TM.servers.length));
  srv.x = 40 + (index % cols) * 240;
  srv.y = 30 + Math.floor(index / cols) * 150;
  tmPersistPositionsSoon();
}

async function tmPersistPositionsSoon() {
  clearTimeout(TM.saveTimer);
  TM.saveTimer = setTimeout(function () {
    var positions = {};
    TM.servers.forEach(function (s) { positions[s.node_id] = { x: Math.round(s.x || 0), y: Math.round(s.y || 0) }; });
    api('/api/tm/graph', { method: 'PUT', body: { positions: positions } }).catch(function () {});
  }, 700);
}

async function tmSaveGraph(extra) {
  var body = { positions: {}, edges: TM.edges };
  TM.servers.forEach(function (s) { body.positions[s.node_id] = { x: Math.round(s.x || 0), y: Math.round(s.y || 0) }; });
  if (extra && extra.active_server !== undefined) body.active_server = extra.active_server;
  try {
    var r = await api('/api/tm/graph', { method: 'PUT', body: body });
    if (r && r.status) { TM.status = r.status; tmRenderStatus(); }
    return r;
  } catch (e) { toast(e.message || t('toast.err'), 'err'); return null; }
}

// ─── Sections ───────────────────────────────────────────────────────────────
function tmSetSection(section) {
  TM.section = section;
  ['server', 'configure'].forEach(function (key) {
    var btn = document.querySelector('[data-tm-section="' + key + '"]');
    if (btn) { btn.classList.toggle('active', key === section); btn.setAttribute('aria-selected', String(key === section)); }
    var pane = $t('tm-section-' + key);
    if (pane) pane.classList.toggle('hidden', key !== section);
  });
  if (section === 'configure') {
    // The canvas has no layout while hidden, so render after it is on screen.
    requestAnimationFrame(function () {
      tmRenderGraph();
      tmFit();
      tmPingAllMissing();     // one pass, on demand only — never a polling loop
      tmGenerate();
    });
  }
}

function tmRenderStatus() {
  var st = TM.status || {};
  var stateEl = $t('tm-sb-state');
  if (stateEl) {
    var label = !TM.edges.length ? t('tm.state.idle')
      : (st.active ? (st.dirty ? t('tm.state.dirty') : t('tm.state.active')) : t('tm.state.idle'));
    stateEl.textContent = label;
    stateEl.className = 'tm-sb-state ' + (st.active && !st.dirty ? 'ok' : st.last_error ? 'err' : 'idle');
  }
  var put = function (id, val) { var e = $t(id); if (e) e.textContent = val; };
  put('tm-sb-servers', String(TM.servers.length));
  put('tm-sb-links', String(TM.edges.length));
  put('tm-sb-entry', st.entry || '—');
  put('tm-sb-exit', st.exit || '—');
  put('tm-sb-apply', st.last_apply ? new Date(st.last_apply).toLocaleString() : '—');
  // The balancer badge only appears when a hop really does fan out.
  var balItem = $t('tm-sb-bal');
  if (balItem) {
    balItem.classList.toggle('hidden', !st.balancer);
    put('tm-sb-chain', (st.levels || []).map(function (n) { return n.length; }).join('×'));
  }
}

// ─── TM → Server ────────────────────────────────────────────────────────────
function tmRenderServers() {
  var wrap = $t('tm-server-list');
  if (!wrap) return;
  if (!TM.servers.length) {
    wrap.innerHTML = '<div class="empty" style="padding:18px;text-align:center">' + esc(t('tm.noServers')) + '</div>';
    return;
  }
  var html = '';
  TM.servers.forEach(function (srv) {
    var online = srv.status === 'online';
    var checking = srv.status === 'checking' || srv.status === 'unauthorized';
    var stCls = online ? 'node-online' : checking ? 'node-checking' : 'node-offline';
    var stText = online ? t('connected') : checking ? t('connecting') : t('disconnected');
    var roleLabel = tmRoleLabel(srv.node_id);
    var isSet = TM.activeServer === srv.node_id;
    var accId = 'tm-srv-' + srv.node_id;
    var inGraph = TM.edges.some(function (e) { return e.from === srv.node_id || e.to === srv.node_id; });
    var clientPing = tmClientPingText(srv.node_id);
    var clientOff = tmClientPingDown(srv.node_id);
    html += '<div class="spd-acc tm-server-acc">' +
      '<button type="button" class="spd-acc-head" onclick="tmToggleServerAcc(this)" aria-expanded="false" aria-controls="acc-' + escAttr(accId) + '">' +
        '<span class="spd-acc-title"><strong>' + esc(srv.country_flag + ' ' + srv.name) + (isSet ? ' ★' : '') + '</strong>' +
          '<span style="font-size:10px;color:var(--txt2);display:block;margin-top:3px" dir="ltr">' + esc(srv.domain || '—') + '</span>' +
          (roleLabel ? '<span class="tm-server-head-role">' + esc(roleLabel) + '</span>' : '') +
        '</span>' +
        // The client ping lives right here in the header — measured from this
        // device — while the backend's own health number stays in the body.
        '<span class="tm-srv-ping" dir="ltr" id="tm-srv-cping-' + escAttr(srv.node_id) + '">' + esc(clientPing) + '</span>' +
        '<span class="node-status-pill ' + stCls + '">' + esc(stText) + '</span>' +
        '<span class="material-symbols-outlined spd-acc-arrow">expand_more</span>' +
      '</button>' +
      '<div class="spd-acc-body" id="acc-' + escAttr(accId) + '" aria-hidden="true">' +
        '<div class="set-row"><span class="set-lbl">' + esc(t('nodeDomain')) + '</span><span dir="ltr" style="font-size:12px">' + esc(srv.domain || '—') + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">' + esc(t('nodeStatusLabel')) + '</span><span class="node-status-pill ' + stCls + '">' + esc(stText) + '</span></div>' +
        // The accordion's ping IS the client ping — the same number this
        // browser measured, shown in the header and on the graph box. The
        // server-to-server number lives on the connection bars, where it
        // describes a link rather than the viewer's own route.
        '<div class="set-row"><span class="set-lbl">' + esc(t('tm.clientPing')) + ' <span class="tm-lat-tag tm-lat-tag-client">' + esc(t('tm.fromThisDevice')) + '</span></span>' +
          '<span class="tm-srv-ping' + (clientOff ? ' off' : '') + '" dir="ltr" id="tm-srv-cping-body-' + escAttr(srv.node_id) + '">' + esc(clientPing) + '</span></div>' +
        '<div class="set-row"><span class="set-lbl">API Key</span><span dir="ltr" style="font-size:11px;font-family:monospace">' + esc(srv.api_key_masked || '—') + '</span></div>' +
        (roleLabel ? '<div class="set-row"><span class="set-lbl">' + esc(t('tm.roleLabel')) + '</span><b>' + esc(roleLabel) + '</b></div>' : '') +
        '<div class="set-row"><span class="set-lbl">TM</span><b>' + esc(inGraph ? (roleLabel || '—') : t('tm.role.idle')) + '</b></div>' +
        (isSet ? '<div class="tm-set-badge"><span class="material-symbols-outlined" style="font-size:13px">check_circle</span>Set</div>' : '') +
        '<div class="tm-server-actions">' +
          '<button type="button" class="btn btn-g btn-sm" onclick="tmOpenServerModal(\'' + escAttr(srv.node_id) + '\')">' + esc(t('tm.editServer')) + '</button>' +
          '<button type="button" class="btn btn-g btn-sm" onclick="tmPingThis(\'' + escAttr(srv.node_id) + '\')">' + esc(t('proxy.pingBtn')) + '</button>' +
          '<button type="button" class="btn btn-g btn-sm" onclick="tmDeleteServer(\'' + escAttr(srv.node_id) + '\')">' + esc(t('delete')) + '</button>' +
        '</div>' +
      '</div></div>';
  });
  wrap.innerHTML = html;
}

// One value, rendered wherever the client-side latency appears.
function tmClientPingText(nodeId) {
  var p = TM.pings[nodeId];
  if (p && p.testing) return t('tm.measuring');
  if (p && p.clientMs != null) return Math.round(p.clientMs) + ' ms';
  if (p && p.clientMs === null) return t('tm.offline');
  return '—';
}
function tmClientPingDown(nodeId) {
  var p = TM.pings[nodeId];
  return !!(p && p.testing) || (p && p.clientMs === null);
}

// Same behaviour as the Node accordions, without depending on their id scheme.
function tmToggleServerAcc(btn) {
  var acc = btn && btn.parentElement;
  if (!acc) return;
  var open = !acc.classList.contains('open');
  acc.classList.toggle('open', open);
  btn.setAttribute('aria-expanded', String(open));
  var body = acc.querySelector('.spd-acc-body');
  if (body) body.setAttribute('aria-hidden', String(!open));
}

var TM_EDIT_SERVER = '';

function tmOpenServerModal(nodeId) {
  TM_EDIT_SERVER = nodeId || '';
  var srv = nodeId ? tmServerById(nodeId) : null;
  var tEl = $t('tm-server-modal-title');
  if (tEl) tEl.textContent = nodeId ? t('tm.editServerTitle') : t('tm.addServerTitle');
  $t('tm-srv-name').value = srv ? (srv.name || '') : '';
  $t('tm-srv-domain').value = srv ? (srv.domain || '') : '';
  var key = $t('tm-srv-key');
  key.value = '';
  key.required = !srv;
  // The stored key is never echoed back: only its masked form, as a placeholder.
  key.placeholder = (srv && srv.api_key_masked) ? srv.api_key_masked : 'spdr_xxxxxxxxx';
  $t('tm-srv-key-hint').textContent = srv
    ? t('tm.keyNeverShown').replace('{{masked}}', srv.api_key_masked || '—')
    : t('tm.keyHint');
  $t('mod-tm-server').classList.add('show');
}

async function tmSubmitServer(e) {
  e.preventDefault();
  var name = $t('tm-srv-name').value.trim();
  var domain = $t('tm-srv-domain').value.trim();
  var apiKey = $t('tm-srv-key').value.trim();
  if (!domain) { toast(t('tm.fillDomain'), 'err'); return; }
  if (!TM_EDIT_SERVER && !apiKey) { toast(t('tm.errorApiKey'), 'err'); return; }
  var btn = $t('tm-srv-submit');
  if (btn) btn.disabled = true;
  try {
    var body = { name: name, domain: domain };
    if (apiKey) body.api_key = apiKey;   // empty while editing = keep the stored key
    var r = TM_EDIT_SERVER
      ? await api('/api/tm/servers/' + encodeURIComponent(TM_EDIT_SERVER), { method: 'PATCH', body: body })
      : await api('/api/tm/servers', { method: 'POST', body: body });
    if (r && r.ok) toast(TM_EDIT_SERVER ? t('tm.serverSaved') : t('tm.serverAdded'), 'ok');
    else toast((r && r.error) || t('tm.serverSavedOffline'), 'warn');
    closeModal('mod-tm-server');
    await loadTM();
    if (typeof loadNodes === 'function') loadNodes();
  } catch (err) { toast(err.message || t('tm.remoteOffline'), 'err'); }
  if (btn) btn.disabled = false;
}

// The backend health probe. Its number is kept apart from the client ping in
// every place it is shown, and it never overwrites a client measurement.
async function tmPingServerHealth(nodeId) {
  var el = $t('tm-srv-ping-' + nodeId);
  if (el) { el.classList.add('busy'); el.textContent = t('tm.pinging'); }
  try {
    var r = await api('/api/tm/servers/' + encodeURIComponent(nodeId) + '/ping', { method: 'POST' });
    if (!r.ok) toast(r.last_error || t('tm.remoteOffline'), 'err');
  } catch (e) { toast(e.message || t('tm.remoteOffline'), 'err'); }
  await loadTM();
}

async function tmDeleteServer(nodeId) {
  if (!confirm(t('tm.confirmDelServer'))) return;
  try {
    await api('/api/tm/servers/' + encodeURIComponent(nodeId), { method: 'DELETE' });
    toast(t('tm.serverDeleted'), 'ok');
    await loadTM();
    if (typeof loadNodes === 'function') loadNodes();
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
}

// "Set" = use this server for generating the configuration. Real backend state.
async function tmSetServer(nodeId) {
  try {
    var r = await api('/api/tm/set-server', { method: 'POST', body: { node_id: nodeId } });
    TM.activeServer = (r.graph && r.graph.active_server) || '';
    TM.status = r.status || TM.status;
    toast(nodeId ? t('tm.setDone') : t('tm.setClear'), 'ok');
    tmRenderGraph();
    tmRenderServers();
    tmRenderStatus();
    tmGenerate();
    tmScheduleApply();        // the chosen entry point changes every hop
  } catch (e) { toast(e.message || t('toast.err'), 'err'); }
}

function tmClearSet() { return tmSetServer(''); }

// One button, two honest numbers: the client ping is measured by this browser
// against the server's own endpoint, the health number by the panel backend.
async function tmPingServer(nodeId) {
  TM.pings[nodeId] = { clientMs: null, testing: true, at: Date.now() };
  tmUpdatePingUI(nodeId);
  await Promise.all([tmPingNode(nodeId, 3), tmPingServerHealth(nodeId)]);
}

// Run once per tab visit for the servers that have no measurement yet.
function tmPingAllMissing() {
  TM.servers.forEach(function (srv) {
    if (!TM.pings[srv.node_id]) tmPingNode(srv.node_id, 3);
  });
}

// ─── Client-side ping (measured by THIS browser, never by the panel) ────────
function tmRemoteOrigin(domain) {
  var d = String(domain || '').trim().replace(/\/+$/, '');
  if (!d) return '';
  if (/^https?:\/\//i.test(d)) return d;
  var host = d.split('/')[0];
  if (/^(localhost|127\.0\.0\.1|\[::1\])(:\d+)?$/i.test(host)) return 'http://' + d;
  return 'https://' + d;
}

function tmMeasureOnce(url) {
  return new Promise(function (resolve) {
    var ctrl = (typeof AbortController !== 'undefined') ? new AbortController() : null;
    var done = false;
    var finish = function (v) { if (!done) { done = true; clearTimeout(timer); resolve(v); } };
    var timer = setTimeout(function () { if (ctrl) { try { ctrl.abort(); } catch (e) {} } finish(null); }, 4500);
    var started = (window.performance && performance.now) ? performance.now() : Date.now();
    var opts = { cache: 'no-store', credentials: 'omit', mode: 'cors' };
    if (ctrl) opts.signal = ctrl.signal;
    fetch(url, opts).then(function (r) {
      return r.text().then(function (txt) {
        var now = (window.performance && performance.now) ? performance.now() : Date.now();
        var ok = false;
        try { var j = JSON.parse(txt); ok = !!(j && j.service === 'white-tm'); } catch (e) { ok = false; }
        // Only a real TM panel answers with this marker — a captive portal or a
        // random host that returns 200 must never look like a successful ping.
        finish((r.ok && ok) ? (now - started) : null);
      });
    }).catch(function () { finish(null); });
  });
}

async function tmPingNode(nodeId, samples) {
  var srv = tmServerById(nodeId);
  if (!srv) return null;
  var origin = tmRemoteOrigin(srv.domain);
  if (!origin) { TM.pings[nodeId] = { clientMs: null, testing: false, at: Date.now() }; tmUpdatePingUI(nodeId); return null; }
  var url = origin + '/api/tm/peer/ping?_=' + Date.now();
  var count = samples || 3;
  var results = [];
  for (var i = 0; i < count; i++) {
    // Each sample is a fresh, cache-busted connection: no keep-alive reuse, so
    // the number describes the network path rather than a warm socket.
    var ms = await tmMeasureOnce(url + '&n=' + i);
    if (ms != null) results.push(ms);
    if (i < count - 1) await tmSleep(110);
  }
  if (!results.length) {
    TM.pings[nodeId] = { clientMs: null, testing: false, at: Date.now() };
    tmUpdatePingUI(nodeId);
    return null;
  }
  results.sort(function (a, b) { return a - b; });
  var median = results[Math.floor(results.length / 2)];
  TM.pings[nodeId] = { clientMs: median, testing: false, at: Date.now(), samples: results.length };
  tmUpdatePingUI(nodeId);
  return median;
}

// Re-render only what shows this node's client latency: the canvas box, the
// Server accordion header and body, and the swap-view row.
function tmUpdatePingUI(nodeId) {
  var txt = tmClientPingText(nodeId);
  var off = tmClientPingDown(nodeId);
  ['tm-node-lat-' + nodeId, 'tm-node-cping-' + nodeId, 'tm-srv-cping-' + nodeId].forEach(function (id) {
    var el = $t(id);
    if (el) { el.textContent = txt; el.classList.toggle('off', off && !el.id.startsWith('tm-node-lat')); }
  });
  var bodyEl = $t('tm-srv-cping-body-' + nodeId);
  if (bodyEl) {
    bodyEl.textContent = txt;
    bodyEl.classList.toggle('off', off);
  }
}

// ─── Canvas geometry ────────────────────────────────────────────────────────
function tmZoom(screenX, screenY, factor) {
  var old = TM.zoom;
  var next = Math.max(0.35, Math.min(2.4, old * factor));
  if (next === old) return;
  if (screenX != null) {
    // Keep the point under the fingers/cursor anchored while zooming.
    TM.panX = screenX - (screenX - TM.panX) * (next / old);
    TM.panY = screenY - (screenY - TM.panY) * (next / old);
  }
  TM.zoom = next;
  tmApplyTransform();
}

function tmZoomButton(factor) {
  var canvas = $t('tm-canvas');
  if (!canvas) return;
  var r = canvas.getBoundingClientRect();
  tmZoom(r.width / 2, r.height / 2, factor);
}

function tmApplyTransform() {
  var tr = 'translate(' + TM.panX + 'px,' + TM.panY + 'px) scale(' + TM.zoom + ')';
  var layer = $t('tm-node-layer');
  if (layer) { layer.style.transformOrigin = '0 0'; layer.style.transform = tr; }
  var svg = $t('tm-edges');
  if (svg) { svg.style.transformOrigin = '0 0'; svg.style.transform = tr; }
}

function tmFit() {
  var canvas = $t('tm-canvas');
  if (!canvas || !TM.servers.length) return;
  var r = canvas.getBoundingClientRect();
  if (!r.width) return;
  var minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity;
  TM.servers.forEach(function (s) {
    var size = TM.sizes[s.node_id] || { w: 180, h: 64 };
    minX = Math.min(minX, s.x); minY = Math.min(minY, s.y);
    maxX = Math.max(maxX, s.x + size.w); maxY = Math.max(maxY, s.y + size.h);
  });
  var pad = 26;
  var w = Math.max(1, maxX - minX), h = Math.max(1, maxY - minY);
  var zoom = Math.min((r.width - pad * 2) / w, (r.height - pad * 2) / h);
  TM.zoom = Math.max(0.35, Math.min(1.15, zoom));
  TM.panX = pad + (r.width - pad * 2 - w * TM.zoom) / 2 - minX * TM.zoom;
  TM.panY = pad + (r.height - pad * 2 - h * TM.zoom) / 2 - minY * TM.zoom;
  tmApplyTransform();
}

function tmCanvasPoint(clientX, clientY) {
  var canvas = $t('tm-canvas');
  var r = canvas.getBoundingClientRect();
  return {
    x: (clientX - r.left - TM.panX) / TM.zoom,
    y: (clientY - r.top - TM.panY) / TM.zoom
  };
}

// Half-extent intersection of the line leaving the box centre in direction (ux,uy).
function tmBoxExit(cx, cy, ux, uy, hw, hh) {
  var tx = Math.abs(ux) < 1e-6 ? Infinity : hw / Math.abs(ux);
  var ty = Math.abs(uy) < 1e-6 ? Infinity : hh / Math.abs(uy);
  var tt = Math.min(tx, ty);
  if (!isFinite(tt)) tt = 0;
  return { x: cx + ux * tt, y: cy + uy * tt };
}

// The line picks its own shape: nearly centred boxes get a straight bar,
// anything offset gets a curve. There is no manual shape control any more —
// the only thing that overrides the automatic choice is a dragged bend.
function tmAutoStyle(a, b) {
  var acx = a.x + a.w / 2, acy = a.y + a.h / 2;
  var bcx = b.x + b.w / 2, bcy = b.y + b.h / 2;
  var dx = Math.abs(bcx - acx), dy = Math.abs(bcy - acy);
  // Same row at roughly the same height → perfectly straight looks best.
  return dy < Math.max(a.h, b.h) * 0.42 && dx > dy ? 'straight' : 'curved';
}

// `bend` is the perpendicular offset of the control point (0 = straight). The
// geometry walks a bar that the operator dragged into a different shape, and
// falls back to its own straight/curved decision when nothing was dragged.
function tmEdgeGeometry(a, b, style, bend) {
  var aw = a.w, ah = a.h, bw = b.w, bh = b.h;
  var acx = a.x + aw / 2, acy = a.y + ah / 2;
  var bcx = b.x + bw / 2, bcy = b.y + bh / 2;
  var dx = bcx - acx, dy = bcy - acy;
  var len = Math.max(1, Math.sqrt(dx * dx + dy * dy));
  var ux = dx / len, uy = dy / len;
  var start = tmBoxExit(acx, acy, ux, uy, aw / 2, ah / 2);
  var end = tmBoxExit(bcx, bcy, -ux, -uy, bw / 2, bh / 2);
  // A small gap on both ends so the bar never touches the boxes.
  start = { x: start.x + ux * 3, y: start.y + uy * 3 };
  end = { x: end.x - ux * 3, y: end.y - uy * 3 };
  var mx = (start.x + end.x) / 2, my = (start.y + end.y) / 2;
  var measured = !!(TM.sizes[a.node_id] || TM.sizes[b.node_id]);
  var shape;
  if (bend != null) {
    // A bar the operator dragged keeps the shape they gave it.
    shape = Math.abs(bend) < 6 ? 'straight' : 'curved';
  } else {
    shape = measured ? tmAutoStyle(a, b) : 'curved';
  }
  if (shape === 'straight') {
    return { shape: 'straight', d: 'M' + start.x + ',' + start.y + ' L' + end.x + ',' + end.y, mid: { x: mx, y: my }, start: start, end: end, ux: ux, uy: uy };
  }
  var off = (bend != null) ? bend : Math.min(70, len * 0.18);
  var cx = mx - uy * off, cy = my + ux * off;
  var d = 'M' + start.x + ',' + start.y + ' Q' + cx + ',' + cy + ' ' + end.x + ',' + end.y;
  // Midpoint of a quadratic curve — where the little control chip sits.
  return {
    shape: 'curved',
    d: d,
    mid: { x: 0.25 * start.x + 0.5 * cx + 0.25 * end.x, y: 0.25 * start.y + 0.5 * cy + 0.25 * end.y },
    control: { x: cx, y: cy },
    start: start, end: end, ux: ux, uy: uy
  };
}

// ─── Graph rendering ────────────────────────────────────────────────────────
function tmRenderGraph() {
  tmRenderNodes();
  tmRenderEdges();
  tmApplyTransform();
  // The link keys are a mode of their own: the bar says which end is still
  // missing instead of hiding the state in a toolbar button.
  var modeEl = $t('tm-mode-label');
  if (modeEl) {
    if (TM.pendingSrc) {
      modeEl.textContent = t('tm.hlFrom').replace('{{src}}', tmNodeName(TM.pendingSrc));
    } else if (TM.pendingDst) {
      modeEl.textContent = t('tm.hlTo').replace('{{dst}}', tmNodeName(TM.pendingDst));
    } else {
      modeEl.textContent = '';
    }
  }
  var canvas = $t('tm-canvas');
  if (canvas) canvas.classList.toggle('connect', !!(TM.pendingSrc || TM.pendingDst));
}

function tmRenderNodes() {
  var layer = $t('tm-node-layer');
  if (!layer) return;
  if (!TM.servers.length) {
    layer.innerHTML = '<div style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;color:var(--txt2);font-size:12.5px">' + esc(t('tm.noServers')) + '</div>';
    return;
  }
  var html = '';
  TM.servers.forEach(function (srv) {
    var role = tmRoleOf(srv.node_id);
    var roleCls = role && role.role ? role.role : '';
    var online = srv.status === 'online' ? 'on' : 'off';
    var p = TM.pings[srv.node_id];
    var pingTxt = (p && p.testing) ? t('tm.measuring')
      : (p && p.clientMs != null) ? Math.round(p.clientMs) + ' ms'
      : p ? t('tm.offline')
      : (srv.server_latency_ms != null ? Math.round(srv.server_latency_ms) + ' ms' : '—');
    var pingOff = !!(p && p.clientMs == null) || (p && p.testing);
    var isSrc = TM.pendingSrc === srv.node_id;
    var isDst = TM.pendingDst === srv.node_id;
    html += '<div class="tm-node' + (TM.activeServer === srv.node_id ? ' set-node' : '') +
        (isSrc || isDst ? ' connect-src' : '') +
        (TM.openNode === srv.node_id ? ' open' : '') + '" data-node="' + escAttr(srv.node_id) + '"' +
        ' style="left:0;top:0;transform:translate(' + Math.round(srv.x) + 'px,' + Math.round(srv.y) + 'px)">' +
      // Source key: sits behind the box, sticking out of its back edge.
      '<button type="button" class="tm-link-key from' + (isSrc ? ' hot' : '') + '"' +
        ' onclick="tmPickLink(event,\'' + escAttr(srv.node_id) + '\',\'from\')"' +
        ' title="' + escAttr(t('tm.keyFrom')) + '" aria-label="' + escAttr(t('tm.keyFrom')) + '">' +
        '<span class="material-symbols-outlined">outbox</span></button>' +
      '<div class="tm-node-head">' +
        '<div style="flex:1;min-width:0">' +
          '<div class="tm-node-title">' + esc(srv.country_flag + ' ' + srv.name) + '</div>' +
          '<div class="tm-node-role ' + escAttr(roleCls) + '">' + esc(tmRoleLabel(srv.node_id) || t('tm.role.idle')) + '</div>' +
        '</div>' +
        // Destination key: sits on the face of the box.
        '<button type="button" class="tm-link-key to' + (isDst ? ' hot' : '') + '"' +
          ' onclick="tmPickLink(event,\'' + escAttr(srv.node_id) + '\',\'to\')"' +
          ' title="' + escAttr(t('tm.keyTo')) + '" aria-label="' + escAttr(t('tm.keyTo')) + '">' +
          '<span class="material-symbols-outlined">login</span></button>' +
        '<span class="tm-node-lat' + (pingOff ? ' off' : '') + '" id="tm-node-lat-' + escAttr(srv.node_id) + '" dir="ltr">' + esc(pingTxt) + '</span>' +
        '<button type="button" class="tm-node-btn" onclick="tmToggleNode(event,\'' + escAttr(srv.node_id) + '\')" aria-label="expand">' +
          '<span class="material-symbols-outlined">expand_more</span></button>' +
      '</div>' +
      '<div class="tm-node-info">' +
        '<div class="tm-server-kv"><span>' + esc(t('nodeDomain')) + '</span><b dir="ltr">' + esc(srv.domain || '—') + '</b></div>' +
        '<div class="tm-server-kv"><span>' + esc(t('tm.clientPing')) + '</span><b class="tm-srv-ping' + (pingOff ? ' off' : '') + '" id="tm-node-cping-' + escAttr(srv.node_id) + '" dir="ltr">' + esc(pingTxt) + '</b></div>' +
        '<div class="tm-server-kv"><span>' + esc(t('nodeStatusLabel')) + '</span><b class="tm-srv-ping ' + online + '">' + esc(srv.status || '—') + '</b></div>' +
        '<div class="tm-node-actions">' +
          '<button type="button" class="btn btn-g btn-sm" onclick="tmOpenServerModal(\'' + escAttr(srv.node_id) + '\')">' + esc(t('tm.editServer')) + '</button>' +
          '<button type="button" class="btn btn-g btn-sm" onclick="tmPingThis(\'' + escAttr(srv.node_id) + '\')">' + esc(t('proxy.pingBtn')) + '</button>' +
          '<button type="button" class="btn ' + (TM.activeServer === srv.node_id ? 'btn-p' : 'btn-g') + ' btn-sm" onclick="tmSetServer(\'' + escAttr(srv.node_id) + '\')">Set</button>' +
        '</div>' +
        (TM.activeServer === srv.node_id ? '<div class="tm-set-badge"><span class="material-symbols-outlined" style="font-size:13px">check_circle</span>Set</div>' : '') +
      '</div>' +
    '</div>';
  });
  layer.innerHTML = html;
  tmMeasureNodes();
}

function tmMeasureNodes() {
  TM.sizes = {};
  TM.servers.forEach(function (srv) {
    var el = document.querySelector('.tm-node[data-node="' + srv.node_id + '"]');
    if (!el) return;
    var w = el.offsetWidth, h = el.offsetHeight;
    if (w > 0 && h > 0) TM.sizes[srv.node_id] = { w: w, h: h };
  });
}

function tmNodeBox(nodeId) {
  var srv = tmServerById(nodeId);
  var size = TM.sizes[nodeId] || { w: 176, h: 62 };
  return { x: (srv ? srv.x : 0), y: (srv ? srv.y : 0), w: size.w, h: size.h, node_id: nodeId };
}

// The bend value a bar carries: how far its control point sits from the cord.
function tmEdgeBend(edge) {
  return (edge.bend == null) ? null : Number(edge.bend);
}

function tmRenderEdges() {
  var g = $t('tm-edge-layer');
  if (!g) return;
  var html = '';
  TM.edges.forEach(function (edge) {
    var a = tmNodeBox(edge.from), b = tmNodeBox(edge.to);
    var geo = tmEdgeGeometry(a, b, edge.style, tmEdgeBend(edge));
    var cls = 'tm-edge ' + (geo.shape === 'straight' ? 'straight' : 'curved');
    cls += (TM.appliedChain.indexOf(edge.from) >= 0 && TM.appliedChain.indexOf(edge.to) >= 0) ? ' active-edge' : ' idle-edge';
    if (TM.selectedEdge === edge.id) cls += ' selected';
    // A wide invisible stroke under the visible one: taps on a 2px line are
    // impossible on a phone, and the visible line stays thin.
    html += '<path class="tm-edge-hit" data-edge="' + escAttr(edge.id) + '" d="' + geo.d + '"></path>';
    html += '<path class="' + cls + '" data-edge="' + escAttr(edge.id) + '" d="' + geo.d + '"></path>';
  });
  // A re-render also drops the dashed cord of a cancelled drag (innerHTML
  // above), and tmBarMove recreates it on the next pointer move.
  g.innerHTML = html;
  tmRenderEdgeChip();
}

// Coalesce the edge redraws that follow a node or bar drag into one per frame.
function tmRenderEdgesSoon() {
  if (TM.edgeFrame) return;
  TM.edgeFrame = requestAnimationFrame(function () {
    TM.edgeFrame = null;
    tmRenderEdges();
  });
}

// Entry → exit latency: the panel asks the entry server to walk its own branch
// to the exit, so the number is the round trip the traffic really makes — never
// a browser measurement and never a single link pretending to be the chain.
var TM_CHAIN_PING_MS = 12000;
var TM_CHAIN_PING_TIMER = null;

async function tmChainPing() {
  var out = $t('tm-sb-chain-ping');
  var btn = $t('tm-sb-chain-ping-btn');
  if (out) { out.textContent = t('tm.measuring'); out.classList.add('off'); }
  if (btn) btn.classList.add('busy');
  try {
    var r = await api('/api/tm/chain-ping', { method: 'POST', body: {} });
    if (out) {
      if (r && r.total_ms != null) {
        out.textContent = Math.round(r.total_ms) + ' ms';
        out.classList.remove('off');
      } else {
        out.textContent = t('tm.offline');
        out.classList.add('off');
        var bad = (r && r.hops || []).filter(function (h) { return !h.ok && h.error; })[0];
        toast(bad ? bad.error : (r && r.error) || t('tm.offline'), 'warn');
      }
    }
    // The number is temporary, like every other measured key.
    clearTimeout(TM_CHAIN_PING_TIMER);
    TM_CHAIN_PING_TIMER = setTimeout(function () {
      var el = $t('tm-sb-chain-ping');
      if (el) { el.textContent = '—'; el.classList.add('off'); }
    }, TM_CHAIN_PING_MS);
  } catch (e) {
    if (out) { out.textContent = t('tm.offline'); out.classList.add('off'); }
    toast(e.message || t('tm.offline'), 'err');
  }
  if (btn) btn.classList.remove('busy');
}

// One shared chip per link, carrying its tiny ping SVG. Tapping the SVG runs a
// chain ping across the link — server-to-server, never a browser number.
// Dragging the chip moves the whole bar; dragging either of its round ends
// pulls that end to another box and snaps onto it.
var TM_CHIP_PING_MS = 12000;   // how long a measured number stays on a chip
var TM_NODE_AUTO_MS = 14000;    // how long an expanded node popover stays open

function tmRenderEdgeChip() {
  var layer = $t('tm-node-layer');
  if (!layer) return;
  document.querySelectorAll('.tm-edge-label, .tm-edge-handle').forEach(function (el) { el.remove(); });
  TM.edges.forEach(function (edge) {
    var a = tmNodeBox(edge.from), b = tmNodeBox(edge.to);
    var geo = tmEdgeGeometry(a, b, edge.style, tmEdgeBend(edge));
    var chip = document.createElement('div');
    chip.className = 'tm-edge-label draggable';
    chip.dataset.edge = edge.id;
    chip.style.left = geo.mid.x + 'px';
    chip.style.top = geo.mid.y + 'px';
    var age = TM.edgePingAt[edge.id] ? (Date.now() - TM.edgePingAt[edge.id]) : Infinity;
    var hidePing = TM.edgePingBusy !== edge.id && age > TM_CHIP_PING_MS;
    var res = (TM.edgePings[edge.id] != null && !hidePing)
      ? '<b dir="ltr">' + Math.round(TM.edgePings[edge.id]) + ' ms</b>'
      : '';
    chip.innerHTML =
      '<button type="button" class="tm-edge-ping' + (TM.edgePingBusy === edge.id ? ' busy' : '') + '"' +
        ' title="' + escAttr(t('tm.chainPing')) + '" onclick="tmPingEdge(event,\'' + escAttr(edge.id) + '\')" aria-label="' + escAttr(t('tm.chainPing')) + '">' +
        '<svg class="tm-link-ping-svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.4" stroke-linecap="round">' +
          '<path d="M2 12h4"/><path d="M18 12h4"/>' +
          '<path d="M8 12a4 4 0 0 1 8 0" opacity=".55"/>' +
          '<path d="M6 12a6 6 0 0 1 12 0" opacity=".3"/>' +
          '<circle cx="12" cy="12" r="1.6" fill="currentColor" stroke="none"/>' +
        '</svg>' + res + '</button>' +
      '<button type="button" title="' + escAttr(t('delete')) + '" onclick="tmDeleteEdge(event,\'' + escAttr(edge.id) + '\')"><span class="material-symbols-outlined">delete</span></button>';
    layer.appendChild(chip);
    // Only the selected bar shows its two ends — otherwise every bar would be
    // covered in handles.
    if (TM.selectedEdge !== edge.id) return;
    ['from', 'to'].forEach(function (which) {
      var handle = document.createElement('div');
      handle.className = 'tm-edge-handle';
      var pt = (which === 'from') ? geo.start : geo.end;
      handle.style.left = pt.x + 'px';
      handle.style.top = pt.y + 'px';
      layer.appendChild(handle);
    });
  });
  tmArmChipTimer();
}

// After a ping the number stays for a while, then disappears by itself.
function tmArmChipTimer() {
  clearTimeout(TM.chipTimer);
  var now = Date.now();
  var soonest = null;
  TM.edges.forEach(function (edge) {
    var at = TM.edgePingAt[edge.id];
    if (!at) return;
    var hideAt = at + TM_CHIP_PING_MS;
    if (hideAt > now && (soonest == null || hideAt < soonest)) soonest = hideAt;
  });
  if (soonest == null) return;
  TM.chipTimer = setTimeout(function () {
    var changed = false;
    TM.edges.forEach(function (edge) {
      var at = TM.edgePingAt[edge.id];
      if (at && Date.now() - at > TM_CHIP_PING_MS) { delete TM.edgePingAt[edge.id]; changed = true; }
    });
    // Re-render only when the chip is not being dragged right now.
    if (changed && !TM.bar) tmRenderEdges();
  }, soonest - now + 30);
}

// The expanded node popover is temporary: it closes itself after a while, so
// the canvas cannot end up covered in open boxes.
function tmArmNodeTimer() {
  clearTimeout(TM.nodeTimer);
  if (!TM.openNode) return;
  TM.nodeTimer = setTimeout(function () {
    if (TM.openNode) { TM.openNode = ''; tmRenderGraph(); }
  }, TM_NODE_AUTO_MS);
}

// Chain ping for exactly one link: this panel pings `from` and asks `from` to
// ping `to`. A pure server-to-server number, shown beside the bar.
async function tmPingEdge(event, edgeId) {
  if (event) { event.stopPropagation(); if (event.preventDefault) event.preventDefault(); }
  var edge = tmEdgeById(edgeId);
  if (!edge) return;
  TM.edgePingBusy = edgeId;
  TM.edgePings[edgeId] = null;
  delete TM.edgePingAt[edgeId];
  tmRenderEdges();
  try {
    var body = { active_server: TM.activeServer, from: edge.from, to: edge.to };
    var r = await api('/api/tm/edge-ping', { method: 'POST', body: body });
    TM.edgePings[edgeId] = (r && r.ms != null) ? r.ms : null;
    TM.edgePingAt[edgeId] = Date.now();
    if (r && r.ms == null) toast(r.error || t('tm.offline'), 'warn');
  } catch (e) {
    TM.edgePings[edgeId] = null;
    delete TM.edgePingAt[edgeId];
    toast(e.message || t('tm.offline'), 'err');
  }
  TM.edgePingBusy = '';
  tmRenderEdges();
}

// ─── Edge editing ───────────────────────────────────────────────────────────
async function tmDeleteEdge(event, edgeId) {
  if (event) { event.stopPropagation(); if (event.preventDefault) event.preventDefault(); }
  var id = edgeId || TM.selectedEdge;
  if (!id) return;
  var edge = tmEdgeById(id);
  if (!edge) return;
  var backup = TM.edges.slice();
  TM.edges = TM.edges.filter(function (e) { return e.id !== id; });
  if (TM.selectedEdge === id) TM.selectedEdge = '';
  var ok = await tmValidateEdges(t('tm.edgeDeleted'));
  if (!ok) {
    TM.edges = backup;
    TM.selectedEdge = id;
    tmRenderGraph();
    return;
  }
  delete TM.edgePings[id];
  delete TM.edgePingAt[id];
  tmRenderGraph();
  toast(t('tm.edgeDeleted'), 'ok');
  tmScheduleApply(0);
}

// The two link keys on a box replace any "Connect" button:
//   * the key ON the face of the box     → this box becomes the DESTINATION
//   * the key BEHIND the box (its port)  → this box becomes the SOURCE
// As soon as both ends are picked the connection is created. With a line
// already selected, a key re-points that line instead — that is "change
// destination" without a separate dialog.
async function tmPickLink(event, nodeId, kind) {
  if (event) { event.stopPropagation(); if (event.preventDefault) event.preventDefault(); }

  var selected = tmEdgeById(TM.selectedEdge);
  if (selected) {
    var from = selected.from, to = selected.to;
    if (kind === 'from' && nodeId !== from) {
      selected.from = nodeId;
      TM.selectedEdge = '';
      return tmRetarget(selected, from, to, 'from');
    }
    if (kind === 'to' && nodeId !== to) {
      selected.to = nodeId;
      TM.selectedEdge = '';
      return tmRetarget(selected, from, to, 'to');
    }
  }

  if (kind === 'from') {
    TM.pendingSrc = (TM.pendingSrc === nodeId) ? '' : nodeId;
    if (TM.pendingSrc) { TM.pendingDst = ''; TM.pendingFromKey = 'from'; }
  } else {
    TM.pendingDst = (TM.pendingDst === nodeId) ? '' : nodeId;
    if (TM.pendingDst) { TM.pendingSrc = ''; TM.pendingFromKey = 'to'; }
  }
  if (TM.pendingSrc && TM.pendingDst) {
    var src = TM.pendingSrc, dst = TM.pendingDst;
    TM.pendingSrc = ''; TM.pendingDst = '';
    tmCancelDraw();
    return tmAddEdge(src, dst);
  }
  if (TM.pendingSrc || TM.pendingDst) tmStartDraw();
  else tmCancelDraw();
  tmRenderGraph();
}

// ─── Live link drawing ──────────────────────────────────────────────────────
// After a link key is pressed the bar follows the pointer immediately — the
// connection is not created only after the second key; the user sees what they
// are connecting while they move.
var TM_DRAW_PX = 26;   // how close the pointer must get for the magnet to lock

function tmDrawAnchor() {
  var id = TM.pendingFromKey === 'to' ? TM.pendingDst : TM.pendingSrc;
  return tmNodeBox(id);
}

function tmStartDraw() {
  tmCancelDraw();
  var svgLayer = $t('tm-edge-layer');
  if (!svgLayer) return;
  var line = document.createElementNS('http://www.w3.org/2000/svg', 'path');
  line.setAttribute('id', 'tm-draw-line');
  line.setAttribute('class', 'tm-edge-dangle');
  svgLayer.appendChild(line);
  var box = tmDrawAnchor();
  var cx = box.x + box.w / 2, cy = box.y + box.h / 2;
  line.setAttribute('d', 'M' + cx + ',' + cy + ' L' + cx + ',' + cy);
  var canvas = $t('tm-canvas');
  if (canvas) {
    canvas.addEventListener('pointermove', tmDrawMove);
    canvas.addEventListener('pointerdown', tmDrawClick, { capture: true });
  }
}

function tmDrawMove(e) {
  var line = $t('tm-draw-line');
  if (!line) return;
  var pt = tmCanvasPoint(e.clientX, e.clientY);
  var box = tmDrawAnchor();
  var cx = box.x + box.w / 2, cy = box.y + box.h / 2;
  line.setAttribute('d', 'M' + cx + ',' + cy + ' L' + pt.x + ',' + pt.y);
  // Magnet: when the pointer hovers over a box that could receive the link,
  // light it up so the user knows where a click will land.
  var fixed = TM.pendingFromKey === 'to' ? TM.pendingDst : TM.pendingSrc;
  var hoverId = '';
  var nodeEl = document.elementFromPoint(e.clientX, e.clientY);
  var hover = nodeEl && nodeEl.closest && nodeEl.closest('.tm-node');
  if (hover) hoverId = hover.dataset.node || '';
  tmShowSnap(hoverId && hoverId !== fixed ? hoverId : '');
}

function tmDrawClick(e) {
  // A click on the canvas completes the link to whatever box is under the
  // pointer; clicking empty space cancels the pending link.
  e.stopPropagation();
  e.preventDefault();
  var fixed = TM.pendingFromKey === 'to' ? TM.pendingDst : TM.pendingSrc;
  var nodeEl = document.elementFromPoint(e.clientX, e.clientY);
  var hover = nodeEl && nodeEl.closest && nodeEl.closest('.tm-node');
  var target = hover ? (hover.dataset.node || '') : '';
  var src, dst;
  if (TM.pendingFromKey === 'to') { src = target; dst = fixed; }
  else { src = fixed; dst = target; }
  tmCancelDraw();
  tmClearSnap();
  if (target && target !== fixed) {
    tmAddEdge(src, dst);
  } else {
    TM.pendingSrc = ''; TM.pendingDst = '';
    tmRenderGraph();
  }
}

function tmCancelDraw() {
  var line = $t('tm-draw-line');
  if (line) line.remove();
  var canvas = $t('tm-canvas');
  if (canvas) {
    canvas.removeEventListener('pointermove', tmDrawMove);
    canvas.removeEventListener('pointerdown', tmDrawClick, { capture: true });
  }
  tmClearSnap();
}

async function tmRetarget(edge, wasFrom, wasTo, which) {
  var backupEdge = { id: edge.id, from: wasFrom, to: wasTo, style: edge.style };
  var ok = await tmValidateEdges(t('tm.destChanged'));
  if (!ok) {
    edge.from = backupEdge.from;
    edge.to = backupEdge.to;
    TM.selectedEdge = edge.id;
    tmRenderGraph();
    return false;
  }
  await tmSaveGraph({});
  tmRenderGraph();
  toast(t('tm.destChanged'), 'ok');
  tmScheduleApply(0);
  return true;
}

// Ask the backend before committing: the graph is validated by the same code
// that will run on Apply, so a bad link is refused here rather than at apply time.
async function tmValidateEdges(failToast) {
  try {
    var r = await api('/api/tm/graph/validate', {
      method: 'POST',
      body: { edges: TM.edges, active_server: TM.activeServer }
    });
    if (r.ok) { TM.status = r.status || TM.status; return true; }
    toast(r.error || (failToast || t('tm.invalid')), 'err');
    return false;
  } catch (e) {
    toast(e.message || t('toast.err'), 'err');
    return false;
  }
}

async function tmAddEdge(from, to) {
  if (!from || !to) return false;
  if (from === to) { toast(t('tm.selfLoop'), 'warn'); tmRenderGraph(); return false; }
  if (tmEdgeFor(from, to)) { toast(t('tm.duplicate'), 'warn'); tmRenderGraph(); return false; }
  if (tmPathExists(to, from)) { toast(t('tm.cycle'), 'err'); tmRenderGraph(); return false; }

  var edge = {
    id: 'e_' + Date.now().toString(36) + Math.random().toString(36).slice(2, 6),
    from: from, to: to,
    style: tmAutoStyle(tmNodeBox(from), tmNodeBox(to))
  };
  var backup = TM.edges.slice();
  TM.edges = TM.edges.concat([edge]);
  tmRenderGraph();

  if (!(await tmValidateEdges(t('tm.edgeAdded')))) {
    TM.edges = backup;
    tmRenderGraph();
    return false;
  }
  await tmSaveGraph({});
  tmRenderGraph();
  toast(t('tm.edgeAdded'), 'ok');
  tmScheduleApply(0);
  return true;
}

// Is there already a path from `from` down to `to`? Used as
// tmPathExists(newTo, newFrom) when adding edge from→to: a cycle exists iff
// `to` can already reach `from`. Each node is visited once, so a graph that is
// already a DAG can never loop this walk.
function tmPathExists(from, to) {
  var seen = {}, stack = [from];
  while (stack.length) {
    var cur = stack.pop();
    if (cur === to) return true;
    if (seen[cur]) continue;
    seen[cur] = true;
    TM.edges.forEach(function (e) { if (e.from === cur) stack.push(e.to); });
  }
  return false;
}

function tmCancelConnect() {
  TM.pendingSrc = '';
  TM.pendingDst = '';
  tmRenderGraph();
}

function tmToggleNode(event, nodeId) {
  if (event) event.stopPropagation();
  TM.openNode = (TM.openNode === nodeId) ? '' : nodeId;
  tmRenderGraph();
  tmArmNodeTimer();
}

async function tmPingThis(nodeId) {
  TM.pings[nodeId] = { clientMs: null, testing: true, at: Date.now() };
  tmUpdatePingUI(nodeId);
  var ms = await tmPingNode(nodeId, 3);
  tmRenderGraph();
  toast(ms != null ? (Math.round(ms) + ' ms') : t('tm.offline'), ms != null ? 'ok' : 'err');
}

// ─── Canvas interaction: pan, pinch, long-press drag ────────────────────────
function tmBindCanvas() {
  var canvas = $t('tm-canvas');
  if (!canvas || canvas.dataset.tmBound) return;
  canvas.dataset.tmBound = '1';

  canvas.addEventListener('pointerdown', tmPointerDown);
  canvas.addEventListener('pointermove', tmPointerMove);
  canvas.addEventListener('pointerup', tmPointerUp);
  canvas.addEventListener('pointercancel', tmPointerUp);
  canvas.addEventListener('wheel', tmWheel, { passive: false });
  canvas.addEventListener('contextmenu', function (e) { e.preventDefault(); });
}

function tmWheel(e) {
  e.preventDefault();
  var canvas = $t('tm-canvas');
  var r = canvas.getBoundingClientRect();
  tmZoom(e.clientX - r.left, e.clientY - r.top, e.deltaY < 0 ? 1.12 : 1 / 1.12);
}

function tmPointerDown(e) {
  var canvas = $t('tm-canvas');
  if (!canvas) return;
  if (e.pointerType === 'mouse' && e.button !== 0) return;
  TM.pointers[e.pointerId] = { x: e.clientX, y: e.clientY };

  // Two fingers: pinch-zoom, never a node drag.
  var ids = Object.keys(TM.pointers);
  if (ids.length === 2) {
    tmCancelPress();
    if (TM.bar) { tmCancelBar(); tmRenderEdges(); }
    TM.drag = null;
    TM.pan = null;
    var a = TM.pointers[ids[0]], b = TM.pointers[ids[1]];
    TM.pinch = { dist: Math.max(1, Math.hypot(a.x - b.x, a.y - b.y)), zoom: TM.zoom };
    return;
  }
  if (ids.length > 2) return;

  var target = e.target;
  var onButton = target.closest && target.closest('button');
  if (onButton) return;              // let the node's own buttons work untouched

  // A bar: grabbed either by its chip (move the whole bar) or by the line
  // itself (also moves it; a tap with no movement just selects it).
  var chip = target.closest && target.closest('.tm-edge-label');
  var hit = target.closest && target.closest('.tm-edge-hit');
  var barId = chip ? chip.dataset.edge : (hit ? hit.dataset.edge : '');
  if (barId) {
    tmActivateBar(canvas, e, barId, 'move', '');
    return;
  }

  var nodeEl = target.closest && target.closest('.tm-node');
  if (nodeEl) {
    var nodeId = nodeEl.dataset.node;
    if (TM.pendingSrc || TM.pendingDst) return;   // handled on pointerup as a tap
    var start = tmCanvasPoint(e.clientX, e.clientY);
    var srv = tmServerById(nodeId);
    TM.press = {
      nodeId: nodeId,
      pointerId: e.pointerId,
      startX: start.x,
      startY: start.y,
      originX: srv ? srv.x : 0,
      originY: srv ? srv.y : 0,
      moved: false,
      activated: false
    };
    if (e.pointerType === 'mouse') {
      tmActivateDrag(canvas, e);
    } else {
      // Long press to drag: a short touch gesture stays a scroll-ish tap and a
      // quick swipe on the canvas pans instead of moving the node.
      TM.press.timer = setTimeout(function () { tmActivateDrag(canvas, e); }, 320);
    }
    return;
  }
  // Empty canvas: pan (and clear any edge selection).
  if (TM.selectedEdge) { TM.selectedEdge = ''; tmRenderEdges(); }
  TM.pan = { pointerId: e.pointerId, x: e.clientX, y: e.clientY, panX: TM.panX, panY: TM.panY };
  canvas.classList.add('grabbing');
}

// ─── Magnetic bars ──────────────────────────────────────────────────────────
// Two ways to move a bar: grab its chip (or the line) to shift the curve, grab
// one of its round ends to pull that end to another box. The end sticks to a
// box as soon as the pointer is close enough, and the link is re-pointed on
// release — validation decides whether the new link is legal.
var TM_SNAP_PX = 62;   // magnet radius in canvas units

function tmSnapNodeAt(pt, exclude) {
  var best = '', bestD = Infinity;
  TM.servers.forEach(function (srv) {
    if (exclude && exclude.indexOf(srv.node_id) >= 0) return;
    var box = tmNodeBox(srv.node_id);
    var cx = box.x + box.w / 2, cy = box.y + box.h / 2;
    var inside = pt.x >= box.x && pt.x <= box.x + box.w && pt.y >= box.y && pt.y <= box.y + box.h;
    var d = inside ? 0 : Math.sqrt((pt.x - cx) * (pt.x - cx) + (pt.y - cy) * (pt.y - cy));
    if (d < bestD) { bestD = d; best = srv.node_id; }
  });
  return (best && bestD <= TM_SNAP_PX) ? best : '';
}

function tmShowSnap(nodeId) {
  if (TM.snapHl === nodeId) return;
  tmClearSnap();
  if (!nodeId) return;
  TM.snapHl = nodeId;
  var el = document.querySelector('.tm-node[data-node="' + nodeId + '"]');
  if (el) el.classList.add('snap-target');
}

function tmClearSnap() {
  if (!TM.snapHl) return;
  var el = document.querySelector('.tm-node[data-node="' + TM.snapHl + '"]');
  if (el) el.classList.remove('snap-target');
  TM.snapHl = '';
}

function tmActivateBar(canvas, e, edgeId, mode, which) {
  var edge = tmEdgeById(edgeId);
  if (!edge) return;
  var start = tmCanvasPoint(e.clientX, e.clientY);
  TM.bar = {
    edgeId: edgeId, pointerId: e.pointerId, mode: mode, end: which,
    start: start, moved: false, snapped: '', bent: tmEdgeBend(edge),
    prevSelected: TM.selectedEdge
  };
  if (mode === 'move') TM.selectedEdge = edgeId;
  try { canvas.setPointerCapture(e.pointerId); } catch (err) {}
  if (navigator.vibrate) { try { navigator.vibrate(8); } catch (err) {} }
}

function tmCancelBar() {
  TM.bar = null;
  tmClearSnap();
}

// Live update of the bar being dragged: the curve follows the pointer, or the
// pulled end dangles behind it while a box lights up as the magnet target.
function tmBarMove(pt) {
  var bar = TM.bar;
  var edge = tmEdgeById(bar.edgeId);
  if (!edge) return;
  var d = Math.sqrt((pt.x - bar.start.x) * (pt.x - bar.start.x) + (pt.y - bar.start.y) * (pt.y - bar.start.y));
  if (d * TM.zoom > 6) bar.moved = true;

  var a = tmNodeBox(edge.from), b = tmNodeBox(edge.to);
  var acx = a.x + a.w / 2, acy = a.y + a.h / 2;
  var bcx = b.x + b.w / 2, bcy = b.y + b.h / 2;
  var dx = bcx - acx, dy = bcy - acy;
  var len = Math.max(1, Math.sqrt(dx * dx + dy * dy));
  var ux = dx / len, uy = dy / len;
  var mx = (acx + bcx) / 2, my = (acy + bcy) / 2;

  if (bar.mode === 'move') {
    if (!bar.moved) return;
    // Reference geometry from the bend the bar had when the drag started, so
    // the very first pixel of movement does not jump.
    var geo = tmEdgeGeometry(a, b, edge.style, bar.bent);
    var nx = -uy, ny = ux;
    // The chip sits at half the control offset (quadratic midpoint weighting),
    // so a chip move of d needs a bend move of 2d. The chord's own offset
    // cancels out because the box exits sit along the chord direction.
    var off0 = 2 * ((geo.mid.x - mx) * nx + (geo.mid.y - my) * ny);
    var p0 = (bar.start.x - mx) * nx + (bar.start.y - my) * ny;
    var p1 = (pt.x - mx) * nx + (pt.y - my) * ny;
    var bend = Math.max(-260, Math.min(260, off0 + 2 * (p1 - p0)));
    if (Math.abs(bend) < 6) bend = 0;
    edge.bend = bend;
    tmRenderEdgesSoon();
    return;
  }

  // Pulling an end: the end follows the pointer, and the nearest box that
  // could receive it lights up.
  if (!bar.moved) return;
  var other = (bar.end === 'from') ? edge.to : edge.from;
  var snap = tmSnapNodeAt(pt, [other]);
  var current = (bar.end === 'from') ? edge.from : edge.to;
  tmShowSnap(snap && snap !== current ? snap : '');
  var handles = document.querySelectorAll('.tm-edge-handle');
  var handle = handles[(bar.end === 'from') ? 0 : 1];
  if (handle) { handle.style.left = pt.x + 'px'; handle.style.top = pt.y + 'px'; }
  bar.snapped = (snap && snap !== current) ? snap : '';

  // The fixed end stays put; a dashed cord shows where the pulled end is going.
  var visible = document.querySelector('.tm-edge[data-edge="' + edge.id + '"]');
  if (visible) visible.style.opacity = '.18';
  var dangle = $t('tm-edge-dangle');
  if (!dangle) {
    dangle = document.createElementNS('http://www.w3.org/2000/svg', 'path');
    dangle.setAttribute('id', 'tm-edge-dangle');
    dangle.setAttribute('class', 'tm-edge-dangle');
    var svgLayer = $t('tm-edge-layer');
    if (svgLayer) svgLayer.appendChild(dangle);
  }
  var fixed = (bar.end === 'from') ? geo.end : geo;
  if (dangle) {
    dangle.setAttribute('d', 'M' + fixed.x + ',' + fixed.y + ' L' + pt.x + ',' + pt.y);
  }
}

// What a released bar does: a tap selects (or deselects) it, a bend drag
// persists the new shape, and a pulled end re-points the link — the backend
// validates that the new link is legal before anything is kept.
async function tmBarDrop() {
  var bar = TM.bar;
  if (!bar) return;
  var edge = tmEdgeById(bar.edgeId);
  var snap = bar.snapped;
  TM.bar = null;
  tmClearSnap();
  if (!edge) { tmRenderEdges(); return; }

  if (!bar.moved) {
    // A tap: select the bar (or deselect it when it already was).
    TM.selectedEdge = (bar.prevSelected === edge.id) ? '' : edge.id;
    tmRenderEdges();
    return;
  }

  if (!bar.end) {
    // Bending changes no topology, only the drawing: save and be done.
    await tmSaveGraph({});
    tmRenderEdges();
    return;
  }

  if (!snap) { tmRenderEdges(); return; }   // the end went nowhere near a box

  var wasFrom = edge.from, wasTo = edge.to;
  if (bar.end === 'from') edge.from = snap;
  else edge.to = snap;
  if (!(await tmValidateEdges(t('tm.destChanged')))) {
    edge.from = wasFrom;
    edge.to = wasTo;
    tmRenderEdges();
    return;
  }
  await tmSaveGraph({});
  tmRenderGraph();
  toast(t('tm.destChanged'), 'ok');
  tmScheduleApply(0);
}

function tmActivateDrag(canvas, e) {
  if (!TM.press || TM.press.activated) return;
  TM.press.activated = true;
  if (!tmServerById(TM.press.nodeId)) { tmCancelPress(); return; }
  var el = document.querySelector('.tm-node[data-node="' + TM.press.nodeId + '"]');
  if (el) el.classList.add('dragging');
  TM.drag = TM.press;
  try { canvas.setPointerCapture(e.pointerId); } catch (err) {}
  if (navigator.vibrate) { try { navigator.vibrate(12); } catch (err) {} }
}

function tmCancelPress() {
  if (TM.press && TM.press.timer) clearTimeout(TM.press.timer);
  TM.press = null;
}

function tmPointerMove(e) {
  if (TM.pointers[e.pointerId]) TM.pointers[e.pointerId] = { x: e.clientX, y: e.clientY };

  if (TM.pinch && Object.keys(TM.pointers).length >= 2) {
    var ids = Object.keys(TM.pointers);
    var a = TM.pointers[ids[0]], b = TM.pointers[ids[1]];
    var dist = Math.max(1, Math.hypot(a.x - b.x, a.y - b.y));
    var canvas = $t('tm-canvas');
    var r = canvas.getBoundingClientRect();
    tmZoom((a.x + b.x) / 2 - r.left, (a.y + b.y) / 2 - r.top, (TM.pinch.zoom * (dist / TM.pinch.dist)) / TM.zoom);
    return;
  }

  // A bar is being carried: bend it, or pull one of its ends onto a box.
  if (TM.bar && TM.bar.pointerId === e.pointerId) {
    tmBarMove(tmCanvasPoint(e.clientX, e.clientY));
    return;
  }

  if (TM.drag && TM.drag.pointerId === e.pointerId) {
    var pt = tmCanvasPoint(e.clientX, e.clientY);
    var srv = tmServerById(TM.drag.nodeId);
    if (!srv) return;
    srv.x = Math.max(-400, Math.min(4000, TM.drag.originX + (pt.x - TM.drag.startX)));
    srv.y = Math.max(-400, Math.min(4000, TM.drag.originY + (pt.y - TM.drag.startY)));
    var el = document.querySelector('.tm-node[data-node="' + TM.drag.nodeId + '"]');
    if (el) el.style.transform = 'translate(' + Math.round(srv.x) + 'px,' + Math.round(srv.y) + 'px)';
    tmRenderEdgesSoon();
    return;
  }

  if (TM.press && TM.press.pointerId === e.pointerId) {
    var now = tmCanvasPoint(e.clientX, e.clientY);
    var dx = now.x - TM.press.startX, dy = now.y - TM.press.startY;
    // Screen-space threshold, so a zoomed-out canvas is not jittery and a
    // zoomed-in one still lets a deliberate slide start a node drag.
    var threshold = 16 / Math.max(0.5, TM.zoom);
    if (Math.sqrt(dx * dx + dy * dy) > threshold) {
      // Moved before the long press fired: this was a pan, not a node drag.
      if (TM.press.timer) clearTimeout(TM.press.timer);
      TM.press = null;
      TM.pan = { pointerId: e.pointerId, x: e.clientX, y: e.clientY, panX: TM.panX, panY: TM.panY };
      var canvasEl = $t('tm-canvas');
      if (canvasEl) canvasEl.classList.add('grabbing');
    }
    return;
  }

  if (TM.pan && TM.pan.pointerId === e.pointerId) {
    TM.panX = TM.pan.panX + (e.clientX - TM.pan.x);
    TM.panY = TM.pan.panY + (e.clientY - TM.pan.y);
    tmApplyTransform();
  }
}

function tmPointerUp(e) {
  delete TM.pointers[e.pointerId];
  if (Object.keys(TM.pointers).length < 2) TM.pinch = null;

  if (TM.bar && TM.bar.pointerId === e.pointerId) {
    // A cancelled gesture must not leave a half-pulled bar or a stale cord.
    if (e.type === 'pointercancel') {
      var cancelled = tmEdgeById(TM.bar.edgeId);
      if (cancelled) cancelled.bend = TM.bar.bent;
      tmCancelBar();
      tmRenderEdges();
      return;
    }
    tmBarDrop();
    return;
  }

  if (TM.drag && TM.drag.pointerId === e.pointerId) {
    var el = document.querySelector('.tm-node[data-node="' + TM.drag.nodeId + '"]');
    if (el) el.classList.remove('dragging');
    TM.drag = null;
    TM.press = null;
    tmPersistPositionsSoon();
    return;
  }

  if (TM.press && TM.press.pointerId === e.pointerId) {
    var wasActivated = TM.press.activated;
    if (TM.press.timer) clearTimeout(TM.press.timer);
    var nodeId = TM.press.nodeId;
    TM.press = null;
    if (!wasActivated) {
      if (e.type !== 'pointercancel') {
        // A plain tap opens the node's own controls (name, ping, Edit, Set).
        // Linking is done with the two keys, never with the body of the box.
        TM.openNode = (TM.openNode === nodeId) ? '' : nodeId;
        tmRenderGraph();
        tmArmNodeTimer();
      }
    }
    return;
  }

  if (TM.pan && TM.pan.pointerId === e.pointerId) {
    TM.pan = null;
    var canvas = $t('tm-canvas');
    if (canvas) canvas.classList.remove('grabbing');
  }
}

// ─── Validate / Apply / Generate ────────────────────────────────────────────
// ─── Automatic apply ────────────────────────────────────────────────────────
// There is no Apply button: any accepted change to the graph is saved and then
// pushed to the servers by itself. Edits are coalesced, so dragging a box or
// finishing a chain produces exactly one push.
var TM_APPLY_TIMER = null;
var TM_APPLY_RUNNING = false;

function tmSetAutoState(kind, text) {
  var el = $t('tm-auto-state');
  if (!el) return;
  el.classList.remove('busy', 'err');
  if (kind) el.classList.add(kind);
  var label = el.querySelector('span:last-child');
  if (label) label.textContent = text;
}

function tmScheduleApply(delay) {
  clearTimeout(TM_APPLY_TIMER);
  TM_APPLY_TIMER = setTimeout(function () { tmAutoApply(); }, delay == null ? 900 : delay);
}

async function tmAutoApply() {
  if (TM_APPLY_RUNNING) { tmScheduleApply(1200); return; }
  if (!TM.edges.length && !TM.activeServer) return;
  TM_APPLY_RUNNING = true;
  tmSetAutoState('busy', t('tm.applying'));
  try {
    await tmSaveGraph({});
    var r = await api('/api/tm/apply', { method: 'POST', body: {} });
    if (r.ok) {
      TM.appliedChain = r.chain || [];
      TM.appliedEdges = TM.edges.slice();
      tmSetAutoState('', t('tm.autoDone'));
      tmRenderProgress(r.steps);
      tmGenerate();
      // A change to the chain invalidates the per-link numbers that were
      // measured against the previous chain.
      tmClearEdgePings();
    } else {
      tmSetAutoState('err', t('tm.applyFailed'));
      tmRenderProgress(r.steps);
      toast(r.error || t('tm.applyFailed'), 'err');
    }
    await loadTM();
  } catch (e) {
    tmSetAutoState('err', t('tm.applyFailed'));
    toast(e.message || t('tm.applyFailed'), 'err');
  }
  TM_APPLY_RUNNING = false;
}

function tmClearEdgePings() {
  TM.edgePings = {};
  TM.edgePingAt = {};
  tmRenderEdges();
}

// The progress list stays as a short, self-expiring report of the last push.
function tmRenderProgress(steps) {
  var el = $t('tm-progress');
  if (!el) return;
  if (!steps || !steps.length) { el.classList.remove('show'); el.innerHTML = ''; return; }
  el.classList.add('show');
  el.innerHTML = steps.map(function (s) {
    var cls = s.ok ? 'ok' : 'fail';
    var icon = s.ok ? 'check_circle' : 'error';
    return '<div class="tm-progress-step ' + cls + '"><span class="material-symbols-outlined">' + icon + '</span><span>' + esc(s.label) + '</span></div>';
  }).join('');
  clearTimeout(TM.progressTimer);
  TM.progressTimer = setTimeout(function () {
    el.classList.remove('show');
    el.innerHTML = '';
  }, 9000);
}

async function tmGenerate() {
  var pre = $t('tm-config-preview');
  if (!pre) return;
  if (!TM.servers.length) { pre.textContent = t('tm.noServers'); return; }
  try {
    var r = await api('/api/tm/generate', { method: 'POST', body: { node_id: TM.activeServer } });
    pre.textContent = r.ok ? r.config : (r.error || '—');
  } catch (e) { pre.textContent = e.message || '—'; }
}

function tmCopyConfig() {
  var pre = $t('tm-config-preview');
  if (!pre || !pre.textContent) return;
  var txt = pre.textContent;
  if (navigator.clipboard && navigator.clipboard.writeText) {
    navigator.clipboard.writeText(txt).then(function () { toast(t('copyOk') || 'Copied', 'ok'); })
      .catch(function () { toast(t('toast.err'), 'err'); });
  } else { toast(t('toast.err'), 'err'); }
}

// ─── Wiring ─────────────────────────────────────────────────────────────────
document.addEventListener('DOMContentLoaded', function () {
  tmBindCanvas();
});
setTimeout(function () { tmBindCanvas(); }, 0);

// Login form submit
var loginFormEl = $id('login-form');
if (loginFormEl) {
  loginFormEl.addEventListener('submit', function(e) { e.preventDefault(); doLogin(); });
}
