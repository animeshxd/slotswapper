import { createFileRoute } from "@tanstack/react-router";
import RegisterComponent from "@/features/auth/register.component";

export const Route = createFileRoute("/(public)/register")({
	component: RegisterComponent,
});

