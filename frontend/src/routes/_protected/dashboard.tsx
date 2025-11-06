import { createFileRoute } from "@tanstack/react-router";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { CreateEventDialog } from "@/features/events/create-event-dialog";
import { EditEventDialog } from "@/features/events/edit-event-dialog";
import { z } from "zod";

import type { Event } from "@/features/events/types";

const createEventSchema = z
	.object({
		title: z.string().min(1, "Title is required"),
		start_time: z.iso.datetime(),
		end_time: z.iso.datetime(),
		status: z.enum(["BUSY", "SWAPPABLE", "SWAP_PENDING"]),
	})
	.refine((data) => new Date(data.end_time) > new Date(data.start_time), {
		message: "End time must be after start time",
		path: ["end_time"], // Assign the error to the endTime field
	});
type CreateEventSchema = z.infer<typeof createEventSchema>;

const updateEventSchema = z
	.object({
		title: z.string().min(1, "Title is required"),
		start_time: z.iso.datetime(),
		end_time: z.iso.datetime(),
	})
	.refine((data) => new Date(data.end_time) > new Date(data.start_time), {
		message: "End time must be after start time",
		path: ["end_time"], // Assign the error to the endTime field
	});
type UpdateEventSchema = z.infer<typeof updateEventSchema>;

// API function to fetch events
async function fetchEvents(): Promise<Event[]> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/user`,
		{ credentials: "include" },
	);
	if (!res.ok) {
		throw new Error("Failed to fetch events");
	}
	return res.json();
}

// API function to create an event
async function createEvent(newEvent: CreateEventSchema): Promise<Event> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events`,
		{
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify(newEvent),
			credentials: "include",
		},
	);
	if (!res.ok) {
		throw new Error("Failed to create event");
	}
	return res.json();
}

// API function to update an event
async function updateEvent({
	id,
	event,
}: {
	id: number;
	event: UpdateEventSchema;
}): Promise<Event> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/${id}`,
		{
			method: "PUT",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ ...event, status: "BUSY" }),
			credentials: "include",
		},
	);
	if (!res.ok) {
		throw new Error("Failed to update event");
	}
	return res.json();
}

// API function to update an event's status
async function updateEventStatus({
	id,
	status,
}: {
	id: number;
	status: string;
}): Promise<Event> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/${id}/status`,
		{
			method: "POST",
			headers: { "Content-Type": "application/json" },
			body: JSON.stringify({ status }),
			credentials: "include",
		},
	);
	if (!res.ok) {
		throw new Error("Failed to update event status");
	}
	return res.json();
}

// API function to delete an event
async function deleteEvent(id: number): Promise<void> {
	const res = await fetch(
		`${import.meta.env.VITE_HTTP_SERVER_URL}/api/events/${id}`,
		{
			method: "DELETE",
			credentials: "include",
		},
	);
	if (!res.ok) {
		throw new Error("Failed to delete event");
	}
}

export const Route = createFileRoute("/_protected/dashboard")({
	component: DashboardComponent,
});

function DashboardComponent() {
	const queryClient = useQueryClient();

	const {
		data: events,
		isLoading,
		isError,
	} = useQuery<Event[]>({
		queryKey: ["events"],
		queryFn: fetchEvents,
	});

	const createEventMutation = useMutation({
		mutationFn: createEvent,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["events"] });
		},
	});

	const updateEventMutation = useMutation({
		mutationFn: updateEvent,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["events"] });
		},
	});

	const updateEventStatusMutation = useMutation({
		mutationFn: updateEventStatus,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["events"] });
		},
	});

	const deleteEventMutation = useMutation({
		mutationFn: deleteEvent,
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ["events"] });
		},
	});

	return (
		<div className="p-4 md:p-8">
			<div className="flex justify-between items-center mb-4">
				<h1 className="text-2xl font-bold">Your Events</h1>
				<CreateEventDialog mutation={createEventMutation} />
			</div>

			{isLoading && <div>Loading events...</div>}
			{isError && <div>Error fetching events.</div>}

			<div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
				{events?.map((event) => (
					<Card key={event.id}>
						<CardHeader>
							<CardTitle>{event.title}</CardTitle>
						</CardHeader>
						<CardContent className="grid gap-2">
							<p className="text-sm">
								<strong>From:</strong>{" "}
								{new Date(event.start_time).toLocaleString()}
							</p>
							<p className="text-sm">
								<strong>To:</strong> {new Date(event.end_time).toLocaleString()}
							</p>
							<p className="text-sm">
								<strong>Status:</strong> {event.status}
							</p>
							<div className="flex gap-2 mt-4">
								<EditEventDialog event={event} mutation={updateEventMutation} />
								<Button
									size="sm"
									onClick={() =>
										updateEventStatusMutation.mutate({
											id: event.id,
											status: event.status === "BUSY" ? "SWAPPABLE" : "BUSY",
										})
									}
									disabled={
										updateEventStatusMutation.isPending ||
										event.status === "SWAP_PENDING"
									}
								>
									{event.status === "BUSY" ? "Make Swappable" : "Make Busy"}
								</Button>
								<Button
									size="sm"
									variant="destructive"
									onClick={() => deleteEventMutation.mutate(event.id)}
									disabled={
										deleteEventMutation.isPending ||
										event.status === "SWAP_PENDING"
									}
								>
									Delete
								</Button>
							</div>
						</CardContent>
					</Card>
				))}
			</div>
		</div>
	);
}
