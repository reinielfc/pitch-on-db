<script lang="ts">
    import type { Icon } from "@lucide/svelte";

    export type BadgeProps = {
        label: string;
        className?: string;
        icon?: typeof Icon;
    };

    let { label, className, icon: BadgeIcon }: BadgeProps = $props();
</script>

<div class="badge badge-xs gap-1 {className}">
    {#if BadgeIcon}
        <BadgeIcon class="inline-block size-[0.92em] align-[-0.12em]" />
    {/if}
    {label}
</div>
