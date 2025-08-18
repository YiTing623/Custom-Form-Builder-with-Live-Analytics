"use client";
import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { getMyForms } from "@/lib/forms";
import type { FormDoc } from "@/lib/types";
import { getToken } from "@/lib/auth";

export default function MyFormsPage() {
  const router = useRouter();
  const [items, setItems] = useState<FormDoc[] | null>(null);
  const [err, setErr] = useState<string | null>(null);

  useEffect(() => {
    if (!getToken()) {
      router.replace("/login");
      return;
    }
    getMyForms().then(setItems).catch((e) => setErr(e.message));
  }, [router]);

  if (err) return <div className="p-6 text-red-600">{err}</div>;

  const isEmpty = !items || items.length === 0;

  return (
    <div className="max-w-3xl mx-auto p-6 space-y-4">
      <h1 className="text-2xl font-bold">My Forms</h1>

      {isEmpty ? (
        <div>
          No forms yet. Go to{" "}
          <Link className="underline" href="/builder">
            Builder
          </Link>
          .
        </div>
      ) : (
        <ul className="divide-y">
          {items.map((f) => (
            <li key={f.id} className="py-3 flex items-center justify-between">
              <div>
                <div className="font-medium">{f.title}</div>
                <div className="text-sm text-gray-600">Status: {f.status}</div>
              </div>
              <div className="flex gap-2">
                <Link className="px-3 py-1 border rounded text-sm" href={`/builder/${f.id}`}>
                  Edit
                </Link>
                <Link className="px-3 py-1 border rounded text-sm" href={`/dashboard/${f.id}`}>
                  Dashboard
                </Link>
                <Link className="px-3 py-1 border rounded text-sm" href={`/form/${f.id}`}>
                  Open
                </Link>
              </div>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
