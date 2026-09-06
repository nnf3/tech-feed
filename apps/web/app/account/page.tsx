import { redirect } from "next/navigation";
import { saveProfile } from "@/app/actions";
import { Header } from "@/components/Header";
import { Tags } from "@/components/Tags";
import { authConfig } from "@/lib/auth/config";
import { getSession } from "@/lib/auth/session";
import { getProfile } from "@/lib/profiles";
import { getUser } from "@/lib/users";

export const dynamic = "force-dynamic";

export default async function AccountPage({
  searchParams,
}: {
  searchParams: Promise<{ profile_error?: string; profile_saved?: string }>;
}) {
  const { profile_error: profileError = "", profile_saved: profileSaved = "" } = await searchParams;
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  const user = (await getUser(session.sub)) ?? {
    id: session.sub,
    email: session.email,
    name: session.name,
  };
  const profile = await getProfile(session.sub);

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

      <p className="meta">プロフィール</p>
      {profileError ? <p className="meta">保存に失敗しました: {profileError}</p> : null}
      {profileSaved ? <p className="meta">プロフィールを保存しました。</p> : null}
      <section className="account-card">
        <form className="profile-form" action={saveProfile}>
          <label>
            <span>関心タグ</span>
            <input
              type="text"
              name="interest_tags"
              defaultValue={profile.interest_tags.join(", ")}
              placeholder="go, rust, kubernetes"
              autoComplete="off"
            />
            <Tags tags={profile.interest_tags} interest={profile.interest_tags} />
          </label>
          <label>
            <span>除外タグ</span>
            <input
              type="text"
              name="exclude_tags"
              defaultValue={profile.exclude_tags.join(", ")}
              placeholder="beginner, poem"
              autoComplete="off"
            />
            <Tags tags={profile.exclude_tags} />
          </label>
          <div className="account-actions">
            <button type="submit">プロフィールを保存</button>
          </div>
          <p className="account-note">
            カンマまたは空白区切りです。除外タグの記事は一覧から外し、関心タグは上に寄ります。
          </p>
        </form>
      </section>
    </main>
  );
}
