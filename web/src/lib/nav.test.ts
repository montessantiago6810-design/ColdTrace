import { describe, expect, it } from 'vitest';
import { navItems, visibleNav } from './nav';
import type { SessionUser } from './session';

const user = (roles: SessionUser['roles']): SessionUser => ({
	id: 'u1',
	name: 'Test User',
	orgSlug: 'nordic-pharma',
	roles
});

const hrefs = (u: SessionUser | null) => visibleNav(u).map((i) => i.href);

describe('visibleNav', () => {
	it('shows only public entries to anonymous visitors', () => {
		expect(hrefs(null)).toEqual(['/']);
	});

	it('filters by role', () => {
		expect(hrefs(user(['dispatcher']))).toEqual(['/', '/shipments']);
		expect(hrefs(user(['auditor']))).toEqual(['/', '/shipments', '/cases']);
		expect(hrefs(user(['device_admin']))).toEqual(['/', '/devices']);
	});

	it('combines roles without duplicates', () => {
		expect(hrefs(user(['quality_reviewer', 'device_admin']))).toEqual(['/', '/shipments', '/cases', '/devices']);
	});

	it('keeps hrefs unique', () => {
		expect(new Set(navItems.map((i) => i.href)).size).toBe(navItems.length);
	});
});
