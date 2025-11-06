// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-nocheck
// biome-ignore-all lint: not a owner

import * as React from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import { LogOut, LayoutDashboard, ShoppingCart, Bell } from "lucide-react";

import { serverLogout, useAuthStore } from "@/features/auth/auth.store";
import { Button } from "@/components/ui/button";
import {
	Sidebar,
	SidebarContent,
	SidebarHeader,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	SidebarFooter,
	SidebarMenuSub,
	SidebarMenuSubItem,
	SidebarMenuSubButton,
} from "@/components/ui/sidebar";

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
	const { user, logout } = useAuthStore();
	const navigate = useNavigate();

	const handleLogout = () => {
		logout();
        serverLogout().finally(() => navigate({ to: "/login" }))
	};

	return (
		<Sidebar {...props}>
			<SidebarHeader>
				<h2 className="text-lg font-semibold">SlotSwapper</h2>
			</SidebarHeader>
			<SidebarContent>
				<SidebarMenu>
					<SidebarMenuItem>
						<SidebarMenuButton asChild>
							<Link to="/dashboard" className="[&.active]:font-bold">
								<LayoutDashboard className="mr-2 h-4 w-4" />
								Dashboard
							</Link>
						</SidebarMenuButton>
					</SidebarMenuItem>
					<SidebarMenuItem>
						<SidebarMenuButton asChild>
							<Link to="/marketplace" className="[&.active]:font-bold">
								<ShoppingCart className="mr-2 h-4 w-4" />
								Marketplace
							</Link>
						</SidebarMenuButton>
					</SidebarMenuItem>
					<SidebarMenuItem>
						<SidebarMenuButton>
							<Bell className="mr-2 h-4 w-4" />
							Requests
						</SidebarMenuButton>
						<SidebarMenuSub>
							<SidebarMenuSubItem>
								<SidebarMenuSubButton asChild>
									<Link
										to="/requests/incoming"
										className="[&.active]:font-bold"
									>
										Incoming
									</Link>
								</SidebarMenuSubButton>
							</SidebarMenuSubItem>
							<SidebarMenuSubItem>
								<SidebarMenuSubButton asChild>
									<Link
										to="/requests/outgoing"
										className="[&.active]:font-bold"
									>
										Outgoing
									</Link>
								</SidebarMenuSubButton>
							</SidebarMenuSubItem>
						</SidebarMenuSub>
					</SidebarMenuItem>
				</SidebarMenu>
			</SidebarContent>
			<SidebarFooter>
				<div className="flex items-center gap-2">
					<div className="font-semibold">{user?.name}</div>
				</div>
				<Button variant="outline" size="sm" onClick={handleLogout}>
					<LogOut className="mr-2 h-4 w-4" />
					Logout
				</Button>
			</SidebarFooter>
		</Sidebar>
	);
}
