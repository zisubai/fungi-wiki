import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Evidence } from '../../types';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { EvidenceManagement } from './EvidenceManagement';

const species = [
  { id: 'species-1', slug: 'bacillus-test', latinName: 'Bacillus test' },
  { id: 'species-2', slug: 'yeast-test', latinName: 'Yeast test' },
];
const existingEvidence: Evidence = { id: 'evidence-1', title: 'Existing paper', authors: '', journal: 'Nature', publicationYear: 2025, doi: '', pmid: '', sourceUrl: '', conclusion: '已有结论', evidenceLevel: 'high', evidenceScore: 90 };

describe('EvidenceManagement', () => {
  let evidences: Evidence[];

  beforeEach(() => {
    evidences = [existingEvidence];
    requestMock.mockReset();
    requestMock.mockImplementation((path: string, options?: RequestInit) => {
      if (path === '/api/admin/species?limit=100') return Promise.resolve({ items: species });
      if (path.endsWith('/evidences') && options?.method === 'POST') return Promise.resolve({ id: 'evidence-2' });
      if (path.includes('/evidences/') && options?.method === 'DELETE') { evidences = []; return Promise.resolve(undefined); }
      if (path.endsWith('/evidences')) return Promise.resolve({ items: evidences });
      return Promise.resolve(undefined);
    });
  });

  it('将文献证据关联到当前菌种并规范化年份', async () => {
    render(<EvidenceManagement />);
    await screen.findByText('Existing paper');
    fireEvent.change(screen.getByLabelText('文献标题 *'), { target: { value: 'New paper' } });
    fireEvent.change(screen.getByLabelText('年份'), { target: { value: '2026' } });
    fireEvent.change(screen.getByLabelText('实验结论 *'), { target: { value: '支持生防功能' } });
    fireEvent.click(screen.getByRole('button', { name: '保存证据' }));

    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/bacillus-test/evidences', expect.objectContaining({
      method: 'POST',
      body: expect.stringContaining('"publicationYear":2026'),
    })));
    expect(await screen.findByText('文献证据已添加')).toBeInTheDocument();
  });

  it('切换菌种后按当前关联删除证据', async () => {
    vi.spyOn(window, 'confirm').mockReturnValue(true);
    render(<EvidenceManagement />);
    await screen.findByText('Existing paper');
    fireEvent.change(screen.getByRole('combobox', { name: '关联菌种' }), { target: { value: 'yeast-test' } });
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/yeast-test/evidences'));
    fireEvent.click(screen.getByRole('button', { name: '删除' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/species/yeast-test/evidences/evidence-1', { method: 'DELETE' }));
    expect(await screen.findByText('暂无文献证据。')).toBeInTheDocument();
  });
});
