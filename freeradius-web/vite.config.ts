import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	// Saat `npm run dev`, frontend :5173 dan backend :3000 beda origin,
	// jadi /api di-proxy ke backend. Saat production (di-serve backend),
	// fetch relatif /api/* same-origin tanpa proxy.
	const env = loadEnv(mode, process.cwd(), '');
	const backend = (env.BACKEND_URL || 'http://localhost:3000').replace(/\/$/, '');
	return {
		plugins: [sveltekit()],
		server: {
			port: 5173,
			proxy: { '/api': { target: backend, changeOrigin: true } }
		}
	};
});
