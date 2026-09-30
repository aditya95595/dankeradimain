<script lang="ts">
	import { ActivityLog, Slash, Person, Gear, Globe, DiscordLogo } from "svelte-radix";
	import { cfg, instances, api } from "$lib/state.svelte";
	import { onMount } from "svelte";

	const routes = [
		{ path: "/#/", label: "Overview", icon: ActivityLog },
		{ path: "/#/commands", label: "Auto Grind", icon: Slash },
		{ path: "/#/accounts", label: "Accounts", icon: Person },
		{ path: "/#/settings", label: "Settings", icon: Gear }
	];

	let selected = $state("");
	let saved = $state(false);
	let onlineCount = $derived(instances.i.filter((i) => ["running", "ready"].includes(i.state)).length);

	function updateSelected() {
		selected = window.location.hash.split("#/")[1] || "";
	}

	async function save() {
		await cfg.save();
		saved = true;
		setTimeout(() => (saved = false), 1800);
	}

	async function logout() {
		await api("/api/logout", { method: "POST" });
		window.location.hash = "#/login";
	}

	onMount(() => {
		updateSelected();
		const hashTimer = () => updateSelected();
		window.addEventListener("hashchange", hashTimer);
		const timer = setInterval(() => instances.fetch(), 15000);
		return () => {
			window.removeEventListener("hashchange", hashTimer);
			clearInterval(timer);
		};
	});
</script>

<aside class="sticky top-14 hidden h-[calc(100vh-3.5rem)] w-60 shrink-0 border-r border-slate-800 bg-slate-950/95 p-3 md:block">
	<div class="mb-4 rounded-xl border border-slate-800 bg-slate-900/70 p-3">
		<div class="flex items-center justify-between">
			<div><p class="text-sm font-semibold">DMG Control</p><p class="text-xs text-slate-500">Browser dashboard</p></div>
			<span class="h-2.5 w-2.5 rounded-full bg-emerald-400"></span>
		</div>
		<p class="mt-3 text-xs text-slate-400">{onlineCount} active instance{onlineCount === 1 ? "" : "s"}</p>
	</div>
	<nav class="space-y-1">
		{#each routes as route}
			<a href={route.path} class="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition {selected === route.path.slice(3) ? "bg-indigo-500/15 text-indigo-300" : "text-slate-400 hover:bg-slate-900 hover:text-white"}">
				<route.icon class="size-4" />{route.label}
			</a>
		{/each}
	</nav>
	<div class="mt-auto space-y-2 border-t border-slate-800 pt-4">
		<button class="w-full rounded-lg bg-indigo-500 px-3 py-2 text-sm font-medium hover:bg-indigo-400" onclick={save}>{saved ? "Saved ✓" : "Save changes"}</button>
		<button class="w-full rounded-lg border border-slate-800 px-3 py-2 text-sm text-slate-400 hover:bg-slate-900 hover:text-white" onclick={logout}>Sign out</button>
		<div class="flex items-center justify-center gap-3 pt-2 text-slate-500">
			<a href="https://discord.com/invite/KTrmQnhCHb" target="_blank" rel="noreferrer"><DiscordLogo class="size-5 hover:text-white" /></a>
			<a href="https://www.dankmemer.tools/" target="_blank" rel="noreferrer"><Globe class="size-5 hover:text-white" /></a>
		</div>
	</div>
</aside>

<div class="fixed bottom-3 left-3 right-3 z-40 flex gap-1 rounded-xl border border-slate-800 bg-slate-950/95 p-1 shadow-xl md:hidden">
	{#each routes as route}
		<a href={route.path} class="flex flex-1 flex-col items-center gap-1 rounded-lg px-1 py-2 text-[10px] {selected === route.path.slice(3) ? "bg-indigo-500/15 text-indigo-300" : "text-slate-500"}"><route.icon class="size-4" />{route.label}</a>
	{/each}
</div>
