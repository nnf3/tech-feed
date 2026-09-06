"use server";

import { revalidatePath } from "next/cache";
import { redirect } from "next/navigation";
import { getSession } from "@/lib/auth/session";
import { parseTags, upsertProfile } from "@/lib/profiles";

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
