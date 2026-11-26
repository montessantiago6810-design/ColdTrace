/** Roles from the charter. The Go API enforces them; the UI uses them for navigation only. */
export const roles = ['dispatcher', 'quality_reviewer', 'site_manager', 'auditor', 'device_admin'] as const;

export type Role = (typeof roles)[number];

export interface SessionUser {
	id: string;
	name: string;
	orgSlug: string;
	roles: Role[];
}

/**
 * Resolves the signed-in user for a request. Session auth (a secure cookie
 * validated by the Go API) is not built yet, so every request is anonymous.
 */
export async function resolveSession(): Promise<SessionUser | null> {
	return null;
}
