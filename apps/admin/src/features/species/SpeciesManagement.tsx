import { FormEvent, useEffect, useMemo, useState } from 'react';
import { request } from '../../api';
import { Messages, PageHeader } from '../../components/Common';
import type { CultureCondition, FunctionTag, ListResponse, Species, SpeciesAlias, SpeciesFunction, SpeciesPayload, SpeciesStatus } from '../../types';

const emptySpecies: SpeciesPayload = { slug: '', latinName: '', chineseName: '', strainNumber: '', sourceEnvironment: '', safetyLevel: 'BSL-1', isModelOrganism: false, summary: '', status: 'draft' };
const emptyCulture: CultureCondition = { mediumName: '', temperatureMin: null, temperatureMax: null, phMin: null, phMax: null, oxygenRequirement: '', cultureTime: '', notes: '' };
const statusLabels: Record<SpeciesStatus, string> = { draft: '草稿', pending_review: '待审核', published: '已发布', archived: '已归档' };

function payloadOf(species: Species): SpeciesPayload {
  const { slug, latinName, chineseName, strainNumber, sourceEnvironment, safetyLevel, isModelOrganism, summary, status } = species;
  return { slug, latinName, chineseName, strainNumber, sourceEnvironment, safetyLevel, isModelOrganism, summary, status };
}

export function SpeciesManagement() {
  const [items, setItems] = useState<Species[]>([]);
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('');
  const [form, setForm] = useState<SpeciesPayload>(emptySpecies);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [functionTags, setFunctionTags] = useState<FunctionTag[]>([]);
  const [selectedTagIds, setSelectedTagIds] = useState<string[]>([]);
  const [culture, setCulture] = useState<CultureCondition>(emptyCulture);
  const [aliasesText, setAliasesText] = useState('');
  const editingTitle = useMemo(() => editingId ? '编辑菌种' : '新增菌种', [editingId]);

  async function loadSpecies() {
    setLoading(true); setError('');
    try {
      const params = new URLSearchParams();
      if (query.trim()) params.set('q', query.trim());
      if (status) params.set('status', status);
      const data = await request<ListResponse<Species>>(`/api/admin/species?${params}`);
      setItems(data.items);
    } catch (e) { setError(e instanceof Error ? e.message : '加载失败'); }
    finally { setLoading(false); }
  }

  useEffect(() => {
    void loadSpecies();
    void request<ListResponse<FunctionTag>>('/api/function-tags?limit=200').then((data) => setFunctionTags(data.items)).catch((e) => setError(e instanceof Error ? e.message : '功能标签加载失败'));
  }, []);

  function update<K extends keyof SpeciesPayload>(key: K, value: SpeciesPayload[K]) { setForm((current) => ({ ...current, [key]: value })); }
  function startCreate() { setEditingId(null); setForm(emptySpecies); setSelectedTagIds([]); setCulture(emptyCulture); setAliasesText(''); setError(''); }
  async function startEdit(species: Species) {
    const id = species.slug || species.id; setEditingId(id); setForm(payloadOf(species)); setError(''); setNotice('');
    try {
      const [functions, conditions, aliases] = await Promise.all([
        request<ListResponse<SpeciesFunction>>(`/api/admin/species/${id}/functions`),
        request<ListResponse<CultureCondition>>(`/api/admin/species/${id}/culture-conditions`),
        request<ListResponse<SpeciesAlias>>(`/api/admin/species/${id}/aliases`),
      ]);
      setSelectedTagIds(functions.items.map((item) => item.functionTagId));
      setCulture(conditions.items[0] ?? emptyCulture);
      setAliasesText(aliases.items.map((item) => item.name).join('\n'));
    } catch (e) { setError(e instanceof Error ? e.message : '菌种关联数据加载失败'); }
  }

  async function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault(); setSaving(true); setError(''); setNotice('');
    try {
      const saved = editingId
        ? await request<Species>(`/api/admin/species/${editingId}`, { method: 'PUT', body: JSON.stringify(form) })
        : await request<Species>('/api/admin/species', { method: 'POST', body: JSON.stringify(form) });
      const id = saved.slug || saved.id;
      await request(`/api/admin/species/${id}/functions`, { method: 'PUT', body: JSON.stringify({ items: selectedTagIds.map((functionTagId) => ({ functionTagId })) }) });
      const hasCulture = Object.values(culture).some((value) => value !== '' && value !== null);
      await request(`/api/admin/species/${id}/culture-conditions`, { method: 'PUT', body: JSON.stringify({ items: hasCulture ? [culture] : [] }) });
      const aliases = aliasesText.split(/\n|;|；/).map((name) => name.trim()).filter(Boolean);
      await request(`/api/admin/species/${id}/aliases`, { method: 'PUT', body: JSON.stringify({ items: aliases.map((name) => ({ name, type: 'synonym' })) }) });
      startCreate(); setNotice(editingId ? '菌种已更新' : '菌种已新增'); await loadSpecies();
    } catch (e) { setError(e instanceof Error ? e.message : '保存失败'); }
    finally { setSaving(false); }
  }

  async function archive(species: Species) {
    if (!window.confirm(`确认归档 ${species.latinName}？`)) return;
    try { await request(`/api/admin/species/${species.slug || species.id}`, { method: 'DELETE' }); setNotice('菌种已归档'); await loadSpecies(); }
    catch (e) { setError(e instanceof Error ? e.message : '归档失败'); }
  }
  async function submitForReview(species: Species) {
    if (!window.confirm(`提交 ${species.latinName} 进行发布审核？`)) return;
    try { await request(`/api/admin/audits/species/${species.slug || species.id}/submit`, { method: 'POST', body: '{}' }); setNotice('已提交审核'); await loadSpecies(); }
    catch (e) { setError(e instanceof Error ? e.message : '提交审核失败'); }
  }

  return <section className="content">
    <PageHeader title="菌种管理" description="维护微生物百科与功能菌数据库的基础菌种主数据。" action={<button className="primary" onClick={startCreate}>新增菌种</button>} />
    <section className="toolbar"><input value={query} onChange={(e) => setQuery(e.target.value)} onKeyDown={(e) => e.key === 'Enter' && loadSpecies()} placeholder="搜索 slug、拉丁名、中文名或摘要" /><select value={status} onChange={(e) => setStatus(e.target.value)}><option value="">全部状态</option><option value="draft">草稿</option><option value="pending_review">待审核</option><option value="published">已发布</option><option value="archived">已归档</option></select><button onClick={loadSpecies} disabled={loading}>{loading ? '加载中...' : '查询'}</button></section>
    <Messages error={error} notice={notice} />
    <section className="mainGrid"><div className="panel tablePanel"><div className="panelTitle"><h3>菌种列表</h3><span>{items.length} 条</span></div><div className="tableWrap"><table><thead><tr><th>拉丁名</th><th>中文名</th><th>安全等级</th><th>状态</th><th>更新时间</th><th>操作</th></tr></thead><tbody>{items.map((species) => <tr key={species.id}><td><strong>{species.latinName}</strong><small>{species.slug}</small></td><td>{species.chineseName || '-'}</td><td>{species.safetyLevel || '-'}</td><td><span className={`status ${species.status}`}>{statusLabels[species.status]}</span></td><td>{new Date(species.updatedAt).toLocaleString()}</td><td className="actions"><button onClick={() => void startEdit(species)}>编辑</button>{species.status === 'draft' && <button onClick={() => submitForReview(species)}>提交审核</button>}<button className="danger" onClick={() => archive(species)}>归档</button></td></tr>)}{!items.length && <tr><td colSpan={6} className="empty">暂无数据，请新增菌种或调整查询条件。</td></tr>}</tbody></table></div></div>
      <form className="panel formPanel" onSubmit={submit}><div className="panelTitle"><h3>{editingTitle}</h3>{editingId && <button type="button" onClick={startCreate}>取消编辑</button>}</div>
        <label><span>Slug *</span><input required value={form.slug} onChange={(e) => update('slug', e.target.value)} placeholder="bacillus-subtilis" /></label>
        <label><span>拉丁名 *</span><input required value={form.latinName} onChange={(e) => update('latinName', e.target.value)} placeholder="Bacillus subtilis" /></label>
        <label><span>中文名</span><input value={form.chineseName} onChange={(e) => update('chineseName', e.target.value)} /></label>
        <label><span>保藏/菌株编号</span><input value={form.strainNumber} onChange={(e) => update('strainNumber', e.target.value)} placeholder="ATCC / DSM / CGMCC" /></label>
        <label><span>来源环境</span><input value={form.sourceEnvironment} onChange={(e) => update('sourceEnvironment', e.target.value)} /></label>
        <div className="twoCols"><label><span>安全等级</span><input value={form.safetyLevel} onChange={(e) => update('safetyLevel', e.target.value)} /></label><label><span>状态</span><input value={statusLabels[form.status]} disabled /></label></div>
        <label className="checkbox"><input type="checkbox" checked={form.isModelOrganism} onChange={(e) => update('isModelOrganism', e.target.checked)} /><span>模式菌 / 常用底盘菌</span></label>
        <label><span>摘要</span><textarea rows={5} value={form.summary} onChange={(e) => update('summary', e.target.value)} /></label>
        <label><span>别名 / 同义词</span><textarea rows={3} value={aliasesText} onChange={(e) => setAliasesText(e.target.value)} placeholder="每行一个" /></label>
        <fieldset className="tagFieldset"><legend>功能标签</legend><div className="tagOptions">{functionTags.map((tag) => <label className="checkbox" key={tag.id}><input type="checkbox" checked={selectedTagIds.includes(tag.id)} onChange={() => setSelectedTagIds((ids) => ids.includes(tag.id) ? ids.filter((id) => id !== tag.id) : [...ids, tag.id])} /><span>{tag.name}</span></label>)}{!functionTags.length && <span className="empty">请先创建功能标签。</span>}</div></fieldset>
        <fieldset className="tagFieldset"><legend>主要培养条件</legend>
          <label><span>培养基</span><input value={culture.mediumName} onChange={(e) => setCulture({ ...culture, mediumName: e.target.value })} /></label>
          <div className="twoCols"><label><span>温度范围 °C</span><div className="inlineInputs"><input type="number" value={culture.temperatureMin ?? ''} onChange={(e) => setCulture({ ...culture, temperatureMin: e.target.value ? Number(e.target.value) : null })} /><input type="number" value={culture.temperatureMax ?? ''} onChange={(e) => setCulture({ ...culture, temperatureMax: e.target.value ? Number(e.target.value) : null })} /></div></label><label><span>pH 范围</span><div className="inlineInputs"><input type="number" step="0.1" value={culture.phMin ?? ''} onChange={(e) => setCulture({ ...culture, phMin: e.target.value ? Number(e.target.value) : null })} /><input type="number" step="0.1" value={culture.phMax ?? ''} onChange={(e) => setCulture({ ...culture, phMax: e.target.value ? Number(e.target.value) : null })} /></div></label></div>
          <label><span>氧需求</span><input value={culture.oxygenRequirement} onChange={(e) => setCulture({ ...culture, oxygenRequirement: e.target.value })} /></label>
          <label><span>培养时间</span><input value={culture.cultureTime} onChange={(e) => setCulture({ ...culture, cultureTime: e.target.value })} /></label>
          <label><span>备注</span><textarea rows={2} value={culture.notes} onChange={(e) => setCulture({ ...culture, notes: e.target.value })} /></label>
        </fieldset><button className="primary full" disabled={saving}>{saving ? '保存中...' : '保存菌种'}</button>
      </form>
    </section>
  </section>;
}
