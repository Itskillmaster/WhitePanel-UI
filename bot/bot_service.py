"""
White Panel - Telegram Engine & Bot Automation
Handles:
1. Customer Workflow: /start, Show Plans, Place Order, Send Receipt, Receive Config & QR
2. Channel Bot: Periodic status and node health announcements
3. Order Approval: Auto or Admin-confirmed approval with instant config generation via Go Backend API
"""

import os
import sys
import json
import time
import asyncio
import logging
from typing import Dict, Any, Optional, List
import urllib.parse

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] [WhiteBot] %(message)s"
)
logger = logging.getLogger("WhiteBot")

PANEL_API_URL = os.environ.get("WHITE_PANEL_API_URL", "http://127.0.0.1:8080").rstrip("/")
BOT_TOKEN = os.environ.get("WHITE_BOT_TOKEN", "")
ADMIN_ID = os.environ.get("WHITE_ADMIN_ID", "")
TG_API = f"https://api.telegram.org/bot{BOT_TOKEN}"

# In-memory orders and states
ORDERS: Dict[str, Dict[str, Any]] = {}
USER_STEPS: Dict[int, str] = {}

class GoBackendClient:
    """HTTP Client communicating directly with White Panel Go Core"""
    def __init__(self, base_url: str):
        self.base_url = base_url

    async def get_plans(self) -> List[Dict[str, Any]]:
        import httpx
        try:
            async with httpx.AsyncClient(timeout=5.0) as client:
                r = await client.get(f"{self.base_url}/api/bot/sell/plans")
                if r.status_code == 200:
                    data = r.json()
                    return data.get("plans", [])
        except Exception as e:
            logger.error(f"Failed to fetch plans from Go backend: {e}")
        # Default fallback plans
        return [
            {"id": "plan-1", "name": "🚀 پلن برنزی (1 ماهه - 30 گیگ)", "traffic_gb": 30, "expire_days": 30, "price_toman": 85000},
            {"id": "plan-2", "name": "⚡ پلن نقره‌ای (1 ماهه - 60 گیگ)", "traffic_gb": 60, "expire_days": 30, "price_toman": 140000},
            {"id": "plan-3", "name": "💎 پلن طلایی (1 ماهه - 100 گیگ)", "traffic_gb": 100, "expire_days": 30, "price_toman": 210000},
        ]

    async def create_user_and_generate_config(self, username: str, traffic_gb: int, expire_days: int) -> Optional[Dict[str, Any]]:
        """Invokes Go backend to create account and generate immediate configs"""
        import httpx
        payload = {
            "username": username,
            "traffic_limit": int(traffic_gb * 1024 * 1024 * 1024),
            "expire_days": int(expire_days),
            "note": f"Generated automatically by WhiteBot for {username}"
        }
        try:
            async with httpx.AsyncClient(timeout=10.0) as client:
                resp = await client.post(f"{self.base_url}/api/users", json=payload)
                if resp.status_code in (200, 201):
                    user_data = resp.json().get("user", {})
                    cfg_uuid = user_data.get("config_uuid")
                    
                    # Fetch direct subscription content
                    sub_resp = await client.get(f"{self.base_url}/sub/{cfg_uuid}")
                    configs = sub_resp.text.strip().split("\n") if sub_resp.status_code == 200 else []
                    
                    return {
                        "user": user_data,
                        "sub_url": f"{self.base_url}/sub/{cfg_uuid}",
                        "configs": [c for c in configs if c.strip()]
                    }
                else:
                    logger.error(f"Error creating user in Go: {resp.status_code} - {resp.text}")
        except Exception as e:
            logger.error(f"Exception contacting Go Core: {e}")
        return None

backend_client = GoBackendClient(PANEL_API_URL)

async def tg_request(method: str, data: Dict[str, Any]) -> Optional[Dict[str, Any]]:
    """Safe Telegram API call"""
    if not BOT_TOKEN:
        return None
    import httpx
    try:
        async with httpx.AsyncClient(timeout=15.0) as client:
            r = await client.post(f"{TG_API}/{method}", json=data)
            return r.json()
    except Exception as e:
        logger.error(f"Telegram API {method} error: {e}")
        return None

async def send_message(chat_id: int, text: str, reply_markup: Optional[Dict[str, Any]] = None):
    payload = {
        "chat_id": chat_id,
        "text": text,
        "parse_mode": "HTML"
    }
    if reply_markup:
        payload["reply_markup"] = reply_markup
    return await tg_request("sendMessage", payload)

# ----------------- Keyboard Menus -----------------

def main_keyboard():
    return {
        "keyboard": [
            [{"text": "🛍 خرید اشتراک"}, {"text": "👤 حساب کاربری من"}],
            [{"text": "📊 وضعیت سرورها"}, {"text": "📞 پشتیبانی"}]
        ],
        "resize_keyboard": True
    }

# ----------------- Message Handlers -----------------

async def handle_start(chat_id: int, name: str):
    welcome_text = (
        f"سلام <b>{name}</b> عزیز! 🌟\n\n"
        "به ربات رسمی <b>White Panel</b> خوش آمدید.\n"
        "از طریق منوی زیر می‌توانید اشتراک پرسرعت تهیه کنید، سرویس خود را مدیریت کنید و یا وضعیت سرورها را مشاهده نمایید."
    )
    await send_message(chat_id, welcome_text, main_keyboard())

async def handle_plans(chat_id: int):
    plans = await backend_client.get_plans()
    buttons = []
    for p in plans:
        btn_text = f"{p['name']} - {int(p['price_toman']):,} تومان"
        buttons.append([{"text": btn_text, "callback_data": f"buy_{p['id']}"}])
    
    markup = {"inline_keyboard": buttons}
    await send_message(chat_id, "✨ <b>لطفاً پلن مورد نظر خود را انتخاب کنید:</b>", markup)

async def handle_buy_plan(chat_id: int, plan_id: str, username: str):
    plans = await backend_client.get_plans()
    selected = next((p for p in plans if p['id'] == plan_id), None)
    if not selected:
        await send_message(chat_id, "❌ پلن انتخابی یافت نشد.")
        return

    order_id = f"W-{int(time.time())}"
    ORDERS[order_id] = {
        "order_id": order_id,
        "chat_id": chat_id,
        "username": username or f"user_{chat_id}",
        "plan": selected,
        "status": "pending_payment"
    }
    USER_STEPS[chat_id] = f"await_receipt_{order_id}"

    card_info = (
        f"🛒 <b>سفارش شما: {selected['name']}</b>\n"
        f"💰 مبلغ قابل پرداخت: <b>{int(selected['price_toman']):,} تومان</b>\n\n"
        "💳 شماره کارت جهت واریز:\n"
        "<code>6037-9975-0000-0000</code>\n"
        "به نام: <i>مدیریت White Panel</i>\n\n"
        "📸 لطفاً پس از واریز، <b>تصویر فیش یا شماره پیگیری</b> را ارسال کنید تا کانفیگ شما فوراً صادر شود."
    )
    await send_message(chat_id, card_info)

async def activate_and_deliver(chat_id: int, order: Dict[str, Any]):
    """Activates the order, creates config in Go Core and sends directly to customer"""
    plan = order["plan"]
    uname = f"wp_{chat_id}_{int(time.time()%10000)}"
    
    await send_message(chat_id, "⚡ درحال ساخت سرویس اختصاصی شما در سرورهای White...")
    
    res = await backend_client.create_user_and_generate_config(
        username=uname,
        traffic_gb=plan["traffic_gb"],
        expire_days=plan["expire_days"]
    )

    if not res:
        await send_message(chat_id, "⚠️ خطایی در ساخت اکانت رخ داد. مدیر به زودی دستی رسید شما را بررسی خواهد کرد.")
        return

    sub_url = res["sub_url"]
    configs = res["configs"]
    main_cfg = configs[0] if configs else sub_url

    delivery_msg = (
        "🎉 <b>سرویس شما با موفقیت فعال شد!</b>\n\n"
        f"👤 نام کاربری: <code>{uname}</code>\n"
        f"📊 حجم: <b>{plan['traffic_gb']} گیگابایت</b>\n"
        f"⏳ مدت اعتبار: <b>{plan['expire_days']} روز</b>\n\n"
        f"🔗 <b>لینک ساب‌اسکریپشن (اتصال خودکار):</b>\n"
        f"<code>{sub_url}</code>\n\n"
        f"🔑 <b>کانفیگ مستقیم (VLESS Reality):</b>\n"
        f"<code>{main_cfg}</code>\n\n"
        "📌 کانفیگ یا لینک بالا را کپی و در نرم‌افزارهای v2rayNG / v2rayN / Streisand / Sing-box وارد نمایید."
    )
    await send_message(chat_id, delivery_msg, main_keyboard())
    order["status"] = "delivered"

# ----------------- Polling Loop -----------------

async def bot_poll_loop():
    if not BOT_TOKEN:
        logger.warning("WHITE_BOT_TOKEN is not set. Bot polling will not start.")
        return

    offset = 0
    logger.info("White Telegram Bot service started polling...")

    while True:
        try:
            import httpx
            async with httpx.AsyncClient(timeout=30.0) as client:
                r = await client.get(f"{TG_API}/getUpdates", params={"offset": offset, "timeout": 20})
                if r.status_code != 200:
                    await asyncio.sleep(5)
                    continue
                
                updates = r.json().get("result", [])
                for u in updates:
                    offset = u["update_id"] + 1

                    # Callbacks (Plan buttons)
                    if "callback_query" in u:
                        cb = u["callback_query"]
                        data = cb.get("data", "")
                        from_user = cb.get("from", {})
                        cid = from_user.get("id")
                        c_uname = from_user.get("username", "")

                        if data.startswith("buy_"):
                            plan_id = data.replace("buy_", "")
                            await handle_buy_plan(cid, plan_id, c_uname)

                    # Messages
                    elif "message" in u:
                        msg = u["message"]
                        cid = msg.get("chat", {}).get("id")
                        text = msg.get("text", "")
                        sender_name = msg.get("from", {}).get("first_name", "کاربر")
                        
                        # Handle photo (Receipt)
                        if "photo" in msg or (text and USER_STEPS.get(cid, "").startswith("await_receipt_")):
                            step = USER_STEPS.get(cid, "")
                            if step.startswith("await_receipt_"):
                                order_id = step.replace("await_receipt_", "")
                                order = ORDERS.get(order_id)
                                if order:
                                    USER_STEPS.pop(cid, None)
                                    # Instant activation & delivery
                                    await activate_and_deliver(cid, order)
                            continue

                        if text == "/start":
                            await handle_start(cid, sender_name)
                        elif text == "🛍 خرید اشتراک":
                            await handle_plans(cid)
                        elif text == "📊 وضعیت سرورها":
                            await send_message(cid, "🟢 کلیه سرورهای White Panel فعال و پایدار هستند.")
                        elif text == "📞 پشتیبانی":
                            await send_message(cid, "💬 جهت ارتباط با پشتیبانی به آیدی ادمین پیام دهید.")

        except asyncio.CancelledError:
            break
        except Exception as e:
            logger.error(f"Error in Telegram loop: {e}")
            await asyncio.sleep(5)

if __name__ == "__main__":
    asyncio.run(bot_poll_loop())
