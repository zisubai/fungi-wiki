import { fireEvent, render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { Species } from '../../types';
import { SpeciesDetailPanel, SpeciesListPanel } from './SpeciesPanels';

const species: Species = { id: 'species-1', slug: 'bacillus-test', latinName: 'Bacillus test', chineseName: '测试芽孢杆菌', strainNumber: 'CGMCC 1.1', sourceEnvironment: '土壤', safetyLevel: 'BSL-1', isModelOrganism: true, summary: '测试摘要', status: 'published', dataQualityScore: 90, createdAt: '2026-01-01', updatedAt: '2026-01-01' };
const emptyFilters = { functionTag: '', temperature: '', ph: '', safetyLevel: '', sourceEnvironment: '' };

describe('SpeciesPanels', () => {
  it('菌种列表支持选择、排序、分页和重置', () => {
    const onSelect = vi.fn(); const onSort = vi.fn(); const onPage = vi.fn(); const onReset = vi.fn();
    render(<SpeciesListPanel items={[species]} selectedId="species-1" loading={false} submittedQuery="芽孢" activeFilterCount={2} page={1} pageSize={10} total={21} sort="updated" appliedFilters={emptyFilters} onReset={onReset} onSort={onSort} onPage={onPage} onSelect={onSelect} />);
    expect(screen.getByText('关键词“芽孢” · 2 个筛选条件')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: /Bacillus test/ }));
    fireEvent.change(screen.getByRole('combobox', { name: '结果排序' }), { target: { value: 'quality' } });
    fireEvent.click(screen.getByRole('button', { name: '下一页' }));
    fireEvent.click(screen.getByRole('button', { name: '重置' }));
    expect(onSelect).toHaveBeenCalledWith('bacillus-test'); expect(onSort).toHaveBeenCalledWith('quality'); expect(onPage).toHaveBeenCalledWith(2); expect(onReset).toHaveBeenCalledOnce();
  });

  it('菌种详情展示别名、培养条件和文献证据', () => {
    render(<SpeciesDetailPanel selected={species} detailLoading={false} routeSlug="bacillus-test" aliases={[{ id: 'alias-1', name: '旧名', type: 'synonym', source: '' }]} functions={[{ functionTagId: 'tag-1', functionTagName: '生防' }]} conditions={[{ id: 'condition-1', mediumName: 'LB', temperatureMin: 25, temperatureMax: 37, phMin: 6, phMax: 8, oxygenRequirement: '好氧', cultureTime: '24 h', notes: '' }]} evidences={[{ id: 'evidence-1', title: 'Evidence paper', authors: '', journal: 'Nature', publicationYear: 2026, doi: '', pmid: '', sourceUrl: 'https://example.com/paper', conclusion: '支持生防功能', evidenceLevel: 'high', evidenceScore: 90 }]} />);
    expect(screen.getByText('旧名')).toBeInTheDocument(); expect(screen.getByText('LB')).toBeInTheDocument(); expect(screen.getByText('温度：25–37 °C')).toBeInTheDocument(); expect(screen.getByRole('link', { name: 'Evidence paper' })).toHaveAttribute('href', 'https://example.com/paper'); expect(screen.getByText('支持生防功能')).toBeInTheDocument();
  });

  it('未选择菌种时显示空状态', () => {
    render(<SpeciesDetailPanel selected={null} detailLoading={false} routeSlug="" aliases={[]} functions={[]} conditions={[]} evidences={[]} />);
    expect(screen.getByText('请选择一个菌种查看详情。')).toBeInTheDocument();
  });
});
