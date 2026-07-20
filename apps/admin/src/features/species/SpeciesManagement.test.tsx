import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { SpeciesManagement } from './SpeciesManagement';

const existingSpecies = { id: 'species-1', slug: 'bacillus-test', latinName: 'Bacillus test', chineseName: '测试芽孢杆菌', strainNumber: 'CGMCC 1.1', sourceEnvironment: '土壤', safetyLevel: 'BSL-1', isModelOrganism: false, summary: '测试菌种', status: 'draft', dataQualityScore: 80, createdAt: '2026-01-01', updatedAt: '2026-01-01' };

describe('SpeciesManagement', () => {
  beforeEach(() => {
    requestMock.mockReset();
    requestMock.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/api/function-tags?limit=200') return Promise.resolve({ items: [{ id: 'tag-1', parentId: '', name: '生防', code: 'biocontrol', description: '', sortOrder: 1, createdAt: '', updatedAt: '' }] });
      if (path.startsWith('/api/admin/species?')) return Promise.resolve({ items: [] });
      if (path === '/api/admin/species' && options?.method === 'POST') return Promise.resolve({ id: 'species-1', slug: 'bacillus-test' });
      return Promise.resolve(undefined);
    });
  });

  it('创建菌种后同步标签、培养条件和别名', async () => {
    render(<SpeciesManagement />);
    fireEvent.change(screen.getByLabelText('Slug *'), { target: { value: 'bacillus-test' } });
    fireEvent.change(screen.getByLabelText('拉丁名 *'), { target: { value: 'Bacillus test' } });
    fireEvent.change(screen.getByLabelText('别名 / 同义词'), { target: { value: '旧名一\n旧名二' } });
    fireEvent.change(screen.getByLabelText('培养基'), { target: { value: 'LB' } });
    fireEvent.click(await screen.findByLabelText('生防'));
    fireEvent.click(screen.getByRole('button', { name: '保存菌种' }));

    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test/functions', expect.objectContaining({ method: 'PUT', body: JSON.stringify({ items: [{ functionTagId: 'tag-1' }] }) })));
    expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test/culture-conditions', expect.objectContaining({ body: expect.stringContaining('"mediumName":"LB"') }));
    expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test/aliases', expect.objectContaining({ body: JSON.stringify({ items: [{ name: '旧名一', type: 'synonym' }, { name: '旧名二', type: 'synonym' }] }) }));
    expect(await screen.findByText('菌种已新增')).toBeInTheDocument();
  });

  it('编辑菌种时加载并保存全部关联数据', async () => {
    requestMock.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/api/function-tags?limit=200') return Promise.resolve({ items: [{ id: 'tag-1', parentId: '', name: '生防', code: 'biocontrol', description: '', sortOrder: 1, createdAt: '', updatedAt: '' }] });
      if (path.startsWith('/api/admin/species?')) return Promise.resolve({ items: [existingSpecies] });
      if (path.endsWith('/functions') && !options) return Promise.resolve({ items: [{ functionTagId: 'tag-1', functionTagName: '生防' }] });
      if (path.endsWith('/culture-conditions') && !options) return Promise.resolve({ items: [{ mediumName: 'LB', temperatureMin: 25, temperatureMax: 37, phMin: 6, phMax: 8, oxygenRequirement: '好氧', cultureTime: '24 h', notes: '' }] });
      if (path.endsWith('/aliases') && !options) return Promise.resolve({ items: [{ id: 'alias-1', name: '旧名', type: 'synonym', source: '' }] });
      if (path === '/api/admin/species/bacillus-test' && options?.method === 'PUT') return Promise.resolve(existingSpecies);
      return Promise.resolve(undefined);
    });
    render(<SpeciesManagement />);
    fireEvent.click(await screen.findByRole('button', { name: '编辑' }));
    expect(await screen.findByDisplayValue('旧名')).toBeInTheDocument();
    expect(screen.getByDisplayValue('LB')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('摘要'), { target: { value: '更新后的摘要' } });
    fireEvent.click(screen.getByRole('button', { name: '保存菌种' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test', expect.objectContaining({ method: 'PUT', body: expect.stringContaining('更新后的摘要') })));
    expect(await screen.findByText('菌种已更新')).toBeInTheDocument();
  });

  it('支持提交发布审核和归档', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    requestMock.mockImplementation((path: string) => {
      if (path === '/api/function-tags?limit=200') return Promise.resolve({ items: [] });
      if (path.startsWith('/api/admin/species?')) return Promise.resolve({ items: [existingSpecies] });
      return Promise.resolve(undefined);
    });
    render(<SpeciesManagement />);
    fireEvent.click(await screen.findByRole('button', { name: '提交审核' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/audits/species/bacillus-test/submit', { method: 'POST', body: '{}' }));
    expect(await screen.findByText('已提交审核')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: '归档' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test', { method: 'DELETE' }));
    expect(await screen.findByText('菌种已归档')).toBeInTheDocument();
  });
});
