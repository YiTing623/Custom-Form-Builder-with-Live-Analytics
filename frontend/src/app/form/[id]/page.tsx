"use client";
import { useEffect, useState } from "react";
import { getForm, submitResponse } from "@/lib/forms";
import type { FormDoc } from "@/lib/types";

export default function FillFormPage({ params }: { params: { id: string } }) {
  const { id } = params;
  const [form, setForm] = useState<FormDoc | null>(null);
  const [answers, setAnswers] = useState<Record<string, any>>({});
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    getForm(id).then(setForm).catch(e => setErr(e.message));
  }, [id]);

  function setAnswer(fid: string, v: any) {
    setAnswers(a => ({ ...a, [fid]: v }));
  }

  async function onSubmit() {
    setSaving(true); setMsg(null); setErr(null);
    try {
      await submitResponse(id, { answers });
      setMsg("Thanks! Your response was submitted.");
      setAnswers({});
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setSaving(false);
    }
  }

  if (err) return <div className="p-6 text-red-600">{err}</div>;
  if (!form) return <div className="p-6">Loading…</div>;
  if (form.status !== "published") return <div className="p-6">This form is not published.</div>;

  return (
    <div className="max-w-2xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">{form.title}</h1>

      <div className="space-y-4">
        {form.fields.map(f => (
          <div key={f.id} className="space-y-2">
            <label className="block font-medium">{f.label}{f.required ? " *" : ""}</label>

            {f.type === "text" && (
              <input
                className="w-full border rounded px-3 py-2"
                value={answers[f.id] ?? ""}
                onChange={e => setAnswer(f.id, e.target.value)}
              />
            )}

            {f.type === "rating" && (
              <input
                type="number"
                min={1}
                max={f.max ?? 5}
                className="w-24 border rounded px-3 py-2"
                value={answers[f.id] ?? ""}
                onChange={e => setAnswer(f.id, Number(e.target.value))}
              />
            )}

            {f.type === "multiple" && (
              <select
                className="w-full border rounded px-3 py-2"
                value={answers[f.id] ?? ""}
                onChange={e => setAnswer(f.id, e.target.value)}
              >
                <option value="">Select…</option>
                {f.options?.map(o => <option key={o} value={o}>{o}</option>)}
              </select>
            )}

            {f.type === "checkbox" && (
              <div className="flex flex-wrap gap-3">
                {f.options?.map(o => {
                  const arr = (answers[f.id] as string[] | undefined) ?? [];
                  const checked = arr.includes(o);
                  return (
                    <label key={o} className="inline-flex items-center gap-2">
                      <input
                        type="checkbox"
                        checked={checked}
                        onChange={(e) => {
                          const next = new Set(arr);
                          e.target.checked ? next.add(o) : next.delete(o);
                          setAnswer(f.id, Array.from(next));
                        }}
                      />
                      <span>{o}</span>
                    </label>
                  );
                })}
              </div>
            )}
          </div>
        ))}
      </div>

      <button
        className="px-4 py-2 rounded bg-black text-white disabled:opacity-50"
        disabled={saving}
        onClick={onSubmit}
      >
        {saving ? "Submitting…" : "Submit"}
      </button>

      {msg && <div className="text-green-600">{msg}</div>}
      {err && <div className="text-red-600">{err}</div>}
    </div>
  );
}
