import { redirect } from '@sveltejs/kit';
import type { LayoutServerLoad } from './$types';
import { COOKIE_TOKEN, COOKIE_USER } from '$lib/server/config.js';
import { verifyToken } from '$lib/server/proxy.js';

// Guard dashboard: wajib cookie fr_token valid (cek ke backend /auth/verify).
export const load: LayoutServerLoad = async ({ cookies, url }) => {
	const token = cookies.get(COOKIE_TOKEN);
	if (!token) throw redirect(302, `/login?next=${encodeURIComponent(url.pathname)}`);
	const { valid, username } = await verifyToken(token);
	if (!valid) {
		cookies.delete(COOKIE_TOKEN, { path: '/' });
		throw redirect(302, '/login');
	}
	return { username: cookies.get(COOKIE_USER) ?? username ?? 'admin' };
};
