import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card.tsx";
import { Button } from "@/components/ui/button.tsx";

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
async function fetchIncomingRequests(): Promise<IncomingSwapRequest[]> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swap-requests/incoming`, { credentials: "include" });
  if (!res.ok) {
    throw new Error('Failed to fetch incoming requests');
  }
  return res.json();
}

async function fetchOutgoingRequests(): Promise<OutgoingSwapRequest[]> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swap-requests/outgoing`, { credentials: "include" });
  if (!res.ok) {
    throw new Error('Failed to fetch outgoing requests');
  }
  return res.json();
}

async function respondToSwapRequest(id: number, status: 'ACCEPTED' | 'REJECTED'): Promise<void> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swap-response/${id}`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
    credentials: 'include',
  });
  if (!res.ok) {
    throw new Error('Failed to respond to swap request');
  }
}

export const Route = createFileRoute("/_protected/requests")({
  component: RequestsComponent,
});

function RequestsComponent() {
  const queryClient = useQueryClient();

  const { data: incoming, isLoading: isLoadingIncoming } = useQuery<IncomingSwapRequest[]>({ 
    queryKey: ['incoming-requests'], 
    queryFn: fetchIncomingRequests 
  });

  const { data: outgoing, isLoading: isLoadingOutgoing } = useQuery<OutgoingSwapRequest[]>({ 
    queryKey: ['outgoing-requests'], 
    queryFn: fetchOutgoingRequests 
  });

  const mutation = useMutation({
    mutationFn: ({ id, status }: { id: number, status: 'ACCEPTED' | 'REJECTED' }) => respondToSwapRequest(id, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['incoming-requests'] });
      queryClient.invalidateQueries({ queryKey: ['outgoing-requests'] });
      queryClient.invalidateQueries({ queryKey: ['swappable-events'] });
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });

  return (
    <div className="p-4 md:p-8 grid gap-8 md:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>Incoming Requests</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4">
          {isLoadingIncoming && <div>Loading...</div>}
          {incoming?.map(req => (
            <div key={req.id} className="p-4 border rounded-md">
              <p><strong>From:</strong> {req.requester_name}</p>
              <p><strong>Their Event:</strong> {req.requester_event_title} ({new Date(req.requester_event_start_time).toLocaleString()})</p>
              <p><strong>Your Event:</strong> {req.responder_event_title} ({new Date(req.responder_event_start_time).toLocaleString()})</p>
              <div className="flex gap-2 mt-4">
                <Button size="sm" onClick={() => mutation.mutate({ id: req.id, status: 'ACCEPTED' })}>Accept</Button>
                <Button size="sm" variant="outline" onClick={() => mutation.mutate({ id: req.id, status: 'REJECTED' })}>Reject</Button>
              </div>
            </div>
          ))}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Outgoing Requests</CardTitle>
        </CardHeader>
        <CardContent className="grid gap-4">
          {isLoadingOutgoing && <div>Loading...</div>}
          {outgoing?.map(req => (
            <div key={req.id} className="p-4 border rounded-md">
              <p><strong>To:</strong> {req.responder_name}</p>
              <p><strong>Your Event:</strong> {req.requester_event_title} ({new Date(req.requester_event_start_time).toLocaleString()})</p>
              <p><strong>Their Event:</strong> {req.responder_event_title} ({new Date(req.responder_event_start_time).toLocaleString()})</p>
              <p className="text-sm text-muted-foreground mt-2">Status: PENDING</p>
              <Button size="sm" variant="destructive" onClick={() => mutation.mutate({ id: req.id, status: 'REJECTED' })}>Cancel</Button>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
