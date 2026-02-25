import type { Metadata } from "next";

export const metadata: Metadata = {
    title: "Dashboard | NexusAsset",
};

export default function DashboardPage() {
    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-bold text-white">Dashboard</h1>
                <p className="text-slate-400 mt-1">
                    Welcome to NexusAsset – your IT catalog overview.
                </p>
            </div>

            {/* Stat cards – populated in Phase 6 */}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                {[
                    { label: "Total Software", value: "–", color: "indigo" },
                    { label: "Servers", value: "–", color: "blue" },
                    { label: "Vendors", value: "–", color: "purple" },
                    { label: "Databases", value: "–", color: "cyan" },
                ].map((stat) => (
                    <div
                        key={stat.label}
                        className="rounded-xl border border-slate-800 bg-slate-900 p-5"
                    >
                        <p className="text-sm text-slate-400">{stat.label}</p>
                        <p className="mt-2 text-3xl font-bold text-white">{stat.value}</p>
                    </div>
                ))}
            </div>

            <div className="rounded-xl border border-slate-800 bg-slate-900 p-6 text-center text-slate-500 text-sm">
                Dashboard charts and analytics will be added in Phase 6.
            </div>
        </div>
    );
}
