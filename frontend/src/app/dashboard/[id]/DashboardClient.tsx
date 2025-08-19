'use client';
import { useEffect, useRef, useState } from "react";
import { getForm, getAnalytics } from "@/lib/forms";
import type { AnalyticsSnapshot, FormDoc } from "@/lib/types";
import { API_URL } from "@/lib/api";
import { Chart, BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend } from "chart.js";

Chart.register(BarController, BarElement, CategoryScale, LinearScale, Tooltip, Legend);

async function downloadFile(url: string, filename: string) {
  const res = await fetch(url, { cache: "no-store" });
  if (!res.ok) throw new Error(await res.text().catch(() => `HTTP ${res.status}`));
  const blob = await res.blob();
  const a = document.createElement("a");
  a.href = URL.createObjectURL(blob);
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(a.href), 1000);
}

export default function DashboardClient({ id }: { id: string }) { 
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
          setAnalytics(msg.analytics as AnalyticsSnapshot);
        }
      } catch {}
    };
    return () => es.close();
  }, [id]);

  useEffect(() => {
    if (!analytics || !form) return;

    const mcqField = form.fields.find((f) => f.type === "multiple");
    if (mcqField && mcqCanvas.current) {
      const mcqData = analytics.fields?.[mcqField.id] as
        | { type: "multiple"; distribution: Record<string, number> }
        | undefined;

      const dist = mcqData?.distribution ?? {};
      const labels = Object.keys(dist);
      const values = labels.map((l) => dist[l] ?? 0);

      if (!mcqChartRef.current) {
        mcqChartRef.current = new Chart(mcqCanvas.current, {
          type: "bar",
          data: { labels, datasets: [{ label: mcqField.label, data: values }] },
          options: { responsive: true, animation: false },
        });
      } else {
        const ch = mcqChartRef.current;
        ch.data.labels = labels;
        (ch.data.datasets[0].data as number[]) = values;
        ch.update();
      }
    }

    const ratingField = form.fields.find((f) => f.type === "rating");
    if (ratingField && ratingCanvas.current) {
      const rData = analytics.fields?.[ratingField.id] as
        | { type: "rating"; distribution: Record<number, number>; average: number }
        | undefined;

      const dist = rData?.distribution || {};
      const labels = Object.keys(dist).sort((a, b) => Number(a) - Number(b));
      const values = labels.map((l) => dist[Number(l)] ?? 0);
      const label = `${ratingField.label} (avg ${Number(rData?.average ?? 0).toFixed(2)})`;

      if (!ratingChartRef.current) {
        ratingChartRef.current = new Chart(ratingCanvas.current, {
          type: "bar",
          data: { labels, datasets: [{ label, data: values }] },
          options: { responsive: true, animation: false },
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
    return () => {
      stop = true;
      clearTimeout(timer);
    };
  }, [id]);

  if (err) return <div className="p-6 text-red-600">{err}</div>;
  if (!form || !analytics) return <div className="p-6">Loading…</div>;

  const cboxField = form.fields.find((f) => f.type === "checkbox");
  const cData = cboxField
    ? (analytics.fields?.[cboxField.id] as
        | { type: "checkbox"; distribution: Record<string, number> }
        | undefined)
    : undefined;
  const cdist = cData?.distribution ?? {};

  const labelOf = (fid: string) =>
    form.fields.find((f) => f.id === fid)?.label || fid;

  const trends = analytics.trends;

  return (
    <div className="max-w-5xl mx-auto p-6 space-y-6">
      <h1 className="text-2xl font-bold">Analytics — {form.title}</h1>
      <div className="text-sm text-gray-600">Total responses: {analytics.count}</div>

      {trends && (
        <section className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="border rounded p-4 bg-white">
            <div className="text-sm text-gray-500">Average rating (all rating fields)</div>
            <div className="text-3xl font-semibold">
              {trends.avgRating !== undefined ? trends.avgRating.toFixed(2) : "—"}
            </div>
          </div>

          <div className="border rounded p-4 bg-white">
            <div className="text-sm text-gray-500 mb-2">Most skipped (top 3)</div>
            {!trends.mostSkipped || trends.mostSkipped.length === 0 ? (
              <div className="text-sm text-gray-600">No data yet</div>
            ) : (
              <ul className="space-y-1">
                {trends.mostSkipped.map((row, i) => (
                  <li key={i} className="flex justify-between text-sm">
                    <span className="truncate mr-2">{row.label}</span>
                    <span className="tabular-nums">
                      {row.skipped}/{row.total}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>

          <div className="border rounded p-4 bg-white">
            <div className="text-sm text-gray-500 mb-2">Most common answers</div>
            {!trends.mostCommon || Object.keys(trends.mostCommon).length === 0 ? (
              <div className="text-sm text-gray-600">No data yet</div>
            ) : (
              <ul className="space-y-1">
                {Object.entries(trends.mostCommon).map(([fid, v]) => (
                  <li key={fid} className="text-sm flex justify-between">
                    <span className="truncate mr-2">{labelOf(fid)}</span>
                    <span className="truncate">
                      {Array.isArray(v) ? v.join(", ") : v}
                    </span>
                  </li>
                ))}
              </ul>
            )}
          </div>
        </section>
      )}

      <div className="flex flex-wrap gap-2">
        <button
          className="px-3 py-1 rounded border"
          onClick={() =>
            downloadFile(`${API_URL}/api/forms/${id}/export?format=csv`, `responses-${id}.csv`)
          }
        >
          Download CSV
        </button>
        <button
          className="px-3 py-1 rounded border"
          onClick={() =>
            downloadFile(`${API_URL}/api/forms/${id}/export?format=pdf`, `responses-${id}.pdf`)
          }
        >
          Download PDF
        </button>
      </div>

      <section className="grid grid-cols-1 md:grid-cols-2 gap-8">
        <div className="p-4 border rounded bg-white">
          <h2 className="font-semibold mb-2">Multiple Choice</h2>
          <canvas ref={mcqCanvas} />
        </div>

        <div className="p-4 border rounded bg-white">
          <h2 className="font-semibold mb-2">Rating</h2>
          <canvas ref={ratingCanvas} />
        </div>
      </section>

      {cboxField && (
        <div className="p-4 border rounded bg-white">
          <h2 className="font-semibold mb-2">{cboxField.label} (checkbox)</h2>
          <table className="min-w-[320px] text-left">
            <thead>
              <tr>
                <th className="py-2 pr-6">Option</th>
                <th className="py-2">Count</th>
              </tr>
            </thead>
            <tbody>
              {Object.keys(cdist).map((k) => (
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
