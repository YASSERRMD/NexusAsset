"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { CanAccess } from "@/components/layout/CanAccess";
import { toast } from "sonner";
import {
    Plus, Search, Pencil, Trash2, Package,
    ExternalLink, Tag,
} from "lucide-react";
import type { ApiResponse, Software } from "@/types";
import { formatDate } from "@/lib/utils";
import Link from "next/link";

const CRITICALITY_COLORS: Record<string, string> = {
    critical: "bg-red-900/40 text-red-400 border border-red-900/60",
    high: "bg-orange-900/40 text-orange-400 border border-orange-900/60",
    medium: "bg-amber-900/40 text-amber-400 border border-amber-900/60",
    low: "bg-emerald-900/40 text-emerald-400 border border-emerald-900/60",
};

const STATUS_COLORS: Record<string, string> = {
    active: "bg-emerald-900/40 text-emerald-400",
    in_development: "bg-blue-900/40 text-blue-400",
    deprecated: "bg-amber-900/40 text-amber-400",
    eol: "bg-red-900/40 text-red-400",
    on_hold: "bg-slate-700 text-slate-400",
    decommissioned: "bg-slate-700 text-slate-400",
};

export default function SoftwarePage() {
    const [search, setSearch] = useState("");
    const [kindFilter, setKindFilter] = useState("");
    const [statusFilter, setStatusFilter] = useState("");
    const [criticalityFilter, setCriticalityFilter] = useState("");
    const [page, setPage] = useState(1);
    const qc = useQueryClient();
    const { isAdmin } = useAuth();

    const { data, isLoading } = useQuery({
        queryKey: ["software", page, search, kindFilter, statusFilter, criticalityFilter],
        queryFn: () =>
            apiClient
                .get<ApiResponse<Software[]>>("/software", {
                    params: { page, limit: 20, search, kind: kindFilter, status: statusFilter, criticality: criticalityFilter },
                })
                .then((r) => r.data),
    });

    const deleteMut = useMutation({
        mutationFn: (id: string) => apiClient.delete(`/software/${id}`),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["software"] });
            toast.success("Software deleted");
        },
        onError: () => toast.error("Failed to delete software"),
    });

    const items: Software[] = data?.data ?? [];
    const total = data?.meta?.total ?? 0;

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white">Software Catalog</h1>
                    <p className="text-slate-400 mt-1">{total} application{total !== 1 ? "s" : ""} registered</p>
                </div>
                <CanAccess roles={["admin", "contributor"]}>
                    <button
                        className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500"
                        onClick={() => toast.info("Software creation form coming soon")}
                    >
                        <Plus className="h-4 w-4" />
                        Add Software
                    </button>
                </CanAccess>
            </div>

            {/* Filters */}
            <div className="flex flex-wrap gap-3">
                <div className="relative flex-1 min-w-52">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                    <input
                        value={search}
                        onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                        placeholder="Search software…"
                        className="w-full rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500"
                    />
                </div>
                <select
                    value={kindFilter}
                    onChange={(e) => { setKindFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All kinds</option>
                    <option value="inhouse">In-house</option>
                    <option value="vendor">Vendor</option>
                </select>
                <select
                    value={criticalityFilter}
                    onChange={(e) => { setCriticalityFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All criticalities</option>
                    {["critical", "high", "medium", "low"].map(c => (
                        <option key={c} value={c}>{c.charAt(0).toUpperCase() + c.slice(1)}</option>
                    ))}
                </select>
                <select
                    value={statusFilter}
                    onChange={(e) => { setStatusFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All statuses</option>
                    {["active", "in_development", "deprecated", "eol", "on_hold", "decommissioned"].map(s => (
                        <option key={s} value={s}>{s.replace(/_/g, " ")}</option>
                    ))}
                </select>
            </div>

            {/* Card grid */}
            {isLoading ? (
                <div className="flex h-48 items-center justify-center text-slate-500 text-sm">Loading…</div>
            ) : items.length === 0 ? (
                <div className="flex h-48 items-center justify-center rounded-xl border border-slate-800 bg-slate-900 text-slate-500 text-sm">
                    No software found
                </div>
            ) : (
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
                    {items.map((sw) => (
                        <div key={sw.id} className="group relative rounded-xl border border-slate-800 bg-slate-900 p-5 transition hover:border-slate-700">
                            {/* Actions */}
                            <div className="absolute top-4 right-4 flex gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                                <CanAccess roles={["admin", "contributor"]}>
                                    <button className="rounded-lg p-1.5 text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition">
                                        <Pencil className="h-3.5 w-3.5" />
                                    </button>
                                </CanAccess>
                                {isAdmin && (
                                    <button
                                        className="rounded-lg p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-900/20 transition"
                                        onClick={() => { if (confirm(`Delete ${sw.name}?`)) deleteMut.mutate(sw.id); }}
                                    >
                                        <Trash2 className="h-3.5 w-3.5" />
                                    </button>
                                )}
                            </div>

                            {/* Header */}
                            <div className="flex items-start gap-3 mb-3">
                                <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-indigo-600/20">
                                    <Package className="h-5 w-5 text-indigo-400" />
                                </div>
                                <div className="flex-1 min-w-0">
                                    <h3 className="font-semibold text-white text-sm truncate">{sw.display_name || sw.name}</h3>
                                    {sw.version && <p className="text-xs text-slate-500">v{sw.version}</p>}
                                </div>
                            </div>

                            {/* Description */}
                            {sw.description && (
                                <p className="text-xs text-slate-500 mb-3 line-clamp-2">{sw.description}</p>
                            )}

                            {/* Badges */}
                            <div className="flex flex-wrap gap-1.5 mb-3">
                                <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${CRITICALITY_COLORS[sw.criticality] ?? "bg-slate-700 text-slate-400"}`}>
                                    {sw.criticality}
                                </span>
                                <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[sw.status] ?? "bg-slate-700 text-slate-400"}`}>
                                    {sw.status.replace(/_/g, " ")}
                                </span>
                                <span className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-slate-700 text-slate-400">
                                    {sw.software_kind === "inhouse" ? "In-house" : "Vendor"}
                                </span>
                            </div>

                            {/* Tags */}
                            {(sw.tags?.length ?? 0) > 0 && (
                                <div className="flex flex-wrap gap-1 mb-3">
                                    {sw.tags!.slice(0, 3).map(t => (
                                        <span key={t} className="inline-flex items-center gap-1 rounded px-1.5 py-0.5 text-xs bg-slate-800 text-slate-400">
                                            <Tag className="h-2.5 w-2.5" />{t}
                                        </span>
                                    ))}
                                    {(sw.tags!.length ?? 0) > 3 && <span className="text-xs text-slate-600">+{sw.tags!.length - 3}</span>}
                                </div>
                            )}

                            {/* Footer */}
                            <div className="flex items-center justify-between border-t border-slate-800 pt-3 mt-3">
                                <span className="text-xs text-slate-600">Updated {formatDate(sw.updated_at)}</span>
                                <Link
                                    href={`/software/${sw.id}`}
                                    className="inline-flex items-center gap-1 text-xs text-indigo-400 hover:text-indigo-300 transition"
                                >
                                    Details <ExternalLink className="h-3 w-3" />
                                </Link>
                            </div>
                        </div>
                    ))}
                </div>
            )}

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
