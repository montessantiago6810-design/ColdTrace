/** Shape of GET /api/v1/health (internal/api.HealthResponse). */
export interface Health {
	status: string;
	service: string;
	version: string;
}

export type HealthResult =
	| { ok: true; health: Health; requestId: string | null }
	| { ok: false; reason: string };

export const defaultApiUrl = 'http://127.0.0.1:8080';

/**
 * Returns the API base URL without a trailing slash. Only http(s) URLs are
 * accepted; anything else falls back to the local default.
 */
export function apiBaseUrl(raw: string | undefined): string {
	const value = raw?.trim();
	if (!value) return defaultApiUrl;
	try {
		const url = new URL(value);
		if (url.protocol !== 'http:' && url.protocol !== 'https:') return defaultApiUrl;
		return url.toString().replace(/\/+$/, '');
	} catch {
		return defaultApiUrl;
	}
}

function isHealth(body: unknown): body is Health {
	if (typeof body !== 'object' || body === null) return false;
	const b = body as Record<string, unknown>;
	return typeof b.status === 'string' && typeof b.service === 'string' && typeof b.version === 'string';
}

/** Calls the API health endpoint. Never throws; failures become `ok: false`. */
export async function fetchHealth(
	fetchFn: typeof fetch,
	baseUrl: string,
	timeoutMs = 2000
): Promise<HealthResult> {
	let response: Response;
	try {
		response = await fetchFn(`${baseUrl}/api/v1/health`, {
			headers: { accept: 'application/json' },
			signal: AbortSignal.timeout(timeoutMs)
		});
	} catch (err) {
		const timedOut = err instanceof DOMException && err.name === 'TimeoutError';
		return { ok: false, reason: timedOut ? `no answer within ${timeoutMs} ms` : 'API unreachable' };
	}

	if (!response.ok) return { ok: false, reason: `API answered ${response.status}` };

	let body: unknown;
	try {
		body = await response.json();
	} catch {
		return { ok: false, reason: 'API returned invalid JSON' };
	}
	if (!isHealth(body)) return { ok: false, reason: 'unexpected health response' };

	return { ok: true, health: body, requestId: response.headers.get('x-request-id') };
}
