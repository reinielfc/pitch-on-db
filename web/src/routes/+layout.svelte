<script lang="ts">
	import favicon from "$lib/assets/favicon.svg";
	import { QueryClientProvider } from "@tanstack/svelte-query";
	import "./layout.css";

	let { data, children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

<QueryClientProvider client={data.queryClient}>
	{@render children()}
</QueryClientProvider>
