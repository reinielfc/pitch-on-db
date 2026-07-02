import { createPigeonsApiClient } from "$lib/api";
import type { Pigeon } from "$lib/types/pigeon";

class PigeonStore {
    pigeons = $state<Pigeon[]>([]);
    #api = createPigeonsApiClient();

    async load() {
        this.pigeons = await this.#api.list();
    }

    async delete(pigeon: Pigeon) {
        await this.#api.delete(pigeon);
        this.pigeons = this.pigeons.filter(p => p.id !== pigeon.id);
    }
}

export const pigeonStore = new PigeonStore();