<script lang="ts">
    import PigeonBadge from "$lib/components/PigeonBadge.svelte";
    import type { BadgeProps } from "$lib/components/PigeonBadge.svelte";
    import type { Pigeon } from "$lib/types/pigeon";
    import { Cake, Columns4, Mars, Venus, Trash2 } from "@lucide/svelte";
    import { pigeonStore } from "$lib/state/pigeons.svelte";

    export type PigeonCardProps = {
        pigeon: Pigeon;
        onDelete?: (pigeon: Pigeon) => void;
    };

    let { pigeon, onDelete }: PigeonCardProps = $props();

    let badges = $derived.by(() => {
        const badges: BadgeProps[] = [];

        if (pigeon.birthDate) {
            badges.push({ label: pigeon.birthDate.toISOString(), icon: Cake, className: "badge-info italic" });
        }

        if (pigeon.captureDate) {
            badges.push({ label: pigeon.captureDate.toISOString(), icon: Columns4, className: "badge-info italic" });
        }

        for (const tag of pigeon.tags ?? []) {
            badges.push({ label: `#${tag}`, className: "badge-soft badge-info" });
        }

        return badges;
    });


    let deleting = $state(false);
    async function handleDelete() {
        deleting = true;
        try {
            const res = await pigeonStore.delete(pigeon);
            if (res.ok) onDelete?.(pigeon);
        } finally {
            deleting = false;
        }
    }

</script>

<div class="card bg-base-100 shadow-xl">
    <div class="card-body">
        <div class="flex flex-row flex-nowrap justify-between items-start gap-4">
            <div class="card-title">
                <h2 class="leading-tight">
                    {pigeon.name}
                    {#if pigeon.sex === "M"}
                        <Mars class="inline-block size-[0.92em] align-[-0.12em] ml-1 text-blue-500" />
                    {:else if pigeon.sex === "F"}
                        <Venus class="inline-block size-[0.92em] align-[-0.12em] ml-1 text-pink-500" />
                    {/if}
                </h2>
            </div>
            <span class="text-xs font-medium text-base-content/25">#{pigeon.id}</span>
        </div>
        {#if badges.length > 0}
            <div class="flex flex-wrap gap-1">
                {#each badges as props}
                    <PigeonBadge {...props} />
                {/each}
            </div>
        {/if}
        <div class="card-actions justify-end">
            <button class="btn btn-error btn-xs" onclick={handleDelete} disabled={deleting}>
                {#if deleting}
                    <span class="loading loading-spinner loading-sm"></span>
                {:else}
                    <Trash2 class="inline-block size-[0.92em] align-[-0.12em]" />
                {/if}
            </button>
        </div>
    </div>
</div>
