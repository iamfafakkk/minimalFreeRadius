import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		// Static output (SPA) served by the Go backend. All non-file routes
		// fall back to index.html; the auth guard runs client-side.
		adapter: adapter({ pages: 'build', assets: 'build', fallback: 'index.html', strict: true }),
		alias: { $lib: './src/lib' }
	}
};

export default config;
