import { useState, useId } from "react";
import { z } from "zod";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
	DialogFooter,
	DialogClose,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import type { TreeifyError } from "@/lib/types.ts";
import type { Event } from "@/features/events/types";
import type { UseMutationResult } from "@tanstack/react-query";

const updateEventSchema = z
	.object({
		title: z.string().min(1, "Title is required"),
		start_time: z.string().datetime(),
		end_time: z.string().datetime(),
	})
	.refine((data) => new Date(data.end_time) > new Date(data.start_time), {
		message: "End time must be after start time",
		path: ["end_time"], // Assign the error to the endTime field
	});
type UpdateEventSchema = z.infer<typeof updateEventSchema>;

export function EditEventDialog({
	event,
	mutation,
}: {
	event: Event;
	mutation: UseMutationResult<
		Event,
		Error,
		{ id: number; event: UpdateEventSchema },
		unknown
	>;
}) {
	const titleId = useId();
	const [title, setTitle] = useState(event.title);
	const [startTime, setStartTime] = useState<Date | undefined>(
		new Date(event.start_time),
	);
	const [endTime, setEndTime] = useState<Date | undefined>(
		new Date(event.end_time),
	);
	const [formErrors, setFormErrors] =
		useState<TreeifyError<UpdateEventSchema> | null>(null);

	const handleSubmit = (e: React.FormEvent) => {
		e.preventDefault();
		if (!startTime || !endTime) {
			setFormErrors({
				errors: [],
				properties: {
					start_time: { errors: ["Start time is required"] },
					end_time: { errors: ["End time is required"] },
				},
			});
			return;
		}
		const updatedEvent = {
			title,
			start_time: startTime.toISOString(),
			end_time: endTime.toISOString(),
		};
		const result = updateEventSchema.safeParse(updatedEvent);
		if (result.success) {
			mutation.mutate({ id: event.id, event: result.data });
			setFormErrors(null);
		} else {
			setFormErrors(z.treeifyError(result.error));
		}
	};

	const titleError = formErrors?.properties?.title?.errors[0];
	const startTimeError = formErrors?.properties?.start_time?.errors[0];
	const endTimeError = formErrors?.properties?.end_time?.errors[0];

	return (
		<Dialog>
			<DialogTrigger>
				<Button variant="outline" size="sm">
					Edit
				</Button>
			</DialogTrigger>
			<DialogContent>
				<DialogHeader>
					<DialogTitle>Edit Event</DialogTitle>
				</DialogHeader>
				<form onSubmit={handleSubmit} className="grid gap-4 py-4">
					<div className="grid gap-2">
						<Label htmlFor={titleId}>Title</Label>
						<Input
							id={titleId}
							value={title}
							onChange={(e) => setTitle(e.target.value)}
							placeholder="Team Meeting"
						/>
						{titleError && (
							<p className="text-red-500 text-xs mt-1">{titleError}</p>
						)}
					</div>
					<div className="grid gap-2">
						<Label>Start Time</Label>
						<DateTimePicker date={startTime} setDate={setStartTime} />
						{startTimeError && (
							<p className="text-red-500 text-xs mt-1">{startTimeError}</p>
						)}
					</div>
					<div className="grid gap-2">
						<Label>End Time</Label>
						<DateTimePicker date={endTime} setDate={setEndTime} />
						{endTimeError && (
							<p className="text-red-500 text-xs mt-1">{endTimeError}</p>
						)}
					</div>
					<DialogFooter>
						<DialogClose>
							<Button variant="outline">Cancel</Button>
						</DialogClose>
						<Button type="submit" disabled={mutation.isPending}>
							{mutation.isPending ? "Updating..." : "Update Event"}
						</Button>
					</DialogFooter>
					{mutation.isError && (
						<p className="text-red-500 text-xs mt-2">
							{mutation.error.message}
						</p>
					)}
				</form>
			</DialogContent>
		</Dialog>
	);
}
