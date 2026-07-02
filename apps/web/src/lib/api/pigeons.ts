import type { Pigeon } from "$lib/types/pigeon";
import { createApiClient } from "./client";


export function createPigeonsApiClient(fetchFn?: typeof fetch) {
    const path = '/pigeons';
    const api = createApiClient(fetchFn);

    

    return {
        list: async (): Promise<Pigeon[]> => api.get(path),
        delete: async (pigeon: Pigeon): Promise<void> => api.delete(`${path}/${pigeon.id}`),
    };
}
