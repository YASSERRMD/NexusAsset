"use client";

import { useEffect } from "react";
import { useRouter } from "next/navigation";
import { Sidebar } from "@/components/layout/Sidebar";
import { useAuth } from "@/hooks/useAuth";

export default function DashboardLayout({
    children,
}: {
    children: React.ReactNode;
}) {
    const { isAuthenticated, isHydrated } = useAuth();
    const router = useRouter();

    useEffect(() => {
        if (isHydrated && !isAuthenticated) {
            router.replace("/login");
        }
    }, [isHydrated, isAuthenticated, router]);

    // Show nothing while hydrating to avoid flash
    if (!isHydrated) {
        return (
            <div className="flex h-screen items-center justify-center bg-slate-950">
                <div className="h-8 w-8 animate-spin rounded-full border-2 border-indigo-600 border-t-transparent" />
            </div>
        );
    }

    if (!isAuthenticated) return null;

    return (
        <div className="flex h-screen overflow-hidden bg-slate-950">
            <Sidebar />
            <main className="flex-1 overflow-y-auto">
                {/* Topbar */}
                <header className="sticky top-0 z-10 flex h-16 items-center justify-between border-b border-slate-800 bg-slate-950/80 px-6 backdrop-blur-sm">
                    <div />
                    <div className="flex items-center gap-3 text-sm text-slate-400">
                        <span>NexusAsset IT Catalog</span>
                    </div>
                </header>

                {/* Page content */}
                <div className="p-6">{children}</div>
            </main>
        </div>
    );
}
