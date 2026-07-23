import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { DataAssetsManagement } from './DataAssetsManagement';

const species = [{ id: 'species-1', slug: 'bacillus-test', latinName: 'Bacillus test', chineseName: '测试芽孢杆菌' }];
const existingCase = { id: 'case-1', speciesId: 'species-1', industry: '农业', scenario: '土传病害防控', problem: '病害', solution: '施用菌剂', resultSummary: '发病率下降', maturityLevel: '中试', source: '试验报告', createdAt: '', updatedAt: '' };
const version = { id: 'version-1', speciesId: 'species-1', versionNumber: 2, changeType: 'insert', sourceTable: 'application_cases', snapshot: {}, createdAt: '2026-07-23T05:00:00Z' };

describe('DataAssetsManagement', () => {
  let cases = [existingCase];
  beforeEach(() => {
    cases = [existingCase]; requestMock.mockReset();
    requestMock.mockImplementation((path: string, options?: RequestInit) => {
      if (path.startsWith('/api/admin/species?')) return Promise.resolve({ items: species });
      if (path.endsWith('/versions')) return Promise.resolve({ items: [version] });
      if (path.endsWith('/application-cases') && options?.method === 'POST') return Promise.resolve({ id: 'case-2' });
      if (path.includes('/application-cases/') && options?.method === 'PUT') return Promise.resolve(existingCase);
      if (path.includes('/application-cases/') && options?.method === 'DELETE') { cases = []; return Promise.resolve(undefined); }
      if (path.endsWith('/application-cases')) return Promise.resolve({ items: cases });
      return Promise.resolve(undefined);
    });
  });

  it('展示应用案例与版本并新增案例', async () => {
    render(<DataAssetsManagement />); await screen.findByText('土传病害防控'); expect(screen.getByText('v2')).toBeInTheDocument();
    fireEvent.change(screen.getByLabelText('行业 *'), { target: { value: '环保' } }); fireEvent.change(screen.getByLabelText('应用场景 *'), { target: { value: '废水处理' } }); fireEvent.click(screen.getByRole('button', { name: '保存案例' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test/application-cases', expect.objectContaining({ method: 'POST', body: expect.stringContaining('废水处理') })));
    expect(await screen.findByText('应用案例已新增')).toBeInTheDocument();
  });

  it('编辑并删除已有案例', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true); render(<DataAssetsManagement />); await screen.findByText('土传病害防控'); fireEvent.click(screen.getByRole('button', { name: '编辑' }));
    fireEvent.change(screen.getByLabelText('应用场景 *'), { target: { value: '更新场景' } }); fireEvent.click(screen.getByRole('button', { name: '保存案例' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test/application-cases/case-1', expect.objectContaining({ method: 'PUT' })));
    fireEvent.click(screen.getByRole('button', { name: '删除' })); await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test/application-cases/case-1', { method: 'DELETE' }));
  });
});
