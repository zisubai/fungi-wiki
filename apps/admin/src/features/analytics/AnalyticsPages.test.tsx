import { act, fireEvent, render, screen } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { RecommendationQualityPage, SearchAnalytics } from './AnalyticsPages';

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
    requestMock.mockResolvedValue({ total: 4, helpful: 3, unhelpful: 1, records: [{ id: 'r1', requirement: '土壤生防', parsedIntent: {}, items: [{ latinName: 'Bacillus subtilis', score: 92 }], modelName: 'rules-v1', riskLevel: 'low', helpfulCount: 3, unhelpfulCount: 1, createdAt: '2026-01-01' }] });
    render(<RecommendationQualityPage />);
    expect(await screen.findByText('75%')).toBeInTheDocument();
    expect(screen.getByText('Bacillus subtilis · 92 分')).toBeInTheDocument();
    await act(async () => { fireEvent.click(screen.getByRole('button', { name: '刷新' })); });
    expect(requestMock).toHaveBeenCalledTimes(2);
  });

  it('展示空搜索数据状态', async () => {
    requestMock.mockResolvedValue({ days: 30, totalSearches: 0, noResultSearches: 0, distinctQueries: 0, popularQueries: [], noResultQueries: [] });
    render(<SearchAnalytics />);
    expect(await screen.findByText('暂无关键词搜索记录。')).toBeInTheDocument();
    expect(screen.getByText('暂无无结果关键词。')).toBeInTheDocument();
  });
});
