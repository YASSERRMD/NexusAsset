"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { serversApi } from "@/lib/catalog-api";
import { useAuth } from "@/hooks/useAuth";
import { CanAccess } from "@/components/layout/CanAccess";
import { toast } from "sonner";
import { Plus, Search, Pencil, Trash2, Server, Cpu, HardDrive } from "lucide-react";
import type { Server as ServerType } from "@/types";

const STATUS_COLORS: Record<string, string> = {
    active: "bg-emerald-900/40 text-emerald-400",
    decommissioned: "bg-red-900/40 text-red-400",
    maintenance: "bg-amber-900/40 text-amber-400",
    reserved: "bg-slate-700 text-slate-400",
};

const TYPE_LABELS: Record<string, string> = {
    bare_metal: "Bare Metal",
    vm: "Virtual Machine",
    container: "Container",
    cloud_instance: "Cloud Instance",
};

export default function ServersPage() {
    const [search, setSearch] = useState("");
    const [statusFilter, setStatusFilter] = useState("");
    const [typeFilter, setTypeFilter] = useState("");
    const [page, setPage] = useState(1);
    const qc = useQueryClient();
    const { isAdmin } = useAuth();

    const { data, isLoading } = useQuery({
        queryKey: ["servers", page, search, statusFilter, typeFilter],
        queryFn: () =>
            serversApi.list({
                page,
                limit: 20,
                search,
                status: statusFilter,
                server_type: typeFilter,
            }),
    });

    const deleteMut = useMutation({
        mutationFn: serversApi.delete,
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["servers"] });
            toast.success("Server deleted");
        },
        onError: () => toast.error("Failed to delete server"),
    });

    const servers: ServerType[] = data?.data ?? [];
    const total = data?.meta?.total ?? 0;

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white">Server Catalog</h1>
                    <p className="text-slate-400 mt-1">
                        {total} server{total !== 1 ? "s" : ""} registered
                    </p>
                </div>
                <CanAccess roles={["admin", "contributor"]}>
                    <button
                        className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500"
                        onClick={() => toast.info("Server creation form coming in Phase 3")}
                    >
                        <Plus className="h-4 w-4" />
                        Add Server
                    </button>
                </CanAccess>
            </div>

            {/* Filters */}
            <div className="flex flex-wrap gap-3">
                <div className="relative flex-1 min-w-48">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                    <input
                        value={search}
                        onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                        placeholder="Search name, hostname, IP…"
                        className="w-full rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500"
                    />
                </div>
                <select
                    value={typeFilter}
                    onChange={(e) => { setTypeFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All types</option>
                    {Object.entries(TYPE_LABELS).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                </select>
                <select
                    value={statusFilter}
                    onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All statuses</option>
                    {["active", "decommissioned", "maintenance", "reserved"].map(s => (
                        <option key={s} value={s}>{s.charAt(0).toUpperCase() + s.slice(1)}</option>
                    ))}
                </select>
            </div>

            {/* Table */}
            <div className="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden">
                {isLoading ? (
                    <div className="flex h-48 items-center justify-center text-slate-500 text-sm">Loading…</div>
                ) : (
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
                                <th className="px-4 py-3 font-medium">Server</th>
                                <th className="px-4 py-3 font-medium">Type</th>
                                <th className="px-4 py-3 font-medium">OS</th>
                                <th className="px-4 py-3 font-medium">Resources</th>
                                <th className="px-4 py-3 font-medium">Provider</th>
                                <th className="px-4 py-3 font-medium">Status</th>
                                <th className="px-4 py-3 font-medium" />
                            </tr>
                        </thead>
                        <tbody>
                            {servers.length === 0 ? (
                                <tr><td colSpan={7} className="px-4 py-10 text-center text-slate-500">No servers found</td></tr>
                            ) : servers.map((s) => (
                                <tr key={s.id} className="border-b border-slate-800/50 hover:bg-slate-800/30 transition-colors">
                                    <td className="px-4 py-3">
                                        <div className="flex items-center gap-2">
                                            <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-slate-800">
                                                <Server className="h-4 w-4 text-slate-400" />
                                            </div>
                                            <div>
                                                <p className="font-medium text-white">{s.name}</p>
                                                <p className="text-xs text-slate-500">{s.hostname ?? s.ip_address ?? "—"}</p>
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-4 py-3 text-slate-400">{TYPE_LABELS[s.server_type] ?? s.server_type}</td>
                                    <td className="px-4 py-3 text-slate-400">{s.os ? `${s.os} ${s.os_version ?? ""}`.trim() : "—"}</td>
                                    <td className="px-4 py-3">
                                        <div className="flex items-center gap-3 text-xs text-slate-500">
                                            {s.cpu_cores && <span className="flex items-center gap-1"><Cpu className="h-3 w-3" />{s.cpu_cores}c</span>}
                                            {s.ram_gb && <span>{s.ram_gb}GB</span>}
                                            {s.disk_gb && <span className="flex items-center gap-1"><HardDrive className="h-3 w-3" />{s.disk_gb}GB</span>}
                                            {!s.cpu_cores && !s.ram_gb && !s.disk_gb && "—"}
                                        </div>
                                    </td>
                                    <td className="px-4 py-3 text-slate-400">{s.cloud_provider ?? s.datacenter ?? "—"}</td>
                                    <td className="px-4 py-3">
                                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[s.status] ?? "bg-slate-700 text-slate-400"}`}>
                                            {s.status}
                                        </span>
                                    </td>
                                    <td className="px-4 py-3">
                                        <div className="flex items-center gap-2 justify-end">
                                            <CanAccess roles={["admin", "contributor"]}>
                                                <button className="rounded-lg p-1.5 text-slate-500 hover:text-slate-300 hover:bg-slate-700 transition" title="Edit">
                                                    <Pencil className="h-4 w-4" />
                                                </button>
                                            </CanAccess>
                                            {isAdmin && (
                                                <button
                                                    className="rounded-lg p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-900/20 transition"
                                                    title="Delete"
                                                    onClick={() => { if (confirm(`Delete server ${s.name}?`)) deleteMut.mutate(s.id); }}
                                                >
                                                    <Trash2 className="h-4 w-4" />
                                                </button>
                                            )}
                                        </div>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                )}
            </div>

            {/* Pagination */}
            {total > 20 && (
                <div className="flex items-center justify-between text-sm text-slate-500">
                    <span>{total} total</span>
                    <div className="flex gap-2">
                        <button disabled={page <= 1} onClick={() => setPage(p => p - 1)} className="px-3 py-1 rounded-lg border border-slate-700 disabled:opacity-40 hover:bg-slate-800 transition">Prev</button>
                        <span className="px-3 py-1">Page {page}</span>
                        <button disabled={page * 20 >= total} onClick={() => setPage(p => p + 1)} className="px-3 py-1 rounded-lg border border-slate-700 disabled:opacity-40 hover:bg-slate-800 transition">Next</button>
                    </div>
                </div>
            )}
        </div>
    );
}
