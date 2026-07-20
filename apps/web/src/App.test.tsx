import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('./api', () => ({ request: requestMock }));
import { App } from './App';

const species = { id: 'species-1', slug: 'bacillus-test', latinName: 'Bacillus test', chineseName: '测试芽孢杆菌', strainNumber: 'CGMCC 1.1', sourceEnvironment: '土壤', safetyLevel: 'BSL-1', isModelOrganism: true, summary: '具有生防潜力', status: 'published', dataQualityScore: 90, createdAt: '2026-01-01', updatedAt: '2026-01-01' };

describe('public App', () => {
  beforeEach(() => {
    requestMock.mockReset();
    requestMock.mockImplementation((path: string, options?: RequestInit) => {
      if (path.startsWith('/api/species?')) return Promise.resolve({ items: [species], total: 1, limit: 10, offset: 0 });
      if (path === '/api/function-tags?limit=200') return Promise.resolve({ items: [{ id: 'tag-1', name: '生防', code: 'biocontrol' }, { id: 'tag-2', name: '发酵', code: 'fermentation' }] });
      if (path === '/api/species/bacillus-test') return Promise.resolve(species);
      if (path.endsWith('/functions')) return Promise.resolve({ items: [{ functionTagId: 'tag-1', functionTagName: '生防' }] });
      if (path.endsWith('/culture-conditions')) return Promise.resolve({ items: [] });
      if (path.endsWith('/evidences')) return Promise.resolve({ items: [] });
      if (path.endsWith('/aliases')) return Promise.resolve({ items: [] });
      if (path === '/api/recommendations/combinations' && options?.method === 'POST') return Promise.resolve({ recordId: 'combination-1', items: [{ members: [{ id: 'species-1', slug: 'bacillus-test', latinName: 'Bacillus test', chineseName: '测试菌', safetyLevel: 'BSL-1', functionTags: ['biocontrol'], evidenceCount: 1 }, { id: 'species-2', slug: 'yeast', latinName: 'Saccharomyces cerevisiae', chineseName: '酿酒酵母', safetyLevel: 'BSL-1', functionTags: ['fermentation'], evidenceCount: 1 }], score: 95, temperatureMin: 25, temperatureMax: 30, phMin: 6, phMax: 6, compatible: true, reasons: ['功能互补'] }], disclaimer: '需要实验验证' });
      if (path === '/api/recommendations/combinations/combination-1/feedback' && options?.method === 'POST') return Promise.resolve({ message: 'ok' });
      if (path === '/api/recommendations' && options?.method === 'POST') {
        if (String(options.body).includes('不存在')) return Promise.resolve({ recordId: 'recommend-empty', parsedIntent: { sourceEnvironment: '不存在' }, items: [], relaxationSuggestions: ['暂无匹配来源环境的菌种，可取消来源限制。'], disclaimer: '仅供科研参考' });
        return Promise.resolve({ recordId: 'recommend-1', parsedIntent: { functionTag: 'biocontrol', temperature: 30, safetyLevel: 'BSL-1' }, items: [{ ...species, score: 92, evidenceCount: 2, evidenceReferences: [{ id: 'evidence-1', title: '生防功能研究', publicationYear: 2026, doi: '10.1000/test', conclusion: '支持生防功能', evidenceLevel: 'high', evidenceScore: 90 }], reasons: ['功能标签匹配'], riskWarning: '' }], disclaimer: '仅供科研参考' });
      }
      return Promise.resolve(undefined);
    });
  });

  it('推荐无结果时展示放宽建议', async () => {
    render(<App />);
    fireEvent.change(screen.getByPlaceholderText(/寻找适合 30°C/), { target: { value: '寻找不存在环境的菌种' } });
    fireEvent.click(screen.getByRole('button', { name: '生成推荐' }));
    expect(await screen.findByText('暂无匹配来源环境的菌种，可取消来源限制。')).toBeInTheDocument();
  });

  it('生成双功能组合菌建议', async () => {
    vi.spyOn(window, 'prompt').mockReturnValue('培养窗口清晰');
    render(<App />);
    await screen.findAllByText('生防', { selector: 'option' });
    fireEvent.change(screen.getByLabelText('第一个目标功能'), { target: { value: 'biocontrol' } });
    fireEvent.change(screen.getByLabelText('第二个目标功能'), { target: { value: 'fermentation' } });
    fireEvent.click(screen.getByRole('button', { name: '生成组合建议' }));
    expect(await screen.findByText('存在共同培养窗口')).toBeInTheDocument();
    expect(screen.getByText('共同温度：25–30 °C · 共同 pH：6–6')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '👍 有帮助' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/recommendations/combinations/combination-1/feedback', { method: 'POST', body: JSON.stringify({ feedbackType: 'helpful', content: '培养窗口清晰' }) }));
  });

  it('组合关键词、功能、温度、pH、安全等级和来源环境进行搜索', async () => {
    render(<App />);
    await screen.findAllByText('Bacillus test');
    fireEvent.change(screen.getByPlaceholderText(/搜索菌种、功能/), { target: { value: '芽孢杆菌' } });
    fireEvent.change(screen.getByLabelText('功能标签'), { target: { value: 'biocontrol' } });
    fireEvent.change(screen.getByLabelText('适宜温度 °C'), { target: { value: '30' } });
    fireEvent.change(screen.getByLabelText('适宜 pH'), { target: { value: '7' } });
    fireEvent.change(screen.getByLabelText('安全等级'), { target: { value: 'BSL-1' } });
    fireEvent.change(screen.getByLabelText('来源环境'), { target: { value: '土壤' } });
    fireEvent.click(screen.getByRole('button', { name: '搜索' }));

    await waitFor(() => {
      const searchPath = requestMock.mock.calls.map(([path]) => String(path)).find((path) => path.includes('q=%E8%8A%BD%E5%AD%A2%E6%9D%86%E8%8F%8C'));
      expect(searchPath).toContain('functionTag=biocontrol');
      expect(searchPath).toContain('temperature=30');
      expect(searchPath).toContain('ph=7');
      expect(searchPath).toContain('safetyLevel=BSL-1');
      expect(searchPath).toContain('sourceEnvironment=%E5%9C%9F%E5%A3%A4');
    });
  });

  it('提交可解释推荐并发送有用反馈', async () => {
    vi.spyOn(window, 'prompt').mockReturnValue('候选理由清晰');
    const { container } = render(<App />);
    await screen.findAllByText('生防', { selector: 'option' });
    fireEvent.change(screen.getByPlaceholderText(/寻找适合 30°C/), { target: { value: '寻找土壤生防菌' } });
    fireEvent.change(container.querySelector('.recommendForm select')!, { target: { value: 'biocontrol' } });
    fireEvent.click(screen.getByRole('button', { name: '生成推荐' }));

    expect(await screen.findByText('功能标签匹配')).toBeInTheDocument();
    expect(screen.getByText('温度：30 °C')).toBeInTheDocument();
    fireEvent.click(screen.getByText('查看推荐证据（1 条）'));
    expect(screen.getByText('生防功能研究')).toBeInTheDocument();
    const recommendCall = requestMock.mock.calls.find(([path, options]) => path === '/api/recommendations' && options?.method === 'POST');
    expect(JSON.parse(String(recommendCall?.[1]?.body))).toEqual(expect.objectContaining({ requirement: '寻找土壤生防菌', functionTag: 'biocontrol', limit: 5 }));
    fireEvent.click(screen.getByRole('button', { name: '👍 有帮助' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/recommendations/recommend-1/feedback', { method: 'POST', body: JSON.stringify({ feedbackType: 'helpful', content: '候选理由清晰' }) }));
    expect(await screen.findByText('感谢反馈，我们会用于改进推荐质量。')).toBeInTheDocument();
  });
});
