/**
 * 指定 plan の __ZAP_SESSION_TOKEN__ を埋め込み、zap-out/<name>.runtime.yaml を書く。
 *
 *   ZAP_SESSION_TOKEN=... node render-plan.mjs              # plan.yaml → plan.runtime.yaml
 *   ZAP_SESSION_TOKEN=... node render-plan.mjs plan-active.yaml
 *
 * トークンはログに出さない。
 */
import { mkdirSync, readFileSync, writeFileSync } from "node:fs";
import { basename, dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const token = process.env.ZAP_SESSION_TOKEN || "";
if (!token) {
  console.error("ZAP_SESSION_TOKEN is required");
  process.exit(1);
}
if (token.includes("\n") || token.includes("\r")) {
  console.error("ZAP_SESSION_TOKEN must be a single line");
  process.exit(1);
}

const here = dirname(fileURLToPath(import.meta.url));
const root = join(here, "..", "..");
const planName = process.argv[2] || "plan.yaml";
if (planName.includes("..") || planName.includes("/") || planName.includes("\\")) {
  console.error("plan name must be a file under .github/zap (no path separators)");
  process.exit(1);
}
const srcPath = join(here, planName);
let src;
try {
  src = readFileSync(srcPath, "utf8");
} catch {
  console.error(`cannot read ${srcPath}`);
  process.exit(1);
}
if (!src.includes("__ZAP_SESSION_TOKEN__")) {
  console.error(`${planName} missing __ZAP_SESSION_TOKEN__ placeholder`);
  process.exit(1);
}

const outDir = join(root, "zap-out");
mkdirSync(outDir, { recursive: true });
const rendered = src.split("__ZAP_SESSION_TOKEN__").join(token);
const base = basename(planName, ".yaml");
const outPath = join(outDir, `${base}.runtime.yaml`);
writeFileSync(outPath, rendered, "utf8");
console.error(`wrote ${outPath}`);
