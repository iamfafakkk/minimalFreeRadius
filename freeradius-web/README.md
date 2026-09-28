# freeradius-web

Dashboard SvelteKit + [shadcn-svelte](https://www.shadcn-svelte.com/) untuk Go backend `freeradius-api`.

## Arsitektur

- Browser ⇄ SvelteKit (`/api/*`, `/dashboard`, `/login`) ⇄ Go backend (`BACKEND_URL`, default `http://localhost:3000`).
- Browser **tidak pernah** memegang JWT. Token disimpan di cookie httpOnly `fr_token` (+ `fr_user` untuk display).
- `src/routes/api/[...path]/+server.ts` proxy semua method ke backend dan menyuntik `Authorization: Bearer <cookie>`.
- Guard dashboard di `src/routes/(app)/+layout.server.ts`: tanpa cookie valid (verifikasi ke `GET /api/v1/auth/verify`) redirect ke `/login`.
- Login di `src/routes/login/+page.server.ts`: POST ke backend, set cookie, redirect `/dashboard`.

## Jalankan

```bash
cp .env.example .env   # sesuaikan BACKEND_URL bila perlu
npm install
npm run dev            # http://localhost:5173
```

Backend harus jalan dulu (`freeradius-api`, default port 3000).

## Build

```bash
npm run build
ORIGIN=http://localhost:5173 PORT=5173 node build   # adapter-node
```

> `ORIGIN` wajib diisi dengan origin publik frontend saat `node build`.
> Tanpa ini, adapter-node mengasumsikan skema `https` dan proteksi CSRF
> SvelteKit menolak semua form POST (login) dengan 403. Alternatif bila di
> belakang nginx: lihat `nginx-example.conf` (`PROTOCOL_HEADER=x-forwarded-proto`).
