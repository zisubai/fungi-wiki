import { afterEach, describe, expect, it, vi } from 'vitest';
import { ApiError, request } from './api';

describe('public request', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('合并 JSON 与自定义请求头并解析响应', async () => {
    const fetchMock = vi.fn().mockResolvedValue(new Response(JSON.stringify({ ok: true }), { status: 200, headers: { 'Content-Type': 'application/json' } }));
    vi.stubGlobal('fetch', fetchMock);
    await expect(request('/api/example', { method: 'POST', headers: { 'X-Trace-ID': 'trace-1', 'X-Request-ID': 'request-web-1' }, body: '{}' })).resolves.toEqual({ ok: true });
    expect(fetchMock).toHaveBeenCalledWith('http://localhost:8080/api/example', { method: 'POST', body: '{}', headers: { 'Content-Type': 'application/json', 'X-Request-ID': 'request-web-1', 'X-Trace-ID': 'trace-1' } });
  });

  it('正确处理 204 空响应', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(null, { status: 204 })));
    await expect(request('/api/feedback', { method: 'POST' })).resolves.toBeUndefined();
  });

  it('优先展示后端错误消息并回退到状态码', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValueOnce(new Response(JSON.stringify({ message: '参数错误' }), { status: 400, headers: { 'Content-Type': 'application/json' } })).mockResolvedValueOnce(new Response('unavailable', { status: 503 })));
    await expect(request('/api/example')).rejects.toThrow('参数错误');
    await expect(request('/api/example')).rejects.toThrow('请求失败：503');
  });

  it('错误对象携带可检索的请求 ID', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response(JSON.stringify({ message: '推荐失败' }), { status: 500, headers: { 'Content-Type': 'application/json', 'X-Request-ID': 'trace-web-1' } })));
    const error = await request('/api/recommendations').catch((cause) => cause);
    expect(error).toBeInstanceOf(ApiError);
    expect(error).toMatchObject({ status: 500, requestId: 'trace-web-1', message: '推荐失败（请求 ID：trace-web-1）' });
  });

  it('网络中断时仍返回调用前确定的请求 ID', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('Network unavailable')));
    const error = await request('/api/species', { headers: { 'X-Request-ID': 'offline-web-1' } }).catch((cause) => cause);
    expect(error).toMatchObject({ status: 0, requestId: 'offline-web-1', message: 'Network unavailable（请求 ID：offline-web-1）' });
  });
});
