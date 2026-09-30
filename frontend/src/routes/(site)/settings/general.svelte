<script lang="ts">
	import { cfg } from "$lib/state.svelte";

	const statusOptions = ["online", "idle", "dnd", "invisible"];

	function cooldown(key: string, field: string) {
		return cfg.c.cooldowns?.[key]?.[field] ?? 0;
	}
</script>

<div class="grid gap-4 lg:grid-cols-2">
	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Grind breaks</h3>
		<p class="mt-1 text-xs text-slate-500">Uses the existing break cooldown and duration logic.</p>
		<div class="mt-5 space-y-5">
			<div><div class="mb-2 flex justify-between text-xs"><span>Grind cooldown</span><span>{cooldown("breakCooldown","minHours").toFixed(1)}–{cooldown("breakCooldown","maxHours").toFixed(1)}h</span></div><input class="w-full" type="range" min="0.2" max="5" step="0.1" value={cooldown("breakCooldown","minHours")} oninput={(e) => cfg.c.cooldowns.breakCooldown.minHours = Number((e.target as HTMLInputElement).value)} /></div>
			<div><div class="mb-2 flex justify-between text-xs"><span>Break duration</span><span>{cooldown("breakDuration","minHours").toFixed(1)}–{cooldown("breakDuration","maxHours").toFixed(1)}h</span></div><input class="w-full" type="range" min="0" max="20" step="0.1" value={cooldown("breakDuration","minHours")} oninput={(e) => cfg.c.cooldowns.breakDuration.minHours = Number((e.target as HTMLInputElement).value)} /></div>
		</div>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Command timing</h3>
		<p class="mt-1 text-xs text-slate-500">These are the same global timing settings used by the grinder.</p>
		<div class="mt-5 grid gap-4 sm:grid-cols-2">
			<label class="text-xs text-slate-400">Command interval<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.commandInterval.minSeconds} /></label>
			<label class="text-xs text-slate-400">Max command interval<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.commandInterval.maxSeconds} /></label>
			<label class="text-xs text-slate-400">Button click delay<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.buttonClickDelay.minSeconds} /></label>
			<label class="text-xs text-slate-400">Max click delay<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.buttonClickDelay.maxSeconds} /></label>
			<label class="text-xs text-slate-400">Account start delay<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.startDelay.minMinutes} /></label>
			<label class="text-xs text-slate-400">Max start delay<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.startDelay.maxMinutes} /></label>
		</div>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Events</h3>
		<p class="mt-1 text-xs text-slate-500">Keep the existing event delay and correctness settings.</p>
		<div class="mt-5 grid gap-4 sm:grid-cols-2">
			<label class="text-xs text-slate-400">Event delay<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.eventDelay.minSeconds} /></label>
			<label class="text-xs text-slate-400">Max event delay<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.eventDelay.maxSeconds} /></label>
			<label class="text-xs text-slate-400 sm:col-span-2">Correct chance ({Math.round((cfg.c.eventsCorrectChance ?? 0) * 100)}%)<input class="mt-2 w-full" type="range" min="0" max="1" step="0.01" bind:value={cfg.c.eventsCorrectChance} /></label>
		</div>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Discord</h3>
		<div class="mt-5 space-y-4">
			<label class="flex items-center justify-between text-sm"><span>Auto read alerts</span><input type="checkbox" bind:checked={cfg.c.readAlerts} /></label>
			<label class="text-xs text-slate-400">Presence<select class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" bind:value={cfg.c.discordStatus}><option value="online">Online</option><option value="idle">Idle</option><option value="dnd">Do Not Disturb</option><option value="invisible">Invisible</option></select></label>
		</div>
	</section>
</div>