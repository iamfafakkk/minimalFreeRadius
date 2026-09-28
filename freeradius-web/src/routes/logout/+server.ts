import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { COOKIE_TOKEN, COOKIE_USER } from '$lib/server/config.js';

export const POST: RequestHandler = ({ cookies }) => {
	cookies.delete(COOKIE_TOKEN, { path: '/' });
	cookies.delete(COOKIE_USER, { path: '/' });
	throw redirect(302, '/login');
};
