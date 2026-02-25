"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { lookupsApi } from "@/lib/catalog-api";
import { useAuth } from "@/hooks/useAuth";
import { toast } from "sonner";
import {
    Tag, Monitor, Globe, Shield, Code, Database, Layers,
    Plus, Pencil, ToggleLeft, ToggleRight, ChevronRight,
} from "lucide-react";

type TableKey =
    | "software-categories"
    | "software-types"
    | "environments"
    | "responsibility-roles"
    | "repo-platforms"
    | "tech-categories";

const SECTIONS: { key: TableKey; label: string; icon: React.ElementType; description: string }[] = [
    { key: "software-categories", label: "Software Categories", icon: Tag, description: "ERP, CRM, AI/ML, etc." },
    { key: "software-types", label: "Software Types", icon: Monitor, description: "Web App, REST API, Batch Job, etc." },
    { key: "environments", label: "Environments", icon: Globe, description: "dev, qa, staging, prod, dr, etc." },
    { key: "responsibility-roles", label: "Responsibility Roles", icon: Shield, description: "product_owner, tech_lead, etc." },
    { key: "repo-platforms", label: "Repo Platforms", icon: Code, description: "GitHub, GitLab, Azure DevOps, etc." },
    { key: "tech-categories", label: "Tech Categories", icon: Layers, description: "language, framework, database, etc." },
];

export default function SettingsPage() {
    const [activeSection, setActiveSection] = useState<TableKey>("software-categories");
    const [newName, setNewName] = useState("");
    const [editId, setEditId] = useState<string | null>(null);
    const [editName, setEditName] = useState("");
    const qc = useQueryClient();
    const { isAdmin } = useAuth();

    const { data, isLoading } = useQuery({
        queryKey: ["lookup", activeSection],
        queryFn: () => lookupsApi.list(activeSection),
    });

    const rows = (data?.data ?? []) as { id: string; name: string; is_active?: boolean }[];

    const createMut = useMutation({
        mutationFn: (name: string) => lookupsApi.create(activeSection, { name }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["lookup", activeSection] });
            setNewName("");
            toast.success("Entry created");
        },
        onError: () => toast.error("Failed to create entry"),
    });

    const updateMut = useMutation({
        mutationFn: ({ id, name }: { id: string; name: string }) => lookupsApi.update(activeSection, id, { name }),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["lookup", activeSection] });
            setEditId(null);
            toast.success("Entry updated");
        },
        onError: () => toast.error("Failed to update entry"),
    });

    const toggleMut = useMutation({
        mutationFn: ({ id, is_active }: { id: string; is_active: boolean }) =>
            lookupsApi.patchActive(activeSection, id, is_active),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["lookup", activeSection] });
            toast.success("Status updated");
        },
        onError: () => toast.error("Failed to toggle status"),
    });

    const supportsActive = activeSection !== "tech-categories";

    return (
        <div className="space-y-6">
            <div>
                <h1 className="text-2xl font-bold text-white">Settings</h1>
                <p className="text-slate-400 mt-1">Manage extensible lookup tables used across all catalogs.</p>
                {!isAdmin && (
                    <p className="mt-2 text-sm text-amber-400">⚠ Only admins can modify lookup tables. You have read-only access.</p>
                )}
            </div>

            <div className="flex gap-6">
                {/* Left rail */}
                <aside className="w-56 shrink-0">
                    <nav className="space-y-1">
                        {SECTIONS.map((s) => {
                            const Icon = s.icon;
                            const isActive = activeSection === s.key;
                            return (
                                <button
                                    key={s.key}
                                    onClick={() => { setActiveSection(s.key); setEditId(null); setNewName(""); }}
                                    className={`w-full flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm text-left transition-colors ${isActive ? "bg-indigo-600/20 text-indigo-400" : "text-slate-400 hover:bg-slate-800 hover:text-white"
                                        }`}
                                >
                                    <Icon className="h-4 w-4 shrink-0" />
                                    <span className="flex-1 truncate">{s.label}</span>
                                    {isActive && <ChevronRight className="h-3 w-3" />}
                                </button>
                            );
                        })}
                    </nav>
                </aside>

                {/* Content */}
                <div className="flex-1 space-y-4">
                    <div className="flex items-center justify-between">
                        <div>
                            <h2 className="text-lg font-semibold text-white">
                                {SECTIONS.find(s => s.key === activeSection)?.label}
                            </h2>
                            <p className="text-sm text-slate-500">
                                {SECTIONS.find(s => s.key === activeSection)?.description}
                            </p>
                        </div>
                        {/* Add new */}
                        {isAdmin && (
                            <form
                                onSubmit={(e) => { e.preventDefault(); if (newName.trim()) createMut.mutate(newName.trim()); }}
                                className="flex gap-2"
                            >
                                <input
                                    value={newName}
                                    onChange={e => setNewName(e.target.value)}
                                    placeholder="New entry name…"
                                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-1.5 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500 w-48"
                                />
                                <button
                                    type="submit"
                                    disabled={!newName.trim() || createMut.isPending}
                                    className="inline-flex items-center gap-1.5 rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:opacity-50"
                                >
                                    <Plus className="h-4 w-4" />
                                    Add
                                </button>
                            </form>
                        )}
                    </div>

                    <div className="rounded-xl border border-slate-800 bg-slate-900 overflow-hidden">
                        {isLoading ? (
                            <div className="flex h-32 items-center justify-center text-slate-500 text-sm">Loading…</div>
                        ) : rows.length === 0 ? (
                            <div className="flex h-32 items-center justify-center text-slate-500 text-sm">No entries yet</div>
                        ) : (
                            <table className="w-full text-sm">
                                <thead>
                                    <tr className="border-b border-slate-800 text-left text-xs text-slate-500 uppercase tracking-wide">
                                        <th className="px-4 py-3 font-medium">Name</th>
                                        {supportsActive && <th className="px-4 py-3 font-medium">Status</th>}
                                        {isAdmin && <th className="px-4 py-3 font-medium text-right">Actions</th>}
                                    </tr>
                                </thead>
                                <tbody>
                                    {rows.map(row => (
                                        <tr key={row.id} className="border-b border-slate-800/50 hover:bg-slate-800/20 transition-colors">
                                            <td className="px-4 py-3">
                                                {editId === row.id ? (
                                                    <form
                                                        onSubmit={(e) => { e.preventDefault(); updateMut.mutate({ id: row.id, name: editName }); }}
                                                        className="flex gap-2"
                                                    >
                                                        <input
                                                            autoFocus
                                                            value={editName}
                                                            onChange={e => setEditName(e.target.value)}
                                                            className="rounded-lg border border-indigo-500 bg-slate-800 px-2 py-1 text-sm text-white outline-none"
                                                        />
                                                        <button type="submit" className="text-xs text-indigo-400 hover:text-indigo-300">Save</button>
                                                        <button type="button" onClick={() => setEditId(null)} className="text-xs text-slate-500 hover:text-slate-300">Cancel</button>
                                                    </form>
                                                ) : (
                                                    <span className="font-medium text-white">{row.name}</span>
                                                )}
                                            </td>
                                            {supportsActive && (
                                                <td className="px-4 py-3">
                                                    <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${row.is_active !== false ? "bg-emerald-900/40 text-emerald-400" : "bg-slate-700 text-slate-400"
                                                        }`}>
                                                        {row.is_active !== false ? "Active" : "Inactive"}
                                                    </span>
                                                </td>
                                            )}
                                            {isAdmin && (
                                                <td className="px-4 py-3">
                                                    <div className="flex items-center gap-2 justify-end">
                                                        <button
                                                            className="rounded-lg p-1.5 text-slate-500 hover:text-slate-300 hover:bg-slate-700 transition"
                                                            title="Edit name"
                                                            onClick={() => { setEditId(row.id); setEditName(row.name); }}
                                                        >
                                                            <Pencil className="h-4 w-4" />
                                                        </button>
                                                        {supportsActive && (
                                                            <button
                                                                className="rounded-lg p-1.5 text-slate-500 hover:text-slate-300 hover:bg-slate-700 transition"
                                                                title={row.is_active !== false ? "Deactivate" : "Activate"}
                                                                onClick={() => toggleMut.mutate({ id: row.id, is_active: !row.is_active })}
                                                            >
                                                                {row.is_active !== false
                                                                    ? <ToggleRight className="h-4 w-4 text-emerald-400" />
                                                                    : <ToggleLeft className="h-4 w-4" />
                                                                }
                                                            </button>
                                                        )}
                                                    </div>
                                                </td>
                                            )}
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}
