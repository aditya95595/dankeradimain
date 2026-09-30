<script lang="ts">
	import { cfg, instances, logs } from "$lib/state.svelte";
	import { onMount } from "svelte";

	let tab = $state("important");
	let now = $state(new Date());

	const tabs = [
		{ key: "important", label: "Important" },
		{ key: "others", label: "System" },
		{ key: "discord", label: "Discord" }
	];

	function activeLogs() {
		if (tab === "important") return logs.importantLogs;
		if (tab === "discord") return logs.discordLogs;
		return logs.othersLogs;
	}

	onMount(() => {
		const timer = setInterval(() => (now = new Date()), 1000);
		return () => clearInterval(timer);
	});
</script>

<div class="space-y-6">
	<div class="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
		<div>
			<p class="text-xs font-medium uppercase tracking-[0.2em] text-indigo-400">Control center</p>
			<h2 class="mt-1 text-2xl font-bold tracking-tight">Overview</h2>
			<p class="mt-1 text-sm text-slate-500">{now.toLocaleTimeString()} · live connection</p>
		</div>
		<div class="rounded-xl border border-slate-800 bg-slate-900/60 px-4 py-3 text-sm">
			<span class="text-slate-500">Grinder</span>
			<span class="ml-2 font-semibold {cfg.c.state ? "text-emerald-300" : "text-red-300"}">{cfg.c.state ? "Running" : "Stopped"}</span>
		</div>
	</div>

	<div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
		<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-4">
			<p class="text-xs text-slate-500">Accounts</p><p class="mt-2 text-2xl font-bold">{cfg.c.accounts?.length ?? 0}</p>
		</div>
		<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-4">
			<p class="text-xs text-slate-500">Instances</p><p class="mt-2 text-2xl font-bold">{instances.i.length}</p>
		</div>
		<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-4">
			<p class="text-xs text-slate-500">Active commands</p><p class="mt-2 text-2xl font-bold">{Object.values(cfg.c.commands ?? {}).filter((x: any) => x?.state).length}</p>
		</div>
		<div class="rounded-2xl border border-slate-800 bg-slate-900/60 p-4">
			<p class="text-xs text-slate-500">Discord status</p><p class="mt-2 text-lg font-bold capitalize">{cfg.c.discordStatus || "unknown"}</p>
		</div>
	</div>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-4">
		<div class="mb-4 flex items-center justify-between">
			<div><h3 class="font-semibold">Instances</h3><p class="text-xs text-slate-500">Live status without exposing account tokens</p></div>
			<button class="rounded-lg border border-slate-700 px-3 py-1.5 text-xs hover:bg-slate-800" onclick={() => instances.restartAll()}>Restart all</button>
		</div>
		<div class="grid gap-3 lg:grid-cols-2">
			{#each instances.i as instance}
				<div class="rounded-xl border border-slate-800 bg-slate-950/60 p-4">
					<div class="flex items-start justify-between gap-3">
						<div>
							<p class="font-medium">{instance.user?.username ?? "Connecting..."}</p>
							<p class="mt-1 text-xs text-slate-500">{instance.accountCfg?.channelID || "No channel"} · {instance.state}</p>
						</div>
						<span class="rounded-full px-2 py-1 text-[11px] {instance.state === "running" ? "bg-emerald-500/10 text-emerald-300" : "bg-slate-800 text-slate-400"}">{instance.state}</span>
					</div>
					<div class="mt-4 flex gap-2">
						<button class="rounded-lg bg-indigo-500 px-3 py-1.5 text-xs font-medium hover:bg-indigo-400" onclick={() => instances.restart(instance.instanceId)}>Restart</button>
						<button class="rounded-lg border border-slate-700 px-3 py-1.5 text-xs hover:bg-slate-800" onclick={() => instances.stop(instance.instanceId)}>Stop</button>
					</div>
				</div>
			{:else}
				<div class="rounded-xl border border-dashed border-slate-800 p-8 text-center text-sm text-slate-500 lg:col-span-2">No running instances.</div>
			{/each}
		</div>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-4">
		<div class="mb-3 flex gap-1 border-b border-slate-800">
			{#each tabs as item}
				<button class="px-3 py-2 text-xs font-medium {tab === item.key ? "border-b-2 border-indigo-400 text-white" : "text-slate-500"}" onclick={() => (tab = item.key)}>{item.label}</button>
			{/each}
		</div>
		<div class="h-[420px] overflow-auto rounded-xl bg-slate-950 p-3 font-mono text-xs">
			{#each activeLogs() as log (log.id)}
				<div class="mb-1 leading-5"><span class="text-slate-600">{log.timestamp}</span> <span class="{log.type === "ERR" ? "text-red-400" : "text-emerald-400"}">{log.type}</span> <span class="text-indigo-300">{log.username}</span> <span class="text-slate-300">{log.message}</span></div>
			{:else}
				<span class="text-slate-600">No logs yet.</span>
			{/each}
		</div>
	</section>
</div>