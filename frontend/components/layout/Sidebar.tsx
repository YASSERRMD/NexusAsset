"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
    LayoutDashboard,
    Monitor,
    Package,
    Server,
    Store,
    Database,
    Network,
    Users,
    Shield,
    Settings,
    ClipboardList,
    LogOut,
    ChevronRight,
} from "lucide-react";
import { cn } from "@/lib/utils";
import { useAuth } from "@/hooks/useAuth";

interface NavItem {
    label: string;
    href: string;
    icon: React.ElementType;
    adminOnly?: boolean;
}

const NAV_ITEMS: NavItem[] = [
    { label: "Dashboard", href: "/", icon: LayoutDashboard },
    { label: "Software", href: "/software", icon: Package },
    { label: "Servers", href: "/servers", icon: Server },
    { label: "Vendors", href: "/vendors", icon: Store },
    { label: "Databases", href: "/databases", icon: Database },
    { label: "Integrations", href: "/integrations", icon: Network },
    { label: "People", href: "/people", icon: Users },
    { label: "Compliance", href: "/compliance", icon: Shield, adminOnly: true },
    { label: "Audit Log", href: "/audit", icon: ClipboardList, adminOnly: true },
    { label: "Settings", href: "/settings", icon: Settings, adminOnly: true },
];

export function Sidebar() {
    const pathname = usePathname();
    const { isAdmin, logout, user, role } = useAuth();

    const visibleItems = NAV_ITEMS.filter((item) => !item.adminOnly || isAdmin);

    return (
        <aside className="flex h-screen w-64 flex-col border-r border-slate-800 bg-slate-950">
            {/* Logo */}
            <div className="flex h-16 items-center gap-3 border-b border-slate-800 px-6">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-indigo-600">
                    <Monitor className="h-4 w-4 text-white" />
                </div>
                <div>
                    <p className="text-sm font-bold text-white">NexusAsset</p>
                    <p className="text-xs text-slate-400">IT Catalog</p>
                </div>
            </div>

            {/* Navigation */}
            <nav className="flex-1 overflow-y-auto py-4 px-3">
                <ul className="space-y-1">
                    {visibleItems.map((item) => {
                        const isActive =
                            item.href === "/"
                                ? pathname === "/"
                                : pathname.startsWith(item.href);
                        const Icon = item.icon;

                        return (
                            <li key={item.href}>
                                <Link
                                    href={item.href}
                                    className={cn(
                                        "group flex items-center gap-3 rounded-lg px-3 py-2 text-sm font-medium transition-all",
                                        isActive
                                            ? "bg-indigo-600 text-white"
                                            : "text-slate-400 hover:bg-slate-800 hover:text-white"
                                    )}
                                >
                                    <Icon className="h-4 w-4 shrink-0" />
                                    <span className="flex-1">{item.label}</span>
                                    {isActive && <ChevronRight className="h-3 w-3" />}
                                </Link>
                            </li>
                        );
                    })}
                </ul>
            </nav>

            {/* User footer */}
            <div className="border-t border-slate-800 p-4">
                <div className="flex items-center gap-3 rounded-lg px-2 py-2">
                    <div className="flex h-8 w-8 items-center justify-center rounded-full bg-indigo-600/20 text-indigo-400 text-xs font-bold uppercase">
                        {user?.username?.[0] ?? "?"}
                    </div>
                    <div className="flex-1 min-w-0">
                        <p className="truncate text-xs font-medium text-white">
                            {user?.username}
                        </p>
                        <span
                            className={cn(
                                "inline-flex items-center rounded-full px-1.5 py-0.5 text-xs font-medium",
                                role === "admin"
                                    ? "bg-indigo-600/20 text-indigo-400"
                                    : role === "contributor"
                                        ? "bg-emerald-600/20 text-emerald-400"
                                        : "bg-slate-600/20 text-slate-400"
                            )}
                        >
                            {role}
                        </span>
                    </div>
                    <button
                        onClick={logout}
                        title="Sign out"
                        className="rounded-lg p-1.5 text-slate-500 hover:bg-slate-800 hover:text-white transition-colors"
                    >
                        <LogOut className="h-4 w-4" />
                    </button>
                </div>
            </div>
        </aside>
    );
}
