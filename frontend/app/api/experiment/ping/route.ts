import { NextRequest, NextResponse } from "next/server";
import { buildPingArgs } from "@/lib/experiment/cmdinject/ping";

/**
 * OS コマンドインジェクション対策の確認用 API。
 * 検証済み argv を返すのみ（OS 実行はしない。実行は backend `/experiment/ping`）。
 */
export async function GET(req: NextRequest) {
  const host = req.nextUrl.searchParams.get("host") ?? "";

  try {
    const args = buildPingArgs(host);
    return NextResponse.json({
      host,
      args,
      warning:
        "experiment endpoint: validated argv only; command is not executed here",
    });
  } catch {
    return NextResponse.json({ error: "bad request" }, { status: 400 });
  }
}
