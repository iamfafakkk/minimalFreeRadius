import type { RequestHandler } from './$types';
import { proxyToBackend } from '$lib/server/proxy.js';

// Proxy semua method ke Go backend. JWT diambil dari cookie httpOnly
// fr_token, browser tidak pernah memegang token.
// Contoh: GET /api/v1/nas -> http://backend:3000/api/v1/nas
export const GET: RequestHandler = ({ cookies, url, params }) =>
	proxyToBackend(cookies, 'GET', `/api/${params.path}`, url.search, undefined);

export const POST: RequestHandler = async ({ cookies, url, params, request }) => {
	const text = await request.text();
	return proxyToBackend(cookies, 'POST', `/api/${params.path}`, url.search, text || undefined);
};

export const PUT: RequestHandler = async ({ cookies, url, params, request }) => {
	const text = await request.text();
	return proxyToBackend(cookies, 'PUT', `/api/${params.path}`, url.search, text || undefined);
};

export const PATCH: RequestHandler = async ({ cookies, url, params, request }) => {
	const text = await request.text();
	return proxyToBackend(cookies, 'PATCH', `/api/${params.path}`, url.search, text || undefined);
};

export const DELETE: RequestHandler = async ({ cookies, url, params, request }) => {
	const text = await request.text();
	return proxyToBackend(cookies, 'DELETE', `/api/${params.path}`, url.search, text || undefined);
};
