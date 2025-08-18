"use client";
import { useCallback, useMemo, useState, useEffect } from "react";
import { createForm, updateForm } from "@/lib/forms";
import type {
  FieldType,
  FormDoc,
  FormField,
  ShowIf,
  ConditionOperator,
} from "@/lib/types";
import { uid } from "@/lib/util";

function ShareRow({ label, path }: { label: string; path: string }) {
  const [copied, setCopied] = useState(false);
  const origin =
    typeof window !== "undefined" && window.location?.origin
      ? window.location.origin
      : "";
  const fullUrl = origin ? `${origin}${path}` : path;
  const doCopy = async () => {
    try {
      await navigator.clipboard.writeText(fullUrl);
      setCopied(true);
      setTimeout(() => setCopied(false), 1500);
    } catch {}
  };
  return (
    <div className="flex w-full gap-2 items-center">
      <span className="text-sm text-gray-600 min-w-20">{label}</span>
      <input readOnly value={fullUrl} className="flex-1 px-2 py-1 border rounded bg-gray-50 text-sm" />
      <button onClick={doCopy} className="px-3 py-1 rounded bg-blue-600 text-white text-sm hover:bg-blue-700 active:scale-[0.99]">
        {copied ? "Copied!" : "Copy"}
      </button>
    </div>
  );
}

type EditField = FormField & { key: string };

const emptyText = (): EditField => ({ key: uid("k"), id: uid("q"), type: "text", label: "New text", required: false });
const emptyMultiple = (): EditField => ({ key: uid("k"), id: uid("q"), type: "multiple", label: "Multiple choice", options: ["A","B","C"] });
const emptyCheckbox = (): EditField => ({ key: uid("k"), id: uid("q"), type: "checkbox", label: "Checkboxes", options: ["X","Y"] });
const emptyRating = (): EditField => ({ key: uid("k"), id: uid("q"), type: "rating", label: "Rating", max: 5 });


export default function BuilderEditor({ initial }: { initial?: FormDoc | null }) {

  const [title, setTitle] = useState(initial?.title ?? "Untitled Form");
  const [fields, setFields] = useState<EditField[]>(
    (initial?.fields ?? []).map(f => ({ ...f, key: uid("k") }))
  );
  const [status, setStatus] = useState<FormDoc["status"]>(initial?.status ?? "draft");
  const [formId, setFormId] = useState<string>(initial?.id ?? "");
  const [saving, setSaving] = useState(false);
  const [msg, setMsg] = useState<string | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    if (!initial) return;
    setTitle(initial.title);
    setFields(initial.fields.map(f => ({ ...f, key: uid("k") })));
    setStatus(initial.status);
    setFormId(initial.id);
  }, [initial]);

  const addField = (t: FieldType) => {
    setFields((f) => [...f,
      t === "text" ? emptyText()
      : t === "multiple" ? emptyMultiple()
      : t === "checkbox" ? emptyCheckbox()
      : emptyRating()
    ]);
  };

  const onDragStart = (idx: number) => (e: React.DragEvent) => {
    e.dataTransfer.setData("text/plain", String(idx));
    e.dataTransfer.effectAllowed = "move";
  };
  const onDrop = (idx: number) => (e: React.DragEvent) => {
    const from = Number(e.dataTransfer.getData("text/plain"));
    e.preventDefault();
    if (Number.isNaN(from) || from === idx) return;
    setFields((list) => {
      const copy = [...list];
      const [moved] = copy.splice(from, 1);
      copy.splice(idx, 0, moved);
      return copy;
    });
  };
  const onDragOver = (e: React.DragEvent) => { e.preventDefault(); e.dataTransfer.dropEffect = "move"; };

  const updateField = (k: string, patch: Partial<EditField>) =>
    setFields((list) => list.map(f => f.key === k ? { ...f, ...patch } : f));

  const updateShowIf = (k: string, patch: Partial<ShowIf> | null) => {
    setFields((list) => list.map(f => {
      if (f.key !== k) return f;
      if (patch === null) {
        const { showIf, ...rest } = f as any;
        delete (rest as any).showIf;
        return { ...rest } as EditField;
      }
      const next: ShowIf = { fieldId: "", op: "eq", value: "", ...(f.showIf || {}), ...patch };
      return { ...f, showIf: next };
    }));
  };

  const removeField = (k: string) => setFields((list) => list.filter(f => f.key !== k));

  const canSave = useMemo(() => {
    if (!title.trim()) return false;
    for (const f of fields) {
      if (!f.label?.trim() || !f.id?.trim()) return false;
      if ((f.type === "multiple" || f.type === "checkbox") && (!f.options || f.options.length === 0)) return false;
      if (f.type === "rating" && (f.max ?? 0) <= 0) return false;
      if (f.showIf) {
        const idxSelf = fields.findIndex(x => x.key === f.key);
        const idxDep = fields.findIndex(x => x.id === f.showIf!.fieldId);
        if (idxDep === -1 || idxDep >= idxSelf) return false;
        const okOps: ConditionOperator[] = ["eq","ne","includes","gt","gte","lt","lte"];
        if (!okOps.includes(f.showIf.op)) return false;
      }
    }
    return fields.length > 0;
  }, [title, fields]);

  const doSave = useCallback(async () => {
    setSaving(true); setMsg(null); setErr(null);
    try {
      const payload = { title: title.trim(), status, fields: fields.map(({ key, ...rest }) => rest) };
      const res = formId ? await updateForm(formId, payload) : await createForm(payload);
      setFormId(res.id);
      setMsg(`Saved! Form ID: ${res.id}`);
    } catch (e: any) {
      setErr(e.message);
    } finally {
      setSaving(false);
    }
  }, [title, status, fields, formId]);

  const dependencyChoices = (selfIdx: number) => fields.slice(0, selfIdx);
  const operatorChoices: { value: ConditionOperator; label: string }[] = [
    { value: "eq", label: "equals" }, { value: "ne", label: "not equals" }, { value: "includes", label: "includes (checkbox)" },
    { value: "gt", label: ">" }, { value: "gte", label: ">=" }, { value: "lt", label: "<" }, { value: "lte", label: "<=" },
  ];

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">{formId ? "Edit Form" : "Form Builder"}</h1>

      <div className="flex flex-col gap-3">
        <label className="font-medium">Form title</label>
        <input className="border rounded px-3 py-2" value={title} onChange={e=>setTitle(e.target.value)} />
      </div>

      <div className="flex items-center gap-2">
        <span className="font-medium">Status:</span>
        <select className="border rounded px-2 py-1" value={status} onChange={e=>setStatus(e.target.value as any)}>
          <option value="draft">draft</option>
          <option value="published">published</option>
        </select>
      </div>

      <div className="flex gap-2">
        <button className="px-3 py-2 border rounded" onClick={()=>addField("text")}>+ Text</button>
        <button className="px-3 py-2 border rounded" onClick={()=>addField("multiple")}>+ Multiple</button>
        <button className="px-3 py-2 border rounded" onClick={()=>addField("checkbox")}>+ Checkbox</button>
        <button className="px-3 py-2 border rounded" onClick={()=>addField("rating")}>+ Rating</button>
      </div>

      <div className="space-y-3">
        {fields.map((f, idx) => (
          <div key={f.key} draggable onDragStart={onDragStart(idx)} onDrop={onDrop(idx)} onDragOver={onDragOver} className="rounded border p-4 bg-white">
            <div className="flex items-center justify-between">
              <div className="text-sm text-gray-500">drag to reorder</div>
              <button onClick={()=>removeField(f.key)} className="text-red-600 text-sm">Remove</button>
            </div>

            <div className="grid sm:grid-cols-2 gap-3 mt-3">
              <div>
                <label className="block text-sm font-medium">Field ID</label>
                <input className="w-full border rounded px-3 py-2" value={f.id} onChange={e=>updateField(f.key, { id: e.target.value })}/>
              </div>
              <div>
                <label className="block text-sm font-medium">Label</label>
                <input className="w-full border rounded px-3 py-2" value={f.label} onChange={e=>updateField(f.key, { label: e.target.value })}/>
              </div>
            </div>

            <div className="grid sm:grid-cols-3 gap-3 mt-3">
              <div>
                <label className="block text-sm font-medium">Type</label>
                <select className="w-full border rounded px-2 py-2" value={f.type} onChange={e=>updateField(f.key, { type: e.target.value as any })}>
                  <option value="text">text</option>
                  <option value="multiple">multiple</option>
                  <option value="checkbox">checkbox</option>
                  <option value="rating">rating</option>
                </select>
              </div>
              <div className="flex items-end gap-2">
                <label className="text-sm">&nbsp;</label>
                <label className="inline-flex items-center gap-2">
                  <input type="checkbox" checked={!!f.required} onChange={e=>updateField(f.key, { required: e.target.checked })}/>
                  <span className="text-sm">required</span>
                </label>
              </div>
              {f.type === "rating" && (
                <div>
                  <label className="block text-sm font-medium">Max</label>
                  <input type="number" min={1} className="w-full border rounded px-3 py-2" value={f.max ?? 5} onChange={e=>updateField(f.key, { max: Number(e.target.value) })}/>
                </div>
              )}
            </div>

            {(f.type === "multiple" || f.type === "checkbox") && (
              <div className="mt-3">
                <label className="block text-sm font-medium">Options (comma separated)</label>
                <input className="w-full border rounded px-3 py-2"
                       value={(f.options ?? []).join(", ")}
                       onChange={e=>updateField(f.key, { options: e.target.value.split(",").map(s=>s.trim()).filter(Boolean) })}/>
              </div>
            )}

            {/* Conditional display */}
            <div className="mt-4 rounded border p-3 bg-gray-50">
              <div className="flex items-center justify-between">
                <div className="font-medium text-sm">Conditional display</div>
                {f.showIf ? (
                  <button className="text-sm text-red-600" onClick={()=>updateShowIf(f.key, null)}>Remove condition</button>
                ) : (
                  <button className="text-sm" onClick={()=>updateShowIf(f.key, { fieldId: "", op: "eq", value: "" })}>+ Add condition</button>
                )}
              </div>
              {f.showIf && (
                <div className="grid sm:grid-cols-3 gap-3 mt-3">
                  <div>
                    <label className="block text-sm font-medium">Depends on</label>
                    <select className="w-full border rounded px-2 py-2" value={f.showIf.fieldId} onChange={e=>updateShowIf(f.key, { fieldId: e.target.value })}>
                      <option value="">(choose a previous question)</option>
                      {fields.slice(0, idx).map(df => (
                        <option key={df.key} value={df.id}>{df.label} ({df.id})</option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium">Operator</label>
                    <select className="w-full border rounded px-2 py-2" value={f.showIf.op} onChange={e=>updateShowIf(f.key, { op: e.target.value as ConditionOperator })}>
                      <option value="eq">equals</option>
                      <option value="ne">not equals</option>
                      <option value="includes">includes (checkbox)</option>
                      <option value="gt">{">"}</option>
                      <option value="gte">{">="}</option>
                      <option value="lt">{"<"}</option>
                      <option value="lte">{"<="}</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium">Value</label>
                    <input className="w-full border rounded px-3 py-2" value={String(f.showIf.value ?? "")} onChange={e=>updateShowIf(f.key, { value: e.target.value })} placeholder="e.g., Yes, 5, UX"/>
                  </div>
                </div>
              )}
              <p className="text-xs text-gray-500 mt-2">
                Show this field only if the “Depends on” answer matches the chosen operator/value. Use “includes” for checkbox arrays.
              </p>
            </div>
          </div>
        ))}
      </div>

      <div className="flex flex-col gap-3">
        <div className="flex flex-wrap gap-2 items-center">
          <button className="px-4 py-2 rounded bg-black text-white disabled:opacity-50" disabled={!canSave || saving} onClick={doSave}>
            {saving ? "Saving…" : "Save form"}
          </button>
        </div>

        {!!formId && (
          <div className="space-y-2 mt-2">
            <span className="text-sm text-gray-600">Share:</span>
            <ShareRow label="Fill" path={`/form/${formId}`} />
            <ShareRow label="Dashboard" path={`/dashboard/${formId}`} />
          </div>
        )}
      </div>

      {msg && <div className="text-green-600">{msg}</div>}
      {err && <div className="text-red-600">{err}</div>}
    </div>
  );
}
