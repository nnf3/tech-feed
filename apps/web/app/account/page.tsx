import { redirect } from "next/navigation";
import { Header } from "@/components/Header";
import { authConfig } from "@/lib/auth/config";
import { getSession } from "@/lib/auth/session";
import { getUser } from "@/lib/users";

export const dynamic = "force-dynamic";

export default async function AccountPage() {
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  const user = (await getUser(session.sub)) ?? {
    id: session.sub,
    email: session.email,
    name: session.name,
  };

  return (
    <main>
      <Header session={session} />
      <p className="meta">アカウント</p>
      <section className="account-card">
        <dl className="account-fields">
          <div>
            <dt>メール</dt>
            <dd>{user.email || "未設定"}</dd>
          </div>
          <div>
            <dt>表示名</dt>
            <dd>{user.name || "未設定"}</dd>
          </div>
        </dl>
        <div className="account-actions">
          <a className="auth-link" href={authConfig().settingsUrl}>
            パスワード・アカウント設定
          </a>
          <form action="/logout" method="post">
            <button type="submit">ログアウト</button>
          </form>
        </div>
        <p className="account-note">
          パスワード変更は IdP（Kratos）の設定画面で行います。完了後はこのページに戻ってください。
        </p>
      </section>
    </main>
  );
}
