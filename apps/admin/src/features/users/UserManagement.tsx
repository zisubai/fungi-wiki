import { FormEvent, useEffect, useState } from 'react';
import { request } from '../../api';
import { Messages, PageHeader } from '../../components/Common';
import type { AuthUser, ListResponse } from '../../types';

const emptyForm = { email: '', password: '', displayName: '', role: 'operator' as AuthUser['role'] };
const roleLabels = { operator: '运营', expert: '专家', admin: '管理员' };

export function UserManagement() {
  const [items, setItems] = useState<AuthUser[]>([]);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [form, setForm] = useState(emptyForm);
  async function load() {
    try { const data = await request<ListResponse<AuthUser>>('/api/admin/users'); setItems(data.items); }
    catch (e) { setError(e instanceof Error ? e.message : '账号加载失败'); }
  }
  useEffect(() => { void load(); }, []);
  async function submit(event: FormEvent) {
    event.preventDefault(); setError('');
    try { await request('/api/admin/users', { method: 'POST', body: JSON.stringify(form) }); setNotice('账号已创建'); setForm(emptyForm); await load(); }
    catch (e) { setError(e instanceof Error ? e.message : '创建失败'); }
  }
  return <section className="content"><PageHeader title="账号管理" description="创建运营、专家和管理员账号，并按角色控制操作权限。" /><Messages error={error} notice={notice} /><section className="mainGrid"><div className="panel tablePanel"><div className="panelTitle"><h3>账号列表</h3><span>{items.length} 个</span></div><div className="tableWrap"><table><thead><tr><th>姓名</th><th>邮箱</th><th>角色</th><th>状态</th></tr></thead><tbody>{items.map((user) => <tr key={user.id}><td><strong>{user.displayName}</strong></td><td>{user.email}</td><td>{roleLabels[user.role]}</td><td>{user.status ?? 'active'}</td></tr>)}</tbody></table></div></div><form className="panel formPanel" onSubmit={submit}><div className="panelTitle"><h3>创建账号</h3></div><label><span>显示名称 *</span><input required value={form.displayName} onChange={(e) => setForm({ ...form, displayName: e.target.value })} /></label><label><span>邮箱 *</span><input type="email" required value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></label><label><span>初始密码 *</span><input type="password" minLength={8} required value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} placeholder="至少 8 位" /></label><label><span>角色</span><select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value as AuthUser['role'] })}><option value="operator">运营：维护数据</option><option value="expert">专家：审核数据</option><option value="admin">管理员：全部权限</option></select></label><button className="primary full">创建账号</button></form></section></section>;
}
