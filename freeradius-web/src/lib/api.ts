// Client-side API helper. Semua request lewat proxy SvelteKit /api/*
// sehingga browser tidak pernah menyimpan/mengirim JWT langsung.
// Cookie httpOnly fr_token dibaca di server (proxy + load functions).

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
