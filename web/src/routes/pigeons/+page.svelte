<script lang="ts">
	import { fetchPigeons } from "$lib/api/pigeons";
	import { createQuery, keepPreviousData } from "@tanstack/svelte-query";

	const limit = 32;

	let page = $state(1);

	const pigeonQuery = createQuery(() => ({
		queryKey: ["pigeons", page, limit],
		queryFn: () => fetchPigeons(page, limit),
		keepPreviousData: keepPreviousData
	}));
</script>

{#if pigeonQuery.isPending}
	<p>Loading pigeons...</p>
{:else if pigeonQuery.isError}
	<p>Error loading pigeons: {pigeonQuery.error.message}</p>
{:else}
	<ul>
		{#each pigeonQuery.data.pigeons as pigeon (pigeon.id)}
			<li>{pigeon.name}</li>
		{/each}
	</ul>
{/if}
