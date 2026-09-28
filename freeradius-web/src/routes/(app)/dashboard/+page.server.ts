import type { PageServerLoad } from './$types';
import { BACKEND_URL, COOKIE_TOKEN } from '$lib/server/config.js';

async function getJSON(token: string, path: string) {
	const res = await fetch(`${BACKEND_URL}${path}`, {
		headers: { Authorization: `Bearer ${token}` }
	});
	return res.json().catch(() => ({})) as Promise<Record<string, unknown>>;
}

export const load: PageServerLoad = async ({ cookies }) => {
	const token = cookies.get(COOKIE_TOKEN) ?? '';
	const [nasStats, userStats, health] = await Promise.all([
		getJSON(token, '/api/v1/nas/stats'),
		getJSON(token, '/api/v1/users/stats'),
		getJSON(token, '/api/v1/auth/health')
	]);
	const nas = (nasStats.data as { total_nas?: number } | undefined)?.total_nas ?? 0;
	const users = (userStats.data as { total_users?: number } | undefined)?.total_users ?? 0;
	const db = (health.data as { database?: string } | undefined)?.database ?? 'unknown';
	return { nas, users, db, healthy: db === 'connected' };
};
