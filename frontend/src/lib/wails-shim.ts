/**
 * Wails runtime shim for web-server mode.
 *
 * Replaces @wailsio/runtime so the existing Svelte components work unchanged
 * when the app is served as a plain HTTP server instead of a Wails desktop app.
 *
 * Events  → Server-Sent Events (SSE) at /api/events
 * Browser → window.open
 */

type EventHandler = (event: { data: unknown }) => void;

const handlers: Record<string, EventHandler[]> = {};

let eventSource: EventSource | null = null;

function connect() {
  if (eventSource) return;

  eventSource = new EventSource('/api/events');

  eventSource.onmessage = (e: MessageEvent) => {
    try {
      const { name, data } = JSON.parse(e.data) as { name: string; data: unknown };
      const list = handlers[name];
      if (list) {
        list.forEach((cb) => cb({ data }));
      }
    } catch {
      // ignore malformed events
    }
  };

  eventSource.onerror = () => {
    // Reconnect after 3 seconds
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }
    setTimeout(connect, 3000);
  };
}

// Start connecting immediately.
connect();

export const Events = {
  /**
   * Register a handler for a named event.
   * The callback receives `{ data }` matching Wails event format.
   */
  On(name: string, cb: EventHandler): void {
    if (!handlers[name]) handlers[name] = [];
    handlers[name].push(cb);
  },

  Off(name: string, cb?: EventHandler): void {
    if (!handlers[name]) return;
    if (cb) {
      handlers[name] = handlers[name].filter((h) => h !== cb);
    } else {
      delete handlers[name];
    }
  },

  Emit(_name: string, ..._args: unknown[]): void {
    // Client-to-server events are handled via REST calls; nothing to do here.
  }
};

export const Browser = {
  OpenURL(url: string): void {
    window.open(url, '_blank', 'noopener,noreferrer');
  }
};
