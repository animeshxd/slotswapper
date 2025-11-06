import * as ReactDOM from "react-dom/client";
import { RouterProvider, createRouter } from "@tanstack/react-router";

import {
	// biome-ignore lint/nursery/noUnresolvedImports: false positive
	StrictMode,
	useEffect,
	useState,
} from "react";

import * as TanStackQueryProvider from "./integrations/tanstack-query/root-provider.tsx";

// Import the generated route tree
import { routeTree } from "./routeTree.gen.ts";
import { useAuthStore } from "@/features/auth/auth.store.ts";

import "./styles.css";
import reportWebVitals from "./reportWebVitals.ts";

// Create a new router instance

const TanStackQueryProviderContext = TanStackQueryProvider.getContext();
const router = createRouter({
	routeTree,
	context: {
		...TanStackQueryProviderContext,
		auth: useAuthStore.getState(),
	},
	defaultPreload: "intent",
	scrollRestoration: true,
	defaultStructuralSharing: true,
	defaultPreloadStaleTime: 0,
});

// Subscribe to the auth store and update the router context
useAuthStore.subscribe((state) => {
	router.options.context.auth = state;
});

// Register the router instance for type safety
declare module "@tanstack/react-router" {
	interface Register {
		router: typeof router;
	}
}

function App() {
	const { isAuthenticated, setUser, logout } = useAuthStore();
	const [isLoading, setIsLoading] = useState(true);

	useEffect(() => {
		const fetchUser = async () => {
			try {
				const res = await fetch(
					`${import.meta.env.VITE_HTTP_SERVER_URL}/api/me`,
					{ credentials: "include" },
				);
				if (!res.ok) {
					throw new Error("Failed to fetch user");
				}
				const user = await res.json();
				setUser(user);
			} catch (error) {
				console.error(error);
				logout();
			}
		};
		console.log({ isAuthenticated, isLoading });
		if (!isAuthenticated) {
			fetchUser().finally(() => setIsLoading(false));
		} else {
			setIsLoading(false);
		}
	}, [isAuthenticated, setUser, logout, isLoading]);

	if (isLoading) {
		return <div>Loading...</div>; // Or a proper loading spinner
	}

	return <RouterProvider router={router} />;
}

const rootElement = document.getElementById("app");
if (rootElement && !rootElement.innerHTML) {
	const root = ReactDOM.createRoot(rootElement);
	root.render(
		<StrictMode>
			<TanStackQueryProvider.Provider {...TanStackQueryProviderContext}>
				<App />
			</TanStackQueryProvider.Provider>
		</StrictMode>,
	);
}

reportWebVitals();
