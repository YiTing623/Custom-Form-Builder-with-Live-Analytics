import "./globals.css";
import type { Metadata } from "next";
import Navbar from "@/components/NavBar";

export const metadata: Metadata = {
  title: "Custom Form Builder",
  description: "Build forms and view live analytics",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        <Navbar />
        <main className="min-h-screen">{children}</main>
      </body>
    </html>
  );
}
