import { format } from "date-fns";
// biome-ignore lint/nursery/noUnresolvedImports: false positive
import { Calendar as CalendarIcon } from "lucide-react";

import { cn } from "@/lib/utils.ts";
import { Calendar } from "@/components/ui/calendar.tsx";
import { Input } from "@/components/ui/input.tsx";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover.tsx";

export function DateTimePicker({
	date,
	setDate,
}: {
	date: Date | undefined;
	setDate: (date: Date | undefined) => void;
}) {
	const handleTimeChange = (e: React.ChangeEvent<HTMLInputElement>) => {
		const time = e.target.value;
		const [hours, minutes] = time.split(":").map(Number);
		const newDate = date ? new Date(date) : new Date();
		newDate.setHours(hours, minutes);
		setDate(newDate);
	};

	return (
		<div className="flex items-center gap-2">
			<Popover>
				<PopoverTrigger>
					<div
						className={cn(
							"w-[240px] justify-start text-left font-normal",
							!date && "text-muted-foreground",
							"inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2",
						)}
					>
						<CalendarIcon className="mr-2 h-4 w-4" />
						{date ? format(date, "PPP") : <span>Pick a date</span>}
					</div>
				</PopoverTrigger>
				<PopoverContent className="w-auto p-0">
					<Calendar
						mode="single"
						selected={date}
						onSelect={setDate}
						initialFocus
					/>
				</PopoverContent>
			</Popover>
			<Input
				type="time"
				step="60"
				value={date ? format(date, "HH:mm") : ""}
				onChange={handleTimeChange}
				className="w-[120px]"
			/>
		</div>
	);
}
