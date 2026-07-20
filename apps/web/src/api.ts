const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export class ApiError extends Error {
  constructor(message: string, public readonly status: number, public readonly requestId: string) {
    super(requestId ? `${message}（请求 ID：${requestId}）` : message);
    this.name = 'ApiError';
  }
}

function createRequestId() {
  return globalThis.crypto?.randomUUID?.() ?? `web-${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`;
}

export async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const requestId = new Headers(options?.headers).get('X-Request-ID') ?? createRequestId();
  let response: Response;
  try {
    response = await fetch(`${apiBaseUrl}${path}`, {
      ...options,
      headers: { 'Content-Type': 'application/json', 'X-Request-ID': requestId, ...options?.headers },
    });
  } catch (cause) {
    throw new ApiError(cause instanceof Error ? cause.message : '网络请求失败', 0, requestId);
  }
  if (!response.ok) {
    const message = await response.json().catch(() => undefined);
    throw new ApiError(message?.message ?? `请求失败：${response.status}`, response.status, response.headers.get('X-Request-ID') ?? requestId);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
