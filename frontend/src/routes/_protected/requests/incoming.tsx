import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

// Define the types for the swap requests
interface IncomingSwapRequest {
	id: number;
	requester_name: string;
	requester_event_title: string;
	requester_event_start_time: string;
	requester_event_end_time: string;
	responder_event_title: string;
	responder_event_start_time: string;
	responder_event_end_time: string;
}

// API functions
async function fetchIncomingRequests(): Promise<IncomingSwapRequest[]> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swap-requests/incoming`,
		{ credentials: "include" },
	);
	if (!res.ok) {
		throw new Error("Failed to fetch incoming requests");
	}
	return res.json();
}

async function respondToSwapRequest(
	id: number,
	status: "ACCEPTED" | "REJECTED",
): Promise<void> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swap-response/${id}`,
		{
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ status }),
			credentials: "include",
		},
	);
	if (!res.ok) {
		throw new Error("Failed to respond to swap request");
	}
}

export const Route = createFileRoute("/_protected/requests/incoming")({
	component: IncomingRequestsComponent,
});

function IncomingRequestsComponent() {
	const queryClient = useQueryClient();

	const { data: incoming, isLoading: isLoadingIncoming } = useQuery<
		IncomingSwapRequest[]
	>({
		queryKey: ["incoming-requests"],
		queryFn: fetchIncomingRequests,
	});

	const mutation = useMutation({
		mutationFn: ({
			id,
			status,
		}: {
			id: number;
			status: "ACCEPTED" | "REJECTED";
		}) => respondToSwapRequest(id, status),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["incoming-requests"] });
			queryClient.invalidateQueries({ queryKey: ["outgoing-requests"] });
			queryClient.invalidateQueries({ queryKey: ["swappable-events"] });
			queryClient.invalidateQueries({ queryKey: ["events"] });
		},
	});

	const isEmpty = !incoming?.length && !isLoadingIncoming;
	return (
		<div className="p-4 md:p-8">
			<h1 className="text-2xl font-bold mb-4">Incoming Requests</h1>

			{isLoadingIncoming && <div>Loading...</div>}

			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				{isEmpty && <p className="">You have no incoming requests.</p>}
				{incoming?.map((req) => (
					<Card key={req.id}>
						<CardHeader>
							<CardTitle>Swap Request</CardTitle>
						</CardHeader>
						<CardContent className="grid gap-2">
							<p>
								<strong>From:</strong> {req.requester_name}
							</p>
							<p>
								<strong>Their Event:</strong> {req.requester_event_title} (
								{new Date(req.requester_event_start_time).toLocaleString()})
							</p>
							<p>
								<strong>Your Event:</strong> {req.responder_event_title} (
								{new Date(req.responder_event_start_time).toLocaleString()})
							</p>
							<div className="flex gap-2 mt-4">
								<Button
									size="sm"
									onClick={() =>
										mutation.mutate({ id: req.id, status: "ACCEPTED" })
									}
								>
									Accept
								</Button>
								<Button
									size="sm"
									variant="outline"
									onClick={() =>
										mutation.mutate({ id: req.id, status: "REJECTED" })
									}
								>
									Reject
								</Button>
							</div>
						</CardContent>
					</Card>
				))}
			</div>
		</div>
	);
}
