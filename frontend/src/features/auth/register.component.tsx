import { Link, useNavigate } from "@tanstack/react-router";
import { useState, useId } from "react";
import { z } from "zod";
import { useMutation } from "@tanstack/react-query";
import { useAuthStore } from "./auth.store";
import type { TreeifyError } from "@/lib/types.ts";
import { Button } from "@/components/ui/button.tsx";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Label } from "@/components/ui/label.tsx";

const registerSchema = z.object({
	name: z.string().min(1, "Name is required"),
	email: z.email(),
	password: z.string().min(8, "Password must be at least 8 characters"),
});

type RegisterSchema = z.infer<typeof registerSchema>;

async function registerUser(values: RegisterSchema) {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/signup`,
		{
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(values),
		},
	);

	if (!res.ok) {
		const error = await res.text();
		throw new Error(error || "Registration failed");
	}

	return res.json();
}


export default function RegisterComponent() {
	const navigate = useNavigate();
	const { setUser } = useAuthStore();
	const [name, setName] = useState("");
	const [email, setEmail] = useState("");
	const [password, setPassword] = useState("");
	const [formErrors, setFormErrors] =
		useState<TreeifyError<RegisterSchema> | null>(null);

	const mutation = useMutation({
		mutationFn: registerUser,
		onSuccess: (data) => {
			setUser(data.user);
			navigate({ to: "/dashboard" });
		},
	});

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		const values = { name, email, password };
		const validationResult = registerSchema.safeParse(values);

		if (!validationResult.success) {
			setFormErrors(z.treeifyError(validationResult.error));
			return;
		}

		setFormErrors(null);
		mutation.mutate(validationResult.data);
	};

	const nameError = formErrors?.properties?.name?.errors[0];
	const emailError = formErrors?.properties?.email?.errors[0];
	const passwordError = formErrors?.properties?.password?.errors[0];

	const nameId = useId();
	const emailId = useId();
	const passwordId = useId();

	return (
		<div className="flex min-h-screen items-center justify-center">
			<Card className="w-full max-w-sm">
				<CardHeader>
					<CardTitle className="text-2xl">Sign Up</CardTitle>
					<CardDescription>
						Enter your information to create an account.
					</CardDescription>
				</CardHeader>
				<CardContent className="grid gap-4">
					<form onSubmit={handleSubmit} className="grid gap-4">
						<div className="grid gap-2">
							<Label htmlFor={nameId}>Name</Label>
							<Input id={nameId} value={name} onChange={(e: React.ChangeEvent<HTMLInputElement>) => setName(e.target.value)} required />
							{nameError && <p className="text-red-500 text-xs">{nameError}</p>}
						</div>
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
							{mutation.isPending ? "Creating account..." : "Create an account"}
						</Button>
						{mutation.isError && <p className="text-red-500 text-xs mt-2">{mutation.error.message}</p>}
					</form>
				</CardContent>
				<CardFooter>
					<div className="mt-4 text-center text-sm">
						Already have an account?{" "}
						<Link to="/login" className="underline">
							Sign in
						</Link>
					</div>
				</CardFooter>
			</Card>
		</div>
	);
}
