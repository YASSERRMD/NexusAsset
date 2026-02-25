"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api";
import {
    Activity, Server, Package, Database, ShieldAlert,
    Clock, Users, Building, ShieldCheck, AlertOctagon
} from "lucide-react";
import Link from "next/link";
import { formatDate } from "@/lib/utils";

// Types matching the backend
interface ExpiryItem {
    id: string;
    name: string;
    days_left: number;
    expiry_on: string;
}

interface AuditSummary {
    action: string;
    entity_type: string;
    entity_id: string;
    user_id?: string;
    created_at: string;
}

interface Stats {
    total_software: number;
    inhouse_software: number;
    vendor_software: number;
    active_software: number;
    deprecated_software: number;
    critical_software: number;
    total_servers: number;
    active_servers: number;
    total_vendors: number;
    total_databases: number;
    total_integrations: number;
    total_persons: number;
    total_teams: number;
    expiring_licenses: ExpiryItem[];
    eol_software: ExpiryItem[];
    recent_audit_logs: AuditSummary[];
}

interface ChartData {
    software_by_criticality: { criticality: string; count: number }[];
    software_by_status: { status: string; count: number }[];
    servers_by_type: { type: string; count: number }[];
}

// Simple internal component for stat cards
function StatCard({ title, value, subtext, icon: Icon, colorClass }: any) {
    return (
        <div className="rounded-xl border border-slate-800 bg-slate-900 p-5 flex flex-col items-start gap-4 hover:border-slate-700 transition">
            <div className={`p-3 rounded-lg ${colorClass}`}>
                <Icon className="h-5 w-5" />
            </div>
            <div>
                <h3 className="text-slate-400 text-sm font-medium">{title}</h3>
                <div className="text-2xl font-bold text-white mt-1">{value}</div>
                {subtext && <p className="text-xs text-slate-500 mt-1">{subtext}</p>}
            </div>
        </div>
    );
}

export default function DashboardPage() {
    const { data, isLoading } = useQuery({
        queryKey: ["dashboard", "stats"],
        queryFn: () => apiClient.get<{ data: Stats }>("/dashboard").then(r => r.data.data)
    });

    const { data: charts, isLoading: chartsLoading } = useQuery({
        queryKey: ["dashboard", "charts"],
        queryFn: () => apiClient.get<{ data: ChartData }>("/dashboard/charts").then(r => r.data.data)
    });

    if (isLoading || chartsLoading) {
        return <div className="flex h-64 items-center justify-center text-slate-500">Loading dashboard…</div>;
    }

    if (!data || !charts) {
        return <div className="flex h-64 items-center justify-center text-red-500">Failed to load dashboard</div>;
    }

    return (
        <div className="space-y-8">
            {/* Header */}
            <div>
                <h1 className="text-2xl font-bold text-white tracking-tight">Executive Dashboard</h1>
                <p className="text-slate-400 mt-1">Overview of IT assets, software portfolio, and compliance status.</p>
            </div>

            {/* Top Level KPIs */}
            <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
                <StatCard
                    title="Total Software"
                    value={data.total_software}
                    subtext={`${data.active_software} active, ${data.critical_software} critical`}
                    icon={Package}
                    colorClass="bg-indigo-500/20 text-indigo-400"
                />
                <StatCard
                    title="Servers Infrastructure"
                    value={data.total_servers}
                    subtext={`${data.active_servers} currently active`}
                    icon={Server}
                    colorClass="bg-blue-500/20 text-blue-400"
                />
                <StatCard
                    title="Third-Party Vendors"
                    value={data.total_vendors}
                    subtext={`${data.vendor_software} deployed vendor apps`}
                    icon={Building}
                    colorClass="bg-emerald-500/20 text-emerald-400"
                />
                <StatCard
                    title="Databases & Data"
                    value={data.total_databases}
                    subtext={`Connected via ${data.total_integrations} integrations`}
                    icon={Database}
                    colorClass="bg-amber-500/20 text-amber-400"
                />
            </div>

            <div className="grid grid-cols-1 xl:grid-cols-3 gap-6">

                {/* Left col: Charts / Breakdowns */}
                <div className="xl:col-span-2 space-y-6">
                    {/* Criticality Breakdown */}
                    <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
                        <h3 className="font-semibold text-white mb-4 flex items-center gap-2">
                            <ShieldCheck className="h-5 w-5 text-indigo-400" />
                            Software Criticality
                        </h3>
                        <div className="flex flex-col gap-3">
                            {charts.software_by_criticality.map((c, i) => (
                                <div key={i} className="flex items-center justify-between">
                                    <span className="text-sm font-medium capitalize text-slate-300">{c.criticality}</span>
                                    <div className="flex items-center gap-3">
                                        <div className="w-48 bg-slate-800 h-2 rounded-full overflow-hidden">
                                            <div
                                                className={`h-full ${c.criticality === 'critical' ? 'bg-red-500' :
                                                        c.criticality === 'high' ? 'bg-orange-500' :
                                                            c.criticality === 'medium' ? 'bg-amber-500' : 'bg-emerald-500'
                                                    }`}
                                                style={{ width: `${(c.count / data.total_software) * 100}%` }}
                                            />
                                        </div>
                                        <span className="text-sm text-slate-400 w-8 text-right">{c.count}</span>
                                    </div>
                                </div>
                            ))}
                        </div>
                    </div>

                    {/* Team & Activity summary */}
                    <div className="grid grid-cols-2 gap-6">
                        <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
                            <h3 className="font-semibold text-white mb-4 flex items-center gap-2">
                                <Users className="h-5 w-5 text-indigo-400" />
                                Organization
                            </h3>
                            <div className="space-y-4">
                                <div className="flex justify-between items-center pb-2 border-b border-slate-800">
                                    <span className="text-slate-400 text-sm">Active Persons</span>
                                    <span className="text-white font-medium">{data.total_persons}</span>
                                </div>
                                <div className="flex justify-between items-center pb-2 border-b border-slate-800">
                                    <span className="text-slate-400 text-sm">Teams & Departments</span>
                                    <span className="text-white font-medium">{data.total_teams}</span>
                                </div>
                            </div>
                            <Link href="/people" className="block text-center text-sm text-indigo-400 mt-4 hover:underline">
                                Manage Organization
                            </Link>
                        </div>

                        <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
                            <h3 className="font-semibold text-white mb-4 flex items-center gap-2">
                                <Activity className="h-5 w-5 text-indigo-400" />
                                Recent Activity
                            </h3>
                            <div className="space-y-3">
                                {data.recent_audit_logs.slice(0, 4).map((audit, idx) => (
                                    <div key={idx} className="flex flex-col gap-1 text-sm border-b border-slate-800 pb-2 last:border-0 last:pb-0">
                                        <div className="flex items-center justify-between">
                                            <span className="text-white font-medium uppercase font-mono text-[10px] tracking-wider px-1.5 py-0.5 rounded bg-slate-800">{audit.action}</span>
                                            <span className="text-slate-500 text-xs">{formatDate(audit.created_at)}</span>
                                        </div>
                                        <div className="text-slate-400">
                                            Modified {audit.entity_type} <span className="text-slate-300 font-mono text-xs">{audit.entity_id.split('-')[0]}</span>
                                        </div>
                                    </div>
                                ))}
                            </div>
                            <Link href="/audit" className="block text-center text-sm text-indigo-400 mt-3 hover:underline">
                                View Full Audit Log
                            </Link>
                        </div>
                    </div>
                </div>

                {/* Right col: Alerts */}
                <div className="space-y-6">
                    {/* License Expiry Alerts */}
                    <div className="rounded-xl border border-orange-900/30 bg-orange-900/10 p-6 relative overflow-hidden">
                        <div className="absolute top-0 right-0 p-4 opacity-10">
                            <Clock className="w-24 h-24" />
                        </div>
                        <h3 className="font-semibold text-orange-400 mb-4 flex items-center gap-2">
                            <Clock className="h-5 w-5" />
                            Expiring Licenses (90 Days)
                        </h3>
                        {data.expiring_licenses.length === 0 ? (
                            <p className="text-sm text-slate-400">No licenses expiring within 90 days.</p>
                        ) : (
                            <div className="space-y-3 relative z-10">
                                {data.expiring_licenses.map(exp => (
                                    <div key={exp.id} className="bg-slate-900/80 rounded-lg p-3 border border-orange-900/40">
                                        <div className="flex justify-between items-start mb-1">
                                            <Link href={`/software/${exp.id}`} className="text-sm font-medium text-white hover:text-orange-400 transition">{exp.name}</Link>
                                            <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${exp.days_left <= 30 ? 'bg-red-900/50 text-red-400' : 'bg-orange-900/50 text-orange-400'}`}>
                                                {exp.days_left}d left
                                            </span>
                                        </div>
                                        <div className="text-xs text-slate-500">Exp: {formatDate(exp.expiry_on)}</div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>

                    {/* EOL Software Alerts */}
                    <div className="rounded-xl border border-red-900/30 bg-red-900/10 p-6 relative overflow-hidden">
                        <div className="absolute top-0 right-0 p-4 opacity-10">
                            <AlertOctagon className="w-24 h-24" />
                        </div>
                        <h3 className="font-semibold text-red-500 mb-4 flex items-center gap-2">
                            <ShieldAlert className="h-5 w-5" />
                            EOL Software (180 Days)
                        </h3>
                        {data.eol_software.length === 0 ? (
                            <p className="text-sm text-slate-400">No software reaching EOL within 180 days.</p>
                        ) : (
                            <div className="space-y-3 relative z-10">
                                {data.eol_software.map(exp => (
                                    <div key={exp.id} className="bg-slate-900/80 rounded-lg p-3 border border-red-900/40">
                                        <div className="flex justify-between items-start mb-1">
                                            <Link href={`/software/${exp.id}`} className="text-sm font-medium text-white hover:text-red-400 transition">{exp.name}</Link>
                                            <span className={`text-xs font-bold px-1.5 py-0.5 rounded ${exp.days_left <= 90 ? 'bg-red-900/50 text-red-400' : 'bg-orange-900/50 text-orange-400'}`}>
                                                {exp.days_left}d left
                                            </span>
                                        </div>
                                        <div className="text-xs text-slate-500">EOL: {formatDate(exp.expiry_on)}</div>
                                    </div>
                                ))}
                            </div>
                        )}
                    </div>
                </div>

            </div>
        </div>
    );
}
