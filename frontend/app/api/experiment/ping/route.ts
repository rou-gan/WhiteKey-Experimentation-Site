import { NextRequest, NextResponse } from "next/server";
import { buildPingArgs } from "@/lib/experiment/cmdinject/ping";

/**
 * OS コマンドインジェクション実験 API。
 * INTENTIONAL: buildPingArgs は sh -c 連結を返す。
 * ここでは argv を返すのみで OS 実行はしない（UT／教育用。実行シンクは backend 側実験を参照）。
 */
export async function GET(req: NextRequest) {
  const host = req.nextUrl.searchParams.get("host") ?? "";

  try {
    const args = buildPingArgs(host);
    return NextResponse.json({
      host,
      args,
      warning:
        "experiment endpoint: intentional OS command injection construction; command is not executed here",
    });
  } catch {
    return NextResponse.json({ error: "bad request" }, { status: 400 });
  }
}
