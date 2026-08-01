import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig, searchForWorkspaceRoot } from "vite";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

// In web-server mode we don't use the Wails vite plugin.
// @wailsio/runtime is aliased to our SSE-based shim.
export default defineConfig({
	server: {
		fs: {
			allow: [
				searchForWorkspaceRoot(process.cwd()),
				"./bindings/*"
			]
		},
		host: true,
		allowedHosts: "all"
	},
	resolve: {
		alias: {
			"@": path.resolve(__dirname, "./"),
			// Redirect the Wails runtime to our SSE shim.
			"@wailsio/runtime": path.resolve(__dirname, "./src/lib/wails-shim.ts")
		}
	},
	plugins: [sveltekit(), tailwindcss()]
});
