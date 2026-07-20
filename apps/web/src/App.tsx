import { FormEvent, useEffect, useMemo, useState } from 'react';
import { request } from './api';
import { RecommendationPanel } from './features/recommend/RecommendationPanel';
import { SearchHero } from './features/search/SearchHero';
import { SpeciesDetailPanel, SpeciesListPanel } from './features/species/SpeciesPanels';
import type { CultureCondition, Evidence, FunctionTag, ListResponse, RecommendationResponse, SearchFilters, Species, SpeciesAlias, SpeciesFunction } from './types';
const emptyFilters: SearchFilters = { functionTag: '', temperature: '', ph: '', safetyLevel: '', sourceEnvironment: '' };

function getRouteSpeciesSlug() {
  const match = window.location.pathname.match(/^\/species\/([^/]+)\/?$/);
  return match ? decodeURIComponent(match[1]) : '';
}

function navigateTo(path: string) {
  window.history.pushState({}, '', path);
  window.dispatchEvent(new PopStateEvent('popstate'));
}

export function App() {
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
  const [aliases, setAliases] = useState<SpeciesAlias[]>([]);
  const [functionTags, setFunctionTags] = useState<FunctionTag[]>([]);
  const [filters, setFilters] = useState<SearchFilters>(emptyFilters);
  const [appliedFilters, setAppliedFilters] = useState<SearchFilters>(emptyFilters);
  const [total, setTotal] = useState(0); const [page, setPage] = useState(1); const pageSize = 10; const [sort, setSort] = useState('updated');
  const [recommendRequirement, setRecommendRequirement] = useState(''); const [recommendFunction, setRecommendFunction] = useState(''); const [recommending, setRecommending] = useState(false); const [recommendation, setRecommendation] = useState<RecommendationResponse | null>(null);
  const [recommendFeedback, setRecommendFeedback] = useState('');

  const stats = useMemo(() => {
    const modelCount = items.filter((item) => item.isModelOrganism).length;
    const safetyLevels = new Set(items.map((item) => item.safetyLevel).filter(Boolean));
    return { modelCount, safetyLevelCount: safetyLevels.size };
  }, [items]);

  async function loadSpecies(search = query, nextFilters = filters, targetPage = 1, nextSort = sort) {
    setLoading(true);
    setError('');
    try {
      const params = new URLSearchParams();
      if (search.trim()) params.set('q', search.trim());
      Object.entries(nextFilters).forEach(([key, value]) => { if (value.trim()) params.set(key, value.trim()); });
      params.set('limit', String(pageSize)); params.set('offset', String((targetPage - 1) * pageSize)); params.set('sort', nextSort);
      const data = await request<ListResponse>(`/api/species?${params.toString()}`);
      setItems(data.items);
      setTotal(data.total); setPage(targetPage); setSort(nextSort);
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
      const [functionData, conditionData, evidenceData, aliasData] = await Promise.all([
        request<{ items: SpeciesFunction[] }>(`/api/species/${data.slug || data.id}/functions`),
        request<{ items: CultureCondition[] }>(`/api/species/${data.slug || data.id}/culture-conditions`),
        request<{ items: Evidence[] }>(`/api/species/${data.slug || data.id}/evidences`),
        request<{ items: SpeciesAlias[] }>(`/api/species/${data.slug || data.id}/aliases`),
      ]);
      setSpeciesFunctions(functionData.items); setCultureConditions(conditionData.items); setEvidences(evidenceData.items);
      setAliases(aliasData.items);
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
    void loadSpecies(query, filters, 1);
  }

  function resetSearch() {
    setQuery(''); setFilters(emptyFilters); setSort('updated'); void loadSpecies('', emptyFilters, 1, 'updated');
  }

  async function submitRecommendation(event: FormEvent) { event.preventDefault(); setRecommending(true); setError(''); setRecommendFeedback(''); try { const payload = { requirement: recommendRequirement, functionTag: recommendFunction || undefined, temperature: filters.temperature ? Number(filters.temperature) : undefined, ph: filters.ph ? Number(filters.ph) : undefined, safetyLevel: filters.safetyLevel || undefined, sourceEnvironment: filters.sourceEnvironment || undefined, limit: 5 }; setRecommendation(await request<RecommendationResponse>('/api/recommendations', { method: 'POST', body: JSON.stringify(payload) })); } catch (err) { setError(err instanceof Error ? err.message : '推荐失败'); } finally { setRecommending(false); } }
  async function sendRecommendationFeedback(feedbackType: 'helpful' | 'unhelpful') { if (!recommendation) return; const content = window.prompt(feedbackType === 'helpful' ? '哪些内容对你有帮助？（可选）' : '推荐哪里不符合需求？（可选）') ?? ''; try { await request(`/api/recommendations/${recommendation.recordId}/feedback`, { method: 'POST', body: JSON.stringify({ feedbackType, content }) }); setRecommendFeedback(feedbackType); } catch (err) { setError(err instanceof Error ? err.message : '反馈提交失败'); } }

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
      <SearchHero routeSlug={routeSlug} query={query} loading={loading} filters={filters} functionTags={functionTags} onBack={() => navigateTo('/')} onQueryChange={setQuery} onFiltersChange={setFilters} onSubmit={submitSearch} />
      <RecommendationPanel requirement={recommendRequirement} functionCode={recommendFunction} recommending={recommending} recommendation={recommendation} feedback={recommendFeedback} functionTags={functionTags} onRequirementChange={setRecommendRequirement} onFunctionChange={setRecommendFunction} onSubmit={submitRecommendation} onOpenSpecies={(slug) => void loadDetail(slug)} onFeedback={(type) => void sendRecommendationFeedback(type)} />

      <section className="stats">
        <article>
          <strong>{total}</strong>
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
        <SpeciesListPanel items={items} selectedId={selected?.id} loading={loading} submittedQuery={submittedQuery} activeFilterCount={activeFilterCount} page={page} pageSize={pageSize} total={total} sort={sort} appliedFilters={appliedFilters} onReset={resetSearch} onSort={(value) => void loadSpecies(submittedQuery, appliedFilters, 1, value)} onPage={(target) => void loadSpecies(submittedQuery, appliedFilters, target)} onSelect={(slug) => void loadDetail(slug)} />
        <SpeciesDetailPanel selected={selected} detailLoading={detailLoading} routeSlug={routeSlug} aliases={aliases} functions={speciesFunctions} conditions={cultureConditions} evidences={evidences} />
      </section>

    </main>
  );
}
