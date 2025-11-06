import { create } from "zustand";

export interface User {
	id: number;
	name: string;
	email: string;
}

interface AuthState {
	user: User | null;
	isAuthenticated: boolean;
	setUser: (user: User | null) => void;
	logout: () => void;
}

export const useAuthStore = create<AuthState>((set) => ({
	user: null,
	isAuthenticated: false,
	setUser: (user) => set({ user, isAuthenticated: !!user }),
	logout: () => set({ user: null, isAuthenticated: false }),
}));

export async function serverLogout() {
	await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/logout`, {
		method: "POST",
		credentials: "include",
	});
}
