"use client";

import { useState } from "react";
import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { CanAccess } from "@/components/layout/CanAccess";
import { toast } from "sonner";
import { Plus, Search, Pencil, Trash2, Store, Globe, Mail, Phone } from "lucide-react";
import type { ApiResponse, Vendor } from "@/types";
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

export default function VendorsPage() {
    const [search, setSearch] = useState("");
    const [page, setPage] = useState(1);
    const [isCreateOpen, setIsCreateOpen] = useState(false);
    const qc = useQueryClient();
    const { isAdmin } = useAuth();

    const { data, isLoading } = useQuery({
        queryKey: ["vendors", page, search],
        queryFn: () => apiClient.get<ApiResponse<Vendor[]>>("/vendors", { params: { page, limit: 20, search } }).then(r => r.data),
    });

    const deleteMut = useMutation({
        mutationFn: (id: string) => apiClient.delete(`/vendors/${id}`),
        onSuccess: () => { qc.invalidateQueries({ queryKey: ["vendors"] }); toast.success("Vendor deleted"); },
        onError: () => toast.error("Failed to delete"),
    });

    const createMut = useMutation({
        mutationFn: (data: Partial<Vendor>) => apiClient.post("/vendors", data),
        onSuccess: () => {
            qc.invalidateQueries({ queryKey: ["vendors"] });
            toast.success("Vendor created");
            setIsCreateOpen(false);
        },
        onError: (err: any) => toast.error(err.response?.data?.error || "Failed to create vendor"),
    });

    const vendors: Vendor[] = data?.data ?? [];
    const total = data?.meta?.total ?? 0;

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white">Vendors</h1>
                    <p className="text-slate-400 mt-1">{total} vendor{total !== 1 ? "s" : ""} registered</p>
                </div>
                <CanAccess roles={["admin", "contributor"]}>
                    <Dialog open={isCreateOpen} onOpenChange={setIsCreateOpen}>
                        <DialogTrigger asChild>
                            <button className="inline-flex items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500">
                                <Plus className="h-4 w-4" /> Add Vendor
                            </button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-md bg-slate-900 border-slate-800 text-white">
                            <DialogHeader>
                                <DialogTitle>Add New Vendor</DialogTitle>
                            </DialogHeader>
                            <form
                                onSubmit={(e) => {
                                    e.preventDefault();
                                    const fd = new FormData(e.currentTarget);
                                    createMut.mutate({
                                        name: fd.get("name") as string,
                                        display_name: fd.get("display_name") as string,
                                        region: fd.get("region") as string,
                                        website: fd.get("website") as string,
                                        contact_email: fd.get("contact_email") as string,
                                        contact_phone: fd.get("contact_phone") as string,
                                    });
                                }}
                                className="space-y-4"
                            >
                                <div className="space-y-2">
                                    <Label htmlFor="name">Vendor ID (unique)</Label>
                                    <Input id="name" name="name" required placeholder="e.g. microsoft" className="bg-slate-800 border-slate-700 font-mono text-sm" />
                                </div>
                                <div className="space-y-2">
                                    <Label htmlFor="display_name">Display Name</Label>
                                    <Input id="display_name" name="display_name" placeholder="Microsoft Corp" className="bg-slate-800 border-slate-700" />
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="region">Region</Label>
                                        <Input id="region" name="region" placeholder="US / EMEA" className="bg-slate-800 border-slate-700" />
                                    </div>
                                    <div className="space-y-2">
                                        <Label htmlFor="website">Website URL</Label>
                                        <Input id="website" name="website" placeholder="https://..." type="url" className="bg-slate-800 border-slate-700" />
                                    </div>
                                </div>
                                <div className="grid grid-cols-2 gap-4">
                                    <div className="space-y-2">
                                        <Label htmlFor="contact_email">Email</Label>
                                        <Input id="contact_email" name="contact_email" placeholder="support@..." type="email" className="bg-slate-800 border-slate-700" />
                                    </div>
                                    <div className="space-y-2">
                                        <Label htmlFor="contact_phone">Phone</Label>
                                        <Input id="contact_phone" name="contact_phone" className="bg-slate-800 border-slate-700" />
                                    </div>
                                </div>
                                <button
                                    type="submit"
                                    disabled={createMut.isPending}
                                    className="w-full mt-4 rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-indigo-500 disabled:opacity-50"
                                >
                                    {createMut.isPending ? "Adding..." : "Add Vendor"}
                                </button>
                            </form>
                        </DialogContent>
                    </Dialog>
                </CanAccess>
            </div>

            <div className="relative max-w-sm">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-slate-500" />
                <input
                    value={search}
                    onChange={(e) => { setSearch(e.target.value); setPage(1); }}
                    placeholder="Search vendors…"
                    className="w-full rounded-lg border border-slate-700 bg-slate-800 pl-9 pr-4 py-2 text-sm text-white placeholder-slate-500 outline-none focus:border-indigo-500"
                />
            </div>

            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-3">
                {isLoading ? (
                    <div className="col-span-3 flex h-48 items-center justify-center text-slate-500 text-sm">Loading…</div>
                ) : vendors.length === 0 ? (
                    <div className="col-span-3 flex h-48 items-center justify-center rounded-xl border border-slate-800 bg-slate-900 text-slate-500 text-sm">No vendors found</div>
                ) : vendors.map((v) => (
                    <div key={v.id} className="group relative rounded-xl border border-slate-800 bg-slate-900 p-5 hover:border-slate-700 transition">
                        <div className="absolute top-4 right-4 flex gap-1.5 opacity-0 group-hover:opacity-100 transition-opacity">
                            <CanAccess roles={["admin", "contributor"]}>
                                <button className="rounded-lg p-1.5 text-slate-500 hover:text-slate-300 hover:bg-slate-800 transition"><Pencil className="h-3.5 w-3.5" /></button>
                            </CanAccess>
                            {isAdmin && (
                                <button className="rounded-lg p-1.5 text-slate-500 hover:text-red-400 hover:bg-red-900/20 transition" onClick={() => { if (confirm(`Delete ${v.name}?`)) deleteMut.mutate(v.id); }}>
                                    <Trash2 className="h-3.5 w-3.5" />
                                </button>
                            )}
                        </div>

                        <div className="flex items-center gap-3 mb-3">
                            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-purple-600/20">
                                <Store className="h-5 w-5 text-purple-400" />
                            </div>
                            <div>
                                <h3 className="font-semibold text-white text-sm">{v.display_name || v.name}</h3>
                                {v.region && <p className="text-xs text-slate-500">{v.region}</p>}
                            </div>
                        </div>

                        <div className="space-y-1.5 text-xs text-slate-500">
                            {v.website && (
                                <a href={v.website} target="_blank" rel="noopener noreferrer" className="flex items-center gap-1.5 hover:text-slate-300 transition">
                                    <Globe className="h-3 w-3 shrink-0" /><span className="truncate">{v.website}</span>
                                </a>
                            )}
                            {v.contact_email && <div className="flex items-center gap-1.5"><Mail className="h-3 w-3 shrink-0" />{v.contact_email}</div>}
                            {v.contact_phone && <div className="flex items-center gap-1.5"><Phone className="h-3 w-3 shrink-0" />{v.contact_phone}</div>}
                        </div>

                        <div className="mt-3 pt-3 border-t border-slate-800">
                            <span className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${v.is_active ? "bg-emerald-900/40 text-emerald-400" : "bg-slate-700 text-slate-400"}`}>
                                {v.is_active ? "Active" : "Inactive"}
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
