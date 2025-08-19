# Custom Form Builder with Live Analytics

A full-stack demo that lets you **build custom forms**, share a **public feedback link**, and see **real-time analytics** as responses arrive.

## Tech Stack
- **Backend**: Go (Fiber) + MongoDB (Atlas or local)
- **Frontend**: Next.js (App Router) + Tailwind + vanilla React hooks (no Formik / RHF)
- **Realtime**: Server-Sent Events (SSE)
- **Auth**: Basic email/password + JWT (owner can view/manage only their forms)
- **Extras**: CSV/PDF export, conditional fields (show/hide by answers), trends (avg rating, most common answers, most-skipped)

---

## Features

### Form Builder
- Text, Multiple choice, Checkboxes, Rating fields
- Drag-and-drop reordering
- Field validation
- Save as draft / publish
- **Conditional fields** (e.g., if “Q3 = Yes” then show Q4)
- Shareable links (copy button)

### Feedback Form
- Unique URL per form: `/form/:id`
- Validates answers server-side

### Analytics Dashboard
- Live updates via SSE (no reload)
- Distributions + avg rating charts
- **Trends**: global avg rating, most-common options, most-skipped questions
- Export **CSV** and **PDF**

### Auth
- Register / Login
- `/my-forms` lists only the current user’s forms
- Edit via `/builder/:id`

---

## Project Structure
```
/backend
  main.go
  internal/
    db/           # Mongo connection + indexes
    handlers/     # auth, forms, responses, analytics, export
    middleware/   # JWT middleware
    models/       # Form, Field, Response, User types
    ws/           # SSE hub
  Dockerfile

/frontend
  src/
    app/          # Next.js routes (builder, form, dashboard, auth, my-forms)
    lib/          # API wrapper, types, auth helpers
    components/   # Navbar, etc.

.env.local.example
```

---

## Local Setup

### 1. Prerequisites
- Go ≥ 1.23
- Node ≥ 18 & pnpm (or npm/yarn)
- MongoDB
  - Local: `mongodb://localhost:27017`
  - or MongoDB Atlas URI

### 2. Backend Setup
```bash
cd backend
cp .env.example .env   # or create .env with the vars below
go mod tidy
go run main.go
```

#### Backend `.env`
```ini
PORT=8080
# Use one of these:
MONGO_URI=mongodb://localhost:27017
# or Atlas:
# MONGO_URI="mongodb+srv://<user>:<pass>@<cluster-url>/?retryWrites=true&w=majority"
MONGO_DB=Custom-Form-Builder-with-Live-Analytics
JWT_SECRET=dev_change_me
```

---

## Using the App

1. **Register** at `/register`, then **login** at `/login`.
   - Navbar shows only Register/Login when logged out, and Builder / My Forms / … after login.

2. **Create a form** at `/builder`
   - Add fields, reorder via drag-and-drop.
   - Optional **Conditional display** per field (depends on a previous field, with operators: eq/ne/includes/gt/gte/lt/lte).
   - **Save** (draft or published).
   - After save, you’ll see share links:
     - Fill: `/form/:id`
     - Dashboard: `/dashboard/:id`

3. **Share the fill link** `/form/:id` and submit responses.

4. **Watch analytics live** at `/dashboard/:id`
   - Charts update in real-time via SSE.
   - Download CSV or PDF exports.

5. **Manage your forms** at `/my-forms`
   - Edit opens `/builder/:id`
   - Quick links to dashboard and public form

---

## API (high-level)

- `POST /api/auth/register` — { email, name, password }
- `POST /api/auth/login` — { email, password }
- `GET /api/auth/me` — current user
- `POST /api/forms` — create form (auth)
- `PUT /api/forms/:id` — update form (auth, owner)
- `GET /api/forms/:id` — public form schema
- `POST /api/forms/:id/response` — submit answers
- `GET /api/forms/:id/analytics` — current snapshot
- `GET /api/sse/:formId` — SSE stream (dashboard)
- `GET /api/forms/:id/export?format=csv|pdf` — downloads
- `GET /api/my/forms` — list my forms (auth)

---

## Real-Time Analytics: How to Test

### Option A: Via the UI
1. Open two tabs:
   - Tab A: `/dashboard/:id`
   - Tab B: `/form/:id`
2. Submit answers in Tab B.  
   → Charts in Tab A update immediately without refresh.

### Option B: Via curl
1. Stream analytics events:
   ```bash
   curl -N http://localhost:8080/api/sse/<FORM_ID>
   ```
   You’ll see `: ping` and `event: message` entries.

2. Post a response:
   ```bash
   curl -X POST http://localhost:8080/api/forms/<FORM_ID>/response      -H "Content-Type: application/json"      -d '{"answers":{"name":"Alice","satisfaction":4,"fav":"UX","tags":["Idea","Praise"]}}'
   ```
   → SSE stream prints `{"type":"response:new","analytics": ... }`.

---

## CSV / PDF Export

In the dashboard, click **Download CSV** or **Download PDF**.  
Or via curl:

```bash
# CSV
curl -o responses.csv "http://localhost:8080/api/forms/<FORM_ID>/export?format=csv"

# PDF
curl -o responses.pdf "http://localhost:8080/api/forms/<FORM_ID>/export?format=pdf"
```

---

## Assumptions & Constraints

- **State management** uses only React hooks; no Formik/RHF.
- **Conditional fields** enforced in client UI; server still validates on submission.
- **SSE** chosen over websockets for simplicity—good for charts and low-frequency updates.
- **PDF export** is a lightweight textual report (no headless browser).
- **Auth** is basic JWT (cookie + Bearer). Not production-grade (no email verification, password reset, etc.).
- **Indexes** on Mongo: `_id`, `ownerId`, `formId/created`, `email` (unique).

---

## Common Pitfalls / Gotchas

- **Frontend doesn’t talk to backend**:
  - Check `NEXT_PUBLIC_API_URL` on frontend
  - CORS on backend (`*` allowed in dev)
  - Cloud env vars (`MONGO_URI`, `JWT_SECRET`)

- **Atlas connection issues** (e.g., “email already registered”):
  - Verify backend logs show Atlas URI, not localhost
  - In Atlas, add server IP to **Network Access**
  - Confirm `users` collection has unique index on `email`

---

## Docker (optional)

### Backend
```bash
# from /backend
docker build -t formbuilder-backend:local .
docker run -p 8080:8080   -e MONGO_URI="mongodb://host.docker.internal:27017"   -e MONGO_DB="Custom-Form-Builder-with-Live-Analytics"   -e JWT_SECRET="dev_change_me"   formbuilder-backend:local
```

### Frontend
Most hosts (Vercel/Netlify) build from source. If you want a container:

```bash
# from /frontend
docker build -t formbuilder-frontend:local .
docker run -p 3000:3000   -e NEXT_PUBLIC_API_URL="http://localhost:8080"   formbuilder-frontend:local
```
