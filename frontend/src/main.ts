import './style.css';

const BACKEND = 'http://localhost:3000';

type View = 'connect' | 'register' | 'registered';

let currentView: View = 'connect';

function render() {
  const app = document.querySelector<HTMLDivElement>('#app')!;
  app.innerHTML = views[currentView]();
  bindEvents();
}

const views: Record<View, () => string> = {
  connect: () => `
    <div class="card">
      <div class="logo-mark"></div>
      <h1>RAM<span class="accent">VPN</span></h1>
      <p class="subtitle">Zero-persistence. Every reboot is a clean slate.</p>
      <div class="form">
        <input
          id="user-id-input"
          type="text"
          inputmode="numeric"
          maxlength="16"
          placeholder="Enter your 16-digit ID"
          autocomplete="off"
          spellcheck="false"
        />
        <button id="connect-btn" class="btn-primary">Connect</button>
      </div>
      <p class="status" id="status"></p>
      <button id="go-register" class="btn-ghost">No account? Register</button>
    </div>
  `,

  register: () => `
    <div class="card">
      <div class="logo-mark"></div>
      <h1>RAM<span class="accent">VPN</span></h1>
      <p class="subtitle">Create an account. No email. No name. Just a number.</p>
      <button id="register-btn" class="btn-primary">Generate Account</button>
      <p class="status" id="status"></p>
      <button id="go-connect" class="btn-ghost">Already have an ID? Connect</button>
    </div>
  `,

  registered: () => `
    <div class="card">
      <div class="logo-mark success"></div>
      <h1>Account <span class="accent">Created</span></h1>
      <p class="subtitle">This is your ID. It is your only credential — save it now.</p>
      <div class="id-display" id="id-display"></div>
      <button id="copy-btn" class="btn-secondary">Copy ID</button>
      <p class="hint">There is no recovery. If you lose this, your account is gone.</p>
      <button id="go-connect-after" class="btn-primary" style="margin-top:8px">Connect Now</button>
    </div>
  `,
};

function setStatus(msg: string, error = false) {
  const el = document.querySelector<HTMLParagraphElement>('#status');
  if (!el) return;
  el.textContent = msg;
  el.className = 'status ' + (error ? 'error' : 'info');
}

function bindEvents() {
  document.querySelector('#go-register')?.addEventListener('click', () => {
    currentView = 'register';
    render();
  });

  document.querySelector('#go-connect')?.addEventListener('click', () => {
    currentView = 'connect';
    render();
  });

  document.querySelector('#go-connect-after')?.addEventListener('click', () => {
    currentView = 'connect';
    render();
  });

  document.querySelector('#connect-btn')?.addEventListener('click', async () => {
    const input = document.querySelector<HTMLInputElement>('#user-id-input')!;
    const id = input.value.trim();
    if (!/^\d{16}$/.test(id)) {
      setStatus('ID must be exactly 16 digits.', true);
      return;
    }
    setStatus('Contacting controller…');
    // TODO: POST to controller node once built
    setStatus('Controller not yet implemented.', true);
  });

  document.querySelector('#register-btn')?.addEventListener('click', async () => {
    const btn = document.querySelector<HTMLButtonElement>('#register-btn')!;
    btn.disabled = true;
    setStatus('Registering…');

    try {
      const res = await fetch(`${BACKEND}/auth/register`, { method: 'POST' });
      if (!res.ok) throw new Error(`Server error ${res.status}`);
      const data: { user_id: string } = await res.json();

      currentView = 'registered';
      render();

      document.querySelector<HTMLDivElement>('#id-display')!.textContent = data.user_id;

      document.querySelector('#copy-btn')?.addEventListener('click', () => {
        navigator.clipboard.writeText(data.user_id);
        const btn = document.querySelector<HTMLButtonElement>('#copy-btn')!;
        btn.textContent = 'Copied!';
        setTimeout(() => (btn.textContent = 'Copy ID'), 2000);
      });
    } catch (e) {
      setStatus('Registration failed. Try again.', true);
      btn.disabled = false;
    }
  });
}

render();
