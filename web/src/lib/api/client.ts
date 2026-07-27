import { PUBLIC_API_CSR_PATH, PUBLIC_API_SSR_URL } from "$env/static/public";
import { browser } from "$app/environment";
import createClient from "openapi-fetch";
import type { paths } from "./schema";

export const client = createClient<paths>({
	baseUrl: browser ? PUBLIC_API_CSR_PATH : PUBLIC_API_SSR_URL
});
