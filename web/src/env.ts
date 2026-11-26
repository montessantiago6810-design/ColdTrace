import { defineEnvVars } from '@sveltejs/kit/env';

export const variables = defineEnvVars({
	COLDTRACE_API_URL: {
		description: 'Base URL of the ColdTrace Go API, used by server-side loads. Defaults to http://127.0.0.1:8080.',
		schema: (value) => value
	}
});
