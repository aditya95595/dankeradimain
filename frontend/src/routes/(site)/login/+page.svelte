<script lang="ts">
	import { api } from "$lib/state.svelte";
	let password = $state("");
	let error = $state("");
	let loading = $state(false);

	async function login() {
		loading = true;
		error = "";
		try {
			const response = await api("/api/login", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify({ password })
			});
			if (!response.ok) throw new Error("Incorrect dashboard password.");
			window.location.hash = "#/";
		} catch (e) {
			error = e instanceof Error ? e.message : "Login failed.";
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>DMG Control — Login</title></svelte:head>

<div class="flex min-h-screen items-center justify-center bg-slate-950 px-4">
	<div class="w-full max-w-sm rounded-3xl border border-slate-800 bg-slate-900 p-7 shadow-2xl">
		<div class="mb-7">
			<div class="mb-4 inline-flex rounded-xl bg-indigo-500/15 px-3 py-2 text-indigo-300">DMG</div>
			<h1 class="text-2xl font-bold">Remote Control</h1>
			<p class="mt-2 text-sm text-slate-500">Sign in to manage your server remotely.</p>
		</div>
		<form onsubmit={(e) => { e.preventDefault(); login(); }} class="space-y-4">
			<input class="w-full rounded-xl border border-slate-700 bg-slate-950 px-4 py-3 text-sm outline-none focus:border-indigo-500" type="password" bind:value={password} placeholder="Dashboard password" autofocus />
			{#if error}<p class="text-xs text-red-300">{error}</p>{/if}
			<button class="w-full rounded-xl bg-indigo-500 px-4 py-3 text-sm font-semibold hover:bg-indigo-400 disabled:opacity-50" disabled={loading || !password}>
				{loading ? "Signing in..." : "Sign in"}
			</button>
		</form>
		<p class="mt-5 text-center text-[11px] text-slate-600">Set DASHBOARD_PASSWORD in your Wispbyte environment variables.</p>
	</div>
</div>