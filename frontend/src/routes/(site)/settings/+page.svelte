<script lang="ts">
	import General from "./general.svelte";
	import AutoBuy from "./auto-buy.svelte";
	import AutoUse from "./auto-use.svelte";
</script>

<div class="space-y-6">
	<div>
		<p class="text-xs font-medium uppercase tracking-[0.2em] text-indigo-400">Configuration</p>
		<h2 class="mt-1 text-2xl font-bold">Settings</h2>
		<p class="mt-1 text-sm text-slate-500">All existing DMG configuration is edited from the browser. Press Save changes when finished.</p>
	</div>
	<General />
	<div class="grid gap-4 lg:grid-cols-2">
		<AutoBuy />
		<AutoUse />
	</div>
</div>