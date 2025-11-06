import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

// Define the types for the swap requests
interface OutgoingSwapRequest {
	id: number;
	responder_name: string;
	requester_event_title: string;
	requester_event_start_time: string;
	requester_event_end_time: string;
	responder_event_title: string;
	responder_event_start_time: string;
	responder_event_end_time: string;
}

// API functions
async function fetchOutgoingRequests(): Promise<OutgoingSwapRequest[]> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swap-requests/outgoing`,
		{ credentials: "include" },
	);
	if (!res.ok) {
		throw new Error("Failed to fetch outgoing requests");
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

export const Route = createFileRoute("/_protected/requests/outgoing")({
	component: OutgoingRequestsComponent,
});

function OutgoingRequestsComponent() {
	const queryClient = useQueryClient();

	const { data: outgoing, isLoading: isLoadingOutgoing } = useQuery<
		OutgoingSwapRequest[]
	>({
		queryKey: ["outgoing-requests"],
		queryFn: fetchOutgoingRequests,
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
	const isEmpty = !outgoing?.length && !isLoadingOutgoing;

	return (
		<div className="p-4 md:p-8">
			<h1 className="text-2xl font-bold mb-4">Outgoing Requests</h1>

			{isLoadingOutgoing && <div>Loading...</div>}

			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				{isEmpty && <p>You have no outgoing requests.</p>}
				{outgoing?.map((req) => (
					<Card key={req.id}>
						<CardHeader>
							<CardTitle>Swap Request</CardTitle>
						</CardHeader>
						<CardContent className="grid gap-2">
							<p>
								<strong>To:</strong> {req.responder_name}
							</p>
							<p>
								<strong>Your Event:</strong> {req.requester_event_title} (
								{new Date(req.requester_event_start_time).toLocaleString()})
							</p>
							<p>
								<strong>Their Event:</strong> {req.responder_event_title} (
								{new Date(req.responder_event_start_time).toLocaleString()})
							</p>
							<p className="text-sm text-muted-foreground mt-2">
								Status: PENDING
							</p>
							<Button
								size="sm"
								variant="destructive"
								onClick={() =>
									mutation.mutate({ id: req.id, status: "REJECTED" })
								}
							>
								Cancel
							</Button>
						</CardContent>
					</Card>
				))}
			</div>
		</div>
	);
}
