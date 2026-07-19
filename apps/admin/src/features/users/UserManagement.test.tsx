import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const { requestMock } = vi.hoisted(() => ({ requestMock: vi.fn() }));
vi.mock('../../api', () => ({ request: requestMock }));
import { UserManagement } from './UserManagement';

describe('UserManagement', () => {
  beforeEach(() => {
    requestMock.mockReset();
    requestMock.mockImplementation((path: string) => path === '/api/admin/users' ? Promise.resolve({ items: [] }) : Promise.resolve(undefined));
  });

  it('创建专家账号并提交所选角色', async () => {
    render(<UserManagement />);
    await screen.findByText('0 个');
    fireEvent.change(screen.getByLabelText('显示名称 *'), { target: { value: '审核专家' } });
    fireEvent.change(screen.getByLabelText('邮箱 *'), { target: { value: 'expert@fungi.local' } });
    fireEvent.change(screen.getByLabelText('初始密码 *'), { target: { value: 'safe-password' } });
    fireEvent.change(screen.getByLabelText('角色'), { target: { value: 'expert' } });
    fireEvent.click(screen.getByRole('button', { name: '创建账号' }));
    await waitFor(() => expect(requestMock).toHaveBeenCalledWith('/api/admin/users', expect.objectContaining({ method: 'POST' })));
    const createCall = requestMock.mock.calls.find(([, options]) => options?.method === 'POST');
    expect(JSON.parse(String(createCall?.[1]?.body))).toEqual({ displayName: '审核专家', email: 'expert@fungi.local', password: 'safe-password', role: 'expert' });
    expect(await screen.findByText('账号已创建')).toBeInTheDocument();
  });
});
