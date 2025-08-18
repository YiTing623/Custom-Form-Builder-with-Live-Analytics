"use client";
import { useEffect, useRef, useState } from "react";
import { getForm, getAnalytics } from "@/lib/forms";
import type { AnalyticsSnapshot, FormDoc } from "@/lib/types";
import { API_URL } from "@/lib/api";
import { Chart, BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend } from "chart.js";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend);

async function downloadFile(url: string, filename: string) {
  const res = await fetch(url, { cache: "no-store" });
  if (!res.ok) throw new Error(await res.text().catch(()=>`HTTP ${res.status}`));
  const blob = await res.blob();
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(a.href), 1000);
}

export default function DashboardPage({ params }: { params: { id: string } }) {
  const id = params.id;
  const [form, setForm] = useState<FormDoc | null>(null);
  const [analytics, setAnalytics] = useState<AnalyticsSnapshot | null>(null);
  const [err, setErr] = useState<string | null>(null);

  const mcqCanvas = useRef<HTMLCanvasElement>(null);
  const ratingCanvas = useRef<HTMLCanvasElement>(null);
  const mcqChartRef = useRef<Chart | null>(null);
  const ratingChartRef = useRef<Chart | null>(null);

  useEffect(() => {
    (async () => {
      try {
        const f = await getForm(id);
        setForm(f);
        const snap = await getAnalytics(id);
        setAnalytics(snap);
      } catch (e: any) {
        setErr(e.message);
      }
    })();
  }, [id]);

  useEffect(() => {
    const es = new EventSource(`${API_URL}/api/sse/${id}`);
    es.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data);
        if (msg?.type === "response:new" && msg.analytics) {
          setAnalytics(msg.analytics);
        }
      } catch {}
    };
    return () => es.close();
  }, [id]);

  useEffect(() => {
    if (!analytics || !form) return;

    const mcqField = form.fields.find(f => f.type === "multiple");
    if (mcqField && mcqCanvas.current) {
      const dist = analytics.fields?.[mcqField.id]?.distribution ?? {};
      const labels = Object.keys(dist);
      const values = labels.map(l => dist[l] ?? 0);

      if (!mcqChartRef.current) {
        mcqChartRef.current = new Chart(mcqCanvas.current, {
          type: "bar",
          data: { labels, datasets: [{ label: mcqField.label, data: values }] },
          options: { responsive: true, animation: false }
        });
      } else {
        const ch = mcqChartRef.current;
        ch.data.labels = labels;
        (ch.data.datasets[0].data as number[]) = values;
        ch.update();
      }
    }

    const ratingField = form.fields.find(f => f.type === "rating");
    if (ratingField && ratingCanvas.current) {
      const r = analytics.fields?.[ratingField.id];
      const dist = r?.distribution || {};
      const labels = Object.keys(dist).sort((a,b)=>Number(a)-Number(b));
      const values = labels.map(l => dist[l] ?? 0);
      const label = `${ratingField.label} (avg ${Number(r?.average ?? 0).toFixed(2)})`;

      if (!ratingChartRef.current) {
        ratingChartRef.current = new Chart(ratingCanvas.current, {
          type: "bar",
          data: { labels, datasets: [{ label, data: values }] },
          options: { responsive: true, animation: false }
        });
      } else {
        const ch = ratingChartRef.current;
        ch.data.labels = labels;
        ch.data.datasets[0].label = label;
        (ch.data.datasets[0].data as number[]) = values;
        ch.update();
      }
    }
  }, [analytics, form]);

  useEffect(() => {
    let stop = false;
    let timer: any;

    async function tick() {
      try {
        const snap = await getAnalytics(id);
        setAnalytics(snap);
      } catch {
      } finally {
        if (!stop) timer = setTimeout(tick, 5000);
      }
    }

    timer = setTimeout(tick, 5000);
    return () => { stop = true; clearTimeout(timer); };
  }, [id]);

  if (err) return <div className="p-6 text-red-600">{err}</div>;
  if (!form || !analytics) return <div className="p-6">Loading…</div>;

  const cboxField = form.fields.find(f => f.type === "checkbox");
  const cdist = cboxField ? (analytics.fields?.[cboxField.id]?.distribution ?? {}) : null;

  return (
    <div className="max-w-4xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">Analytics — {form.title}</h1>
      <div className="text-sm text-gray-600">Total responses: {analytics.count}</div>

      <div className="flex flex-wrap gap-2">
        <button
          className="px-3 py-1 rounded border"
          onClick={() => downloadFile(
            `${API_URL}/api/forms/${id}/export?format=csv`,
            `responses-${id}.csv`
          )}
        >
          Download CSV
        </button>
        <button
          className="px-3 py-1 rounded border"
          onClick={() => downloadFile(
            `${API_URL}/api/forms/${id}/export?format=pdf`,
            `responses-${id}.pdf`
          )}
        >
          Download PDF
        </button>
      </div>

      <section className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div className="p-4 border rounded">
          <h2 className="font-semibold mb-2">Multiple Choice</h2>
          <canvas ref={mcqCanvas} />
        </div>

        <div className="p-4 border rounded">
          <h2 className="font-semibold mb-2">Rating</h2>
          <canvas ref={ratingCanvas} />
        </div>
      </section>

      {cboxField && cdist && (
        <div className="p-4 border rounded">
          <h2 className="font-semibold mb-2">{cboxField.label} (checkbox)</h2>
          <table className="min-w-[320px] text-left">
            <thead><tr><th className="py-2 pr-6">Option</th><th className="py-2">Count</th></tr></thead>
            <tbody>
              {Object.keys(cdist).map(k => (
                <tr key={k} className="border-t">
                  <td className="py-2 pr-6">{k}</td>
                  <td className="py-2">{cdist[k]}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
    </div>
  );
}
