import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = ({ locals }) => {
	if (locals.user) redirect(303, '/');
	return {};
};

export const actions: Actions = {
	default: async ({ request }) => {
		const form = await request.formData();
		const email = String(form.get('email') ?? '').trim();
		const password = String(form.get('password') ?? '');

		if (!email || !password) {
			return fail(400, { email, message: 'Enter your email and password.' });
		}
		// Session auth is implemented in the Go API in a later step. Until then
		// nothing is checked or stored, and the password is never echoed back.
		return fail(503, { email, message: 'Sign-in is not available yet.' });
	}
};
