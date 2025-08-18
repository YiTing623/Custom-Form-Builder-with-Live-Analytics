import { api } from "./api";
import type { FormDoc, AnalyticsSnapshot, SubmitResponseBody } from "./types";

function getToken(): string | null {
  if (typeof window === "undefined") return null;
  return localStorage.getItem("token");
}

function withAuthHeaders(init?: RequestInit): RequestInit {
  const token = getToken();
  const headers: Record<string, string> = {
    ...(init?.headers as Record<string, string> | undefined),
  };
  if (token) headers.Authorization = `Bearer ${token}`;
  return { ...init, headers };
}

function withJson(init?: RequestInit): RequestInit {
  const headers: Record<string, string> = {
    "Content-Type": "application/json",
    ...(init?.headers as Record<string, string> | undefined),
  };
  return { ...init, headers };
}



export const createForm = (doc: Partial<FormDoc>) =>
  api<FormDoc>(
    "/api/forms",
    withAuthHeaders(
      withJson({ method: "POST", body: JSON.stringify(doc) })
    )
  );

export const updateForm = (id: string, doc: Partial<FormDoc>) =>
  api<FormDoc>(
    `/api/forms/${id}`,
    withAuthHeaders(
      withJson({ method: "PUT", body: JSON.stringify(doc) })
    )
  );

export const getForm = (id: string) =>
  api<FormDoc>(`/api/forms/${id}`, withAuthHeaders());

export const getMyForms = () =>
  api<FormDoc[]>(`/api/my/forms`, withAuthHeaders());

export const submitResponse = (id: string, body: SubmitResponseBody) =>
  api(
    `/api/forms/${id}/response`,
    withAuthHeaders(
      withJson({ method: "POST", body: JSON.stringify(body) })
    )
  );

export const getAnalytics = (id: string) =>
  api<AnalyticsSnapshot>(`/api/forms/${id}/analytics`);
