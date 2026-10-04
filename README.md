<div align="center">

<img src="svg/hero.svg" alt="WhitePanel — control panel hero" width="100%" />

# WHITE PANEL

**One panel. Every lever. Users, subscriptions, nodes, scanners and Telegram sales — driven from a single black & teal console.**

`English` · [فارسی](README.fa.md) · [العربية](README.ar.md) · [Русский](README.ru.md)

<br>

<img src="https://img.shields.io/badge/FastAPI-009485?style=for-the-badge&logo=fastapi&logoColor=white" alt="FastAPI" />
<img src="https://img.shields.io/badge/Python-3776AB?style=for-the-badge&logo=python&logoColor=white" alt="Python" />
<img src="https://img.shields.io/badge/Xray-00E1C1?style=for-the-badge&logoColor=black" alt="Xray" />
<img src="https://img.shields.io/badge/Port-8080-00E1C1?style=for-the-badge&logoColor=black" alt="Port 8080" />
<img src="https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white" alt="Docker" />
<img src="https://img.shields.io/badge/Railway-0B0B0F?style=for-the-badge&logo=railway&logoColor=00E1C1" alt="Railway" />

</div>

---

<details>
<summary><b>📚 Table of contents</b></summary>

- [Why WhitePanel?](#-why-whitepanel)
- [Deployment](#-deployment)
- [Public endpoint & subscription flow](#-public-endpoint--subscription-flow)
- [Telegram automation](#-telegram-automation)
- [Account, expiry & delivery behavior](#-account-expiry--delivery-behavior)
- [Scanner](#-scanner)
- [Nodes & workers](#-nodes--workers)
- [Project structure](#-project-structure)
- [Runtime & API](#-runtime--api)
- [UI details](#-ui-details)
- [Security notes](#-security-notes)
- [Development](#-development)
- [Creator](#creator)

</details>

---

## ✦ Why WhitePanel?

Most proxy stacks are a pile of half-connected scripts: one thing for users, another for subs, a third for the bot, and a spreadsheet for the rest.

WhitePanel collapses all of it into **one operational surface** — and it stays useful whether you install it on a bare VPS or ship it into a container platform where the public hostname is only known *after* deploy.

<div align="center">
  <img src="svg/features.svg" alt="WhitePanel feature overview" width="100%" />
</div>

### Core capabilities

| Area | What it gives you |
|---|---|
| 👤 **Users** | Create/manage accounts, limits, traffic, links and per-user actions |
| 🔗 **Subscriptions** | Subscription pages, config links, QR codes and copy actions |
| 🛰 **Nodes** | Node management, health checks, refresh/sync operations |
| 🧪 **Scanner** | TCP / IP / SNI tooling, result lists and copy-friendly outputs |
| 🤖 **Telegram Bot** | Bot configuration, channel automation and sales workflows |
| 🛒 **Sell Bot** | Plans, receipts, manual approval and subscription delivery |
| 🟢 **Expiry** | Automatic expiration handling and cleanup of expired users |
| 🛠 **Tools** | Runtime info, workers, tunnel status and utility endpoints |

> **Fixed panel port: `8080`.** Remember it once, never hunt for it again.

---

# 🚀 Deployment

<div align="center">
  <img src="svg/deploy.svg" alt="WhitePanel deployment flow" width="100%" />
</div>

## Option A — One-command VPS install

Run the installer on a supported Linux VPS:

```bash
bash start.sh install
```

The installer handles dependency setup, application files and runtime management. The panel itself listens on **port `8080`**.

### Useful installer variables

| Variable | Default | Purpose |
|---|---|---|
| `WHITE_APP_DIR` | `/opt/WhitePanel` | Application directory |
| `WHITE_REPO` | *(unset)* | Git repository to deploy |
| `WHITE_BRANCH` | `main` | Branch to deploy |
| `WHITE_INSTALLER_URL` | project `start.sh` URL | Installer source |
| `WHITE_UV_VERSION` | `0.12.9` | uv release used by installer |
| `WHITE_XRAY_VERSION` | `26.3.27` | Xray release used by installer |

```bash
# deploy a specific branch into a custom directory
WHITE_APP_DIR=/opt/WhitePanel WHITE_BRANCH=main bash start.sh install
```

> 🔒 Keep secrets out of the repository. Prefer environment variables or protected server configuration for any sensitive deployment data.

---

## Option B — Railway

1. Fork the repository into your GitHub account.
2. Create a new Railway project from the GitHub repository.
3. Let Railway build from the included `Dockerfile` / `railway.toml` configuration.
4. Expose the application on **port `8080`**.
5. Generate a public domain.
6. Open the generated domain and continue into the panel.

### Railway notes

- The application process listens on `0.0.0.0:8080`.
- `railway.toml` runs Uvicorn on the same fixed port.
- Public-domain discovery matters: subscription URLs and user-facing links are built from it.
- Never hard-code `localhost` as a public subscription hostname.

---

# 🧭 Public endpoint & subscription flow

WhitePanel ships runtime/public-endpoint helpers so it can work when the final public hostname is assigned *after* deployment.

```text
Panel
  ├─ /white                → main panel
  ├─ /login                → login page
  ├─ /dashboard            → dashboard page
  ├─ /link/<uuid>          → user link view
  ├─ /p/<uuid_key>         → public subscription page
  └─ /api/sub/<uuid_key>   → subscription endpoint
```

QR endpoints are also exposed for subscription / user configuration delivery.

---

# 🤖 Telegram automation

WhitePanel includes Telegram-oriented workflows for administration, channel operations and sales.

### Channel Bot

The Channel Bot flow can create a new user, generate the subscription + QR, publish the generated content to the configured channel, and then perform the related cleanup step after successful delivery.

### Sell Bot

The customer-facing flow is intentionally boring — boring means conversion:

```text
🟢 My Account   →   🛍 Products   →   💬 Support
```

Plans are managed independently and can carry their own name, price, inbound, traffic quota, duration and user prefix.

A typical purchase flow is:

```text
/start
   ↓
🛍 Products
   ↓
Select Plan
   ↓
Payment info
   ↓
Receipt upload
   ↓
Admin review
   ├─ ✅ Approve → Subscription + QR
   └─ ❌ Reject
```

Also supported: forced channel membership checks, numeric Telegram admin IDs, plan management and expiration handling.

---

# 🛡 Account, expiry & delivery behavior

### Account identity

The Telegram sales flow associates a customer with a **Telegram Numeric User ID**. Existing accounts are updated rather than blindly duplicated, and order activation is designed to avoid double-applying the same purchase.

### Expiration

When a user's configured expiration is reached, the expiry sweeper attempts to remove the account, clean its Telegram mapping, then report the expiration to the user.

### Delivery safety

Activation and delivery are deliberately separated, so a temporary Telegram delivery failure does **not** mean the account was duplicated or that a previous account was wrongly reported as removed.

---

# 🔎 Scanner

The scanner area includes tooling around:

- TCP / IP checks
- Cloudflare subnet data
- SNI lists and checks
- IP / SNI scan result handling
- Batch ping operations
- Fastest-result helpers

The mobile UI is tuned for compact, touch-friendly results so a domain or IP can be tapped and copied directly.

---

# 🌐 Nodes & workers

Node and worker tooling covers health checks, refresh/sync operations, worker setup and heartbeat/health helpers.

<div align="center">
  <img src="svg/architecture.svg" alt="WhitePanel architecture" width="100%" />
</div>

At a high level:

```text
Web UI
  │
  ▼
FastAPI application
  ├── users / auth
  ├── subscriptions / QR
  ├── nodes / workers
  ├── scanner
  ├── Telegram bot
  └── runtime helpers
        │
        ├── Xray / proxy runtime
        └── application data
```

---

# 📁 Project structure

```text
.
├── data/
│   ├── cf_subnets.txt
│   ├── endpoint.txt
│   ├── sni-list.txt
│   ├── sni_reality.txt
│   └── sni_reality_for_scan.txt
├── static/
│   ├── index.html
│   ├── login.html
│   ├── sub.html
│   ├── white-logo.svg
│   ├── img/
│   └── musix/
├── svg/                    # README / project SVG artwork (theme-matched)
│   ├── logo.svg
│   ├── hero.svg
│   ├── features.svg
│   ├── deploy.svg
│   └── architecture.svg
├── worker/
│   ├── worker.js
│   └── _worker.js
├── Dockerfile
├── railway.toml
├── requirements.txt
├── start.sh
├── main.py
└── README.md
```

---

# ⚙️ Runtime & API

The backend is built with **FastAPI / Uvicorn**. The repository contains endpoints for authentication, users, subscriptions, nodes, scanners, workers, Telegram bot configuration and runtime information.

The Docker image exposes the panel on:

```text
0.0.0.0:8080
```

For a clean deployment, map your platform's public hostname to that application port.

---

# 🎨 UI details

The interface is built around:

- responsive dark/light presentation — pure black base, teal `#00e1c1` accent
- compact touch interactions
- copy-friendly configuration actions
- collapsible sections where appropriate
- subscription pages with QR delivery
- Telegram glass-button style flows
- desktop + mobile layouts

The `svg/` directory contains the original documentation artwork used by these READMEs. They are intentionally standalone SVG files so the repository keeps its visual identity without relying on raster screenshots — and they are recolored to the panel's live theme.

---

# 🔐 Security notes

- Never commit Telegram bot tokens, passwords, API keys or private deployment credentials.
- Put sensitive values into protected environment/server configuration.
- Use HTTPS for production panel access.
- Restrict administrative access to trusted operators.
- Review generated subscription links before sharing them publicly.

---

# 🧑‍💻 Development

For local experimentation, install the Python dependencies and run the FastAPI application with Uvicorn:

```bash
python -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt
python -m uvicorn main:app --host 0.0.0.0 --port 8080
```

Then open:

```text
http://127.0.0.1:8080/white
```

For production or public subscriptions, use the real public hostname rather than `127.0.0.1` / `localhost` in user-facing links.

---

<div align="center">
  <img src="svg/logo.svg" alt="WhitePanel logo" width="360" />

  <h3>WHITE PANEL</h3>
  <p><i>Deploy fast. Manage cleanly. Keep every configuration within reach.</i></p>
</div>

---

## Creator

Developed & maintained by **[Itskillmaster](https://github.com/Itskillmaster/White-Panel-Railway)**.

- Repository: <https://github.com/Itskillmaster/White-Panel-Railway>
- The panel's built-in update checker (**Settings → بروزرسانی پنل**) compares your install against this repository's latest commit.

<div align="center">
  <img src="svg/logo.svg" alt="WhitePanel" width="320" />
  <br><br>
  <sub><b>English</b> · <a href="README.fa.md">فارسی</a> · <a href="README.ar.md">العربية</a> · <a href="README.ru.md">Русский</a></sub>
</div>
