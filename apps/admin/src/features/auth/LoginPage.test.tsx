import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { authTokenKey } from '../../api';
import { LoginPage } from './LoginPage';

describe('LoginPage', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('登录成功后保存令牌并返回用户', async () => {
    const user = { id: '1', email: 'admin@fungi.local', displayName: '管理员', role: 'admin' as const };
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ token: 'token-1', expiresAt: '2099-01-01', user }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);
    const onLogin = vi.fn();
    render(<LoginPage onLogin={onLogin} />);
    fireEvent.change(screen.getByLabelText('密码'), { target: { value: 'password-123' } });
    fireEvent.click(screen.getByRole('button', { name: '登录' }));
    await waitFor(() => expect(onLogin).toHaveBeenCalledWith(user));
    expect(localStorage.getItem(authTokenKey)).toBe('token-1');
    expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining('/api/auth/login'), expect.objectContaining({ method: 'POST', body: JSON.stringify({ email: 'admin@fungi.local', password: 'password-123' }) }));
  });

  it('展示后端返回的登录错误', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ message: '邮箱或密码错误' }), { status: 401, headers: { 'Content-Type': 'application/json' } })));
    render(<LoginPage onLogin={vi.fn()} />);
    fireEvent.change(screen.getByLabelText('密码'), { target: { value: 'wrong-password' } });
    fireEvent.click(screen.getByRole('button', { name: '登录' }));
    expect(await screen.findByText(/邮箱或密码错误（请求 ID：.+）/)).toBeInTheDocument();
  });
});
