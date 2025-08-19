import { API_URL } from "./api";

const TOKEN_KEY = "token";
const AUTH_EVENT = "auth-changed";

export function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
}

export function setToken(token: string) {
  if (typeof window === "undefined") return;
  localStorage.setItem("token", token);
  window.dispatchEvent(new Event(AUTH_EVENT));
}

export function clearToken() {
  if (typeof window === "undefined") return;
  localStorage.removeItem("token");
  window.dispatchEvent(new Event(AUTH_EVENT));
}

export function onAuthChanged(handler: () => void) {
  if (typeof window === "undefined") return () => {};
  window.addEventListener(AUTH_EVENT, handler);
  const storageHandler = (e: StorageEvent) => {
    if (e.key === "token") handler();
  };
  window.addEventListener("storage", storageHandler);
  return () => {
    window.removeEventListener(AUTH_EVENT, handler);
    window.removeEventListener("storage", storageHandler);
  };
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
