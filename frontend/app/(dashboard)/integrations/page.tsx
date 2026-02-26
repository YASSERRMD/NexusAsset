"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { CanAccess } from "@/components/layout/CanAccess";
import { toast } from "sonner";
import { Plus, Search, Pencil, Trash2, Network, ArrowRight } from "lucide-react";
import type { ApiResponse, Integration } from "@/types";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

const STATUS_COLORS: Record<string, string> = {
    active: "bg-emerald-900/40 text-emerald-400",
    inactive: "bg-slate-700 text-slate-400",
    deprecated: "bg-amber-900/40 text-amber-400",
    maintenance: "bg-blue-900/40 text-blue-400",
};

const KIND_COLORS: Record<string, string> = {
    api: "bg-indigo-900/40 text-indigo-400",
    event: "bg-purple-900/40 text-purple-400",
    file: "bg-cyan-900/40 text-cyan-400",
    database: "bg-emerald-900/40 text-emerald-400",
    webhook: "bg-amber-900/40 text-amber-400",
    sync: "bg-orange-900/40 text-orange-400",
};

export default function IntegrationsPage() {
    const [search, setSearch] = useState("");
    const [kindFilter, setKindFilter] = useState("");
    const [page, setPage] = useState(1);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const [editingIntegration, setEditingIntegration] = useState<Integration | null>(null);
    const qc = useQueryClient();
    const { isAdmin } = useAuth();

    const { data, isLoading } = useQuery({
        queryKey: ["integrations", page, search, kindFilter],
        queryFn: () =>
            apiClient.get<ApiResponse<Integration[]>>("/integrations", {
                params: { page, limit: 20, search, kind: kindFilter },
            }).then(r => r.data),
    });

    const deleteMut = useMutation({
        mutationFn: (id: string) => apiClient.delete(`/integrations/${id}`),
        onSuccess: () => { qc.invalidateQueries({ queryKey: ["integrations"] }); toast.success("Integration deleted"); },
        onError: () => toast.error("Failed to delete"),
    });

    const createMut = useMutation({
        mutationFn: (data: Partial<Integration>) => apiClient.post("/integrations", data),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["integrations"] });
            toast.success("Integration created");
            setIsCreateOpen(false);
        },
        onError: (err: any) => toast.error(err.response?.data?.error || "Failed to create integration"),
    });

    const updateMut = useMutation({
        mutationFn: ({ id, data }: { id: string; data: Partial<Integration> }) => apiClient.put(`/integrations/${id}`, data),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["integrations"] });
            toast.success("Integration updated");
            setEditingIntegration(null);
        },
        onError: (err: any) => toast.error(err.response?.data?.error || "Failed to update integration"),
    });

    const items: Integration[] = data?.data ?? [];
    const total = data?.meta?.total ?? 0;

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white">Integrations</h1>
                    <p className="text-slate-400 mt-1">{total} integration{total !== 1 ? "s" : ""} registered</p>
                </div>
                <CanAccess roles={["admin", "contributor"]}>
                    <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                        <DialogTrigger asChild>
                            <button className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500">
                                <Plus className="h-4 w-4" /> Add Integration
                            </button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-md bg-slate-900 border-slate-800 text-white">
                            <DialogHeader>
                                <DialogTitle>New Integration Link</DialogTitle>
                            </DialogHeader>
                            <form
                                onSubmit={(e) => {
                                    e.preventDefault();
                                    const fd = new FormData(e.currentTarget);
                                    createMut.mutate({
                                        name: fd.get("name") as string,
                                        description: fd.get("description") as string,
                                        integration_kind: fd.get("integration_kind") as any,
                                        protocol: fd.get("protocol") as string,
                                        source_software_id: (fd.get("source_software_id") as string) || undefined,
                                        target_software_id: (fd.get("target_software_id") as string) || undefined,
                                    });
                                }}
                                className="space-y-4"
                            >
                                <div className="space-y-2">
                                    <Label htmlFor="name">Integration Name</Label>
                                    <Input id="name" name="name" required placeholder="User Sync API" className="bg-slate-800 border-slate-700" />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="description">Description</Label>
                                    <Input id="description" name="description" className="bg-slate-800 border-slate-700" />
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="integration_kind">Integration Kind</Label>
                                        <select id="integration_kind" name="integration_kind" className="w-full h-10 rounded-md border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500">
                                            <option value="api">API</option>
                                            <option value="event">Event (Kafka/MQ)</option>
                                            <option value="file">File Transfer (SFTP)</option>
                                            <option value="database">Database Link</option>
                                            <option value="webhook">Webhook</option>
                                            <option value="sync">Data Sync</option>
                                        </select>
                                    </div>
                                    <div className="space-y-2">
                                        <Label htmlFor="protocol">Protocol / Tech</Label>
                                        <Input id="protocol" name="protocol" placeholder="REST, GraphQL..." className="bg-slate-800 border-slate-700" />
                                    </div>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="source_software_id">Source SW ID (opt)</Label>
                                        <Input id="source_software_id" name="source_software_id" placeholder="UUID" className="bg-slate-800 border-slate-700 font-mono text-xs" />
                                    </div>
                                    <div className="space-y-2">
                                        <Label htmlFor="target_software_id">Target SW ID (opt)</Label>
                                        <Input id="target_software_id" name="target_software_id" placeholder="UUID" className="bg-slate-800 border-slate-700 font-mono text-xs" />
                                    </div>
                                </div>
                                <button
                                    type="submit"
                                    disabled={createMut.isPending}
                                    className="w-full mt-4 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:opacity-50"
                                >
                                    {createMut.isPending ? "Creating..." : "Create Integration"}
                                </button>
                            </form>
                        </DialogContent>
                    </Dialog>
                </CanAccess>
            </div>

            {/* Edit Dialog */}
            <Dialog open={!!editingIntegration} onOpenChange={(open) => !open && setEditingIntegration(null)}>
                <DialogContent className="sm:max-w-md bg-slate-900 border-slate-800 text-white">
                    <DialogHeader>
                        <DialogTitle>Edit Integration</DialogTitle>
                    </DialogHeader>
                    {editingIntegration && (
                        <form
                            onSubmit={(e) => {
                                e.preventDefault();
                                const fd = new FormData(e.currentTarget);
                                updateMut.mutate({
                                    id: editingIntegration.id,
                                    data: {
                                        name: fd.get("name") as string,
                                        description: fd.get("description") as string,
                                        integration_kind: fd.get("integration_kind") as any,
                                        protocol: fd.get("protocol") as string,
                                        source_software_id: (fd.get("source_software_id") as string) || undefined,
                                        target_software_id: (fd.get("target_software_id") as string) || undefined,
                                    }
                                });
                            }}
                            className="space-y-4"
                        >
                            <div className="space-y-2">
                                <Label htmlFor="edit_name">Integration Name</Label>
                                <Input id="edit_name" name="name" defaultValue={editingIntegration.name} required className="bg-slate-800 border-slate-700" />
                            </div>
                            <div className="space-y-2">
                                <Label htmlFor="edit_description">Description</Label>
                                <Input id="edit_description" name="description" defaultValue={editingIntegration.description || ""} className="bg-slate-800 border-slate-700" />
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="edit_integration_kind">Integration Kind</Label>
                                    <select id="edit_integration_kind" name="integration_kind" defaultValue={editingIntegration.integration_kind} className="w-full h-10 rounded-md border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500">
                                        <option value="api">API</option>
                                        <option value="event">Event (Kafka/MQ)</option>
                                        <option value="file">File Transfer (SFTP)</option>
                                        <option value="database">Database Link</option>
                                        <option value="webhook">Webhook</option>
                                        <option value="sync">Data Sync</option>
                                    </select>
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="edit_protocol">Protocol / Tech</Label>
                                    <Input id="edit_protocol" name="protocol" defaultValue={editingIntegration.protocol || ""} className="bg-slate-800 border-slate-700" />
                                </div>
                            </div>
                            <div className="grid grid-cols-2 gap-4">
                                <div className="space-y-2">
                                    <Label htmlFor="edit_source_software_id">Source SW ID (opt)</Label>
                                    <Input id="edit_source_software_id" name="source_software_id" defaultValue={editingIntegration.source_software_id || ""} className="bg-slate-800 border-slate-700 font-mono text-xs" />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="edit_target_software_id">Target SW ID (opt)</Label>
                                    <Input id="edit_target_software_id" name="target_software_id" defaultValue={editingIntegration.target_software_id || ""} className="bg-slate-800 border-slate-700 font-mono text-xs" />
                                </div>
                            </div>
                            <button
                                type="submit"
                                disabled={updateMut.isPending}
                                className="w-full mt-4 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:opacity-50"
                            >
                                {updateMut.isPending ? "Saving..." : "Save Changes"}
                            </button>
                        </form>
                    )}
                </DialogContent>
            </Dialog>

            <div className="flex gap-3">
                <div className="relative flex-1 max-w-sm">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                    <input
                        value={search}
                        onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                        placeholder="Search integrations…"
                        className="w-full rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500"
                    />
                </div>
                <select
                    value={kindFilter}
                    onChange={(e) => { setKindFilter(e.target.value); setPage(1); }}
                    className="rounded-lg border border-slate-700 bg-slate-800 px-3 py-2 text-sm text-white outline-none focus:border-indigo-500"
                >
                    <option value="">All kinds</option>
                    {["api", "event", "file", "database", "webhook", "sync"].map(k => (
                        <option key={k} value={k}>{k.charAt(0).toUpperCase() + k.slice(1)}</option>
                    ))}
                </select>
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex h-48 items-center justify-center text-slate-500 text-sm">Loading…</div>
                ) : items.length === 0 ? (
                    <div className="col-span-3 flex h-48 items-center justify-center rounded-xl border border-slate-800 bg-slate-900 text-slate-500 text-sm">No integrations found</div>
                ) : items.map((intg) => (
                    <div key={intg.id} className="group relative rounded-xl border border-slate-800 bg-slate-900 p-5 hover:border-slate-700 transition">
                        <div className="absolute top-4 right-4 flex gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                            <CanAccess roles={["admin", "contributor"]}>
                                <button
                                    className="rounded-lg p-1.5 text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition"
                                    onClick={() => setEditingIntegration(intg)}
                                >
                                    <Pencil className="h-3.5 w-3.5" />
                                </button>
                            </CanAccess>
                            {isAdmin && (
                                <button className="rounded-lg p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-900/20 transition" onClick={() => { if (confirm(`Delete ${intg.name}?`)) deleteMut.mutate(intg.id); }}>
                                    <Trash2 className="h-3.5 w-3.5" />
                                </button>
                            )}
                        </div>

                        <div className="flex items-center gap-3 mb-3">
                            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-slate-800">
                                <Network className="h-5 w-5 text-slate-400" />
                            </div>
                            <div className="flex-1 min-w-0">
                                <h3 className="font-semibold text-white text-sm truncate">{intg.name}</h3>
                                {intg.description && <p className="text-xs text-slate-500 line-clamp-1">{intg.description}</p>}
                            </div>
                        </div>

                        {/* Source → Target */}
                        <div className="flex items-center gap-2 mb-3 text-xs text-slate-500">
                            <span className="rounded bg-slate-800 px-2 py-1 font-mono">{intg.source_software_id ? `SW:${intg.source_software_id.slice(0, 8)}` : "External"}</span>
                            <ArrowRight className="h-3 w-3 shrink-0" />
                            <span className="rounded bg-slate-800 px-2 py-1 font-mono">{intg.target_software_id ? `SW:${intg.target_software_id.slice(0, 8)}` : "External"}</span>
                        </div>

                        <div className="flex flex-wrap gap-1.5">
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${KIND_COLORS[intg.integration_kind] ?? "bg-slate-700 text-slate-400"}`}>
                                {intg.integration_kind}
                            </span>
                            {intg.protocol && (
                                <span className="inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium bg-slate-700 text-slate-400">
                                    {intg.protocol}
                                </span>
                            )}
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${STATUS_COLORS[intg.status] ?? "bg-slate-700 text-slate-400"}`}>
                                {intg.status}
                            </span>
                        </div>
                    </div>
                ))}
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
