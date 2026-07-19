import React, { FormEvent, useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

type SpeciesStatus = 'draft' | 'pending_review' | 'published' | 'archived';
type ActiveMenu = '菌种管理' | '功能标签' | '批量导入' | '文献证据' | '数据审核' | '搜索分析' | '账号管理' | '推荐质量';

type Species = {
  id: string;
  slug: string;
  latinName: string;
  chineseName: string;
  strainNumber: string;
  sourceEnvironment: string;
  safetyLevel: string;
  isModelOrganism: boolean;
  summary: string;
  status: SpeciesStatus;
  dataQualityScore: number;
  createdAt: string;
  updatedAt: string;
  publishedAt?: string;
};

type SpeciesPayload = {
  slug: string;
  latinName: string;
  chineseName: string;
  strainNumber: string;
  sourceEnvironment: string;
  safetyLevel: string;
  isModelOrganism: boolean;
  summary: string;
  status: SpeciesStatus;
};

type FunctionTag = {
  id: string;
  parentId: string;
  name: string;
  code: string;
  description: string;
  sortOrder: number;
  createdAt: string;
  updatedAt: string;
};

type FunctionTagPayload = {
  parentId: string;
  name: string;
  code: string;
  description: string;
  sortOrder: number;
};

type SpeciesFunction = {
  functionTagId: string;
  functionTagName: string;
};

type CultureCondition = {
  mediumName: string; temperatureMin: number | null; temperatureMax: number | null;
  phMin: number | null; phMax: number | null; oxygenRequirement: string; cultureTime: string; notes: string;
};

type Evidence = {
  id: string; title: string; authors: string; journal: string; publicationYear: number | null;
  doi: string; pmid: string; sourceUrl: string; conclusion: string; evidenceLevel: string; evidenceScore: number;
};

type AuditRecord = {
  id: string; entityId: string; entityName: string; action: string; status: string;
  comment: string; submittedAt: string; reviewedAt?: string;
};
type ImportRowResult = { rowNumber: number; slug: string; status: string; errors?: string[]; };
type ImportBatch = { id: string; sourceFilename: string; totalRows: number; successRows: number; failedRows: number; status: string; createdAt: string; rows?: ImportRowResult[]; };
type AuthUser = { id: string; email: string; displayName: string; role: 'operator' | 'expert' | 'admin'; status?: string; };
type LoginResponse = { token: string; expiresAt: string; user: AuthUser; };
type SearchReport = { days: number; totalSearches: number; noResultSearches: number; distinctQueries: number; popularQueries: { query: string; count: number; averageResults: number }[]; noResultQueries: { query: string; count: number }[]; };

type ListResponse<T> = {
  items: T[];
};

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
const menus: ActiveMenu[] = ['菌种管理', '功能标签', '批量导入', '文献证据', '数据审核', '搜索分析', '账号管理', '推荐质量'];

const emptySpeciesForm: SpeciesPayload = {
  slug: '',
  latinName: '',
  chineseName: '',
  strainNumber: '',
  sourceEnvironment: '',
  safetyLevel: 'BSL-1',
  isModelOrganism: false,
  summary: '',
  status: 'draft',
};

const emptyTagForm: FunctionTagPayload = {
  parentId: '',
  name: '',
  code: '',
  description: '',
  sortOrder: 0,
};

const emptyCultureCondition: CultureCondition = {
  mediumName: '', temperatureMin: null, temperatureMax: null, phMin: null, phMax: null,
  oxygenRequirement: '', cultureTime: '', notes: '',
};

function toSpeciesPayload(species: Species): SpeciesPayload {
  return {
    slug: species.slug,
    latinName: species.latinName,
    chineseName: species.chineseName,
    strainNumber: species.strainNumber,
    sourceEnvironment: species.sourceEnvironment,
    safetyLevel: species.safetyLevel,
    isModelOrganism: species.isModelOrganism,
    summary: species.summary,
    status: species.status,
  };
}

function toTagPayload(tag: FunctionTag): FunctionTagPayload {
  return {
    parentId: tag.parentId,
    name: tag.name,
    code: tag.code,
    description: tag.description,
    sortOrder: tag.sortOrder,
  };
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const isFormData = options?.body instanceof FormData;
  const token = localStorage.getItem('fungi_admin_token');
  const response = await fetch(`${apiBaseUrl}${path}`, {
    headers: { ...(!isFormData ? { 'Content-Type': 'application/json' } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}), ...options?.headers },
    ...options,
  });

  if (!response.ok) {
    if (response.status === 401 && path !== '/api/auth/login') { localStorage.removeItem('fungi_admin_token'); window.dispatchEvent(new Event('auth-expired')); }
    const message = await response.json().catch(() => undefined);
    throw new Error(message?.message ?? `请求失败：${response.status}`);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

function statusLabel(status: SpeciesStatus) {
  const labels: Record<SpeciesStatus, string> = {
    draft: '草稿',
    pending_review: '待审核',
    published: '已发布',
    archived: '已归档',
  };
  return labels[status];
}

function App() {
  const [activeMenu, setActiveMenu] = useState<ActiveMenu>('菌种管理');
  const [user, setUser] = useState<AuthUser | null>(null);
  const [checkingAuth, setCheckingAuth] = useState(true);
  useEffect(() => { const token = localStorage.getItem('fungi_admin_token'); if (!token) { setCheckingAuth(false); return; } void request<AuthUser>('/api/auth/me').then(setUser).catch(() => setUser(null)).finally(() => setCheckingAuth(false)); }, []);
  useEffect(() => { const expire = () => setUser(null); window.addEventListener('auth-expired', expire); return () => window.removeEventListener('auth-expired', expire); }, []);
  if (checkingAuth) return <main className="loginPage"><div className="loginCard">正在验证登录状态…</div></main>;
  if (!user) return <LoginPage onLogin={setUser} />;
  function logout() { localStorage.removeItem('fungi_admin_token'); setUser(null); }
  const visibleMenus = user.role === 'admin' ? menus : user.role === 'expert' ? (['数据审核'] as ActiveMenu[]) : menus.filter((menu) => !['数据审核', '账号管理'].includes(menu));
  const currentMenu = visibleMenus.includes(activeMenu) ? activeMenu : visibleMenus[0];

  return (
    <main className="layout">
      <aside className="sidebar">
        <h1>运营端</h1>
        <div className="currentUser"><strong>{user.displayName}</strong><span>{user.email} · {user.role}</span></div>
        {visibleMenus.map((menu) => (
          <button className={menu === currentMenu ? 'active' : ''} key={menu} onClick={() => setActiveMenu(menu)}>{menu}</button>
        ))}
        <button className="logoutButton" onClick={logout}>退出登录</button>
      </aside>

      {currentMenu === '菌种管理' && <SpeciesManagement />}
      {currentMenu === '功能标签' && <FunctionTagManagement />}
      {currentMenu === '批量导入' && <ImportManagement />}
      {currentMenu === '文献证据' && <EvidenceManagement />}
      {currentMenu === '数据审核' && <AuditManagement />}
      {currentMenu === '搜索分析' && <SearchAnalytics />}
      {currentMenu === '账号管理' && <UserManagement />}
      {currentMenu === '推荐质量' && <Placeholder title={currentMenu} />}
    </main>
  );
}

function LoginPage({ onLogin }: { onLogin: (user: AuthUser) => void }) {
  const [email, setEmail] = useState('admin@fungi.local'); const [password, setPassword] = useState(''); const [error, setError] = useState(''); const [loading, setLoading] = useState(false);
  async function login(event: FormEvent) { event.preventDefault(); setLoading(true); setError(''); try { const data = await request<LoginResponse>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }); localStorage.setItem('fungi_admin_token', data.token); onLogin(data.user); } catch (e) { setError(e instanceof Error ? e.message : '登录失败'); } finally { setLoading(false); } }
  return <main className="loginPage"><form className="loginCard" onSubmit={login}><p className="eyebrow">Fungi Wiki Admin</p><h1>登录运营端</h1><p>使用运营、专家或管理员账号登录。</p>{error && <div className="message error">{error}</div>}<label><span>邮箱</span><input type="email" required value={email} onChange={(e) => setEmail(e.target.value)} /></label><label><span>密码</span><input type="password" required value={password} onChange={(e) => setPassword(e.target.value)} /></label><button className="primary full" disabled={loading}>{loading ? '登录中…' : '登录'}</button></form></main>;
}

function SearchAnalytics() {
  const [days, setDays] = useState(30); const [report, setReport] = useState<SearchReport | null>(null); const [error, setError] = useState('');
  async function load(nextDays = days) { try { setReport(await request<SearchReport>(`/api/admin/search-analytics?days=${nextDays}`)); } catch (e) { setError(e instanceof Error ? e.message : '搜索分析加载失败'); } }
  useEffect(() => { void load(); }, []);
  return <section className="content"><PageHeader title="搜索分析" description="查看用户真实搜索需求和无结果查询，为数据补充与同义词建设提供依据。" action={<select value={days} onChange={(e) => { const value = Number(e.target.value); setDays(value); void load(value); }}><option value="7">最近 7 天</option><option value="30">最近 30 天</option><option value="90">最近 90 天</option></select>} />{error && <div className="message error">{error}</div>}{report ? <><section className="analyticsStats"><article><strong>{report.totalSearches}</strong><span>搜索次数</span></article><article><strong>{report.distinctQueries}</strong><span>不同关键词</span></article><article><strong>{report.noResultSearches}</strong><span>无结果搜索</span></article></section><section className="analyticsGrid"><div className="panel"><div className="panelTitle"><h3>热门关键词</h3></div>{report.popularQueries.map((item) => <div className="queryRow" key={item.query}><strong>{item.query}</strong><span>{item.count} 次 · 平均 {item.averageResults.toFixed(1)} 条</span></div>)}{!report.popularQueries.length && <div className="empty">暂无关键词搜索记录。</div>}</div><div className="panel"><div className="panelTitle"><h3>无结果关键词</h3></div>{report.noResultQueries.map((item) => <div className="queryRow warning" key={item.query}><strong>{item.query}</strong><span>{item.count} 次</span></div>)}{!report.noResultQueries.length && <div className="empty">暂无无结果关键词。</div>}</div></section></> : <div className="panel placeholderPanel">分析数据加载中…</div>}</section>;
}

function UserManagement() {
  const [items, setItems] = useState<AuthUser[]>([]); const [error, setError] = useState(''); const [notice, setNotice] = useState('');
  const [form, setForm] = useState({ email: '', password: '', displayName: '', role: 'operator' as AuthUser['role'] });
  async function load() { try { const data = await request<ListResponse<AuthUser>>('/api/admin/users'); setItems(data.items); } catch (e) { setError(e instanceof Error ? e.message : '账号加载失败'); } }
  useEffect(() => { void load(); }, []);
  async function submit(event: FormEvent) { event.preventDefault(); setError(''); try { await request('/api/admin/users', { method: 'POST', body: JSON.stringify(form) }); setNotice('账号已创建'); setForm({ email: '', password: '', displayName: '', role: 'operator' }); await load(); } catch (e) { setError(e instanceof Error ? e.message : '创建失败'); } }
  const roleLabel = { operator: '运营', expert: '专家', admin: '管理员' };
  return <section className="content"><PageHeader title="账号管理" description="创建运营、专家和管理员账号，并按角色控制操作权限。" /><Messages error={error} notice={notice} /><section className="mainGrid"><div className="panel tablePanel"><div className="panelTitle"><h3>账号列表</h3><span>{items.length} 个</span></div><div className="tableWrap"><table><thead><tr><th>姓名</th><th>邮箱</th><th>角色</th><th>状态</th></tr></thead><tbody>{items.map((user) => <tr key={user.id}><td><strong>{user.displayName}</strong></td><td>{user.email}</td><td>{roleLabel[user.role]}</td><td>{'status' in user ? String((user as AuthUser & { status: string }).status) : 'active'}</td></tr>)}</tbody></table></div></div><form className="panel formPanel" onSubmit={submit}><div className="panelTitle"><h3>创建账号</h3></div><label><span>显示名称 *</span><input required value={form.displayName} onChange={(e) => setForm({ ...form, displayName: e.target.value })} /></label><label><span>邮箱 *</span><input type="email" required value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} /></label><label><span>初始密码 *</span><input type="password" minLength={8} required value={form.password} onChange={(e) => setForm({ ...form, password: e.target.value })} placeholder="至少 8 位" /></label><label><span>角色</span><select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value as AuthUser['role'] })}><option value="operator">运营：维护数据</option><option value="expert">专家：审核数据</option><option value="admin">管理员：全部权限</option></select></label><button className="primary full">创建账号</button></form></section></section>;
}

function PageHeader({ title, description, action }: { title: string; description: string; action?: React.ReactNode }) {
  return (
    <header className="header">
      <div>
        <p className="eyebrow">Admin Console</p>
        <h2>{title}</h2>
        <p>{description}</p>
      </div>
      {action}
    </header>
  );
}

function SpeciesManagement() {
  const [items, setItems] = useState<Species[]>([]);
  const [query, setQuery] = useState('');
  const [status, setStatus] = useState('');
  const [form, setForm] = useState<SpeciesPayload>(emptySpeciesForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');
  const [functionTags, setFunctionTags] = useState<FunctionTag[]>([]);
  const [selectedFunctionTagIds, setSelectedFunctionTagIds] = useState<string[]>([]);
  const [cultureCondition, setCultureCondition] = useState<CultureCondition>(emptyCultureCondition);

  const editingTitle = useMemo(() => editingId ? '编辑菌种' : '新增菌种', [editingId]);

  async function loadSpecies() {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams();
      if (query.trim()) params.set('q', query.trim());
      if (status) params.set('status', status);
      const data = await request<ListResponse<Species>>(`/api/admin/species?${params.toString()}`);
      setItems(data.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }

  async function loadFunctionTags() {
    try {
      const data = await request<ListResponse<FunctionTag>>('/api/function-tags?limit=200');
      setFunctionTags(data.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : '功能标签加载失败');
    }
  }

  useEffect(() => {
    void loadSpecies();
    void loadFunctionTags();
  }, []);

  function updateField<K extends keyof SpeciesPayload>(key: K, value: SpeciesPayload[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function startCreate() {
    setEditingId(null);
    setForm(emptySpeciesForm);
    setNotice('');
    setError('');
    setSelectedFunctionTagIds([]);
    setCultureCondition(emptyCultureCondition);
  }

  async function startEdit(species: Species) {
    setEditingId(species.slug || species.id);
    setForm(toSpeciesPayload(species));
    setNotice('');
    setError('');
    try {
      const [functions, conditions] = await Promise.all([
        request<ListResponse<SpeciesFunction>>(`/api/admin/species/${species.slug || species.id}/functions`),
        request<ListResponse<CultureCondition>>(`/api/admin/species/${species.slug || species.id}/culture-conditions`),
      ]);
      setSelectedFunctionTagIds(functions.items.map((item) => item.functionTagId));
      setCultureCondition(conditions.items[0] ?? emptyCultureCondition);
    } catch (err) {
      setError(err instanceof Error ? err.message : '菌种功能加载失败');
    }
  }

  function toggleFunctionTag(tagId: string) {
    setSelectedFunctionTagIds((current) => current.includes(tagId)
      ? current.filter((id) => id !== tagId)
      : [...current, tagId]);
  }

  async function submitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError('');
    setNotice('');

    try {
      let saved: Species;
      if (editingId) {
        saved = await request<Species>(`/api/admin/species/${editingId}`, { method: 'PUT', body: JSON.stringify(form) });
        setNotice('菌种已更新');
      } else {
        saved = await request<Species>('/api/admin/species', { method: 'POST', body: JSON.stringify(form) });
        setNotice('菌种已新增');
      }
      await request(`/api/admin/species/${saved.slug || saved.id}/functions`, {
        method: 'PUT',
        body: JSON.stringify({ items: selectedFunctionTagIds.map((functionTagId) => ({ functionTagId })) }),
      });
      const hasCondition = Object.values(cultureCondition).some((value) => value !== '' && value !== null);
      await request(`/api/admin/species/${saved.slug || saved.id}/culture-conditions`, {
        method: 'PUT', body: JSON.stringify({ items: hasCondition ? [cultureCondition] : [] }),
      });
      startCreate();
      await loadSpecies();
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存失败');
    } finally {
      setSaving(false);
    }
  }

  async function archiveSpecies(species: Species) {
    if (!window.confirm(`确认归档 ${species.latinName}？`)) return;
    setError('');
    setNotice('');
    try {
      await request<void>(`/api/admin/species/${species.slug || species.id}`, { method: 'DELETE' });
      setNotice('菌种已归档');
      await loadSpecies();
    } catch (err) {
      setError(err instanceof Error ? err.message : '归档失败');
    }
  }

  async function submitForReview(species: Species) {
    if (!window.confirm(`提交 ${species.latinName} 进行发布审核？`)) return;
    try {
      await request(`/api/admin/audits/species/${species.slug || species.id}/submit`, { method: 'POST', body: '{}' });
      setNotice('已提交审核'); await loadSpecies();
    } catch (err) { setError(err instanceof Error ? err.message : '提交审核失败'); }
  }

  return (
    <section className="content">
      <PageHeader
        title="菌种管理"
        description="维护微生物百科与功能菌数据库的基础菌种主数据。"
        action={<button className="primary" onClick={startCreate}>新增菌种</button>}
      />

      <section className="toolbar">
        <input value={query} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => event.key === 'Enter' && loadSpecies()} placeholder="搜索 slug、拉丁名、中文名或摘要" />
        <select value={status} onChange={(event) => setStatus(event.target.value)}>
          <option value="">全部状态</option>
          <option value="draft">草稿</option>
          <option value="pending_review">待审核</option>
          <option value="published">已发布</option>
          <option value="archived">已归档</option>
        </select>
        <button onClick={loadSpecies} disabled={loading}>{loading ? '加载中...' : '查询'}</button>
      </section>

      <Messages error={error} notice={notice} />

      <section className="mainGrid">
        <div className="panel tablePanel">
          <div className="panelTitle"><h3>菌种列表</h3><span>{items.length} 条</span></div>
          <div className="tableWrap">
            <table>
              <thead>
                <tr><th>拉丁名</th><th>中文名</th><th>安全等级</th><th>状态</th><th>更新时间</th><th>操作</th></tr>
              </thead>
              <tbody>
                {items.map((species) => (
                  <tr key={species.id}>
                    <td><strong>{species.latinName}</strong><small>{species.slug}</small></td>
                    <td>{species.chineseName || '-'}</td>
                    <td>{species.safetyLevel || '-'}</td>
                    <td><span className={`status ${species.status}`}>{statusLabel(species.status)}</span></td>
                    <td>{new Date(species.updatedAt).toLocaleString()}</td>
                    <td className="actions"><button onClick={() => void startEdit(species)}>编辑</button>{species.status === 'draft' && <button onClick={() => submitForReview(species)}>提交审核</button>}<button className="danger" onClick={() => archiveSpecies(species)}>归档</button></td>
                  </tr>
                ))}
                {!items.length && <tr><td colSpan={6} className="empty">暂无数据，请新增菌种或调整查询条件。</td></tr>}
              </tbody>
            </table>
          </div>
        </div>

        <form className="panel formPanel" onSubmit={submitForm}>
          <div className="panelTitle"><h3>{editingTitle}</h3>{editingId && <button type="button" onClick={startCreate}>取消编辑</button>}</div>
          <label><span>Slug *</span><input value={form.slug} onChange={(event) => updateField('slug', event.target.value)} required placeholder="bacillus-subtilis" /></label>
          <label><span>拉丁名 *</span><input value={form.latinName} onChange={(event) => updateField('latinName', event.target.value)} required placeholder="Bacillus subtilis" /></label>
          <label><span>中文名</span><input value={form.chineseName} onChange={(event) => updateField('chineseName', event.target.value)} placeholder="枯草芽孢杆菌" /></label>
          <label><span>保藏/菌株编号</span><input value={form.strainNumber} onChange={(event) => updateField('strainNumber', event.target.value)} placeholder="ATCC / DSM / CGMCC" /></label>
          <label><span>来源环境</span><input value={form.sourceEnvironment} onChange={(event) => updateField('sourceEnvironment', event.target.value)} placeholder="土壤、海洋、发酵食品等" /></label>
          <div className="twoCols">
            <label><span>安全等级</span><input value={form.safetyLevel} onChange={(event) => updateField('safetyLevel', event.target.value)} placeholder="BSL-1" /></label>
            <label><span>状态</span><input value={statusLabel(form.status)} disabled /></label>
          </div>
          <label className="checkbox"><input type="checkbox" checked={form.isModelOrganism} onChange={(event) => updateField('isModelOrganism', event.target.checked)} /><span>模式菌 / 常用底盘菌</span></label>
          <label><span>摘要</span><textarea value={form.summary} onChange={(event) => updateField('summary', event.target.value)} rows={5} placeholder="简要描述该菌种的功能、用途和特点" /></label>
          <fieldset className="tagFieldset">
            <legend>功能标签</legend>
            <div className="tagOptions">
              {functionTags.map((tag) => (
                <label className="checkbox" key={tag.id}>
                  <input type="checkbox" checked={selectedFunctionTagIds.includes(tag.id)} onChange={() => toggleFunctionTag(tag.id)} />
                  <span>{tag.name}</span>
                </label>
              ))}
              {!functionTags.length && <span className="empty">请先创建功能标签。</span>}
            </div>
          </fieldset>
          <fieldset className="tagFieldset">
            <legend>主要培养条件</legend>
            <label><span>培养基</span><input value={cultureCondition.mediumName} onChange={(e) => setCultureCondition({ ...cultureCondition, mediumName: e.target.value })} placeholder="LB、YPD 等" /></label>
            <div className="twoCols">
              <label><span>温度范围 °C</span><div className="inlineInputs"><input type="number" value={cultureCondition.temperatureMin ?? ''} onChange={(e) => setCultureCondition({ ...cultureCondition, temperatureMin: e.target.value ? Number(e.target.value) : null })} /><input type="number" value={cultureCondition.temperatureMax ?? ''} onChange={(e) => setCultureCondition({ ...cultureCondition, temperatureMax: e.target.value ? Number(e.target.value) : null })} /></div></label>
              <label><span>pH 范围</span><div className="inlineInputs"><input type="number" step="0.1" value={cultureCondition.phMin ?? ''} onChange={(e) => setCultureCondition({ ...cultureCondition, phMin: e.target.value ? Number(e.target.value) : null })} /><input type="number" step="0.1" value={cultureCondition.phMax ?? ''} onChange={(e) => setCultureCondition({ ...cultureCondition, phMax: e.target.value ? Number(e.target.value) : null })} /></div></label>
            </div>
            <label><span>氧需求</span><input value={cultureCondition.oxygenRequirement} onChange={(e) => setCultureCondition({ ...cultureCondition, oxygenRequirement: e.target.value })} placeholder="好氧、厌氧、兼性厌氧" /></label>
            <label><span>培养时间</span><input value={cultureCondition.cultureTime} onChange={(e) => setCultureCondition({ ...cultureCondition, cultureTime: e.target.value })} placeholder="24–48 h" /></label>
            <label><span>备注</span><textarea rows={2} value={cultureCondition.notes} onChange={(e) => setCultureCondition({ ...cultureCondition, notes: e.target.value })} /></label>
          </fieldset>
          <button className="primary full" disabled={saving}>{saving ? '保存中...' : '保存菌种'}</button>
        </form>
      </section>
    </section>
  );
}

function FunctionTagManagement() {
  const [items, setItems] = useState<FunctionTag[]>([]);
  const [query, setQuery] = useState('');
  const [form, setForm] = useState<FunctionTagPayload>(emptyTagForm);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [notice, setNotice] = useState('');

  const editingTitle = useMemo(() => editingId ? '编辑功能标签' : '新增功能标签', [editingId]);

  async function loadTags() {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams();
      if (query.trim()) params.set('q', query.trim());
      const data = await request<ListResponse<FunctionTag>>(`/api/admin/function-tags?${params.toString()}`);
      setItems(data.items);
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载标签失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    void loadTags();
  }, []);

  function updateField<K extends keyof FunctionTagPayload>(key: K, value: FunctionTagPayload[K]) {
    setForm((current) => ({ ...current, [key]: value }));
  }

  function startCreate() {
    setEditingId(null);
    setForm(emptyTagForm);
    setNotice('');
    setError('');
  }

  function startEdit(tag: FunctionTag) {
    setEditingId(tag.code || tag.id);
    setForm(toTagPayload(tag));
    setNotice('');
    setError('');
  }

  async function submitForm(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSaving(true);
    setError('');
    setNotice('');

    try {
      if (editingId) {
        await request<FunctionTag>(`/api/admin/function-tags/${editingId}`, { method: 'PUT', body: JSON.stringify(form) });
        setNotice('功能标签已更新');
      } else {
        await request<FunctionTag>('/api/admin/function-tags', { method: 'POST', body: JSON.stringify(form) });
        setNotice('功能标签已新增');
      }
      startCreate();
      await loadTags();
    } catch (err) {
      setError(err instanceof Error ? err.message : '保存标签失败');
    } finally {
      setSaving(false);
    }
  }

  async function deleteTag(tag: FunctionTag) {
    if (!window.confirm(`确认删除功能标签「${tag.name}」？`)) return;
    setError('');
    setNotice('');
    try {
      await request<void>(`/api/admin/function-tags/${tag.code || tag.id}`, { method: 'DELETE' });
      setNotice('功能标签已删除');
      await loadTags();
    } catch (err) {
      setError(err instanceof Error ? err.message : '删除标签失败。请确认没有菌种功能关联正在使用该标签。');
    }
  }

  return (
    <section className="content">
      <PageHeader
        title="功能标签管理"
        description="维护促生、生防、固氮、解磷、降解、发酵等标准功能标签。"
        action={<button className="primary" onClick={startCreate}>新增标签</button>}
      />

      <section className="toolbar compactToolbar">
        <input value={query} onChange={(event) => setQuery(event.target.value)} onKeyDown={(event) => event.key === 'Enter' && loadTags()} placeholder="搜索标签名称、编码或描述" />
        <button onClick={loadTags} disabled={loading}>{loading ? '加载中...' : '查询'}</button>
      </section>

      <Messages error={error} notice={notice} />

      <section className="mainGrid">
        <div className="panel tablePanel">
          <div className="panelTitle"><h3>功能标签列表</h3><span>{items.length} 条</span></div>
          <div className="tableWrap">
            <table>
              <thead>
                <tr><th>名称</th><th>编码</th><th>描述</th><th>排序</th><th>更新时间</th><th>操作</th></tr>
              </thead>
              <tbody>
                {items.map((tag) => (
                  <tr key={tag.id}>
                    <td><strong>{tag.name}</strong>{tag.parentId && <small>父级：{tag.parentId}</small>}</td>
                    <td><code>{tag.code}</code></td>
                    <td>{tag.description || '-'}</td>
                    <td>{tag.sortOrder}</td>
                    <td>{new Date(tag.updatedAt).toLocaleString()}</td>
                    <td className="actions"><button onClick={() => startEdit(tag)}>编辑</button><button className="danger" onClick={() => deleteTag(tag)}>删除</button></td>
                  </tr>
                ))}
                {!items.length && <tr><td colSpan={6} className="empty">暂无功能标签，请新增或调整查询条件。</td></tr>}
              </tbody>
            </table>
          </div>
        </div>

        <form className="panel formPanel" onSubmit={submitForm}>
          <div className="panelTitle"><h3>{editingTitle}</h3>{editingId && <button type="button" onClick={startCreate}>取消编辑</button>}</div>
          <label><span>名称 *</span><input value={form.name} onChange={(event) => updateField('name', event.target.value)} required placeholder="促生" /></label>
          <label><span>编码 *</span><input value={form.code} onChange={(event) => updateField('code', event.target.value)} required placeholder="plant-growth-promotion" /></label>
          <label><span>父级标签 ID</span><input value={form.parentId} onChange={(event) => updateField('parentId', event.target.value)} placeholder="可选，填写父级标签 UUID" /></label>
          <label><span>排序</span><input type="number" value={form.sortOrder} onChange={(event) => updateField('sortOrder', Number(event.target.value))} /></label>
          <label><span>描述</span><textarea value={form.description} onChange={(event) => updateField('description', event.target.value)} rows={5} placeholder="描述该功能标签的适用范围和判定标准" /></label>
          <button className="primary full" disabled={saving}>{saving ? '保存中...' : '保存标签'}</button>
        </form>
      </section>
    </section>
  );
}

function ImportManagement() {
  const [file, setFile] = useState<File | null>(null); const [uploading, setUploading] = useState(false);
  const [result, setResult] = useState<ImportBatch | null>(null); const [history, setHistory] = useState<ImportBatch[]>([]);
  const [error, setError] = useState(''); const [notice, setNotice] = useState('');
  async function loadHistory() { try { const data = await request<ListResponse<ImportBatch>>('/api/admin/imports?limit=20'); setHistory(data.items); } catch (e) { setError(e instanceof Error ? e.message : '导入记录加载失败'); } }
  useEffect(() => { void loadHistory(); }, []);
  async function upload(event: FormEvent) { event.preventDefault(); if (!file) { setError('请选择 CSV 或 XLSX 文件'); return; } setUploading(true); setError(''); setNotice(''); const body = new FormData(); body.append('file', file); try { const data = await request<ImportBatch>('/api/admin/imports/species', { method: 'POST', body }); setResult(data); setNotice(`导入完成：成功 ${data.successRows} 行，失败 ${data.failedRows} 行。成功数据已进入待审核队列。`); await loadHistory(); } catch (e) { setError(e instanceof Error ? e.message : '导入失败'); } finally { setUploading(false); } }
  function downloadTemplate() { const headers = ['slug','latin_name','chinese_name','strain_number','source_environment','safety_level','is_model_organism','summary','function_tags','medium_name','temperature_min','temperature_max','ph_min','ph_max','oxygen_requirement','culture_time']; const example = ['bacillus-demo','Bacillus demo','示例芽孢杆菌','CGMCC 1.1','土壤','BSL-1','否','示例数据','biocontrol;plant-growth-promotion','LB','25','37','6.0','8.0','好氧','24 h']; const csv = `\uFEFF${headers.join(',')}\n${example.join(',')}\n`; const url = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' })); const link = document.createElement('a'); link.href = url; link.download = 'species-import-template.csv'; link.click(); URL.revokeObjectURL(url); }
  return <section className="content"><PageHeader title="批量导入" description="上传 CSV 或 Excel，逐行校验后进入数据审核流程。" action={<button className="primary" onClick={downloadTemplate}>下载 CSV 模板</button>} />
    <Messages error={error} notice={notice} />
    <section className="importGrid"><form className="panel uploadPanel" onSubmit={upload}><h3>上传菌种数据</h3><p>支持 .csv、.xlsx、.xlsm，最大 10MB、5000 行。功能标签填写编码或名称，多个标签用分号分隔。</p><label className="filePicker"><input type="file" accept=".csv,.xlsx,.xlsm" onChange={(e) => setFile(e.target.files?.[0] ?? null)} /><span>{file?.name ?? '选择文件'}</span></label><button className="primary full" disabled={uploading}>{uploading ? '解析导入中…' : '开始导入并提交审核'}</button></form>
      <div className="panel"><div className="panelTitle"><h3>最近导入</h3><span>{history.length} 个批次</span></div><div className="batchList">{history.map((batch) => <article key={batch.id}><div><strong>{batch.sourceFilename}</strong><small>{new Date(batch.createdAt).toLocaleString()}</small></div><span className={batch.failedRows ? 'countWarning' : 'countSuccess'}>{batch.successRows} 成功 / {batch.failedRows} 失败</span></article>)}{!history.length && <div className="empty">暂无导入记录。</div>}</div></div>
    </section>
    {result?.rows && <section className="panel importResult"><div className="panelTitle"><h3>本次逐行结果</h3><span>{result.totalRows} 行</span></div><div className="tableWrap"><table><thead><tr><th>行号</th><th>Slug</th><th>结果</th><th>问题</th></tr></thead><tbody>{result.rows.map((row) => <tr key={row.rowNumber}><td>{row.rowNumber}</td><td><code>{row.slug || '-'}</code></td><td>{row.status === 'imported' ? '已导入待审核' : '失败'}</td><td className="rowErrors">{row.errors?.join('；') || '-'}</td></tr>)}</tbody></table></div></section>}
  </section>;
}

function EvidenceManagement() {
  const [species, setSpecies] = useState<Species[]>([]); const [selected, setSelected] = useState('');
  const [items, setItems] = useState<Evidence[]>([]); const [error, setError] = useState(''); const [notice, setNotice] = useState('');
  const [form, setForm] = useState({ title: '', authors: '', journal: '', publicationYear: '', doi: '', pmid: '', sourceUrl: '', conclusion: '', evidenceLevel: 'medium', evidenceScore: 50 });
  useEffect(() => { void request<ListResponse<Species>>('/api/admin/species?limit=100').then((x) => { setSpecies(x.items); if (x.items[0]) setSelected(x.items[0].slug); }).catch((e) => setError(e.message)); }, []);
  async function load(id = selected) { if (!id) return; try { const x = await request<ListResponse<Evidence>>(`/api/admin/species/${id}/evidences`); setItems(x.items); } catch (e) { setError(e instanceof Error ? e.message : '加载失败'); } }
  useEffect(() => { void load(selected); }, [selected]);
  async function submit(e: FormEvent) { e.preventDefault(); try { await request(`/api/admin/species/${selected}/evidences`, { method: 'POST', body: JSON.stringify({ ...form, publicationYear: form.publicationYear ? Number(form.publicationYear) : null }) }); setNotice('文献证据已添加'); setForm({ ...form, title: '', authors: '', journal: '', publicationYear: '', doi: '', pmid: '', sourceUrl: '', conclusion: '' }); await load(); } catch (x) { setError(x instanceof Error ? x.message : '保存失败'); } }
  async function remove(id: string) { if (!window.confirm('确认删除该证据关联？')) return; await request(`/api/admin/species/${selected}/evidences/${id}`, { method: 'DELETE' }); await load(); }
  return <section className="content">
    <PageHeader title="文献证据" description="维护文献来源、实验结论和证据等级。" />
    <section className="toolbar compactToolbar"><select value={selected} onChange={(e) => setSelected(e.target.value)}>{species.map((x) => <option key={x.id} value={x.slug}>{x.latinName}</option>)}</select><button onClick={() => load()}>刷新</button></section>
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

function AuditManagement() {
  const [items, setItems] = useState<AuditRecord[]>([]); const [error, setError] = useState(''); const [notice, setNotice] = useState('');
  async function load() { try { const x = await request<ListResponse<AuditRecord>>('/api/admin/audits?status=pending'); setItems(x.items); } catch (e) { setError(e instanceof Error ? e.message : '加载失败'); } }
  useEffect(() => { void load(); }, []);
  async function review(id: string, action: 'approve' | 'reject') { const comment = window.prompt(action === 'approve' ? '审核意见（可选）' : '请输入驳回原因') ?? ''; if (action === 'reject' && !comment) return; try { await request(`/api/admin/audits/${id}/${action}`, { method: 'POST', body: JSON.stringify({ comment }) }); setNotice(action === 'approve' ? '已审核通过并发布' : '已驳回为草稿'); await load(); } catch (e) { setError(e instanceof Error ? e.message : '审核失败'); } }
  return <section className="content"><PageHeader title="数据审核" description="审核菌种发布申请；通过后才会在用户端展示。" /><Messages error={error} notice={notice} /><div className="panel auditList">{items.map((x) => <article key={x.id}><div><strong>{x.entityName}</strong><small>提交于 {new Date(x.submittedAt).toLocaleString()}</small></div><div className="actions"><button onClick={() => review(x.id, 'approve')}>通过并发布</button><button className="danger" onClick={() => review(x.id, 'reject')}>驳回</button></div></article>)}{!items.length && <div className="empty">没有待审核数据。</div>}</div></section>;
}

function Messages({ error, notice }: { error: string; notice: string }) {
  return (
    <>
      {error && <div className="message error">{error}</div>}
      {notice && <div className="message success">{notice}</div>}
    </>
  );
}

function Placeholder({ title }: { title: string }) {
  return (
    <section className="content">
      <PageHeader title={title} description="该模块还在规划中，后续会继续接入真实业务接口。" />
      <div className="panel placeholderPanel">{title} 页面待建设。</div>
    </section>
  );
}

createRoot(document.getElementById('root')!).render(<App />);
