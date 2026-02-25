"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api";
import { Search, Shield, Filter, Clock } from "lucide-react";
import { formatDate } from "@/lib/utils";
import type { ApiResponse } from "@/types";

interface AuditLog {
    id: string;
    user_id?: string;
    entity_type: string;
    entity_id: string;
    action: string;
    changes?: Record<string, { old: unknown; new: unknown }>;
    ip_address?: string;
    user_agent?: string;
    created_at: string;
}

export default function AuditLogPage() {
    const [page, setPage] = useState(1);
    const [actionFilter, setActionFilter] = useState("");
    const [entityTypeFilter, setEntityTypeFilter] = useState("");
    const [entityIdSearch, setEntityIdSearch] = useState("");

    const { data, isLoading } = useQuery({
        queryKey: ["audit", page, actionFilter, entityTypeFilter, entityIdSearch],
        queryFn: () => apiClient.get<ApiResponse<AuditLog[]>>("/audit", {
            params: {
                page,
                limit: 20,
                action: actionFilter,
                entity_type: entityTypeFilter,
                entity_id: entityIdSearch
            }
        }).then(r => r.data)
    });

    const logs = data?.data ?? [];
    const total = data?.meta?.total ?? 0;

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white flex items-center gap-2">
                        <Shield className="h-6 w-6 text-indigo-400" />
                        Audit Log
                    </h1>
                    <p className="text-slate-400 mt-1">System-wide immutable record of mutations and administrative actions.</p>
                </div>
            </div>

            {/* Filters */}
            <div className="flex flex-wrap gap-3">
                <div className="relative flex-1 min-w-48">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                    <input
                        value={entityIdSearch}
                        onChange={(e) => { setEntityIdSearch(e.target.value); setPage(1); }}
                        placeholder="Search by Entity ID..."
                        className="w-full rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500"
                    />
                </div>

                <div className="flex items-center gap-2 rounded-lg border border-slate-700 bg-slate-800 px-3 py-2">
                    <Filter className="h-4 w-4 text-slate-500" />
                    <select
                        value={entityTypeFilter}
                        onChange={(e) => { setEntityTypeFilter(e.target.value); setPage(1); }}
                        className="bg-transparent text-sm text-white outline-none"
                    >
                        <option value="">All Entities</option>
                        <option value="/api/v1/software">Software</option>
                        <option value="/api/v1/vendors">Vendors</option>
                        <option value="/api/v1/databases">Databases</option>
                        <option value="/api/v1/integrations">Integrations</option>
                        <option value="/api/v1/servers">Servers</option>
                        <option value="/api/v1/persons">Persons</option>
                        <option value="/api/v1/teams">Teams</option>
                        <option value="/api/v1/users">Users</option>
                    </select>
                </div>

                <select
                    value={actionFilter}
                    onChange={(e) => { setActionFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All Actions</option>
                    <option value="POST">POST (Create)</option>
                    <option value="PUT">PUT (Update)</option>
                    <option value="PATCH">PATCH (Modify)</option>
                    <option value="DELETE">DELETE (Remove)</option>
                </select>
            </div>

            {/* Table */}
            <div className="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden text-sm">
                <table className="w-full text-left">
                    <thead className="border-b border-slate-800 bg-slate-900/50 text-slate-400">
                        <tr>
                            <th className="px-4 py-3 font-medium">Timestamp</th>
                            <th className="px-4 py-3 font-medium">Action</th>
                            <th className="px-4 py-3 font-medium">Entity Type / Route</th>
                            <th className="px-4 py-3 font-medium">Entity ID</th>
                            <th className="px-4 py-3 font-medium">Actor (User ID)</th>
                            <th className="px-4 py-3 font-medium text-right">Context</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-800 text-slate-300">
                        {isLoading ? (
                            <tr>
                                <td colSpan={6} className="px-4 py-8 text-center text-slate-500">Loading audit logs...</td>
                            </tr>
                        ) : logs.length === 0 ? (
                            <tr>
                                <td colSpan={6} className="px-4 py-8 text-center text-slate-500">No logs found matching your filters.</td>
                            </tr>
                        ) : (
                            logs.map(log => (
                                <tr key={log.id} className="hover:bg-slate-800/50 transition">
                                    <td className="px-4 py-3 text-slate-400 whitespace-nowrap flex items-center gap-1.5">
                                        <Clock className="h-3 w-3" />
                                        {formatDate(log.created_at)}
                                    </td>
                                    <td className="px-4 py-3">
                                        <span className={`inline-flex px-1.5 py-0.5 rounded text-[10px] font-mono font-bold uppercase tracking-wider ${log.action === 'DELETE' ? 'bg-red-900/40 text-red-400' :
                                                log.action === 'POST' ? 'bg-emerald-900/40 text-emerald-400' :
                                                    'bg-blue-900/40 text-blue-400'
                                            }`}>
                                            {log.action}
                                        </span>
                                    </td>
                                    <td className="px-4 py-3 font-mono text-xs text-indigo-300 truncate max-w-[200px]" title={log.entity_type}>
                                        {log.entity_type}
                                    </td>
                                    <td className="px-4 py-3 font-mono text-xs text-slate-400">
                                        {log.entity_id === "unknown_or_new" ? <span className="text-slate-500 italic">new entity</span> : log.entity_id.split('-')[0] + '...'}
                                    </td>
                                    <td className="px-4 py-3 font-mono text-xs text-slate-400">
                                        {log.user_id ? log.user_id.split('-')[0] + '...' : <span className="text-emerald-500/50">system</span>}
                                    </td>
                                    <td className="px-4 py-3 font-mono text-[10px] text-right text-slate-500 truncate max-w-[150px]" title={log.ip_address || "system"}>
                                        {log.ip_address || "internal"}
                                    </td>
                                </tr>
                            ))
                        )}
                    </tbody>
                </table>
            </div>

            {/* Pagination */}
            {total > 20 && (
                <div className="flex items-center justify-between text-sm text-slate-500">
                    <span>{total} total records</span>
                    <div className="flex gap-2">
                        <button disabled={page <= 1} onClick={() => setPage(p => p - 1)} className="px-3 py-1 rounded-lg border border-slate-700 disabled:opacity-40 hover:bg-slate-800 transition">Prev</button>
                        <span className="px-3 py-1 bg-slate-800 rounded-lg text-white">Page {page}</span>
                        <button disabled={page * 20 >= total} onClick={() => setPage(p => p + 1)} className="px-3 py-1 rounded-lg border border-slate-700 disabled:opacity-40 hover:bg-slate-800 transition">Next</button>
                    </div>
                </div>
            )}
        </div>
    );
}
