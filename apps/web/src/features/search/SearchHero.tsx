import type { FormEvent } from 'react';
import type { FunctionTag, SearchFilters } from '../../types';

type Props = {
  routeSlug: string;
  query: string;
  loading: boolean;
  filters: SearchFilters;
  functionTags: FunctionTag[];
  onBack: () => void;
  onQueryChange: (value: string) => void;
  onFiltersChange: (filters: SearchFilters) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
};

export function SearchHero({ routeSlug, query, loading, filters, functionTags, onBack, onQueryChange, onFiltersChange, onSubmit }: Props) {
  return <section className="hero">
    {routeSlug && <button className="backLink" onClick={onBack}>← 返回菌种列表</button>}
    <p className="eyebrow">Microbial Knowledge Platform</p><h1>微生物百科</h1>
    <p className="summary">从菌种百科开始，逐步建设功能菌数据库、智能搜索和 AI 推荐能力。</p>
    <form className="searchBox" onSubmit={onSubmit}><input value={query} onChange={(e) => onQueryChange(e.target.value)} placeholder="搜索菌种、功能、应用场景，例如：枯草芽孢杆菌" /><button disabled={loading}>{loading ? '搜索中...' : '搜索'}</button></form>
    <section className="filterPanel"><div className="filterTitle"><strong>多条件筛选</strong><span>所有已填写条件同时满足</span></div><div className="filterGrid">
      <label><span>功能标签</span><select value={filters.functionTag} onChange={(e) => onFiltersChange({ ...filters, functionTag: e.target.value })}><option value="">全部功能</option>{functionTags.map((tag) => <option key={tag.id} value={tag.code}>{tag.name}</option>)}</select></label>
      <label><span>适宜温度 °C</span><input type="number" step="0.1" value={filters.temperature} onChange={(e) => onFiltersChange({ ...filters, temperature: e.target.value })} placeholder="例如 30" /></label>
      <label><span>适宜 pH</span><input type="number" min="0" max="14" step="0.1" value={filters.ph} onChange={(e) => onFiltersChange({ ...filters, ph: e.target.value })} placeholder="例如 7.0" /></label>
      <label><span>安全等级</span><select value={filters.safetyLevel} onChange={(e) => onFiltersChange({ ...filters, safetyLevel: e.target.value })}><option value="">全部等级</option>{['BSL-1', 'BSL-2', 'BSL-3', 'BSL-4'].map((level) => <option key={level} value={level}>{level}</option>)}</select></label>
      <label><span>来源环境</span><input value={filters.sourceEnvironment} onChange={(e) => onFiltersChange({ ...filters, sourceEnvironment: e.target.value })} placeholder="土壤、海洋、食品…" /></label>
    </div></section>
  </section>;
}
