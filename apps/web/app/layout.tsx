import type { Metadata } from "next";
import "./globals.css";

export const metadata: Metadata = {
  title: "Tech-Feed",
  description: "今の自分にジャストフィットする技術情報",
};

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="ja">
      <body>{children}</body>
    </html>
  );
}
