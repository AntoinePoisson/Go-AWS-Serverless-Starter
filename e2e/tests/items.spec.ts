import { expect, test } from "@playwright/test";

import { apiClient, Item, publicClient } from "../lib/client";

test("an item can be created, read, listed and deleted", async () => {
  const client = await apiClient();

  const created = await client.post("/items", {
    data: { name: "starter item", tags: ["e2e"] },
  });
  expect(created.status()).toBe(201);

  const item = (await created.json()) as Item;
  expect(item.id).toBeTruthy();
  expect(item.name).toBe("starter item");

  const read = await client.get(`/items/${item.id}`);
  expect(read.status()).toBe(200);
  expect((await read.json()) as Item).toMatchObject({ id: item.id });

  const listed = await client.get("/items?limit=100");
  expect(listed.status()).toBe(200);
  const { items } = (await listed.json()) as { items: Item[] };
  expect(items.map((i) => i.id)).toContain(item.id);

  const deleted = await client.delete(`/items/${item.id}`);
  expect(deleted.status()).toBe(204);

  const missing = await client.get(`/items/${item.id}`);
  expect(missing.status()).toBe(404);

  await client.dispose();
});

test("an item created through the api is readable through the public function", async () => {
  const client = await apiClient();
  const anonymous = await publicClient();

  const created = await client.post("/items", { data: { name: "shared item" } });
  const item = (await created.json()) as Item;

  const read = await anonymous.get(`/public/items/${item.id}`);
  expect(read.status()).toBe(200);
  expect((await read.json()) as Item).toMatchObject({ name: "shared item" });

  await client.delete(`/items/${item.id}`);
  await client.dispose();
  await anonymous.dispose();
});

test("the api rejects a request without a valid key", async () => {
  const client = await apiClient("wrong-key");

  const response = await client.get("/items");

  expect(response.status()).toBe(401);
  expect(await response.json()).toMatchObject({ code: "unauthorized" });

  await client.dispose();
});

test("the api rejects an item without a name", async () => {
  const client = await apiClient();

  const response = await client.post("/items", { data: { name: "" } });

  expect(response.status()).toBe(400);
  expect(await response.json()).toMatchObject({ code: "invalid_input" });

  await client.dispose();
});
