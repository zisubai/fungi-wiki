import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { AuditManagement } from './AuditManagement';

describe('AuditManagement', () => {
  beforeEach(() => {
    requestMock.mockReset();
    let loads = 0;
    requestMock.mockImplementation((path: string) => {
      if (path === '/api/admin/audits?status=pending') {
        loads += 1;
        return Promise.resolve({ items: loads === 1 ? [{ id: 'audit-1', entityId: 'species-1', entityName: 'Bacillus test', action: 'publish', status: 'pending', comment: '', submittedAt: '2026-01-01' }] : [] });
      }
      return Promise.resolve(undefined);
    });
    vi.spyOn(window, 'prompt').mockReturnValue('证据完整');
  });

  it('审核通过后发布并刷新待审列表', async () => {
    render(<AuditManagement />);
    fireEvent.click(await screen.findByRole('button', { name: '通过并发布' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/audits/audit-1/approve', { method: 'POST', body: JSON.stringify({ comment: '证据完整' }) }));
    expect(await screen.findByText('已审核通过并发布')).toBeInTheDocument();
    expect(await screen.findByText('没有待审核数据。')).toBeInTheDocument();
  });
});
