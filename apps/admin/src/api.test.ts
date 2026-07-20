import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError, authTokenKey, request } from './api';

describe('request', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('同时发送认证头、JSON 头和调用方自定义头', async () => {
    localStorage.setItem(authTokenKey, 'token-1');
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);

    await request('/api/example', { method: 'POST', headers: { 'X-Trace-ID': 'trace-1', 'X-Request-ID': 'request-1' }, body: '{}' });

    expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/example', expect.objectContaining({
      method: 'POST',
      headers: { 'Content-Type': 'application/json', Authorization: 'Bearer token-1', 'X-Trace-ID': 'trace-1', 'X-Request-ID': 'request-1' },
    }));
  });

  it('上传 FormData 时不手动设置 Content-Type', async () => {
    localStorage.setItem(authTokenKey, 'token-1');
    const fetchMock = vi.fn().mockResolvedValue(new Response(null, { status: 204 }));
    vi.stubGlobal('fetch', fetchMock);
    const body = new FormData(); body.append('file', new File(['data'], 'species.csv'));

    await request('/api/upload', { method: 'POST', body });

    expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/upload', expect.objectContaining({ body, headers: expect.objectContaining({ Authorization: 'Bearer token-1', 'X-Request-ID': expect.any(String) }) }));
  });

  it('非登录接口返回 401 时清除令牌并发出过期事件', async () => {
    localStorage.setItem(authTokenKey, 'expired-token');
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ message: '登录已过期' }), { status: 401, headers: { 'Content-Type': 'application/json' } })));
    const listener = vi.fn(); window.addEventListener('auth-expired', listener);

    await expect(request('/api/admin/species')).rejects.toThrow('登录已过期');

    expect(localStorage.getItem(authTokenKey)).toBeNull();
    expect(listener).toHaveBeenCalledOnce();
    window.removeEventListener('auth-expired', listener);
  });

  it('无 JSON 错误体时返回包含状态码的默认消息', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('service unavailable', { status: 503 })));
    await expect(request('/api/health')).rejects.toThrow('请求失败：503');
  });

  it('错误对象包含状态码和后端请求 ID', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ message: '保存失败' }), { status: 422, headers: { 'Content-Type': 'application/json', 'X-Request-ID': 'trace-admin-1' } })));
    const error = await request('/api/admin/species').catch((cause) => cause);
    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 422, requestId: 'trace-admin-1', message: '保存失败（请求 ID：trace-admin-1）' });
  });

  it('网络失败时保留前端生成的请求 ID', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Failed to fetch')));
    const error = await request('/api/admin/species', { headers: { 'X-Request-ID': 'offline-admin-1' } }).catch((cause) => cause);
    expect(error).toMatchObject({ status: 0, requestId: 'offline-admin-1', message: 'Failed to fetch（请求 ID：offline-admin-1）' });
  });
});
