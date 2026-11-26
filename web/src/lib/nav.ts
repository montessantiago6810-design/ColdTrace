import type { Role, SessionUser } from './session';

export interface NavItem {
	href: string;
	label: string;
	/** Empty: visible to everyone. Otherwise: signed-in users holding any of these roles. */
	roles: readonly Role[];
}

export const navItems: readonly NavItem[] = [
	{ href: '/', label: 'Overview', roles: [] },
	{ href: '/shipments', label: 'Shipments', roles: ['dispatcher', 'quality_reviewer', 'site_manager', 'auditor'] },
	{ href: '/cases', label: 'Cases', roles: ['quality_reviewer', 'auditor'] },
	{ href: '/devices', label: 'Devices', roles: ['device_admin'] }
];

/** Navigation entries the user may see. The API still checks every request. */
export function visibleNav(user: SessionUser | null): NavItem[] {
	return navItems.filter(
		(item) => item.roles.length === 0 || (user !== null && item.roles.some((r) => user.roles.includes(r)))
	);
}
