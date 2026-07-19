import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { ImportManagement } from './ImportManagement';

describe('ImportManagement', () => {
  beforeEach(() => {
    requestMock.mockReset();
    requestMock.mockImplementation((path: string) => path === '/api/admin/imports/species'
      ? Promise.resolve({ id: 'batch-1', sourceFilename: 'species.csv', totalRows: 2, successRows: 1, failedRows: 1, status: 'completed', createdAt: '2026-01-01', rows: [{ rowNumber: 2, slug: 'valid-row', status: 'imported' }, { rowNumber: 3, slug: '', status: 'failed', errors: ['拉丁名不能为空'] }] })
      : Promise.resolve({ items: [] }));
  });

  it('上传文件并展示逐行校验结果', async () => {
    const { container } = render(<ImportManagement />);
    const input = container.querySelector('input[type="file"]') as HTMLInputElement;
    fireEvent.change(input, { target: { files: [new File(['slug,latin_name'], 'species.csv', { type: 'text/csv' })] } });
    fireEvent.click(screen.getByRole('button', { name: '开始导入并提交审核' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/imports/species', expect.objectContaining({ method: 'POST', body: expect.any(FormData) })));
    expect(await screen.findByText('已导入待审核')).toBeInTheDocument();
    expect(screen.getByText('拉丁名不能为空')).toBeInTheDocument();
    expect(screen.getByText(/成功 1 行，失败 1 行/)).toBeInTheDocument();
  });
});
