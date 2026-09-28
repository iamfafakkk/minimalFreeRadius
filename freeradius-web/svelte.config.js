import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
	preprocess: vitePreprocess(),
	kit: {
		// Output statis (SPA) yang di-serve backend Go. Semua route
		// non-file fallback ke index.html, guard auth jalan client-side.
		adapter: adapter({ pages: 'build', assets: 'build', fallback: 'index.html', strict: true }),
		alias: { $lib: './src/lib' }
	}
};

export default config;
