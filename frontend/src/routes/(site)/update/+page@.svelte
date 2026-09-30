<script lang="ts">
	import { api } from "$lib/state.svelte";
	let message = $state("Web mode uses your deployment process for updates.");
	async function check() {
		const response = await api("/api/check-updates");
		const data = await response.json();
		message = data.hasUpdate ? "A newer version is available. Pull the latest repository code and restart the Wispbyte server." : "You are on the current configured version.";
	}
</script>

<div class="mx-auto max-w-2xl space-y-6 py-8">
	<div><p class="text-xs uppercase tracking-[0.2em] text-indigo-400">Maintenance</p><h2 class="mt-1 text-2xl font-bold">Updates</h2><p class="mt-2 text-sm text-slate-500">The browser dashboard does not download or replace the running binary itself.</p></div>
	<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<p class="text-sm text-slate-300">{message}</p>
		<button class="mt-5 rounded-lg bg-indigo-500 px-4 py-2 text-sm font-medium hover:bg-indigo-400" onclick={check}>Check for updates</button>
	</div>
</div>