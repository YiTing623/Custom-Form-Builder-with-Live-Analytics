"use client";
import { useEffect, useState } from "react";
import { getToken } from "@/lib/auth";
import { getForm } from "@/lib/forms";
import type { FormDoc } from "@/lib/types";
import BuilderEditor from "@/components/BuilderEditor";

export default function EditBuilderPage({ params }: { params: { id: string } }) {
  const { id } = params;
  const [isAuthed, setIsAuthed] = useState<boolean | null>(null);
  const [initial, setInitial] = useState<FormDoc | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => { setIsAuthed(!!getToken()); }, []);
  useEffect(() => {
    (async () => {
      try {
        const doc = await getForm(id);
        setInitial(doc);
      } catch (e: any) {
        setErr(e.message || "Failed to load form");
      }
    })();
  }, [id]);

  if (isAuthed === null) return <div className="p-6">Loading…</div>;
  if (!isAuthed) {
    return (
      <div className="max-w-md mx-auto p-6 space-y-4">
        <h1 className="text-2xl font-bold">Form Builder</h1>
        <p className="text-gray-700">Please sign in to edit your form.</p>
        <div className="flex gap-2">
          <a className="px-4 py-2 rounded bg-black text-white" href="/login">Sign in</a>
          <a className="px-4 py-2 rounded border" href="/register">Create account</a>
        </div>
      </div>
    );
  }
  if (err) return <div className="p-6 text-red-600">{err}</div>;
  if (!initial) return <div className="p-6">Loading form…</div>;

  return <BuilderEditor initial={initial} />;
}
