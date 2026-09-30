<script lang="ts">
	import { cfg, instances, api } from "$lib/state.svelte";

	let token = $state("");
	let channelID = $state("");
	let message = $state("");

	function mask(index: number) {
		return "Account " + (index + 1);
	}

	async function addAccount() {
		message = "";
		const response = await api("/api/accounts", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ token, channelID, state: true })
		});
		if (!response.ok) {
			message = await response.text();
			return;
		}
		token = "";
		channelID = "";
		message = "Account added and started.";
		await cfg.fetch();
		await instances.fetch();
	}
</script>

<div class="space-y-6">
	<div>
		<p class="text-xs font-medium uppercase tracking-[0.2em] text-indigo-400">Connections</p>
		<h2 class="mt-1 text-2xl font-bold">Accounts</h2>
		<p class="mt-1 text-sm text-slate-500">Tokens stay on the server. The dashboard only receives safe account metadata.</p>
	</div>

	<div class="grid gap-4 lg:grid-cols-[1fr_360px]">
		<section class="space-y-3">
			{#each cfg.c.accounts ?? [] as account, index}
				{@const instance = instances.i.find((item) => item.instanceId && item.accountCfg?.channelID === account.channelID)}
				<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-4">
					<div class="flex items-start justify-between gap-4">
						<div>
							<p class="font-semibold">{mask(index)}</p>
							<p class="mt-1 text-xs text-slate-500">Channel {account.channelID || "not set"}</p>
						</div>
						<button class="rounded-full px-3 py-1 text-xs {account.state ? "bg-emerald-500/15 text-emerald-300" : "bg-slate-800 text-slate-500"}" onclick={() => (account.state = !account.state)}>
							{account.state ? "Enabled" : "Disabled"}
						</button>
					</div>
					<div class="mt-4 flex flex-wrap items-center gap-2 text-xs">
						<span class="rounded-md bg-slate-950 px-2 py-1 text-slate-400">Instance: {instance?.state ?? "not running"}</span>
						{#if instance?.instanceId}
							<button class="rounded-md border border-slate-700 px-2 py-1 hover:bg-slate-800" onclick={() => instances.restart(instance.instanceId)}>Restart</button>
							<button class="rounded-md border border-slate-700 px-2 py-1 hover:bg-slate-800" onclick={() => instances.stop(instance.instanceId)}>Stop</button>
						{/if}
					</div>
				</div>
			{:else}
				<div class="rounded-2xl border border-dashed border-slate-800 p-10 text-center text-sm text-slate-500">No accounts configured.</div>
			{/each}
		</section>

		<section class="h-fit rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
			<h3 class="font-semibold">Add account</h3>
			<p class="mt-1 text-xs text-slate-500">The token is sent directly to your authenticated server session and is never returned by the API.</p>
			<div class="mt-4 space-y-3">
				<input class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="password" bind:value={token} placeholder="Discord token" autocomplete="off" />
				<input class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" bind:value={channelID} placeholder="Channel ID" />
				<button class="w-full rounded-lg bg-indigo-500 px-3 py-2 text-sm font-medium hover:bg-indigo-400 disabled:opacity-50" disabled={!token || !channelID} onclick={addAccount}>Add & start</button>
				{#if message}<p class="text-xs text-slate-400">{message}</p>{/if}
			</div>
		</section>
	</div>
</div>