import { APIRequestContext, request } from "@playwright/test";

export const apiURL = process.env.E2E_API_URL ?? "http://localhost:8080";
export const publicURL = process.env.E2E_PUBLIC_URL ?? "http://localhost:8081";
export const apiKey = process.env.E2E_API_KEY ?? "local-dev-key";

export interface Item {
  id: string;
  name: string;
  tags?: string[];
  metadata?: Record<string, string>;
  createdAt: string;
}

export function apiClient(key: string = apiKey): Promise<APIRequestContext> {
  return request.newContext({
    baseURL: apiURL,
    extraHTTPHeaders: { "X-Api-Key": key },
  });
}

export function publicClient(): Promise<APIRequestContext> {
  return request.newContext({ baseURL: publicURL });
}
