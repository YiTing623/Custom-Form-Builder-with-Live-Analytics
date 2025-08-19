"use client";

import Link from "next/link";
import { useEffect, useState } from "react";
import { getToken, clearToken, onAuthChanged } from "@/lib/auth";
import { useRouter, usePathname } from "next/navigation";

export default function Navbar() {
  const [isAuthed, setIsAuthed] = useState(false);
  const router = useRouter();
  const pathname = usePathname();

  useEffect(() => {
    setIsAuthed(!!getToken());

    const unsubscribe = onAuthChanged(() => setIsAuthed(!!getToken()));
    return unsubscribe;
  }, []);
  useEffect(() => {
    setIsAuthed(!!getToken());
  }, [pathname]);

  const onLogout = () => {
    clearToken();
    setIsAuthed(false);
    router.replace("/login");
  };

  return (
    <nav className="w-full border-b bg-white">
      <div className="max-w-5xl mx-auto px-4 py-3 flex items-center justify-between">
        <Link href="/" className="font-semibold">FormBuilder</Link>

        <div className="flex items-center gap-3">
          {!isAuthed ? (
            <>
              <Link href="/register" className="px-3 py-1 rounded border">Register</Link>
              <Link href="/login" className="px-3 py-1 rounded bg-black text-white">Login</Link>
            </>
          ) : (
            <>
              <Link href="/builder" className="px-3 py-1 rounded border">Builder</Link>
              <Link href="/my-forms" className="px-3 py-1 rounded border">My Forms</Link>
              <button onClick={onLogout} className="px-3 py-1 rounded bg-black text-white">
                Logout
              </button>
            </>
          )}
        </div>
      </div>
    </nav>
  );
}
