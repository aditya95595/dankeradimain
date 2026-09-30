<script lang="ts">
	import { page } from "$app/state";
	import { cfg, instances } from "$lib/state.svelte";
	import { Moon, Sun } from "svelte-radix";

	let dark = $state(true);
	let title = $derived((page.url.hash.split("#/")[1] || "overview").split("/")[0] || "overview");

	function toggleTheme() {
		dark = !dark;
		document.documentElement.classList.toggle("dark", dark);
		cfg.c.gui.theme = dark ? "dark" : "light";
	}

	function toggleBot() {
		cfg.c.state = !cfg.c.state;
	}

	function toggleCommands(value: boolean) {
		for (const command of Object.values(cfg.c.commands ?? {})) {
			if (command && typeof command === "object" && "state" in command) command.state = value;
		}
	}
</script>

<header class="sticky top-0 z-50 flex h-14 items-center justify-between border-b border-slate-800 bg-slate-950/90 px-4 backdrop-blur md:px-6">
	<div>
		<h1 class="text-sm font-semibold capitalize text-white md:text-base">{title}</h1>
		<p class="hidden text-[11px] text-slate-500 sm:block">Remote control · live events enabled</p>
	</div>
	<div class="flex items-center gap-2">
		{#if title === "commands"}
			<button class="hidden rounded-lg border border-slate-800 px-3 py-1.5 text-xs text-slate-300 hover:bg-slate-900 sm:block" onclick={() => toggleCommands(true)}>Enable all</button>
			<button class="hidden rounded-lg border border-red-900/50 px-3 py-1.5 text-xs text-red-300 hover:bg-red-950/30 sm:block" onclick={() => toggleCommands(false)}>Disable all</button>
		{/if}
		<button class="rounded-lg border border-slate-800 p-2 hover:bg-slate-900" onclick={toggleTheme} title="Toggle theme">
			{#if dark}<Moon class="size-4" />{:else}<Sun class="size-4" />{/if}
		</button>
		<button class="rounded-lg px-3 py-1.5 text-xs font-medium {cfg.c.state ? "bg-emerald-500/15 text-emerald-300" : "bg-red-500/15 text-red-300"}" onclick={toggleBot}>
			{cfg.c.state ? "Enabled" : "Disabled"}
		</button>
	</div>
</header>