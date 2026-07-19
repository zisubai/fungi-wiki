import { FormEvent, useEffect, useMemo, useState } from 'react';
import { request } from '../../api';
import { Messages, PageHeader } from '../../components/Common';
import type { FunctionTag, FunctionTagPayload, ListResponse } from '../../types';

const emptyTag: FunctionTagPayload = { parentId: '', name: '', code: '', description: '', sortOrder: 0 };

export function FunctionTagManagement() {
  const [items, setItems] = useState<FunctionTag[]>([]);
  const [query, setQuery] = useState('');
  const [form, setForm] = useState<FunctionTagPayload>(emptyTag);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const editingTitle = useMemo(() => editingId ? '编辑功能标签' : '新增功能标签', [editingId]);
  async function load() {
    setLoading(true); setError('');
    try { const params = new URLSearchParams(); if (query.trim()) params.set('q', query.trim()); const data = await request<ListResponse<FunctionTag>>(`/api/admin/function-tags?${params}`); setItems(data.items); }
    catch (e) { setError(e instanceof Error ? e.message : '加载标签失败'); }
    finally { setLoading(false); }
  }
  useEffect(() => { void load(); }, []);
  function reset() { setEditingId(null); setForm(emptyTag); setError(''); }
  async function submit(event: FormEvent) {
    event.preventDefault(); setSaving(true); setError(''); setNotice('');
    try { await request(editingId ? `/api/admin/function-tags/${editingId}` : '/api/admin/function-tags', { method: editingId ? 'PUT' : 'POST', body: JSON.stringify(form) }); setNotice(editingId ? '功能标签已更新' : '功能标签已新增'); reset(); await load(); }
    catch (e) { setError(e instanceof Error ? e.message : '保存标签失败'); }
    finally { setSaving(false); }
  }
  async function remove(tag: FunctionTag) {
    if (!window.confirm(`确认删除功能标签「${tag.name}」？`)) return;
    try { await request(`/api/admin/function-tags/${tag.code || tag.id}`, { method: 'DELETE' }); setNotice('功能标签已删除'); await load(); }
    catch (e) { setError(e instanceof Error ? e.message : '删除标签失败。请确认标签未被菌种使用。'); }
  }
  function edit(tag: FunctionTag) { setEditingId(tag.code || tag.id); setForm({ parentId: tag.parentId, name: tag.name, code: tag.code, description: tag.description, sortOrder: tag.sortOrder }); }
  return <section className="content"><PageHeader title="功能标签管理" description="维护促生、生防、固氮、解磷、降解、发酵等标准功能标签。" action={<button className="primary" onClick={reset}>新增标签</button>} /><section className="toolbar compactToolbar"><input value={query} onChange={(e) => setQuery(e.target.value)} onKeyDown={(e) => e.key === 'Enter' && load()} placeholder="搜索标签名称、编码或描述" /><button onClick={load} disabled={loading}>{loading ? '加载中...' : '查询'}</button></section><Messages error={error} notice={notice} /><section className="mainGrid"><div className="panel tablePanel"><div className="panelTitle"><h3>功能标签列表</h3><span>{items.length} 条</span></div><div className="tableWrap"><table><thead><tr><th>名称</th><th>编码</th><th>描述</th><th>排序</th><th>更新时间</th><th>操作</th></tr></thead><tbody>{items.map((tag) => <tr key={tag.id}><td><strong>{tag.name}</strong>{tag.parentId && <small>父级：{tag.parentId}</small>}</td><td><code>{tag.code}</code></td><td>{tag.description || '-'}</td><td>{tag.sortOrder}</td><td>{new Date(tag.updatedAt).toLocaleString()}</td><td className="actions"><button onClick={() => edit(tag)}>编辑</button><button className="danger" onClick={() => remove(tag)}>删除</button></td></tr>)}{!items.length && <tr><td colSpan={6} className="empty">暂无功能标签。</td></tr>}</tbody></table></div></div><form className="panel formPanel" onSubmit={submit}><div className="panelTitle"><h3>{editingTitle}</h3>{editingId && <button type="button" onClick={reset}>取消编辑</button>}</div><label><span>名称 *</span><input required value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></label><label><span>编码 *</span><input required value={form.code} onChange={(e) => setForm({ ...form, code: e.target.value })} /></label><label><span>父级标签 ID</span><input value={form.parentId} onChange={(e) => setForm({ ...form, parentId: e.target.value })} /></label><label><span>排序</span><input type="number" value={form.sortOrder} onChange={(e) => setForm({ ...form, sortOrder: Number(e.target.value) })} /></label><label><span>描述</span><textarea rows={5} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} /></label><button className="primary full" disabled={saving}>{saving ? '保存中...' : '保存标签'}</button></form></section></section>;
}
