import { api } from "./api";
import type { FormDoc, AnalyticsSnapshot, SubmitResponseBody } from "./types";

export const createForm = (doc: Partial<FormDoc>) =>
  api<FormDoc>("/api/forms", { method: "POST", body: JSON.stringify(doc) });

export const updateForm = (id: string, doc: Partial<FormDoc>) =>
  api<FormDoc>(`/api/forms/${id}`, { method: "PUT", body: JSON.stringify(doc) });

export const getForm = (id: string) => api<FormDoc>(`/api/forms/${id}`);

export const submitResponse = (id: string, body: SubmitResponseBody) =>
  api(`/api/forms/${id}/response`, { method: "POST", body: JSON.stringify(body) });

export const getAnalytics = (id: string) =>
  api<AnalyticsSnapshot>(`/api/forms/${id}/analytics`);
