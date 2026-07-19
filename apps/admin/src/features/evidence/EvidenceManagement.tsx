import { FormEvent, useEffect, useState } from 'react';
import { request } from '../../api';
import { Messages, PageHeader } from '../../components/Common';
import type { Evidence, ListResponse, Species } from '../../types';

const emptyForm = { title: '', authors: '', journal: '', publicationYear: '', doi: '', pmid: '', sourceUrl: '', conclusion: '', evidenceLevel: 'medium', evidenceScore: 50 };

export function EvidenceManagement() {
  const [species, setSpecies] = useState<Species[]>([]);
  const [selected, setSelected] = useState('');
  const [items, setItems] = useState<Evidence[]>([]);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [form, setForm] = useState(emptyForm);

  useEffect(() => {
    void request<ListResponse<Species>>('/api/admin/species?limit=100')
      .then((x) => { setSpecies(x.items); if (x.items[0]) setSelected(x.items[0].slug); })
      .catch((e) => setError(e instanceof Error ? e.message : '菌种加载失败'));
  }, []);

  async function load(id = selected) {
    if (!id) return;
    try { const x = await request<ListResponse<Evidence>>(`/api/admin/species/${id}/evidences`); setItems(x.items); }
    catch (e) { setError(e instanceof Error ? e.message : '加载失败'); }
  }
  useEffect(() => { void load(selected); }, [selected]);

  async function submit(event: FormEvent) {
    event.preventDefault();
    try {
      await request(`/api/admin/species/${selected}/evidences`, { method: 'POST', body: JSON.stringify({ ...form, publicationYear: form.publicationYear ? Number(form.publicationYear) : null }) });
      setNotice('文献证据已添加'); setForm(emptyForm); await load();
    } catch (e) { setError(e instanceof Error ? e.message : '保存失败'); }
  }
  async function remove(id: string) {
    if (!window.confirm('确认删除该证据关联？')) return;
    try { await request(`/api/admin/species/${selected}/evidences/${id}`, { method: 'DELETE' }); await load(); }
    catch (e) { setError(e instanceof Error ? e.message : '删除失败'); }
  }

  return <section className="content">
    <PageHeader title="文献证据" description="维护文献来源、实验结论和证据等级。" />
    <section className="toolbar compactToolbar"><select aria-label="关联菌种" value={selected} onChange={(e) => setSelected(e.target.value)}>{species.map((x) => <option key={x.id} value={x.slug}>{x.latinName}</option>)}</select><button onClick={() => load()}>刷新</button></section>
    <Messages error={error} notice={notice} />
    <section className="mainGrid"><div className="panel tablePanel"><div className="panelTitle"><h3>证据列表</h3><span>{items.length} 条</span></div><div className="evidenceList">{items.map((x) => <article key={x.id}><strong>{x.title}</strong><p>{x.conclusion}</p><small>{x.journal} {x.publicationYear ?? ''} · {x.evidenceLevel} · {x.evidenceScore} 分</small><button className="danger" onClick={() => remove(x.id)}>删除</button></article>)}{!items.length && <div className="empty">暂无文献证据。</div>}</div></div>
      <form className="panel formPanel" onSubmit={submit}><div className="panelTitle"><h3>新增文献证据</h3></div>
        <label><span>文献标题 *</span><input required value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} /></label>
        <label><span>作者</span><input value={form.authors} onChange={(e) => setForm({ ...form, authors: e.target.value })} /></label>
        <div className="twoCols"><label><span>期刊</span><input value={form.journal} onChange={(e) => setForm({ ...form, journal: e.target.value })} /></label><label><span>年份</span><input type="number" value={form.publicationYear} onChange={(e) => setForm({ ...form, publicationYear: e.target.value })} /></label></div>
        <div className="twoCols"><label><span>DOI</span><input value={form.doi} onChange={(e) => setForm({ ...form, doi: e.target.value })} /></label><label><span>PMID</span><input value={form.pmid} onChange={(e) => setForm({ ...form, pmid: e.target.value })} /></label></div>
        <label><span>来源链接</span><input type="url" value={form.sourceUrl} onChange={(e) => setForm({ ...form, sourceUrl: e.target.value })} /></label>
        <label><span>实验结论 *</span><textarea required rows={4} value={form.conclusion} onChange={(e) => setForm({ ...form, conclusion: e.target.value })} /></label>
        <div className="twoCols"><label><span>证据等级</span><select value={form.evidenceLevel} onChange={(e) => setForm({ ...form, evidenceLevel: e.target.value })}><option value="low">低</option><option value="medium">中</option><option value="high">高</option><option value="expert_verified">专家确认</option></select></label><label><span>证据分</span><input type="number" min="0" max="100" value={form.evidenceScore} onChange={(e) => setForm({ ...form, evidenceScore: Number(e.target.value) })} /></label></div>
        <button className="primary full">保存证据</button></form></section>
  </section>;
}
