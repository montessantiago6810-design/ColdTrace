import { describe, expect, it } from 'vitest';
import { actions } from './+page.server';

type Action = NonNullable<(typeof actions)['default']>;
const login = actions.default as Action;

function submit(fields: Record<string, string>) {
	const body = new FormData();
	for (const [k, v] of Object.entries(fields)) body.set(k, v);
	const request = new Request('http://localhost/login', { method: 'POST', body });
	return login({ request } as Parameters<Action>[0]);
}

describe('login action (placeholder)', () => {
	it('requires email and password', async () => {
		const result = await submit({ email: ' qa@example.test ' });
		expect(result).toMatchObject({ status: 400, data: { email: 'qa@example.test' } });
	});

	it('answers 503 until session auth exists and never echoes the password', async () => {
		const result = await submit({ email: 'qa@example.test', password: 'hunter2' });
		expect(result).toMatchObject({ status: 503, data: { message: 'Sign-in is not available yet.' } });
		expect(JSON.stringify(result)).not.toContain('hunter2');
	});
});
