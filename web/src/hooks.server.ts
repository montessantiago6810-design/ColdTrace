import type { Handle } from '@sveltejs/kit/hooks';
import { resolveSession } from '#lib/session.ts';

export const handle: Handle = async ({ event, resolve }) => {
	event.locals.user = await resolveSession();
	const response = await resolve(event);
	response.headers.set('x-content-type-options', 'nosniff');
	response.headers.set('referrer-policy', 'same-origin');
	response.headers.set('x-frame-options', 'DENY');
	return response;
};
