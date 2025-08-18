"use client";
import { useState } from "react";
import { register } from "@/lib/auth";
import { useRouter } from "next/navigation";

export default function RegisterPage() {
  const r = useRouter();
  const [email, setEmail] = useState("");
  const [name, setName] = useState("");
  const [password, setPassword] = useState("");
  const [err, setErr] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setBusy(true); setErr(null);
    try {
      await register({ email, name, password });
      r.push("/my-forms");
    } catch (e:any) {
      setErr(e.message || "Register failed");
    } finally {
      setBusy(false);
    }
  }

  return (
    <div className="max-w-md mx-auto p-6 space-y-4">
      <h1 className="text-2xl font-bold">Create account</h1>
      <form onSubmit={onSubmit} className="space-y-3">
        <input className="w-full border rounded px-3 py-2" placeholder="Name" value={name} onChange={e=>setName(e.target.value)} />
        <input className="w-full border rounded px-3 py-2" placeholder="Email" value={email} onChange={e=>setEmail(e.target.value)} />
        <input className="w-full border rounded px-3 py-2" placeholder="Password (min 6)" type="password" value={password} onChange={e=>setPassword(e.target.value)} />
        <button className="px-4 py-2 rounded bg-black text-white disabled:opacity-50" disabled={busy}>{busy?"Creating…":"Create account"}</button>
      </form>
      {err && <div className="text-red-600">{err}</div>}
      <div className="text-sm">Already have an account? <a className="underline" href="/login">Sign in</a></div>
    </div>
  );
}
