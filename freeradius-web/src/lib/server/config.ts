import { env } from '$env/dynamic/private';

export const BACKEND_URL = (env.BACKEND_URL ?? 'http://localhost:3000').replace(/\/$/, '');
export const API_PREFIX = (env.API_PREFIX ?? '/api/v1').startsWith('/')
	? (env.API_PREFIX ?? '/api/v1')
	: `/${env.API_PREFIX ?? 'api/v1'}`;
export const COOKIE_TOKEN = 'fr_token';
export const COOKIE_USER = 'fr_user';
export const COOKIE_MAX_AGE = 60 * 60 * 24; // 24 jam, selaras JWT_EXPIRES_IN default backend

export function backendUrl(path: string): string {
	const p = path.startsWith('/') ? path : `/${path}`;
	// path dari proxy sudah termasuk /api/v1/..., teruskan apa adanya
	if (p.startsWith(API_PREFIX) || p === '/health' || p === '/swagger.json') return `${BACKEND_URL}${p}`;
	return `${BACKEND_URL}${API_PREFIX}${p}`;
}
