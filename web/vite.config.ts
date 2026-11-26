import adapter from '@sveltejs/adapter-node';
import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vitest/config';

export default defineConfig({
	plugins: [sveltekit({ adapter: adapter() })],
	server: {
		host: '127.0.0.1',
		port: 5173,
		strictPort: true
	},
	test: {
		include: ['src/**/*.test.ts'],
		environment: 'node'
	}
});
