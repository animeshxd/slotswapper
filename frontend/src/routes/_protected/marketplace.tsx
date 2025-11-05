import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card.tsx";
import { Button } from "@/components/ui/button.tsx";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger, DialogFooter, DialogClose } from "@/components/ui/dialog.tsx";
import { useState } from "react";

// Define the SwappableEvent type based on the backend response
interface SwappableEvent {
  id: number;
  title: string;
  start_time: string;
  end_time: string;
  owner_name: string;
}

// Define the type for the user's own swappable events
interface MySwappableEvent {
  id: number;
  title: string;
}

// API function to fetch swappable events
async function fetchSwappableEvents(): Promise<SwappableEvent[]> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swappable-slots`, { credentials: "include" });
  if (!res.ok) {
    throw new Error('Failed to fetch swappable events');
  }
  return res.json();
}

// API function to fetch the user's own swappable events
async function fetchMySwappableEvents(): Promise<MySwappableEvent[]> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/user?status=SWAPPABLE`, { credentials: "include" });
  if (!res.ok) {
    throw new Error('Failed to fetch your swappable events');
  }
  return res.json();
}

// API function to create a swap request
async function createSwapRequest(requesterSlotId: number, responderSlotId: number): Promise<void> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/swap-request`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ requester_slot_id: requesterSlotId, responder_slot_id: responderSlotId }),
    credentials: 'include',
  });
  if (!res.ok) {
    throw new Error('Failed to create swap request');
  }
}

export const Route = createFileRoute("/_protected/marketplace")({
  component: MarketplaceComponent,
});

function MarketplaceComponent() {
  const { data: events, isLoading, isError } = useQuery<SwappableEvent[]>({ 
    queryKey: ['swappable-events'], 
    queryFn: fetchSwappableEvents 
  });

  const { data: myEvents } = useQuery<MySwappableEvent[]>({ 
    queryKey: ['my-swappable-events'], 
    queryFn: fetchMySwappableEvents 
  });

  return (
    <div className="p-4 md:p-8">
      <h1 className="text-2xl font-bold mb-4">Marketplace</h1>
      <p className="text-muted-foreground mb-4">Browse events that other users have made available for swapping.</p>
      
      {isLoading && <div>Loading available slots...</div>}
      {isError && <div>Error fetching available slots.</div>}

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
        {events?.map(event => (
          <SwapRequestDialog key={event.id} event={event} myEvents={myEvents} />
        ))}
      </div>
    </div>
  );
}

function SwapRequestDialog({ event, myEvents }: { event: SwappableEvent, myEvents: MySwappableEvent[] | undefined }) {
  const queryClient = useQueryClient();
  const [selectedSlot, setSelectedSlot] = useState<number | null>(null);

  const mutation = useMutation({
    mutationFn: (requesterSlotId: number) => createSwapRequest(requesterSlotId, event.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['swappable-events'] });
      // Optionally, close the dialog and show a success message
    },
  });

  const hasSwappableEvents = myEvents && myEvents.length > 0;

  return (
    <Dialog>
      <DialogTrigger disabled={!hasSwappableEvents}>
        <div className="cursor-pointer">
          <Card className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle>{event.title}</CardTitle>
            </CardHeader>
            <CardContent className="grid gap-2">
              <p className="text-sm"><strong>Owner:</strong> {event.owner_name}</p>
              <p className="text-sm"><strong>From:</strong> {new Date(event.start_time).toLocaleString()}</p>
              <p className="text-sm"><strong>To:</strong> {new Date(event.end_time).toLocaleString()}</p>
              <Button className="mt-4 w-full" disabled={!hasSwappableEvents}>
                {hasSwappableEvents ? "Request Swap" : "No swappable events"}
              </Button>
            </CardContent>
          </Card>
        </div>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Request a Swap</DialogTitle>
        </DialogHeader>
        <div className="grid gap-4 py-4">
          <p>You are requesting to swap for <strong>{event.title}</strong>.</p>
          <p>Choose one of your available slots to offer:</p>
          <div className="grid gap-2">
            {myEvents?.map(myEvent => (
              <button 
                type="button"
                key={myEvent.id} 
                className={`p-2 border rounded-md cursor-pointer text-left ${selectedSlot === myEvent.id ? 'border-primary' : ''}`}
                onClick={() => setSelectedSlot(myEvent.id)}
              >
                {myEvent.title}
              </button>
            ))}
          </div>
        </div>
        <DialogFooter>
          <DialogClose>
            <Button variant="outline">Cancel</Button>
          </DialogClose>
          <Button onClick={() => selectedSlot && mutation.mutate(selectedSlot)} disabled={!selectedSlot || mutation.isPending}>
            {mutation.isPending ? 'Sending Request...' : 'Confirm Swap'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
