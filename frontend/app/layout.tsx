import type { Metadata } from "next";
import { Fraunces, Source_Sans_3 } from "next/font/google";
import { connection } from "next/server";
import "./globals.css";

const display = Fraunces({
  subsets: ["latin"],
  variable: "--font-display-loaded",
});

const body = Source_Sans_3({
  subsets: ["latin"],
  variable: "--font-body-loaded",
});

export const metadata: Metadata = {
  title: "しろかぎ",
  description: "セキュリティ実験場 — Design-by-Security",
};

export default async function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  // CSP nonce はリクエスト単位のため、静的生成せず動的描画する。
  await connection();

  return (
    <html lang="ja">
      <body className={`${display.variable} ${body.variable}`}>{children}</body>
    </html>
  );
}
