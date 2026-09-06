"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import { addBookmark, removeBookmark, type BookmarkInput } from "@/lib/bookmarks";
import { getProfile, parseTags, toggleTagLists, upsertProfile } from "@/lib/profiles";

export async function toggleProfileTag(tag: string, kind: "interest" | "exclude") {
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  try {
    const profile = await getProfile(session.sub);
    const next = toggleTagLists(profile.interest_tags, profile.exclude_tags, tag, kind);
    await upsertProfile(session.sub, {
      interest_tags: next.interest,
      exclude_tags: next.exclude,
    });
  } catch (err) {
    const message = err instanceof Error ? err.message : "保存に失敗しました";
    redirect(`/?profile_error=${encodeURIComponent(message)}`);
  }
  revalidatePath("/");
  revalidatePath("/account");
}

export async function toggleBookmark(input: BookmarkInput & { saved: boolean; from?: string }) {
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  const back = input.from === "bookmarks" ? "/bookmarks" : "/";
  try {
    if (input.saved) {
      await removeBookmark(session.sub, input.article_id);
    } else {
      await addBookmark(session.sub, input);
    }
  } catch (err) {
    const message = err instanceof Error ? err.message : "保存に失敗しました";
    redirect(`${back}?bookmark_error=${encodeURIComponent(message)}`);
  }
  revalidatePath("/");
  revalidatePath("/bookmarks");
}

export async function saveProfile(formData: FormData) {
  const session = await getSession();
  if (!session) {
    redirect("/login");
  }

  try {
    await upsertProfile(session.sub, {
      interest_tags: parseTags(formData.get("interest_tags")),
      exclude_tags: parseTags(formData.get("exclude_tags")),
    });
  } catch (err) {
    const message = err instanceof Error ? err.message : "保存に失敗しました";
    redirect(`/account?profile_error=${encodeURIComponent(message)}`);
  }
  revalidatePath("/");
  redirect("/account?profile_saved=1");
}
