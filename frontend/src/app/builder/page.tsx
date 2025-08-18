"use client";
import { useEffect, useState } from "react";
import { getToken } from "@/lib/auth";
import BuilderEditor from "@/components/BuilderEditor";

export default function BuilderPage() {
  const [isAuthed, setIsAuthed] = useState<boolean | null>(null);
  useEffect(() => { setIsAuthed(!!getToken()); }, []);
  if (isAuthed === null) return <div className="p-6">Loading…</div>;
  if (!isAuthed) {
    return (
      <div className="max-w-md mx-auto p-6 space-y-4">
        <h1 className="text-2xl font-bold">Form Builder</h1>
        <p className="text-gray-700">Please sign in to create or edit your forms.</p>
        <div className="flex gap-2">
          <a className="px-4 py-2 rounded bg-black text-white" href="/login">Sign in</a>
          <a className="px-4 py-2 rounded border" href="/register">Create account</a>
        </div>
      </div>
    );
  }
  return <BuilderEditor initial={null} />;
}
