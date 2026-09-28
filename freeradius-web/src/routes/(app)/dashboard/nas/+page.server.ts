import type { PageServerLoad } from './$types';
import { BACKEND_URL, COOKIE_TOKEN } from '$lib/server/config.js';

export const load: PageServerLoad = async ({ cookies, url }) => {
	const token = cookies.get(COOKIE_TOKEN) ?? '';
	const search = url.searchParams.get('search') ?? '';
	const page = url.searchParams.get('page') ?? '1';
	const res = await fetch(
		`${BACKEND_URL}/api/v1/nas/?page=${encodeURIComponent(page)}&limit=20&search=${encodeURIComponent(search)}`,
		{ headers: { Authorization: `Bearer ${token}` } }
	);
	const json = (await res.json().catch(() => ({}))) as {
		data?: unknown[];
		count?: number;
		pagination?: { page?: number; limit?: number; total?: number; pages?: number };
	};
	return { items: (json.data ?? []) as Record<string, unknown>[], pagination: json.pagination ?? {}, search };
};
