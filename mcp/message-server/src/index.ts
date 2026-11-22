import { McpServer } from "@modelcontextprotocol/server";
import { StdioServerTransport } from "@modelcontextprotocol/server/stdio";
import * as z from "zod/v4";

function requireApiKey(): string {
  const key = process.env.MCP_API_KEY?.trim() ?? "";
  if (!key) {
    console.error("MCP_API_KEY is required (shared-secret gate for this experiment)");
    process.exit(1);
  }
  return key;
}

function apiBaseUrl(): string {
  const raw = process.env.API_BASE_URL?.trim() || "http://localhost:8080";
  return raw.replace(/\/+$/, "");
}

async function fetchRootMessage(baseUrl: string): Promise<string> {
  const res = await fetch(`${baseUrl}/`);
  if (!res.ok) {
    throw new Error(`backend returned HTTP ${res.status}`);
  }

  let body: unknown;
  try {
    body = await res.json();
  } catch {
    throw new Error("backend response is not JSON");
  }

  if (
    typeof body !== "object" ||
    body === null ||
    typeof (body as { message?: unknown }).message !== "string"
  ) {
    throw new Error('backend response missing string field "message"');
  }

  return (body as { message: string }).message;
}

async function main(): Promise<void> {
  // Auth gate: stdio clients cannot send Bearer headers, so only a process
  // started with MCP_API_KEY may expose tools.
  requireApiKey();
  const baseUrl = apiBaseUrl();

  const server = new McpServer({
    name: "whitekey-message",
    version: "0.1.0",
  });

  server.registerTool(
    "get_message",
    {
      description:
        "Fetch the root message from the WhiteKey backend (GET /).",
      inputSchema: z.object({}),
    },
    async () => {
      try {
        const message = await fetchRootMessage(baseUrl);
        return {
          content: [{ type: "text" as const, text: message }],
        };
      } catch (err) {
        const text = err instanceof Error ? err.message : String(err);
        return {
          content: [{ type: "text" as const, text }],
          isError: true,
        };
      }
    },
  );

  const transport = new StdioServerTransport();
  await server.connect(transport);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
