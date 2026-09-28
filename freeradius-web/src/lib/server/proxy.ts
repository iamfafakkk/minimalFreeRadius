import type { Cookies } from '@sveltejs/kit';
import { BACKEND_URL, COOKIE_TOKEN } from './config.js';

/** Baca JWT dari cookie httpOnly. */
export function getToken(cookies: Cookies): string | undefined {
	return cookies.get(COOKIE_TOKEN);
}

/** Forward request ke Go backend dengan header Authorization dari cookie. */
export async function proxyToBackend(
	cookies: Cookies,
	method: string,
	upstreamPath: string,
	search: string,
	bodyText: string | undefined
): Promise<Response> {
	const headers: Record<string, string> = {};
	const token = getToken(cookies);
	if (token) headers['Authorization'] = `Bearer ${token}`;
	if (bodyText !== undefined) headers['Content-Type'] = 'application/json';

	let res: Response;
	try {
		res = await fetch(`${BACKEND_URL}${upstreamPath}${search}`, {
			method,
			headers,
			body: bodyText
		});
	} catch {
		return Response.json(
			{ success: false, message: 'Backend tidak dapat dihubungi.' },
			{ status: 502 }
		);
	}

	const buf = new Uint8Array(await res.arrayBuffer());
	const outHeaders = new Headers();
	const ct = res.headers.get('content-type');
	if (ct) outHeaders.set('content-type', ct);
	return new Response(buf, { status: res.status, headers: outHeaders });
}

/** Verifikasi token ke backend. Dipakai guard dashboard. */
export async function verifyToken(token: string): Promise<{ valid: boolean; username?: string }> {
	try {
		const res = await fetch(`${BACKEND_URL}/api/v1/auth/verify`, {
			headers: { Authorization: `Bearer ${token}` }
		});
		if (!res.ok) return { valid: false };
		const json = (await res.json()) as { data?: { user?: { username?: string } } };
		return { valid: true, username: json.data?.user?.username };
	} catch {
		return { valid: false };
	}
}
