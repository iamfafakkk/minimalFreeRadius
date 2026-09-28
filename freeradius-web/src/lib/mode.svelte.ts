// Theme mode store — Svelte 5 runes. Default: dark.
// `mode.current` = selected choice ('light'|'dark'|'system'), `mode.resolved` = effective light/dark.
// Initial value is read from the <html class="dark"> set by app.html before paint (anti-flash).
import { browser } from '$app/environment';

const STORAGE_KEY = 'theme';

export type ThemeMode = 'light' | 'dark' | 'system';

function readMode(): ThemeMode {
	if (!browser) return 'dark';
	const v = localStorage.getItem(STORAGE_KEY);
	return v === 'light' || v === 'dark' || v === 'system' ? v : 'dark';
}

function prefersDark(): boolean {
	return browser && window.matchMedia('(prefers-color-scheme: dark)').matches;
}

class Mode {
	current = $state<ThemeMode>(readMode());
	#systemDark = $state(prefersDark());

	get resolved(): 'light' | 'dark' {
		if (this.current === 'system') return this.#systemDark ? 'dark' : 'light';
		return this.current;
	}

	constructor() {
		if (!browser) return;
		this.#apply();
		window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
			this.#systemDark = e.matches;
			if (this.current === 'system') this.#apply();
		});
	}

	set(mode: ThemeMode) {
		this.current = mode;
		localStorage.setItem(STORAGE_KEY, mode);
		this.#apply();
	}

	#apply() {
		document.documentElement.classList.toggle('dark', this.resolved === 'dark');
	}
}

// Singleton lintas komponen.
export const mode = new Mode();
