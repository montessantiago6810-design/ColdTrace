import { describe, expect, it, vi } from 'vitest';
import { apiBaseUrl, defaultApiUrl, fetchHealth } from './api';

describe('apiBaseUrl', () => {
	it('falls back to the local API when unset or invalid', () => {
		expect(apiBaseUrl(undefined)).toBe(defaultApiUrl);
		expect(apiBaseUrl('  ')).toBe(defaultApiUrl);
		expect(apiBaseUrl('not a url')).toBe(defaultApiUrl);
		expect(apiBaseUrl('file:///etc/passwd')).toBe(defaultApiUrl);
	});

	it('trims trailing slashes', () => {
		expect(apiBaseUrl('https://api.example.test/')).toBe('https://api.example.test');
		expect(apiBaseUrl(' http://127.0.0.1:9090// ')).toBe('http://127.0.0.1:9090');
	});
});

function jsonResponse(body: unknown, status = 200, headers: Record<string, string> = {}): Response {
	return new Response(JSON.stringify(body), {
		status,
		headers: { 'content-type': 'application/json', ...headers }
	});
}

describe('fetchHealth', () => {
	it('returns the health body and request id', async () => {
		const fetchFn = vi.fn(async () =>
			jsonResponse({ status: 'ok', service: 'coldtrace-api', version: 'v0.1.0' }, 200, { 'x-request-id': 'req-1' })
		);
		const result = await fetchHealth(fetchFn as unknown as typeof fetch, 'http://api');
		expect(fetchFn).toHaveBeenCalledWith('http://api/api/v1/health', expect.objectContaining({ signal: expect.any(AbortSignal) }));
		expect(result).toEqual({
			ok: true,
			health: { status: 'ok', service: 'coldtrace-api', version: 'v0.1.0' },
			requestId: 'req-1'
		});
	});

	it('reports non-2xx answers', async () => {
		const result = await fetchHealth(async () => jsonResponse({}, 503), 'http://api');
		expect(result).toEqual({ ok: false, reason: 'API answered 503' });
	});

	it('reports network errors without throwing', async () => {
		const result = await fetchHealth(async () => {
			throw new TypeError('fetch failed');
		}, 'http://api');
		expect(result).toEqual({ ok: false, reason: 'API unreachable' });
	});

	it('reports timeouts', async () => {
		const result = await fetchHealth(async () => {
			throw new DOMException('timed out', 'TimeoutError');
		}, 'http://api', 50);
		expect(result).toEqual({ ok: false, reason: 'no answer within 50 ms' });
	});

	it('rejects invalid JSON and unexpected shapes', async () => {
		const invalid = await fetchHealth(async () => new Response('<html>', { status: 200 }), 'http://api');
		expect(invalid).toEqual({ ok: false, reason: 'API returned invalid JSON' });

		const wrongShape = await fetchHealth(async () => jsonResponse({ status: 'ok' }), 'http://api');
		expect(wrongShape).toEqual({ ok: false, reason: 'unexpected health response' });
	});
});
