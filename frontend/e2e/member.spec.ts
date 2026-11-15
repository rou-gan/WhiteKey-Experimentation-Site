import { expect, test, type Page } from "@playwright/test";

function uniqueCreds() {
  const suffix = `${Date.now().toString(36)}${Math.random().toString(36).slice(2, 6)}`;
  // loginId: 半角英数字 8〜15
  const loginId = `m${suffix}`.slice(0, 15);
  return {
    loginId,
    password: "Abcd123!",
    nickname: `名${suffix}`.slice(0, 50),
  };
}

async function register(page: Page, creds: ReturnType<typeof uniqueCreds>) {
  await page.goto("/members/register");
  await page.getByLabel("ログインID").fill(creds.loginId);
  await page.getByLabel("パスワード").fill(creds.password);
  await page.getByLabel("ニックネーム").fill(creds.nickname);
  await page.getByRole("button", { name: "登録" }).click();
  await expect(page).toHaveURL(/\/members\/login/);
}

async function login(page: Page, creds: ReturnType<typeof uniqueCreds>) {
  await page.goto("/members/login");
  await page.getByLabel("ログインID").fill(creds.loginId);
  await page.getByLabel("パスワード").fill(creds.password);
  await page.getByRole("button", { name: "ログイン" }).click();
}

test.describe("IT-MEM-3 member e2e", () => {
  // シナリオ: IT-MEM-3-S01
  // 目的: 登録からログインしてニックネームが見えること。
  test("S01 登録からログインしてニックネームが見える", async ({ page }) => {
    const creds = uniqueCreds();
    await register(page, creds);
    await login(page, creds);
    await expect(page).toHaveURL(/\/members\/me/);
    await expect(page.getByText(creds.nickname)).toBeVisible();
  });

  // シナリオ: IT-MEM-3-S02
  // 目的: ログアウト後に会員ホームを使えないこと。
  test("S02 ログアウト後に会員ホームを使えない", async ({ page }) => {
    const creds = uniqueCreds();
    await register(page, creds);
    await login(page, creds);
    await expect(page).toHaveURL(/\/members\/me/);
    await page.getByRole("button", { name: "ログアウト" }).click();
    await expect(page).toHaveURL(/\/members\/login/);
    await page.goto("/members/me");
    await expect(page).toHaveURL(/\/members\/login/);
  });

  // シナリオ: IT-MEM-3-S03
  // 目的: 退会後に同じ資格情報でログインできないこと。
  test("S03 退会後に同じ資格情報でログインできない", async ({ page }) => {
    const creds = uniqueCreds();
    await register(page, creds);
    await login(page, creds);
    await expect(page).toHaveURL(/\/members\/me/);

    page.once("dialog", (dialog) => dialog.accept());
    await page.getByRole("button", { name: "退会" }).click();
    await expect(page).toHaveURL(/\/members\/login/);

    await login(page, creds);
    await expect(page.getByText("ログインできませんでした。")).toBeVisible();
    await expect(page).toHaveURL(/\/members\/login/);
  });
});
