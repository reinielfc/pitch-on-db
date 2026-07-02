<script lang="ts">
	import PigeonCard from "$lib/components/PigeonCard.svelte";
	import { pigeonStore } from "$lib/state/pigeons.svelte";
	import type { PageData } from "./$types";

	let { data }: { data: PageData } = $props();

	$effect(() => {
		pigeonStore.pigeons = data.pigeons;
	});
</script>

<section class="mx-auto max-w-5xl p-6">
	<div class="mb-6">
		<h1 class="text-3xl font-bold">Pigeons</h1>
		<p class="text-base-content/70 mt-1">Work in progress...</p>
	</div>

	<div class="grid grid-cols-1 gap-4 mb-6">
		{#each pigeonStore.pigeons as pigeon}
			<PigeonCard {pigeon} />
		{/each}
	</div>
</section>
