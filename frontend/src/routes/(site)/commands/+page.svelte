<script lang="ts">
	import { cfg } from "$lib/state.svelte";
	import { Slider } from "$lib/components/ui/slider";

	const enumValues: Record<string, string[]> = {
		adventureOption: ["brazil", "space", "vacation", "west"],
		fishLocation: ["Vertigo Beach", "Wily River", "Underwater Sanctuary", "Camp Guillermo", "Scurvy Waters", "Northpoint Cabin"]
	};

	function label(key: string) {
		return key.replace(/([A-Z])/g, " $1").replace(/^./, (s) => s.toUpperCase());
	}

	function valuesFor(key: string, value: any) {
		if (key === "adventureOption") return enumValues.adventureOption;
		if (key === "fishLocation") return enumValues.fishLocation;
		return Array.isArray(value) ? value : [];
	}

	function toggleFishOnly(value: boolean) {
		cfg.c.fish.fishOnly = value;
		if (value) {
			for (const [key, command] of Object.entries(cfg.c.commands ?? {})) {
				if (key !== "fish" && command && typeof command === "object") (command as any).state = false;
			}
			cfg.c.fish.state = true;
		}
	}

	function updateArray(command: any, key: string, event: Event) {
		const value = (event.target as HTMLInputElement).value;
		command[key] = value.split(",").map((item) => item.trim()).filter(Boolean);
	}
</script>

<div class="space-y-5">
	<div>
		<p class="text-xs font-medium uppercase tracking-[0.2em] text-indigo-400">Automation</p>
		<h2 class="mt-1 text-2xl font-bold">Auto Grind Commands</h2>
		<p class="mt-1 max-w-3xl text-sm text-slate-500">The original DMG command configuration is retained. This page only changes how you control it from the browser.</p>
	</div>

	<div class="grid gap-4 xl:grid-cols-2">
		{#each Object.entries(cfg.c.commands ?? {}) as [commandKey, command] (commandKey)}
			{@const cmd = command as any}
			<section class="rounded-2xl border {cmd.state ? "border-indigo-500/40 bg-indigo-500/[0.04]" : "border-slate-800 bg-slate-900/60"} p-4">
				<div class="flex items-center justify-between gap-3">
					<div>
						<h3 class="font-semibold">{label(commandKey)}</h3>
						<p class="text-xs text-slate-500">Minimum interval: {cmd.delay ?? 0}s</p>
					</div>
					<button
						class="rounded-full px-3 py-1 text-xs {cmd.state ? "bg-emerald-500/15 text-emerald-300" : "bg-slate-800 text-slate-500"}"
						disabled={cfg.c.fish?.fishOnly && commandKey.toLowerCase() !== "fish"}
						onclick={() => {
							cmd.state = !cmd.state;
							if (cmd.state && cfg.c.fish?.fishOnly && commandKey.toLowerCase() !== "fish") cmd.state = false;
						}}
					>{cmd.state ? "Enabled" : "Disabled"}</button>
				</div>

				<div class="mt-4 grid gap-3 sm:grid-cols-2">
					{#each Object.entries(cmd) as [key, value] (key)}
						{#if key !== "state" && key !== "delay" && key !== "fishOnlyDelay"}
							<div class="space-y-1.5 {Array.isArray(value) ? "sm:col-span-2" : ""}">
								<label class="text-xs text-slate-500">{label(key)}</label>
								{#if typeof value === "boolean"}
									<button class="block rounded-lg border border-slate-700 px-3 py-2 text-xs {value ? "border-indigo-500/50 bg-indigo-500/10 text-indigo-300" : "text-slate-400"}" onclick={() => (cmd[key] = !value)}>{value ? "On" : "Off"}</button>
								{:else if typeof value === "number"}
									<input class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-indigo-500" type="number" step="0.1" value={value} oninput={(e) => (cmd[key] = Number((e.target as HTMLInputElement).value))} />
								{:else if Array.isArray(value)}
									<input class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm outline-none focus:border-indigo-500" value={value.join(", ")} oninput={(e) => updateArray(cmd, key, e)} />
								{:else if enumValues[key]}
									<select class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" value={value} onchange={(e) => (cmd[key] = (e.target as HTMLSelectElement).value)}>
										{#each valuesFor(key, value) as option}<option value={option}>{option}</option>{/each}
									</select>
								{:else}
									<input class="w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm" value={value} oninput={(e) => (cmd[key] = (e.target as HTMLInputElement).value)} />
								{/if}
							</div>
						{:else if key === "fishOnlyDelay"}
							<div class="sm:col-span-2">
								<label class="text-xs text-slate-500">Fish-only delay</label>
								<Slider type="multiple" bind:value={[cmd.fishOnlyDelay.minSeconds, cmd.fishOnlyDelay.maxSeconds]} max={20} min={0} step={0.1} />
							</div>
						{/if}
					{/each}
				</div>

				{#if commandKey.toLowerCase() === "fish"}
					<div class="mt-4 rounded-xl border border-slate-800 bg-slate-950/50 p-3">
						<div class="flex items-center justify-between">
							<div><p class="text-sm font-medium">Fish Only Mode</p><p class="text-xs text-slate-500">Preserves the existing fish-only behavior.</p></div>
							<button class="rounded-full px-3 py-1 text-xs {cmd.fishOnly ? "bg-indigo-500/15 text-indigo-300" : "bg-slate-800 text-slate-500"}" onclick={() => toggleFishOnly(!cmd.fishOnly)}>{cmd.fishOnly ? "Enabled" : "Disabled"}</button>
						</div>
					</div>
				{/if}
			</section>
		{/each}
	</div>

	<p class="text-xs text-slate-600">Changes are held locally until you press <strong class="text-slate-400">Save changes</strong>.</p>
</div>