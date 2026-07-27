import { browser } from "$app/environment";
import { QueryClient } from "@tanstack/svelte-query";

export async function load() {
	return { queryClient: new QueryClient({ defaultOptions: { queries: { enabled: browser } } }) };
}
