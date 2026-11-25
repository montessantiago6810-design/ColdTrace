import type { SessionUser } from '#lib/session.ts';

declare global {
	namespace App {
		interface Error {
			message: string;
		}
		interface Locals {
			user: SessionUser | null;
		}
		interface PageData {
			user: SessionUser | null;
		}
	}
}

export {};
