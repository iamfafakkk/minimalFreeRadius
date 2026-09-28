# freeradius-web

Dashboard SvelteKit + [shadcn-svelte](https://www.shadcn-svelte.com/) untuk Go backend `freeradius-api`.
Di-build statis (`adapter-static`) dan **di-serve langsung oleh backend Go** — satu port `:3000`
untuk API + panel web.

## Arsitektur

- Browser ⇄ backend Go `:3000`: `/` + `/login` + `/dashboard/*` (file statis / SPA fallback),
  `/api/*` (REST), `/health`, `/swagger.json`.
- Auth via cookie sesi `fr_token` (httpOnly, diset backend saat login) + `fr_user` (display).
  Guard dashboard jalan client-side: `GET /api/v1/auth/verify`, redirect `/login` bila 401.
- Tidak ada server Node di production; tidak ada masalah CORS/CSRF antar-origin.

## Dev

```bash
cp .env.example .env   # BACKEND_URL menunjuk backend lokal
npm install
npm run dev            # http://localhost:5173, /api di-proxy vite ke backend
```

Backend harus jalan dulu (`freeradius-api`, default port 3000).

## Build + serve via backend

```bash
npm run build          # output: freeradius-web/build/
cd ../freeradius-api
WEB_DIR=../freeradius-web/build ./freeradius-api   # auto-detect bila dikosongkan
```

Buka `http://localhost:3000/` untuk panel, `http://localhost:3000/api/v1/...` untuk API.
Lihat `nginx-example.conf` untuk deploy di belakang nginx.
