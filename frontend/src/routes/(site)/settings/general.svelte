<script lang="ts">
	import { cfg, api } from "$lib/state.svelte";

	const cooldown = (key: string, field: string) => cfg.c.cooldowns?.[key]?.[field] ?? 0;

	async function updateStatus(status: string) {
		cfg.c.discordStatus = status;
		const response = await api("/api/discord-status", {
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ status })
		});
		if (!response.ok) {
			await cfg.fetch();
		}
	}

	function rangeFields(key: string, min: string, max: string) {
		return {
			min: cooldown(key, min),
			max: cooldown(key, max)
		};
	}
</script>

<div class="grid gap-4 lg:grid-cols-2">
	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5 lg:col-span-2">
		<h3 class="font-semibold">Captcha solver API</h3>
		<p class="mt-1 text-xs text-slate-500">Optional DMG API key. It is never returned by the server after saving.</p>
		<label class="mt-4 block text-xs text-slate-400">API key
			<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="password" autocomplete="new-password" placeholder="Leave blank to keep the existing key" bind:value={cfg.c.apiKey} />
		</label>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Grind breaks</h3>
		<p class="mt-1 text-xs text-slate-500">The same break cooldown and duration values used by the original grinder.</p>
		<div class="mt-5 space-y-4">
			{@const breakCooldown = rangeFields("breakCooldown", "minHours", "maxHours")}
			{@const breakDuration = rangeFields("breakDuration", "minHours", "maxHours")}
			<label class="block text-xs text-slate-400">Grind cooldown (hours)
				<div class="mt-1 grid grid-cols-2 gap-2">
					<input class="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" min="0" step="0.1" bind:value={cfg.c.cooldowns.breakCooldown.minHours} />
					<input class="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" min="0" step="0.1" bind:value={cfg.c.cooldowns.breakCooldown.maxHours} />
				</div>
			</label>
			<label class="block text-xs text-slate-400">Break duration (hours)
				<div class="mt-1 grid grid-cols-2 gap-2">
					<input class="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" min="0" step="0.1" bind:value={cfg.c.cooldowns.breakDuration.minHours} />
					<input class="rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" min="0" step="0.1" bind:value={cfg.c.cooldowns.breakDuration.maxHours} />
				</div>
			</label>
		</div>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Command timing</h3>
		<p class="mt-1 text-xs text-slate-500">Global timing settings used by the existing command engine.</p>
		<div class="mt-5 grid gap-4 sm:grid-cols-2">
			<label class="text-xs text-slate-400">Command interval min<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.commandInterval.minSeconds} /></label>
			<label class="text-xs text-slate-400">Command interval max<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.commandInterval.maxSeconds} /></label>
			<label class="text-xs text-slate-400">Button click delay min<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.buttonClickDelay.minSeconds} /></label>
			<label class="text-xs text-slate-400">Button click delay max<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.buttonClickDelay.maxSeconds} /></label>
			<label class="text-xs text-slate-400">Account start delay min<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.startDelay.minMinutes} /></label>
			<label class="text-xs text-slate-400">Account start delay max<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.startDelay.maxMinutes} /></label>
		</div>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Events</h3>
		<p class="mt-1 text-xs text-slate-500">Existing event timing and correctness settings.</p>
		<div class="mt-5 grid gap-4 sm:grid-cols-2">
			<label class="text-xs text-slate-400">Event delay min<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.eventDelay.minSeconds} /></label>
			<label class="text-xs text-slate-400">Event delay max<input class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" type="number" step="0.1" bind:value={cfg.c.cooldowns.eventDelay.maxSeconds} /></label>
			<label class="text-xs text-slate-400 sm:col-span-2">Events correct chance ({Math.round((cfg.c.eventsCorrectChance ?? 0) * 100)}%)
				<input class="mt-2 w-full" type="range" min="0" max="1" step="0.01" bind:value={cfg.c.eventsCorrectChance} />
			</label>
		</div>
	</section>

	<section class="rounded-2xl border border-slate-800 bg-slate-900/60 p-5">
		<h3 class="font-semibold">Discord</h3>
		<div class="mt-5 space-y-4">
			<label class="flex items-center justify-between text-sm"><span>Auto read alerts</span><input type="checkbox" bind:checked={cfg.c.readAlerts} /></label>
			<label class="text-xs text-slate-400">Presence
				<select class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" value={cfg.c.discordStatus} onchange={(e) => updateStatus((e.target as HTMLSelectElement).value)}>
					<option value="online">Online</option><option value="idle">Idle</option><option value="dnd">Do Not Disturb</option><option value="invisible">Invisible</option>
				</select>
			</label>
		</div>
	</section>
</div>