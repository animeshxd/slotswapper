import {
	createFileRoute,
	Outlet,
	redirect,
	type LoaderFnContext,
  Link,
} from "@tanstack/react-router";

export const Route = createFileRoute("/_protected")({
	beforeLoad: ({ context, location }: LoaderFnContext) => {
		// If the user is not authenticated, redirect them to the login page
		if (!context.auth.isAuthenticated) {
			throw redirect({
				to: "/login",
				search: {
					redirect: location.href,
				},
			});
		}
	},
	component: () => (
    <div>
      <div className="p-2 flex gap-2">
        <Link to="/dashboard" className="[&.active]:font-bold">
          Dashboard
        </Link>{" "}
        <Link to="/marketplace" className="[&.active]:font-bold">
          Marketplace
        </Link>
        <Link to="/requests" className="[&.active]:font-bold">
          Requests
        </Link>
      </div>
      <hr />
      <Outlet />
    </div>
  ),
});