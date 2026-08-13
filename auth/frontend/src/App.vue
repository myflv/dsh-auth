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
    // 点击登录这一刻才向服务端要一次性 token（永远新鲜，不存在过期问题）
    const csrfRes = await fetch('csrf')
    const { token } = await csrfRes.json()

    const fd = new FormData()
    fd.append('username', username.value.trim())
    fd.append('password', password.value)
    fd.append('csrf', token)
    const res = await fetch('login', { method: 'POST', body: fd })
    const data = await res.json().catch(() => ({}))

    if (res.ok && data.session) {
      // 前端直接种会话 cookie（不依赖 Set-Cookie 处理，浏览器行为差异全部绕开）
      const secure = window.location.protocol === 'https:' ? '; Secure' : ''
      document.cookie = `dsh_session=${data.session}; Path=/; Max-Age=43200; SameSite=Strict${secure}`
      window.location.href = '/'
      return
    }
    error.value = data.error || '登录失败，请重试'
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
          <!-- dsh 官方鲸鱼 wordmark（取自 deepseek-harness 前端资源） -->
          <svg viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <g clip-path="url(#dsh-wordmark-whale-clip)">
              <path d="M23.0584 4.95203C22.8129 4.83203 22.7074 5.06103 22.5639 5.17704C22.5149 5.21454 22.4734 5.26354 22.4319 5.30854C22.0734 5.69155 21.6543 5.94306 21.1073 5.91306C20.3073 5.86806 19.6243 6.11957 19.0203 6.73158C18.8918 5.97706 18.4652 5.52655 17.8162 5.23754C17.4767 5.08753 17.1332 4.93703 16.8952 4.61052C16.7292 4.37801 16.6837 4.11901 16.6007 3.8635C16.5477 3.70949 16.4952 3.55199 16.3177 3.52549C16.1252 3.49549 16.0497 3.65699 15.9742 3.792C15.6722 4.34401 15.5552 4.95203 15.5667 5.56805C15.5932 6.95359 16.1782 8.05712 17.3407 8.84215C17.4727 8.93215 17.5067 9.02215 17.4652 9.15366C17.3857 9.42416 17.2917 9.68667 17.2087 9.95718C17.1557 10.1297 17.0767 10.1677 16.8917 10.0922C16.2537 9.82568 15.7027 9.43117 15.2156 8.95465C14.3891 8.15513 13.6416 7.2726 12.7096 6.58158C12.4906 6.42007 12.2716 6.27007 12.045 6.12707C11.094 5.20354 12.1696 4.44502 12.4186 4.35501C12.6791 4.26101 12.5091 3.938 11.6675 3.942C10.826 3.9455 10.056 4.22751 9.07446 4.60302C8.93096 4.65952 8.77995 4.70052 8.62545 4.73452C7.73492 4.56552 6.80989 4.52802 5.84386 4.63702C4.02481 4.83953 2.57177 5.69955 1.50373 7.1676C0.220694 8.93215 -0.0813148 10.9372 0.288196 13.0283C0.676708 15.2323 1.80174 17.0569 3.53029 18.4834C5.32285 19.9625 7.38741 20.6875 9.74298 20.5485C11.1735 20.466 12.7661 20.2745 14.5626 18.7539C15.0156 18.9795 15.4912 19.0695 16.2797 19.137C16.8872 19.1935 17.4722 19.107 17.9252 19.013C18.6347 18.8629 18.5857 18.2059 18.3292 18.0854C16.2497 17.1169 16.7062 17.5109 16.2912 17.1919C17.3477 15.9419 18.9618 13.7198 19.4598 10.6942C19.5088 10.3602 19.5713 9.88968 19.5638 9.61917C19.5598 9.45417 19.5978 9.39016 19.7863 9.37116C20.3073 9.31116 20.8128 9.16866 21.2773 8.91315C22.6249 8.17713 23.1684 6.96809 23.2964 5.51905C23.3154 5.29754 23.2924 5.06853 23.0584 4.95203ZM11.3165 17.9954C9.30097 16.4109 8.32344 15.8894 7.91992 15.9119C7.54241 15.9344 7.61042 16.3664 7.69342 16.6479C7.78042 16.9259 7.89342 17.1174 8.05193 17.3614C8.16143 17.5229 8.23694 17.7629 7.94243 17.9434C7.29341 18.3449 6.16487 17.8084 6.11187 17.7819C4.79833 17.0084 3.7003 15.9874 2.92628 14.5908C2.17875 13.2468 1.74474 11.8047 1.67324 10.2657C1.65424 9.89418 1.76374 9.76267 2.13375 9.69517C2.62077 9.60517 3.12278 9.58617 3.6093 9.65767C5.66636 9.95818 7.41741 10.8777 8.88545 12.3348C9.72348 13.1643 10.3575 14.1558 11.0105 15.1243C11.705 16.1529 12.4521 17.1329 13.4036 17.9364C13.7396 18.2179 14.0076 18.4319 14.2641 18.5899C13.4906 18.6764 12.1996 18.6949 11.3165 17.9964V17.9954ZM12.2826 11.7817C12.2826 11.6167 12.4146 11.4852 12.5806 11.4852C12.6181 11.4852 12.6521 11.4927 12.6826 11.5037C12.7241 11.5187 12.7621 11.5412 12.7921 11.5752C12.8451 11.6277 12.8751 11.7027 12.8751 11.7817C12.8751 11.9467 12.7431 12.0782 12.5771 12.0782C12.4111 12.0782 12.2826 11.9467 12.2826 11.7817ZM15.2831 13.3208C15.0906 13.3998 14.8981 13.4673 14.7131 13.4748C14.4261 13.4898 14.1131 13.3733 13.9431 13.2308C13.6791 13.0093 13.4901 12.8853 13.4111 12.4988C13.3771 12.3338 13.3961 12.0782 13.4261 11.9317C13.4941 11.6162 13.4186 11.4137 13.1961 11.2297C13.0151 11.0797 12.7846 11.0382 12.5316 11.0382C12.4371 11.0382 12.3506 10.9967 12.2861 10.9632C12.1806 10.9107 12.0936 10.7792 12.1766 10.6177C12.2031 10.5652 12.3316 10.4377 12.3616 10.4152C12.7051 10.2197 13.1011 10.2837 13.4676 10.4302C13.8071 10.5692 14.0641 10.8242 14.4336 11.1847C14.8111 11.6202 14.8791 11.7402 15.0941 12.0672C15.2641 12.3228 15.4186 12.5853 15.5247 12.8858C15.5887 13.0733 15.5057 13.2268 15.2831 13.3208Z" fill="currentColor"/>
            </g>
            <defs>
              <clipPath id="dsh-wordmark-whale-clip">
                <rect width="23.16" height="17.0435" fill="white" transform="translate(0.141602 3.52185)"></rect>
              </clipPath>
            </defs>
          </svg>
        </div>
        <h1>dsh web</h1>
        <p class="sub">登录以继续访问</p>
      </div>

      <!-- 登录全程 JS 驱动：token 点击时获取、错误卡片内展示、cookie 前端种 -->
      <form @submit.prevent="submit">
        <label class="field-label" for="username">用户名</label>
        <input id="username" class="field" name="username" v-model="username" autocomplete="username" required placeholder="请输入用户名" autofocus />

        <label class="field-label" for="password">密码</label>
        <input id="password" class="field" name="password" v-model="password" type="password" autocomplete="current-password" required placeholder="请输入密码" />

        <p v-if="error" class="error" role="alert">{{ error }}</p>

        <button class="btn" type="submit" :disabled="loading">
          <span v-if="!loading">登录</span>
          <span v-else class="spinner" aria-hidden="true"></span>
        </button>
      </form>
    </div>
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
  align-items: center;
  justify-content: center;
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
  margin: 0 auto 12px;
  color: var(--brand);
  display: flex;
  align-items: center;
  justify-content: center;
}
.logo svg { width: 38px; height: 38px; }

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

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s var(--ease), transform 0.2s var(--ease); }
.fade-enter-from, .fade-leave-to { opacity: 0; transform: translateY(-2px); }
</style>
