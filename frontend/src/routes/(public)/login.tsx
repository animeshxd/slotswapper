import LoginComponent, {
	LoginSearchSchema,
} from "@/features/auth/login.component";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/(public)/login")({
	validateSearch: LoginSearchSchema,
	component: LoginComponent,
});
