type EventHandler = (event: { data: any }) => void;

const handlers: Record<string, EventHandler[]> = {};
let eventSource: EventSource | null = null;
let reconnectTimer: number | undefined;

function connect() {
	if (typeof window === "undefined" || eventSource) return;
	eventSource = new EventSource("/api/events");
	eventSource.onmessage = (event) => {
		try {
			const payload = JSON.parse(event.data);
			handlers[payload.name]?.forEach((handler) => handler({ data: payload.data }));
		} catch {}
	};
	eventSource.onerror = () => {
		eventSource?.close();
		eventSource = null;
		if (reconnectTimer) window.clearTimeout(reconnectTimer);
		reconnectTimer = window.setTimeout(connect, 3000);
	};
}

export const Events = {
	On(name: string, callback: EventHandler) {
		if (!handlers[name]) handlers[name] = [];
		handlers[name].push(callback);
		connect();
	},
	Off(name: string, callback?: EventHandler) {
		if (!handlers[name]) return;
		if (callback) handlers[name] = handlers[name].filter((item) => item !== callback);
		else delete handlers[name];
	}
};

export const Browser = {
	OpenURL(url: string) {
		if (typeof window !== "undefined") window.open(url, "_blank", "noopener,noreferrer");
	}
};
