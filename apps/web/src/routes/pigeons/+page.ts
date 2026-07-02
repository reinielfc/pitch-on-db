import type { PageLoad } from './$types';
import { createPigeonsApiClient } from '$lib/api';

export const load: PageLoad = async ({ fetch }) => {
	const pigeonsApi = createPigeonsApiClient(fetch);

	const pigeons = await pigeonsApi.list();
	return { pigeons };
};
