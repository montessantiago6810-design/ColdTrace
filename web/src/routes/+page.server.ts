import { COLDTRACE_API_URL } from '$app/env/private';
import { apiBaseUrl, fetchHealth } from '#lib/api.ts';
import type { PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ fetch }) => ({
	health: await fetchHealth(fetch, apiBaseUrl(COLDTRACE_API_URL))
});
