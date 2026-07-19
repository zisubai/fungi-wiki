import React, { FormEvent, useEffect, useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

type SpeciesStatus = 'draft' | 'pending_review' | 'published' | 'archived';

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

type ListResponse = {
  items: Species[];
};

type SpeciesFunction = {
  functionTagId: string;
  functionTagName: string;
};
type CultureCondition = { id: string; mediumName: string; temperatureMin: number | null; temperatureMax: number | null; phMin: number | null; phMax: number | null; oxygenRequirement: string; cultureTime: string; notes: string; };
type Evidence = { id: string; title: string; authors: string; journal: string; publicationYear: number | null; doi: string; pmid: string; sourceUrl: string; conclusion: string; evidenceLevel: string; evidenceScore: number; };
type FunctionTag = { id: string; name: string; code: string; };
type SearchFilters = { functionTag: string; temperature: string; ph: string; safetyLevel: string; sourceEnvironment: string; };
const emptyFilters: SearchFilters = { functionTag: '', temperature: '', ph: '', safetyLevel: '', sourceEnvironment: '' };

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

function getRouteSpeciesSlug() {
  const match = window.location.pathname.match(/^\/species\/([^/]+)\/?$/);
  return match ? decodeURIComponent(match[1]) : '';
}

function navigateTo(path: string) {
  window.history.pushState({}, '', path);
  window.dispatchEvent(new PopStateEvent('popstate'));
}

async function request<T>(path: string): Promise<T> {
  const response = await fetch(`${apiBaseUrl}${path}`);

  if (!response.ok) {
    const message = await response.json().catch(() => undefined);
    throw new Error(message?.message ?? `请求失败：${response.status}`);
  }

  return response.json() as Promise<T>;
}

function App() {
  const [items, setItems] = useState<Species[]>([]);
  const [selected, setSelected] = useState<Species | null>(null);
  const [query, setQuery] = useState('');
  const [submittedQuery, setSubmittedQuery] = useState('');
  const [loading, setLoading] = useState(false);
  const [detailLoading, setDetailLoading] = useState(false);
  const [error, setError] = useState('');
  const [routeSlug, setRouteSlug] = useState(() => getRouteSpeciesSlug());
  const [speciesFunctions, setSpeciesFunctions] = useState<SpeciesFunction[]>([]);
  const [cultureConditions, setCultureConditions] = useState<CultureCondition[]>([]);
  const [evidences, setEvidences] = useState<Evidence[]>([]);
  const [functionTags, setFunctionTags] = useState<FunctionTag[]>([]);
  const [filters, setFilters] = useState<SearchFilters>(emptyFilters);
  const [appliedFilters, setAppliedFilters] = useState<SearchFilters>(emptyFilters);

  const stats = useMemo(() => {
    const modelCount = items.filter((item) => item.isModelOrganism).length;
    const safetyLevels = new Set(items.map((item) => item.safetyLevel).filter(Boolean));
    return { modelCount, safetyLevelCount: safetyLevels.size };
  }, [items]);

  async function loadSpecies(search = query, nextFilters = filters) {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams();
      if (search.trim()) params.set('q', search.trim());
      Object.entries(nextFilters).forEach(([key, value]) => { if (value.trim()) params.set(key, value.trim()); });
      const data = await request<ListResponse>(`/api/species?${params.toString()}`);
      setItems(data.items);
      setSubmittedQuery(search.trim());
      setAppliedFilters(nextFilters);
      if (!routeSlug && data.items.length > 0) {
        await loadDetail(data.items[0].slug || data.items[0].id, false);
      } else if (!routeSlug) {
        setSelected(null);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载菌种失败');
    } finally {
      setLoading(false);
    }
  }

  async function loadDetail(idOrSlug: string, shouldNavigate = true) {
    setDetailLoading(true);
    setError('');
    try {
      const data = await request<Species>(`/api/species/${idOrSlug}`);
      setSelected(data);
      const [functionData, conditionData, evidenceData] = await Promise.all([
        request<{ items: SpeciesFunction[] }>(`/api/species/${data.slug || data.id}/functions`),
        request<{ items: CultureCondition[] }>(`/api/species/${data.slug || data.id}/culture-conditions`),
        request<{ items: Evidence[] }>(`/api/species/${data.slug || data.id}/evidences`),
      ]);
      setSpeciesFunctions(functionData.items); setCultureConditions(conditionData.items); setEvidences(evidenceData.items);
      if (shouldNavigate) {
        navigateTo(`/species/${encodeURIComponent(data.slug || data.id)}`);
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载详情失败');
    } finally {
      setDetailLoading(false);
    }
  }

  function submitSearch(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    void loadSpecies(query);
  }

  function resetSearch() {
    setQuery(''); setFilters(emptyFilters); void loadSpecies('', emptyFilters);
  }

  useEffect(() => {
    const handlePopState = () => setRouteSlug(getRouteSpeciesSlug());
    window.addEventListener('popstate', handlePopState);
    return () => window.removeEventListener('popstate', handlePopState);
  }, []);

  useEffect(() => {
    void loadSpecies('');
    void request<{ items: FunctionTag[] }>('/api/function-tags?limit=200').then((data) => setFunctionTags(data.items)).catch(() => undefined);
  }, []);

  const activeFilterCount = Object.values(appliedFilters).filter(Boolean).length;

  useEffect(() => {
    if (routeSlug) {
      void loadDetail(routeSlug, false);
    }
  }, [routeSlug]);

  return (
    <main className="page">
      <section className="hero">
        {routeSlug && <button className="backLink" onClick={() => navigateTo('/')}>← 返回菌种列表</button>}
        <p className="eyebrow">Microbial Knowledge Platform</p>
        <h1>微生物百科</h1>
        <p className="summary">从菌种百科开始，逐步建设功能菌数据库、智能搜索和 AI 推荐能力。</p>
        <form className="searchBox" onSubmit={submitSearch}>
          <input
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="搜索菌种、功能、应用场景，例如：枯草芽孢杆菌"
          />
          <button disabled={loading}>{loading ? '搜索中...' : '搜索'}</button>
        </form>
        <section className="filterPanel">
          <div className="filterTitle"><strong>多条件筛选</strong><span>所有已填写条件同时满足</span></div>
          <div className="filterGrid">
            <label><span>功能标签</span><select value={filters.functionTag} onChange={(e) => setFilters({ ...filters, functionTag: e.target.value })}><option value="">全部功能</option>{functionTags.map((tag) => <option key={tag.id} value={tag.code}>{tag.name}</option>)}</select></label>
            <label><span>适宜温度 °C</span><input type="number" step="0.1" value={filters.temperature} onChange={(e) => setFilters({ ...filters, temperature: e.target.value })} placeholder="例如 30" /></label>
            <label><span>适宜 pH</span><input type="number" min="0" max="14" step="0.1" value={filters.ph} onChange={(e) => setFilters({ ...filters, ph: e.target.value })} placeholder="例如 7.0" /></label>
            <label><span>安全等级</span><select value={filters.safetyLevel} onChange={(e) => setFilters({ ...filters, safetyLevel: e.target.value })}><option value="">全部等级</option><option value="BSL-1">BSL-1</option><option value="BSL-2">BSL-2</option><option value="BSL-3">BSL-3</option><option value="BSL-4">BSL-4</option></select></label>
            <label><span>来源环境</span><input value={filters.sourceEnvironment} onChange={(e) => setFilters({ ...filters, sourceEnvironment: e.target.value })} placeholder="土壤、海洋、食品…" /></label>
          </div>
        </section>
      </section>

      <section className="stats">
        <article>
          <strong>{items.length}</strong>
          <span>已发布菌种</span>
        </article>
        <article>
          <strong>{stats.modelCount}</strong>
          <span>模式菌 / 底盘菌</span>
        </article>
        <article>
          <strong>{stats.safetyLevelCount}</strong>
          <span>安全等级类型</span>
        </article>
      </section>

      {error && <div className="message error">{error}</div>}

      <section className={`contentGrid ${routeSlug ? 'detailRoute' : ''}`}>
        <section className="panel listPanel">
          <div className="panelTitle">
            <div>
              <h2>菌种列表</h2>
              <p>{submittedQuery || activeFilterCount ? `${submittedQuery ? `关键词“${submittedQuery}” · ` : ''}${activeFilterCount} 个筛选条件` : '展示已发布菌种'}</p>
            </div>
            <button className="ghost" onClick={resetSearch}>重置</button>
          </div>

          <div className="speciesList">
            {items.map((item) => (
              <button
                className={`speciesItem ${selected?.id === item.id ? 'selected' : ''}`}
                key={item.id}
                onClick={() => loadDetail(item.slug || item.id)}
              >
                <span>
                  <strong>{item.latinName}</strong>
                  <small>{item.chineseName || item.slug}</small>
                </span>
                <em>{item.safetyLevel || '未标注'}</em>
              </button>
            ))}

            {!items.length && !loading && (
              <div className="empty">暂无匹配菌种。可以在运营端新增并发布菌种。</div>
            )}
          </div>
        </section>

        <section className="panel detailPanel">
          <div className="panelTitle">
            <div>
              <h2>菌种详情</h2>
              <p>{detailLoading ? '详情加载中...' : routeSlug ? `直达详情页：/species/${routeSlug}` : '来自 GET /api/species/:idOrSlug'}</p>
            </div>
          </div>

          {selected ? (
            <article className="detail">
              <div className="titleRow">
                <div>
                  <h3>{selected.latinName}</h3>
                  <p>{selected.chineseName || '暂无中文名'}</p>
                </div>
                <span className="badge">{selected.safetyLevel || '未标注'}</span>
              </div>

              <p className="description">{selected.summary || '暂无摘要。'}</p>

              <section className="functionSection">
                <h4>功能标签</h4>
                <div className="functionTags">
                  {speciesFunctions.map((item) => <span key={item.functionTagId}>{item.functionTagName}</span>)}
                  {!speciesFunctions.length && <small>暂无已关联功能</small>}
                </div>
              </section>

              <section className="functionSection">
                <h4>培养条件</h4>
                {cultureConditions.map((item) => <div className="conditionCard" key={item.id}>
                  <strong>{item.mediumName || '未指定培养基'}</strong>
                  <span>温度：{item.temperatureMin ?? '-'}–{item.temperatureMax ?? '-'} °C</span>
                  <span>pH：{item.phMin ?? '-'}–{item.phMax ?? '-'}</span>
                  <span>氧需求：{item.oxygenRequirement || '-'}</span><span>时间：{item.cultureTime || '-'}</span>
                </div>)}
                {!cultureConditions.length && <small>暂无培养条件</small>}
              </section>

              <section className="functionSection">
                <h4>文献证据</h4>
                <div className="evidenceCards">{evidences.map((item) => <article key={item.id}>
                  <strong>{item.sourceUrl ? <a href={item.sourceUrl} target="_blank" rel="noreferrer">{item.title}</a> : item.title}</strong>
                  <p>{item.conclusion}</p><small>{item.journal} {item.publicationYear ?? ''} · 证据等级：{item.evidenceLevel}</small>
                </article>)}</div>
                {!evidences.length && <small>暂无文献证据</small>}
              </section>

              <dl>
                <div>
                  <dt>Slug</dt>
                  <dd>{selected.slug}</dd>
                </div>
                <div>
                  <dt>菌株编号</dt>
                  <dd>{selected.strainNumber || '-'}</dd>
                </div>
                <div>
                  <dt>来源环境</dt>
                  <dd>{selected.sourceEnvironment || '-'}</dd>
                </div>
                <div>
                  <dt>模式菌 / 底盘菌</dt>
                  <dd>{selected.isModelOrganism ? '是' : '否'}</dd>
                </div>
                <div>
                  <dt>数据质量分</dt>
                  <dd>{selected.dataQualityScore}</dd>
                </div>
                <div>
                  <dt>更新时间</dt>
                  <dd>{new Date(selected.updatedAt).toLocaleString()}</dd>
                </div>
              </dl>
            </article>
          ) : (
            <div className="empty">请选择一个菌种查看详情。</div>
          )}
        </section>
      </section>
    </main>
  );
}

createRoot(document.getElementById('root')!).render(<App />);
