import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { SpeciesManagement } from './SpeciesManagement';

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
});
