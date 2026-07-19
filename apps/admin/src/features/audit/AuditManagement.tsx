import { useEffect, useState } from 'react';
import { request } from '../../api';
import { Messages, PageHeader } from '../../components/Common';
import type { AuditRecord, ListResponse } from '../../types';

export function AuditManagement() {
  const [items, setItems] = useState<AuditRecord[]>([]);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  async function load() {
    try { const x = await request<ListResponse<AuditRecord>>('/api/admin/audits?status=pending'); setItems(x.items); }
    catch (e) { setError(e instanceof Error ? e.message : '加载失败'); }
  }
  useEffect(() => { void load(); }, []);
  async function review(id: string, action: 'approve' | 'reject') {
    const comment = window.prompt(action === 'approve' ? '审核意见（可选）' : '请输入驳回原因') ?? '';
    if (action === 'reject' && !comment) return;
    try {
      await request(`/api/admin/audits/${id}/${action}`, { method: 'POST', body: JSON.stringify({ comment }) });
      setNotice(action === 'approve' ? '已审核通过并发布' : '已驳回为草稿'); await load();
    } catch (e) { setError(e instanceof Error ? e.message : '审核失败'); }
  }
  return <section className="content"><PageHeader title="数据审核" description="审核菌种发布申请；通过后才会在用户端展示。" /><Messages error={error} notice={notice} /><div className="panel auditList">{items.map((x) => <article key={x.id}><div><strong>{x.entityName}</strong><small>提交于 {new Date(x.submittedAt).toLocaleString()}</small></div><div className="actions"><button onClick={() => review(x.id, 'approve')}>通过并发布</button><button className="danger" onClick={() => review(x.id, 'reject')}>驳回</button></div></article>)}{!items.length && <div className="empty">没有待审核数据。</div>}</div></section>;
}
