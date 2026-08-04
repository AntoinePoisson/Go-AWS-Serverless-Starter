import { expect, test } from "@playwright/test";

import { publicClient } from "../lib/client";

test("the public function reports its health without a key", async () => {
  const client = await publicClient();

  const response = await client.get("/health");

  expect(response.status()).toBe(200);
  expect(await response.json()).toMatchObject({ status: "ok" });

  await client.dispose();
});

test("an unknown path returns a 404", async () => {
  const client = await publicClient();

  const response = await client.get("/does-not-exist");

  expect(response.status()).toBe(404);

  await client.dispose();
});
