export type MessageResponse = {
  message: string;
};

export function apiBaseUrl(): string {
  // Prefer server runtime env so Docker/Render can change the API host without rebuild.
  const raw =
    process.env.API_BASE_URL?.trim() ||
    process.env.NEXT_PUBLIC_API_BASE_URL?.trim();
  if (!raw) {
    return "http://localhost:8080";
  }
  return raw.replace(/\/$/, "");
}

export async function fetchRootMessage(): Promise<MessageResponse> {
  const res = await fetch(`${apiBaseUrl()}/`, {
    cache: "no-store",
  });
  if (!res.ok) {
    throw new Error("failed to fetch message");
  }
  return (await res.json()) as MessageResponse;
}
