import { client } from "./client";
import { ApiError } from "./errors";

export async function fetchPigeons(page: number, limit = 32, fetchFn: typeof fetch = fetch) {
	const { data, error, response } = await client.GET("/v1/pigeons", {
		params: { query: { page, limit } },
		fetch: fetchFn
	});
	if (error) throw new ApiError(response.status, error, error.detail);
	return data;
}

export async function addPigeon(fetchFn: typeof fetch = fetch) {
	const { data, error, response } = await client.POST("/v1/pigeons", { fetch: fetchFn });
	if (error) throw new ApiError(response.status, error, error.detail);
	return data;
}
