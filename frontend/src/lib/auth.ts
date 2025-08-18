import { API_URL } from "./api";

const TOKEN_KEY = "token";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem(TOKEN_KEY);
}
export function setToken(t: string) {
  if (typeof window === "undefined") return;
  localStorage.setItem(TOKEN_KEY, t);
}
export function clearToken() {
  if (typeof window === "undefined") return;
  localStorage.removeItem(TOKEN_KEY);
}

export function authHeaders(): Record<string, string> {
  const h: Record<string, string> = {};
  const t = getToken();
  if (t) h.Authorization = `Bearer ${t}`;
  return h;
}

export async function register(data: { email: string; password: string; name: string }) {
  const res = await fetch(`${API_URL}/api/auth/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error(await res.text());
  const j = await res.json();
  setToken(j.token);
  return j.user as { id: string; email: string; name: string };
}

export async function login(data: { email: string; password: string }) {
  const res = await fetch(`${API_URL}/api/auth/login`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(data),
  });
  if (!res.ok) throw new Error(await res.text());
  const j = await res.json();
  setToken(j.token);
  return j.user as { id: string; email: string; name: string };
}

export async function me() {
  const res = await fetch(`${API_URL}/api/me`, { headers: { ...authHeaders() } });
  if (!res.ok) throw new Error(await res.text());
  return res.json();
}
