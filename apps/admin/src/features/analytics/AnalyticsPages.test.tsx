import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { DataQualityPage, RecommendationQualityPage, SearchAnalytics } from './AnalyticsPages';

describe('Analytics pages', () => {
  beforeEach(() => requestMock.mockReset());

  it('展示搜索指标并按时间范围重新加载', async () => {
    requestMock.mockResolvedValue({ days: 30, totalSearches: 120, noResultSearches: 8, distinctQueries: 32, popularQueries: [{ query: '生防菌', count: 18, averageResults: 4.5 }], noResultQueries: [{ query: '耐高盐固氮菌', count: 3 }] });
    render(<SearchAnalytics />);
    expect(await screen.findByText('120')).toBeInTheDocument();
    expect(screen.getByText('生防菌')).toBeInTheDocument();
    expect(screen.getByText('耐高盐固氮菌')).toBeInTheDocument();
    await act(async () => { fireEvent.change(screen.getByRole('combobox'), { target: { value: '7' } }); });
    expect(requestMock).toHaveBeenCalledWith('/api/admin/search-analytics?days=7');
  });

  it('计算推荐有帮助率并展示候选结果', async () => {
    requestMock.mockResolvedValue({ total: 4, helpful: 3, unhelpful: 1, records: [{ id: 'r1', requirement: '土壤生防', parsedIntent: {}, items: [{ latinName: 'Bacillus subtilis', score: 92 }], modelName: 'rules-v1', riskLevel: 'low', helpfulCount: 3, unhelpfulCount: 1, createdAt: '2026-01-01' }], combinations: [{ id: 'c1', functionTags: ['biocontrol', 'fermentation'], safetyLevel: 'BSL-1', items: [{ members: [{ latinName: 'Bacillus subtilis' }, { latinName: 'Saccharomyces cerevisiae' }], score: 100, compatible: true }], modelName: 'combination-rules-v2', riskLevel: 'low', helpfulCount: 2, unhelpfulCount: 0, createdAt: '2026-01-02' }] });
    render(<RecommendationQualityPage />);
    expect(await screen.findByText('75%')).toBeInTheDocument();
    expect(screen.getByText('Bacillus subtilis · 92 分')).toBeInTheDocument();
    expect(screen.getByText('#1 Bacillus subtilis + Saccharomyces cerevisiae · 100 分 · 兼容')).toBeInTheDocument();
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新' })); });
    expect(requestMock).toHaveBeenCalledTimes(2);
  });

  it('提交组合菌共培养实验结果', async () => {
    requestMock
      .mockResolvedValueOnce({ total: 0, helpful: 0, unhelpful: 0, records: [], combinations: [{ id: 'c1', functionTags: ['biocontrol', 'fermentation'], safetyLevel: 'BSL-1', items: [{ members: [{ latinName: 'Bacillus subtilis' }, { latinName: 'Saccharomyces cerevisiae' }], score: 90, compatible: true }], modelName: 'combination-rules-v2', riskLevel: 'low', helpfulCount: 0, unhelpfulCount: 0, experiments: [], createdAt: '2026-01-02' }] })
      .mockResolvedValueOnce({ id: 'e1' })
      .mockResolvedValueOnce({ total: 0, helpful: 0, unhelpful: 0, records: [], combinations: [] });
    render(<RecommendationQualityPage />);
    expect(await screen.findByText('实验历史（0）')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('实验温度'), { target: { value: '30' } });
    fireEvent.change(screen.getByLabelText('实验 pH'), { target: { value: '6.5' } });
    fireEvent.change(screen.getByLabelText('实验备注'), { target: { value: '共培养 48 小时无拮抗圈' } });
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '保存结果' })); });
    expect(requestMock).toHaveBeenCalledWith('/api/admin/recommendations/combinations/c1/experiments', expect.objectContaining({
      method: 'POST', body: JSON.stringify({ candidateIndex: 0, outcome: 'compatible', temperature: 30, ph: 6.5, notes: '共培养 48 小时无拮抗圈' }),
    }));
  });

  it('展示空搜索数据状态', async () => {
    requestMock.mockResolvedValue({ days: 30, totalSearches: 0, noResultSearches: 0, distinctQueries: 0, popularQueries: [], noResultQueries: [] });
    render(<SearchAnalytics />);
    expect(await screen.findByText('暂无关键词搜索记录。')).toBeInTheDocument();
    expect(screen.getByText('暂无无结果关键词。')).toBeInTheDocument();
  });

  it('展示数据质量分布和高频缺失项', async () => {
    requestMock.mockResolvedValue({ total: 10, averageScore: 62.5, complete: 2, needsCompletion: 3, incomplete: 5, missing: [{ key: 'evidences', label: '文献证据', count: 8 }], prioritySpecies: [{ id: 's1', slug: 'bacillus-test', latinName: 'Bacillus test', status: 'draft', score: 35 }] });
    render(<DataQualityPage />);
    expect(await screen.findByText('62.5')).toBeInTheDocument();
    expect(screen.getByText('8 个缺失')).toBeInTheDocument();
    expect(screen.getByText('Bacillus test')).toBeInTheDocument();
    expect(requestMock).toHaveBeenCalledWith('/api/admin/data-quality');
  });
});
