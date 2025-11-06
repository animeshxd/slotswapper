import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import { useState, useId } from "react";
import { z } from "zod";
import { useMutation } from "@tanstack/react-query";
import { useAuthStore } from "@/features/auth/auth.store.ts";
import type { TreeifyError } from "@/lib/types.ts";
import { Button } from "@/components/ui/button.tsx"
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card.tsx"
import { Input } from "@/components/ui/input.tsx"
import { Label } from "@/components/ui/label.tsx"

const loginSchema = z.object({
	email: z.string().email(),
	password: z.string().min(1, "Password is required"),
});

type LoginSchema = z.infer<typeof loginSchema>;

export const LoginSearchSchema = z.object({
	redirect: z.string().optional(),
});

async function loginUser(values: LoginSchema) {
	const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/login`, {
		method: "POST",
		headers: { "Content-Type": "application/json" },
		body: JSON.stringify(values),
        credentials: "include",
	});

	if (!res.ok) {
		const error = await res.text();
		throw new Error(error || "Login failed");
	}

	return res.json();
}


export default function LoginComponent() {
	const navigate = useNavigate();
	const { redirect } = useSearch({
        strict: false,
    });
	const { setUser } = useAuthStore();
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [formErrors, setFormErrors] =
		useState<TreeifyError<LoginSchema> | null>(null);

	const mutation = useMutation({
		mutationFn: loginUser,
		onSuccess: (data) => {
			setUser(data.user);
			console.log(data);
			navigate({ to: redirect || "/dashboard" });
		},
		onError: (error) => {
			// You might want to display this error to the user
			console.error("Login failed:", error);
		},
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		const values = { email, password };
		const validationResult = loginSchema.safeParse(values);

		if (!validationResult.success) {
			setFormErrors(z.treeifyError(validationResult.error));
			return;
		}

		setFormErrors(null);
		mutation.mutate(validationResult.data);
	};

	const emailError = formErrors?.properties?.email?.errors[0];
	const passwordError = formErrors?.properties?.password?.errors[0];

	const emailId = useId();
	const passwordId = useId();

	return (
		<div className="flex min-h-screen items-center justify-center">
			<Card className="w-full max-w-sm">
				<CardHeader>
					<CardTitle className="text-2xl">Login</CardTitle>
					<CardDescription>
						Enter your email below to login to your account.
					</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4">
					<form onSubmit={handleSubmit} className="grid gap-4">
						<div className="grid gap-2">
							<Label htmlFor={emailId}>Email</Label>
							<Input id={emailId} type="email" value={email} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setEmail(e.target.value)} required />
							{emailError && <p className="text-red-500 text-xs">{emailError}</p>}
						</div>
						<div className="grid gap-2">
							<Label htmlFor={passwordId}>Password</Label>
							<Input id={passwordId} type="password" value={password} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setPassword(e.target.value)} required />
							{passwordError && <p className="text-red-500 text-xs">{passwordError}</p>}
						</div>
						<Button type="submit" className="w-full" disabled={mutation.isPending}>
							{mutation.isPending ? "Signing in..." : "Sign in"}
						</Button>
						{mutation.isError && <p className="text-red-500 text-xs mt-2">{mutation.error.message}</p>}
					</form>
				</CardContent>
				<CardFooter>
					<div className="mt-4 text-center text-sm">
						Don't have an account?{" "}
						<Link to="/register" className="underline">
							Sign up
						</Link>
					</div>
				</CardFooter>
			</Card>
		</div>
	);
}