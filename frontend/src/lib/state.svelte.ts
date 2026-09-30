import { Events } from "$lib/wails-shim";

export type LogEntry = {
	level: string;
	type: string;
	username: string;
	message: string;
	timestamp: string;
	id: number;
};

const emptyConfig = {
	state: false,
	apiKey: "",
	gui: { theme: "dark" },
	readAlerts: false,
	discordStatus: "online",
	eventsCorrectChance: 0.65,
	cooldowns: {
		buttonClickDelay: { minSeconds: 0, maxSeconds: 0 },
		commandInterval: { minSeconds: 0, maxSeconds: 0 },
		breakCooldown: { minHours: 0, maxHours: 0 },
		breakDuration: { minHours: 0, maxHours: 0 },
		startDelay: { minMinutes: 0, maxMinutes: 0 },
		eventDelay: { minSeconds: 0, maxSeconds: 0 }
	},
	accounts: [],
	autoBuy: {},
	autoUse: {},
	commands: {},
	adventure: {}
};

class Cfg {
	c: any = $state(structuredClone(emptyConfig));
	loading = $state(true);
	saving = $state(false);
	error = $state("");

	constructor() {
		Events.On("configUpdate", (event) => {
			if (event.data && typeof event.data === "object") this.c = event.data;
		});
		void this.fetch();
	}

	async fetch() {
		this.loading = true;
		try {
			const response = await api("/api/config");
			if (response.status === 401) {
				window.location.hash = "#/login";
				return;
			}
			if (!response.ok) throw new Error(await response.text());
			this.c = await response.json();
		} catch (error) {
			this.error = error instanceof Error ? error.message : String(error);
		} finally {
			this.loading = false;
		}
	}

	async save() {
		this.saving = true;
		this.error = "";
		try {
			const response = await api("/api/config", {
				method: "POST",
				headers: { "Content-Type": "application/json" },
				body: JSON.stringify(this.c)
			});
			if (!response.ok) throw new Error(await response.text());
		} catch (error) {
			this.error = error instanceof Error ? error.message : String(error);
		} finally {
			this.saving = false;
		}
	}
}

class Instances {
	i = $state<any[]>([]);
	constructor() {
		Events.On("instanceUpdate", (event) => this.upsert(event.data));
		void this.fetch();
	}
	async fetch() {
		try {
			const response = await api("/api/instances");
			if (response.ok) this.i = await response.json();
		} catch {}
	}
	upsert(instance: any) {
		if (!instance) return;
		const id = instance.instanceId;
		const index = this.i.findIndex((item) => item.instanceId === id);
		if (index === -1) this.i.push(instance);
		else this.i[index] = instance;
	}
	async restart(id: string) {
		return api("/api/instances/" + encodeURIComponent(id) + "/restart", { method: "POST" });
	}
	async stop(id: string) {
		return api("/api/instances/" + encodeURIComponent(id), {
			method: "DELETE",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ restarting: false })
		});
	}
	async restartAll() {
		return api("/api/instances/restart", { method: "POST" });
	}
}

class Logs {
	importantLogs = $state<LogEntry[]>([]);
	othersLogs = $state<LogEntry[]>([]);
	discordLogs = $state<LogEntry[]>([]);
	nextId = 0;
	constructor() {
		Events.On("log", (event) => {
			const data = event.data as any;
			const entry: LogEntry = {
				level: data.level ?? "",
				type: data.type ?? "",
				username: data.username ?? "",
				message: String(data.message ?? ""),
				timestamp: new Date().toLocaleTimeString([], { hour: "numeric", minute: "2-digit" }),
				id: this.nextId++
			};
			if (entry.level === "important") this.importantLogs.push(entry);
			else if (entry.level === "discord") this.discordLogs.push(entry);
			else this.othersLogs.push(entry);
		});
	}
}

export async function api(path: string, init: RequestInit = {}) {
	return fetch(path, { ...init, credentials: "same-origin" });
}

export const cfg = $state(new Cfg());
export const instances = $state(new Instances());
export const logs = $state(new Logs());
