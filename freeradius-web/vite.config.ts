import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig, loadEnv } from 'vite';

export default defineConfig(({ mode }) => {
	// During `npm run dev`, frontend :5173 and backend :3000 are different origins,
	// so /api is proxied to the backend. In production (served by the backend),
	// relative /api/* fetches are same-origin with no proxy.
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
