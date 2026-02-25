"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { personsApi, teamsApi } from "@/lib/catalog-api";
import { useAuth } from "@/hooks/useAuth";
import { CanAccess } from "@/components/layout/CanAccess";
import { toast } from "sonner";
import { Plus, Search, Pencil, Trash2, Users, User } from "lucide-react";
import type { Person, Team } from "@/types";
import { formatDate } from "@/lib/utils";

export default function PeoplePage() {
    const [tab, setTab] = useState<"persons" | "teams">("persons");
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const qc = useQueryClient();

    const { isAdmin } = useAuth();

    // ─── Persons data ──────────────────────────────────────────────────────────
    const personsQ = useQuery({
        queryKey: ["persons", page, search],
        queryFn: () => personsApi.list({ page, limit: 20, search }),
        enabled: tab === "persons",
    });

    const teamsQ = useQuery({
        queryKey: ["teams", page, search],
        queryFn: () => teamsApi.list({ page, limit: 20, search }),
        enabled: tab === "teams",
    });

    const deletePersonMut = useMutation({
        mutationFn: personsApi.delete,
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["persons"] });
            toast.success("Person deleted");
        },
        onError: () => toast.error("Failed to delete person"),
    });

    const deleteTeamMut = useMutation({
        mutationFn: teamsApi.delete,
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["teams"] });
            toast.success("Team deleted");
        },
        onError: () => toast.error("Failed to delete team"),
    });

    const persons: Person[] = personsQ.data?.data ?? [];
    const teams: Team[] = teamsQ.data?.data ?? [];
    const total = (tab === "persons" ? personsQ.data : teamsQ.data)?.meta?.total ?? 0;
    const isLoading = tab === "persons" ? personsQ.isLoading : teamsQ.isLoading;

    return (
        <div className="space-y-6">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white">People</h1>
                    <p className="text-slate-400 mt-1">Manage persons and teams in your organization.</p>
                </div>
                <CanAccess roles={["admin", "contributor"]}>
                    <button
                        className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500"
                        onClick={() => toast.info(`Create ${tab === "persons" ? "person" : "team"} form coming in Phase 3`)}
                    >
                        <Plus className="h-4 w-4" />
                        Add {tab === "persons" ? "Person" : "Team"}
                    </button>
                </CanAccess>
            </div>

            {/* Tabs */}
            <div className="flex gap-2 border-b border-slate-800">
                {(["persons", "teams"] as const).map((t) => (
                    <button
                        key={t}
                        onClick={() => { setTab(t); setPage(1); setSearch(""); }}
                        className={`flex items-center gap-2 px-4 py-2.5 text-sm font-medium border-b-2 -mb-px transition-colors ${tab === t
                                ? "border-indigo-500 text-indigo-400"
                                : "border-transparent text-slate-500 hover:text-slate-300"
                            }`}
                    >
                        {t === "persons" ? <User className="h-4 w-4" /> : <Users className="h-4 w-4" />}
                        {t === "persons" ? "Persons" : "Teams"}
                    </button>
                ))}
            </div>

            {/* Search */}
            <div className="relative w-full max-w-sm">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                    value={search}
                    onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                    placeholder={`Search ${tab}…`}
                    className="w-full rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500"
                />
            </div>

            {/* Table */}
            <div className="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden">
                {isLoading ? (
                    <div className="flex h-48 items-center justify-center text-slate-500 text-sm">Loading…</div>
                ) : tab === "persons" ? (
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
                                <th className="px-4 py-3 font-medium">Name</th>
                                <th className="px-4 py-3 font-medium">Email</th>
                                <th className="px-4 py-3 font-medium">Title</th>
                                <th className="px-4 py-3 font-medium">Department</th>
                                <th className="px-4 py-3 font-medium">Status</th>
                                <th className="px-4 py-3 font-medium">Joined</th>
                                <th className="px-4 py-3 font-medium" />
                            </tr>
                        </thead>
                        <tbody>
                            {persons.length === 0 ? (
                                <tr><td colSpan={7} className="px-4 py-10 text-center text-slate-500">No persons found</td></tr>
                            ) : persons.map((p) => (
                                <tr key={p.id} className="border-b border-slate-800/50 hover:bg-slate-800/30 transition-colors">
                                    <td className="px-4 py-3 font-medium text-white">{p.full_name}</td>
                                    <td className="px-4 py-3 text-slate-400">{p.email}</td>
                                    <td className="px-4 py-3 text-slate-400">{p.title ?? "—"}</td>
                                    <td className="px-4 py-3 text-slate-400">{p.department ?? "—"}</td>
                                    <td className="px-4 py-3">
                                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${p.is_active ? "bg-emerald-900/40 text-emerald-400" : "bg-slate-700 text-slate-400"
                                            }`}>{p.is_active ? "Active" : "Inactive"}</span>
                                    </td>
                                    <td className="px-4 py-3 text-slate-500">{formatDate(p.created_at)}</td>
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
                                                    onClick={() => { if (confirm("Delete this person?")) deletePersonMut.mutate(p.id); }}
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
                ) : (
                    <table className="w-full text-sm">
                        <thead>
                            <tr className="border-b border-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
                                <th className="px-4 py-3 font-medium">Team</th>
                                <th className="px-4 py-3 font-medium">Department</th>
                                <th className="px-4 py-3 font-medium">Email</th>
                                <th className="px-4 py-3 font-medium">Status</th>
                                <th className="px-4 py-3 font-medium">Created</th>
                                <th className="px-4 py-3 font-medium" />
                            </tr>
                        </thead>
                        <tbody>
                            {teams.length === 0 ? (
                                <tr><td colSpan={6} className="px-4 py-10 text-center text-slate-500">No teams found</td></tr>
                            ) : teams.map((t) => (
                                <tr key={t.id} className="border-b border-slate-800/50 hover:bg-slate-800/30 transition-colors">
                                    <td className="px-4 py-3 font-medium text-white">{t.name}</td>
                                    <td className="px-4 py-3 text-slate-400">{t.department}</td>
                                    <td className="px-4 py-3 text-slate-400">{t.team_email ?? "—"}</td>
                                    <td className="px-4 py-3">
                                        <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${t.is_active ? "bg-emerald-900/40 text-emerald-400" : "bg-slate-700 text-slate-400"
                                            }`}>{t.is_active ? "Active" : "Inactive"}</span>
                                    </td>
                                    <td className="px-4 py-3 text-slate-500">{formatDate(t.created_at)}</td>
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
                                                    onClick={() => { if (confirm("Delete this team?")) deleteTeamMut.mutate(t.id); }}
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
