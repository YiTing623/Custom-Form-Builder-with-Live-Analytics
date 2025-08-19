export default function Home() {
  return (
    <main className="space-y-4">
      <h1 className="text-2xl font-bold">Custom Form Builder — Demo</h1>
      <p className="text-sm text-gray-600">
        Build a form, share a public fill link, and see analytics update live.
      </p>

      <ul className="list-disc ml-6 space-y-1">
        <li><code>/builder</code> — Create & edit forms (drag &amp; drop)</li>
        <li><code>/form/&lt;FORM_ID&gt;</code> — Public feedback form</li>
        <li><code>/dashboard/&lt;FORM_ID&gt;</code> — Live analytics</li>
        <li><code>/my-forms</code> — Your saved forms</li>
        <li><code>/login</code> / <code>/register</code> — Auth</li>
      </ul>
    </main>
  );
}
