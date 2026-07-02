import { PUBLIC_API_URL } from '$env/static/public';
import { error } from '@sveltejs/kit';
import { HTTP_STATUS } from '$lib/http';

export function createApiClient(fetchFn: typeof fetch = fetch) {
    async function request<T>(path: string, options?: RequestInit): Promise<Response> {
        const res = await fetchFn(`${PUBLIC_API_URL}${path}`, {
            headers: {
                'Content-Type': 'application/json',
                ...options?.headers
            },
            ...options
        });

        if (!res.ok) {
            error(res.status, `Request failed: ${res.status}`);
        }

        return res;
    }

    async function jsonRequest<T>(path: string, options?: RequestInit): Promise<T> {
        try {
            return await res.json();
        } catch {
            error(HTTP_STATUS.INTERNAL_SERVER_ERROR, 'Received invalid JSON from API');
        }
    }

    return {
        get: <T>(path: string, options?: RequestInit) => request<T>(path, { ...options, method: 'GET' }),
        post: <T>(path: string, body: any, options?: RequestInit) => request<T>(path, { ...options, method: 'POST', body: JSON.stringify(body) }),
        put: <T>(path: string, body: any, options?: RequestInit) => request<T>(path, { ...options, method: 'PUT', body: JSON.stringify(body) }),
        patch: <T>(path: string, body: any, options?: RequestInit) => request<T>(path, { ...options, method: 'PATCH', body: JSON.stringify(body) }),
        delete: <T>(path: string, options?: RequestInit) => request<T>(path, { ...options, method: 'DELETE' }),
    };
}
