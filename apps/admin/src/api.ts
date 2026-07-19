const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';

export const authTokenKey = 'fungi_admin_token';

export async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const isFormData = options?.body instanceof FormData;
  const token = localStorage.getItem(authTokenKey);
  const response = await fetch(`${apiBaseUrl}${path}`, {
    headers: { ...(!isFormData ? { 'Content-Type': 'application/json' } : {}), ...(token ? { Authorization: `Bearer ${token}` } : {}), ...options?.headers },
    ...options,
  });
  if (!response.ok) {
    if (response.status === 401 && path !== '/api/auth/login') {
      localStorage.removeItem(authTokenKey);
      window.dispatchEvent(new Event('auth-expired'));
    }
    const message = await response.json().catch(() => undefined);
    throw new Error(message?.message ?? `请求失败：${response.status}`);
  }
  if (response.status === 204) return undefined as T;
  return response.json() as Promise<T>;
}
