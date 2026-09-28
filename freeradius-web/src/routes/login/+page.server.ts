import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { BACKEND_URL, COOKIE_MAX_AGE, COOKIE_TOKEN, COOKIE_USER } from '$lib/server/config.js';
import { verifyToken } from '$lib/server/proxy.js';

export const load: PageServerLoad = async ({ cookies }) => {
	const token = cookies.get(COOKIE_TOKEN);
	if (token && (await verifyToken(token)).valid) throw redirect(302, '/dashboard');
	return {};
};

export const actions: Actions = {
	login: async ({ cookies, request, url }) => {
		const form = await request.formData();
		const username = String(form.get('username') ?? '').trim();
		const password = String(form.get('password') ?? '');
		if (!username || !password) return fail(400, { error: 'Username dan password wajib diisi.' });

		let res: Response;
		try {
			res = await fetch(`${BACKEND_URL}/api/v1/auth/login`, {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ username, password })
			});
		} catch {
			return fail(503, { error: 'Backend tidak dapat dihubungi.' });
		}

		const json = (await res.json().catch(() => ({}))) as {
			success?: boolean;
			message?: string;
			data?: { token?: string; user?: { username?: string } };
		};
		if (!res.ok || !json.data?.token) {
			return fail(res.status === 401 ? 401 : 400, { error: json.message ?? 'Login gagal.' });
		}

		const secure = url.protocol === 'https:';
		cookies.set(COOKIE_TOKEN, json.data.token, {
			path: '/',
			httpOnly: true,
			sameSite: 'lax',
			secure,
			maxAge: COOKIE_MAX_AGE
		});
		cookies.set(COOKIE_USER, json.data.user?.username ?? username, {
			path: '/',
			httpOnly: false,
			sameSite: 'lax',
			secure,
			maxAge: COOKIE_MAX_AGE
		});
		throw redirect(302, '/dashboard');
	}
};
