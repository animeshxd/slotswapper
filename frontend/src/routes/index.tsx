import { createFileRoute, redirect } from "@tanstack/react-router";
import { useAuthStore, type User } from "../lib/store.ts";

async function fetchMe(): Promise<User | { user: User }> {
	const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/me`);
	if (!res.ok) {
		throw new Error("Not authenticated");
	}
	return res.json();
}

export const Route = createFileRoute("/")({
	beforeLoad: async () => {
		try {
			const data = await fetchMe();
			const user = "user" in data ? data.user : data; // Handle both { user: User } and User directly
			useAuthStore.getState().setUser(user);
			throw redirect({
				to: "/dashboard",
			});
		} catch (_error) {
			useAuthStore.getState().setUser(null);
			throw redirect({
				to: "/login",
				search: {
					redirect: "/dashboard",
				},
			});
		}
	},
	component: () => <div>Loading...</div>, // This will be shown briefly before redirect
});
