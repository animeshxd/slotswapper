import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useAuthStore } from "@/features/auth/auth.store.ts";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { useState, useId } from "react";
import { z } from "zod";
import { Button } from "@/components/ui/button.tsx";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card.tsx";
import { Input } from "@/components/ui/input.tsx";
import { Label } from "@/components/ui/label.tsx";
import { DateTimePicker } from "@/components/ui/datetime-picker.tsx";
import type { TreeifyError } from "@/lib/types.ts";

import type { UseMutationResult } from "@tanstack/react-query";

// Define the Event type based on the backend
interface Event {
  id: number;
  title: string;
  start_time: string;
  end_time: string;
  status: 'BUSY' | 'SWAPPABLE' | 'SWAP_PENDING';
}

// Define a generic type for the mutation props
type Mutation<T, U> = UseMutationResult<T, Error, U, unknown>;

// API functions (replace with your actual API calls)
async function fetchEvents(): Promise<Event[]> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/user`, {credentials: "include"});
  if (!res.ok) {
    throw new Error('Failed to fetch events');
  }
  return res.json();
}

const createEventSchema = z.object({
  title: z.string().min(1, 'Title is required'),
  start_time: z.string().datetime(),
  end_time: z.string().datetime(),
  status: z.enum(['BUSY', 'SWAPPABLE', 'SWAP_PENDING']),
}).refine((data) => new Date(data.end_time) > new Date(data.start_time), {
  message: "End time must be after start time",
  path: ["end_time"], // Assign the error to the endTime field
});
type CreateEventSchema = z.infer<typeof createEventSchema>;

async function createEvent(newEvent: CreateEventSchema): Promise<Event> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(newEvent),
    credentials: 'include',
  });
  if (!res.ok) {
    throw new Error('Failed to create event');
  }
  return res.json();
}

async function updateEventStatus(eventId: number, status: Event['status']): Promise<Event> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/${eventId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ status }),
    credentials: 'include',
  });
  if (!res.ok) {
    throw new Error('Failed to update event status');
  }
  return res.json();
}

async function deleteEvent(eventId: number): Promise<void> {
  const res = await fetch(`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/${eventId}`, {
    method: 'DELETE',
    credentials: 'include',
  });
  if (!res.ok) {
    throw new Error('Failed to delete event');
  }
}

export const Route = createFileRoute("/_protected/dashboard")({
	component: DashboardComponent,
});

function DashboardComponent() {
	const { user, setUser } = useAuthStore();
	const navigate = useNavigate();
  const queryClient = useQueryClient();

  const { data: events, isLoading, isError } = useQuery<Event[]>({ 
    queryKey: ['events'], 
    queryFn: fetchEvents 
  });

  const createMutation = useMutation({
    mutationFn: createEvent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ eventId, status }: { eventId: number, status: Event['status'] }) => updateEventStatus(eventId, status),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });

  const deleteMutation = useMutation({
    mutationFn: deleteEvent,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['events'] });
    },
  });

	const handleLogout = () => {
		setUser(null);
		navigate({ to: "/login" });
	};

	return (
		<div className="p-4 md:p-8">
			<div className="flex justify-between items-center mb-4">
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p>Welcome, {user?.name}!</p>
				<Button variant="outline" onClick={handleLogout}>Logout</Button>
			</div>

      <div className="grid gap-8 md:grid-cols-2">
        <CreateEventForm mutation={createMutation} />
        <EventsList 
          events={events} 
          isLoading={isLoading} 
          isError={isError} 
          updateStatusMutation={updateStatusMutation}
          deleteMutation={deleteMutation}
        />
      </div>
		</div>
	);
}

function CreateEventForm({ mutation }: { mutation: Mutation<Event, CreateEventSchema> }) {
  const titleId = useId();
  const [title, setTitle] = useState('');
  const [startTime, setStartTime] = useState<Date | undefined>();
  const [endTime, setEndTime] = useState<Date | undefined>();
  const [formErrors, setFormErrors] = useState<TreeifyError<CreateEventSchema> | null>(null);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!startTime || !endTime) {
      setFormErrors({ errors: [], properties: { start_time: { errors: ["Start time is required"] }, end_time: { errors: ["End time is required"] } } });
      return;
    }
    const event = { title, start_time: startTime.toISOString(), end_time: endTime.toISOString(), status: 'BUSY' as const };
    const result = createEventSchema.safeParse(event);
    if (result.success) {
      mutation.mutate(result.data);
      setTitle('');
      setStartTime(undefined);
      setEndTime(undefined);
      setFormErrors(null);
    } else {
      setFormErrors(z.treeifyError(result.error));
    }
  };

  const titleError = formErrors?.properties?.title?.errors[0];
  const startTimeError = formErrors?.properties?.start_time?.errors[0];
  const endTimeError = formErrors?.properties?.end_time?.errors[0];

  return (
    <Card>
      <CardHeader>
        <CardTitle>Create New Event</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit} className="grid gap-4">
          <div className="grid gap-2">
            <Label htmlFor={titleId}>Title</Label>
            <Input id={titleId} value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Team Meeting" />
            {titleError && <p className="text-red-500 text-xs mt-1">{titleError}</p>}
          </div>
          <div className="grid gap-2">
            <Label>Start Time</Label>
            <DateTimePicker date={startTime} setDate={setStartTime} />
            {startTimeError && <p className="text-red-500 text-xs mt-1">{startTimeError}</p>}
          </div>
          <div className="grid gap-2">
            <Label>End Time</Label>
            <DateTimePicker date={endTime} setDate={setEndTime} />
            {endTimeError && <p className="text-red-500 text-xs mt-1">{endTimeError}</p>}
          </div>
          <Button type="submit" disabled={mutation.isPending}>
            {mutation.isPending ? 'Creating...' : 'Create Event'}
          </Button>
          {mutation.isError && <p className="text-red-500 text-xs mt-2">{mutation.error.message}</p>}
        </form>
      </CardContent>
    </Card>
  );
}

function EventsList({ events, isLoading, isError, updateStatusMutation, deleteMutation }: { events: Event[] | undefined, isLoading: boolean, isError: boolean, updateStatusMutation: Mutation<Event, { eventId: number; status: Event["status"]; }>, deleteMutation: Mutation<void, number> }) {
  if (isLoading) return <div>Loading events...</div>;
  if (isError) return <div>Error fetching events.</div>;

  return (
    <Card>
      <CardHeader>
        <CardTitle>My Events</CardTitle>
      </CardHeader>
      <CardContent className="grid gap-4">
        {events && events.length > 0 ? (
          events.map(event => (
            <div key={event.id} className="p-2 border rounded-md flex justify-between items-center">
              <div>
                <p className="font-semibold">{event.title}</p>
                <p className="text-sm text-gray-500">{new Date(event.start_time).toLocaleString()} - {new Date(event.end_time).toLocaleString()}</p>
                <p className="text-sm font-medium">Status: {event.status}</p>
              </div>
              <div className="flex gap-2">
                {event.status === 'BUSY' && (
                  <Button variant="outline" size="sm" onClick={() => updateStatusMutation.mutate({ eventId: event.id, status: 'SWAPPABLE' })}>
                    Make Swappable
                  </Button>
                )}
                {event.status === 'SWAPPABLE' && (
                  <Button variant="outline" size="sm" onClick={() => updateStatusMutation.mutate({ eventId: event.id, status: 'BUSY' })}>
                    Make Busy
                  </Button>
                )}
                <Button variant="destructive" size="sm" onClick={() => deleteMutation.mutate(event.id)}>
                  Delete
                </Button>
              </div>
            </div>
          ))
        ) : (
          <p>You have no events.</p>
        )}
      </CardContent>
    </Card>
  );
}