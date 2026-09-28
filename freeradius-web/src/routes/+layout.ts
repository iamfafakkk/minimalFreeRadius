// Static SPA: prerender the shell, no SSR. Auth guard + data fetching
// run client-side (onMount) using the same-origin session cookie.
export const prerender = true;
export const ssr = false;
