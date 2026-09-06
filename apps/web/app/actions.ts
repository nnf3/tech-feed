"use server";

import { revalidatePath } from "next/cache";
import { ingestArticles } from "@/lib/feed";

export async function refreshFeed() {
  await ingestArticles();
  revalidatePath("/");
}
