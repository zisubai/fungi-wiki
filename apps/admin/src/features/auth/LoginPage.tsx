import { FormEvent, useState } from 'react';
import { authTokenKey, request } from '../../api';
import type { AuthUser, LoginResponse } from '../../types';

export function LoginPage({ onLogin }: { onLogin: (user: AuthUser) => void }) {
  const [email, setEmail] = useState('admin@fungi.local');
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  async function login(event: FormEvent) {
    event.preventDefault(); setLoading(true); setError('');
    try {
      const data = await request<LoginResponse>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) });
      localStorage.setItem(authTokenKey, data.token); onLogin(data.user);
    } catch (cause) { setError(cause instanceof Error ? cause.message : '登录失败'); }
    finally { setLoading(false); }
  }

  return <main className="loginPage"><form className="loginCard" onSubmit={login}><p className="eyebrow">Fungi Wiki Admin</p><h1>登录运营端</h1><p>使用运营、专家或管理员账号登录。</p>{error && <div className="message error">{error}</div>}<label><span>邮箱</span><input type="email" required value={email} onChange={(event) => setEmail(event.target.value)} /></label><label><span>密码</span><input type="password" required value={password} onChange={(event) => setPassword(event.target.value)} /></label><button className="primary full" disabled={loading}>{loading ? '登录中…' : '登录'}</button></form></main>;
}
