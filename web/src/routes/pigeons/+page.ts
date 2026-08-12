import { fetchPigeons } from "$lib/api/pigeons";

export async function load({ parent, fetch: fetchFn }) {
	const { queryClient } = await parent();
	const page = 1;
	const limit = 32;

	await queryClient.prefetchQuery({
		queryKey: ["pigeons", page, limit],
		queryFn: () => fetchPigeons(page, limit, fetchFn)
	});
}
