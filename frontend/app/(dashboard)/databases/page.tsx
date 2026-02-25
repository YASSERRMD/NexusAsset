"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { CanAccess } from "@/components/layout/CanAccess";
import { toast } from "sonner";
import { Plus, Search, Pencil, Trash2, Database } from "lucide-react";
import type { ApiResponse, DatabaseInstance } from "@/types";

const STATUS_COLORS: Record<string, string> = {
    active: "bg-emerald-900/40 text-emerald-400",
    decommissioned: "bg-red-900/40 text-red-400",
    maintenance: "bg-amber-900/40 text-amber-400",
};

export default function DatabasesPage() {
    const [search, setSearch] = useState("");
    const [engineFilter, setEngineFilter] = useState("");
    const [page, setPage] = useState(1);
    const qc = useQueryClient();
    const { isAdmin } = useAuth();

    const { data, isLoading } = useQuery({
        queryKey: ["databases", page, search, engineFilter],
        queryFn: () =>
            apiClient.get<ApiResponse<DatabaseInstance[]>>("/databases", {
                params: { page, limit: 20, search, engine: engineFilter },
            }).then(r => r.data),
    });

    const deleteMut = useMutation({
        mutationFn: (id: string) => apiClient.delete(`/databases/${id}`),
        onSuccess: () => { qc.invalidateQueries({ queryKey: ["databases"] }); toast.success("Database deleted"); },
        onError: () => toast.error("Failed to delete"),
    });

    const items: DatabaseInstance[] = data?.data ?? [];
    const total = data?.meta?.total ?? 0;

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white">Database Instances</h1>
                    <p className="text-slate-400 mt-1">{total} database{total !== 1 ? "s" : ""} registered</p>
                </div>
                <CanAccess roles={["admin", "contributor"]}>
                    <button
                        className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500"
                        onClick={() => toast.info("Database creation form coming soon")}
                    >
                        <Plus className="h-4 w-4" />
                        Add Database
                    </button>
                </CanAccess>
            </div>

            <div className="flex gap-3">
                <div className="relative flex-1 max-w-sm">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                    <input
                        value={search}
                        onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                        placeholder="Search databases…"
                        className="w-full rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500"
                    />
                </div>
                <select
                    value={engineFilter}
                    onChange={(e) => { setEngineFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All engines</option>
                    {["PostgreSQL", "MySQL", "MSSQL", "Oracle", "MongoDB", "Redis", "Elasticsearch"].map((e) => (
                        <option key={e} value={e}>{e}</option>
                    ))}
                </select>
            </div>

            <div className="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden">
                {isLoading ? (
                    <div className="flex h-48 items-center justify-center text-slate-500 text-sm">Loading…</div>
                ) : (
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
                                <th className="px-4 py-3 font-medium">Name</th>
                                <th className="px-4 py-3 font-medium">Engine</th>
                                <th className="px-4 py-3 font-medium">Host</th>
                                <th className="px-4 py-3 font-medium">Port</th>
                                <th className="px-4 py-3 font-medium">Cloud</th>
                                <th className="px-4 py-3 font-medium">Status</th>
                                <th className="px-4 py-3 font-medium" />
                            </tr>
                        </thead>
                        <tbody>
                            {items.length === 0 ? (
                                <tr><td colSpan={7} className="px-4 py-10 text-center text-slate-500">No databases found</td></tr>
                            ) : items.map((db) => (
                                <tr key={db.id} className="border-b border-slate-800/50 hover:bg-slate-800/30 transition-colors">
                                    <td className="px-4 py-3">
                                        <div className="flex items-center gap-2">
                                            <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-slate-800">
                                                <Database className="h-3.5 w-3.5 text-cyan-400" />
                                            </div>
                                            <div>
                                                <p className="font-medium text-white">{db.name}</p>
                                                {db.database_name && <p className="text-xs text-slate-500">{db.database_name}</p>}
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-4 py-3 text-slate-400">{db.engine}{db.engine_version ? ` ${db.engine_version}` : ""}</td>
                                    <td className="px-4 py-3 text-slate-400 font-mono text-xs">{db.hostname ?? "—"}</td>
                                    <td className="px-4 py-3 text-slate-400">{db.port ?? "—"}</td>
                                    <td className="px-4 py-3 text-slate-400">{db.cloud_provider ?? (db.is_managed ? "Managed" : "Self-hosted")}</td>
                                    <td className="px-4 py-3">
                                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[db.status] ?? "bg-slate-700 text-slate-400"}`}>
                                            {db.status}
                                        </span>
                                    </td>
                                    <td className="px-4 py-3">
                                        <div className="flex items-center gap-1.5 justify-end">
                                            <CanAccess roles={["admin", "contributor"]}>
                                                <button className="rounded-lg p-1.5 text-slate-500 hover:text-slate-300 hover:bg-slate-700 transition"><Pencil className="h-4 w-4" /></button>
                                            </CanAccess>
                                            {isAdmin && (
                                                <button className="rounded-lg p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-900/20 transition" onClick={() => { if (confirm(`Delete ${db.name}?`)) deleteMut.mutate(db.id); }}>
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
