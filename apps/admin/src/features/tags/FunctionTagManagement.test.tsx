import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { FunctionTag } from '../../types';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { FunctionTagManagement } from './FunctionTagManagement';

const tag: FunctionTag = { id: 'tag-1', parentId: '', name: '生防', code: 'biocontrol', description: '抑制病原菌', sortOrder: 10, createdAt: '2026-01-01', updatedAt: '2026-01-01' };

describe('FunctionTagManagement', () => {
  let items: FunctionTag[];
  beforeEach(() => {
    items = [tag];
    requestMock.mockReset();
    requestMock.mockImplementation((path: string, options?: RequestInit) => {
      if (path.startsWith('/api/admin/function-tags?')) return Promise.resolve({ items });
      if (path === '/api/admin/function-tags' && options?.method === 'POST') return Promise.resolve(tag);
      if (path === '/api/admin/function-tags/biocontrol' && options?.method === 'DELETE') { items = []; return Promise.resolve(undefined); }
      return Promise.resolve(undefined);
    });
  });

  it('提交标准化功能标签', async () => {
    render(<FunctionTagManagement />);
    await screen.findByText('biocontrol');
    fireEvent.change(screen.getByLabelText('名称 *'), { target: { value: '固氮' } });
    fireEvent.change(screen.getByLabelText('编码 *'), { target: { value: 'nitrogen-fixation' } });
    fireEvent.change(screen.getByLabelText('排序'), { target: { value: '20' } });
    fireEvent.click(screen.getByRole('button', { name: '保存标签' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/function-tags', {
      method: 'POST',
      body: JSON.stringify({ parentId: '', name: '固氮', code: 'nitrogen-fixation', description: '', sortOrder: 20 }),
    }));
    expect(await screen.findByText('功能标签已新增')).toBeInTheDocument();
  });

  it('确认后删除功能标签并刷新列表', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    render(<FunctionTagManagement />);
    fireEvent.click(await screen.findByRole('button', { name: '删除' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/function-tags/biocontrol', { method: 'DELETE' }));
    expect(await screen.findByText('暂无功能标签。')).toBeInTheDocument();
  });
});
