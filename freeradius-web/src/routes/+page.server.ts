import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { COOKIE_TOKEN } from '$lib/server/config.js';
import { verifyToken } from '$lib/server/proxy.js';

export const load: PageServerLoad = async ({ cookies }) => {
	const token = cookies.get(COOKIE_TOKEN);
	if (token && (await verifyToken(token)).valid) throw redirect(302, '/dashboard');
	throw redirect(302, '/login');
};
