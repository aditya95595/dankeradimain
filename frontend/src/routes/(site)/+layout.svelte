<script lang="ts">
	import "../../app.css";
	import Nav from "$lib/components/Nav.svelte";
	import Header from "$lib/components/Header.svelte";
	import { onMount } from "svelte";

	interface Props { children?: import("svelte").Snippet; }
	let { children }: Props = $props();
	let isLogin = $state(false);

	onMount(() => {
		const update = () => (isLogin = window.location.hash.includes("#/login"));
		update();
		window.addEventListener("hashchange", update);
		return () => window.removeEventListener("hashchange", update);
	});
</script>

<div class="min-h-screen bg-slate-950 text-slate-100 antialiased">
	{#if isLogin}
		{@render children?.()}
	{:else}
		<Header />
		<div class="flex min-h-[calc(100vh-3.5rem)]">
			<Nav />
			<main class="min-w-0 flex-1 p-4 md:p-6 lg:p-8">
				<div class="mx-auto max-w-7xl">
					{@render children?.()}
				</div>
			</main>
		</div>
	{/if}
</div>
