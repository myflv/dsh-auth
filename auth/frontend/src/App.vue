<script setup>
import { ref } from 'vue'

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  if (loading.value) return
  loading.value = true
  error.value = ''
  try {
    const fd = new FormData()
    fd.append('username', username.value.trim())
    fd.append('password', password.value)
    fd.append('csrf', window.__CSRF__ || '')
    const res = await fetch('login', { method: 'POST', body: fd, redirect: 'manual' })
    // 登录成功：服务端 302 + 种会话 cookie，跳进应用
    if (res.type === 'opaqueredirect' || res.status === 302 || res.redirected) {
      window.location.href = '/'
      return
    }
    error.value = (await res.text()) || '登录失败，请重试'
  } catch {
    error.value = '网络异常，请稍后重试'
  }
  loading.value = false
}
</script>

<template>
  <div class="page">
    <div class="card">
      <div class="head">
        <div class="logo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">
            <path d="M21 11.5a8.38 8.38 0 0 1-.9 3.8 8.5 8.5 0 0 1-7.6 4.7 8.38 8.38 0 0 1-3.8-.9L3 21l1.9-5.7a8.38 8.38 0 0 1-.9-3.8 8.5 8.5 0 0 1 4.7-7.6 8.38 8.38 0 0 1 3.8-.9h.5a8.48 8.48 0 0 1 8 8v.5z"/>
          </svg>
        </div>
        <h1>dsh web</h1>
        <p class="sub">登录以继续访问</p>
      </div>

      <form @submit.prevent="submit">
        <label class="field-label" for="username">用户名</label>
        <input id="username" class="field" v-model="username" autocomplete="username" required placeholder="请输入用户名" autofocus />

        <label class="field-label" for="password">密码</label>
        <input id="password" class="field" v-model="password" type="password" autocomplete="current-password" required placeholder="请输入密码" />

        <transition name="fade">
          <p v-if="error" class="error" role="alert">{{ error }}</p>
        </transition>

        <button class="btn" type="submit" :disabled="loading">
          <span v-if="!loading">登录</span>
          <span v-else class="spinner" aria-hidden="true"></span>
        </button>
      </form>
    </div>

    <p class="foot">goauth-proxy · 认证入口每次启动自动更换</p>
  </div>
</template>

<style>
/* 设计 tokens 取自 deepseek-harness 的 design-platform.css（light + dark）：
   packages/client/ui-theme/src/styles/design-platform.css */
* { margin: 0; padding: 0; box-sizing: border-box; }

body {
  --bg: rgb(249, 250, 251);                    /* neutral-bluish-50 */
  --card: rgb(255, 255, 255);                  /* neutral-bluish-00 */
  --card-border: rgba(0, 0, 0, 0.1);           /* border-l2 */
  --card-shadow: rgba(15, 17, 21, 0.04);
  --text-1: rgb(15, 17, 21);                   /* neutral-bluish-1000 */
  --text-2: rgb(97, 102, 107);                 /* neutral-bluish-700 */
  --text-3: rgb(173, 178, 184);                /* neutral-bluish-400 */
  --input-bg: rgb(249, 250, 251);              /* specific-login-input */
  --input-hover: rgb(241, 243, 245);           /* neutral-bluish-75 */
  --brand: rgb(65, 118, 230);                  /* deepseek-500 */
  --brand-hover: rgb(103, 158, 254);           /* deepseek-400 */
  --btn: rgb(15, 17, 21);                      /* brand-primary（深色主按钮） */
  --btn-hover: rgb(67, 69, 74);                /* neutral-bluish-750 */
  --error: rgb(236, 19, 19);                   /* red-600 */
  --hover: rgba(38, 49, 72, 0.06);             /* interactive-bg-hover */
  --ease: cubic-bezier(0.4, 0, 0.2, 1);        /* ds-ease-in-out */
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC',
    'Hiragino Sans GB', 'Microsoft YaHei', 'Helvetica Neue', Helvetica, Arial, sans-serif;
  color: var(--text-1);
}

@media (prefers-color-scheme: dark) {
  body {
    --bg: rgb(21, 21, 23);                     /* neutral-bluish-950 */
    --card: rgb(35, 35, 36);                   /* neutral-bluish-875 */
    --card-border: rgba(255, 255, 255, 0.1);
    --card-shadow: rgba(0, 0, 0, 0.5);
    --text-1: rgb(235, 238, 242);              /* neutral-bluish-100 */
    --text-2: rgb(151, 157, 166);              /* neutral-bluish-400 */
    --text-3: rgb(101, 103, 107);              /* neutral-bluish-550 */
    --input-bg: rgb(44, 44, 46);               /* neutral-bluish-850 */
    --input-hover: rgb(53, 54, 56);            /* neutral-bluish-800 */
    --btn: rgb(249, 250, 251);                 /* 深色模式主按钮反转为浅色 */
    --btn-hover: rgb(225, 229, 238);           /* neutral-bluish-200 */
    --hover: rgba(255, 255, 255, 0.06);
  }
}

.page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 20px;
  background: var(--bg);
  padding: 24px;
  transition: background 0.2s var(--ease);
}

.card {
  width: 360px;
  max-width: calc(100vw - 48px);
  background: var(--card);
  border: 1px solid var(--card-border);
  border-radius: 12px;
  box-shadow: 0 12px 32px var(--card-shadow);
  padding: 32px 28px 28px;
  animation: rise 0.3s var(--ease);
}

@keyframes rise {
  from { opacity: 0; transform: translateY(8px); }
  to { opacity: 1; transform: none; }
}

.head { text-align: center; margin-bottom: 24px; }

.logo {
  width: 44px;
  height: 44px;
  margin: 0 auto 14px;
  border-radius: 10px;
  background: var(--brand);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 4px 12px rgba(65, 118, 230, 0.25);
}
.logo svg { width: 22px; height: 22px; }

h1 { font-size: 17px; font-weight: 600; letter-spacing: 0.2px; }
.sub { font-size: 13px; color: var(--text-2); margin-top: 4px; }

.field-label {
  display: block;
  font-size: 13px;
  color: var(--text-2);
  margin: 14px 0 6px;
}

.field {
  width: 100%;
  height: 38px;
  padding: 0 12px;
  font-size: 14px;
  color: var(--text-1);
  background: var(--input-bg);
  border: 1px solid transparent;
  border-radius: 8px;
  outline: none;
  transition: all 0.2s var(--ease);
}
.field::placeholder { color: var(--text-3); }
.field:hover { background: var(--input-hover); }
.field:focus {
  background: var(--card);
  border-color: var(--brand);
  box-shadow: 0 0 0 3px rgba(65, 118, 230, 0.12);
}

.error {
  font-size: 13px;
  color: var(--error);
  margin: 10px 0 2px;
}

.btn {
  width: 100%;
  height: 38px;
  margin-top: 18px;
  border: none;
  border-radius: 8px;
  background: var(--btn);
  color: var(--card);
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s var(--ease);
}
.btn:hover { background: var(--btn-hover); }
.btn:active { opacity: 0.85; }
.btn:disabled { opacity: 0.55; cursor: default; }

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top-color: currentColor;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }

.foot { font-size: 12px; color: var(--text-3); }

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s var(--ease), transform 0.2s var(--ease); }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(-2px); }
</style>
