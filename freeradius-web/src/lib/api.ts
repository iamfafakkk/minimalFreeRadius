// Client-side API helper (static SPA). Production: same-origin to the Go backend
// that serves this frontend; the fr_token session cookie is sent automatically.
// Dev (`npm run dev`): /api is proxied by vite to BACKEND_URL.

export interface ApiResult<T = unknown> {
	ok: boolean;
	status: number;
	data?: T;
	message?: string;
	errors?: { field: string; message: string }[];
}

async function req<T>(method: string, path: string, body?: unknown): Promise<ApiResult<T>> {
	const res = await fetch(`/api${path.startsWith('/') ? path : `/${path}`}`, {
		method,
		credentials: 'same-origin',
		headers: body !== undefined ? { 'Content-Type': 'application/json' } : undefined,
		body: body !== undefined ? JSON.stringify(body) : undefined
	});
	let json: Record<string, unknown> = {};
	try {
		json = (await res.json()) as Record<string, unknown>;
	} catch {
		/* abaikan body non-JSON */
	}
	return {
		ok: res.ok,
		status: res.status,
		data: (json.data ?? json) as T,
		message: typeof json.message === 'string' ? json.message : undefined,
		errors: Array.isArray(json.errors) ? (json.errors as ApiResult['errors']) : undefined
	};
}

export const api = {
	get: <T>(p: string) => req<T>('GET', p),
	post: <T>(p: string, b?: unknown) => req<T>('POST', p, b),
	put: <T>(p: string, b?: unknown) => req<T>('PUT', p, b),
	del: <T>(p: string, b?: unknown) => req<T>('DELETE', p, b)
};

export interface NAS {
	id: number;
	name: string;
	ip: string;
	secret: string;
	type: string;
	ports: number | null;
	community: string;
	description: string;
}

export interface RadiusUser {
	id: number;
	user: string;
	password: string;
	profile: string | null;
}

/** Check the session cookie against the backend. Used by the dashboard guard + root redirect. */
export async function verifySession(): Promise<{ valid: boolean; username?: string }> {
	try {
		const r = await api.get<{ user?: { username?: string }; valid?: boolean }>('/v1/auth/verify');
		if (!r.ok) return { valid: false };
		const username = (r.data as { user?: { username?: string } } | undefined)?.user?.username;
		return { valid: true, username };
	} catch {
		return { valid: false };
	}
}

function readCookie(name: string): string | undefined {
	for (const part of document.cookie.split(';')) {
		const [k, ...v] = part.trim().split('=');
		if (k === name) return decodeURIComponent(v.join('='));
	}
	return undefined;
}

/** Display name from the fr_user cookie (non-httpOnly, set by the backend on login). */
export function sessionUsername(): string {
	return readCookie('fr_user') || 'admin';
}
