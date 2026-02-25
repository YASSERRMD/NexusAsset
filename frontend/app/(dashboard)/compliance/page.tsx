"use client";

import { useQuery } from "@tanstack/react-query";
import apiClient from "@/lib/api";
import { ShieldCheck, Download, FileJson, FileSpreadsheet } from "lucide-react";

interface ComplianceSummary {
    criticality: string;
    status: string;
    count: number;
}

export default function CompliancePage() {
    const { data: summary, isLoading } = useQuery({
        queryKey: ["compliance", "summary"],
        queryFn: () => apiClient.get<{ data: ComplianceSummary[] }>("/compliance/summary").then(r => r.data.data)
    });

    const handleExport = (format: "json" | "csv") => {
        // We trigger download by opening the export URL in a new window/tab
        // or using an anchor tag. For simplicity, we create a temporary anchor.
        const tokenStr = document.cookie.split('; ').find(row => row.startsWith('access_token='));
        const token = tokenStr ? tokenStr.split('=')[1] : '';

        // Use fetch to hit the API with Auth header, then create a local blob URL
        fetch(`${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1"}/compliance/export?format=${format}`, {
            headers: {
                "Authorization": `Bearer ${token}`
            }
        })
            .then(res => res.blob())
            .then(blob => {
                const url = window.URL.createObjectURL(blob);
                const a = document.createElement("a");
                a.style.display = "none";
                a.href = url;
                a.download = `compliance_export_${new Date().toISOString().split('T')[0]}.${format}`;
                document.body.appendChild(a);
                a.click();
                window.URL.revokeObjectURL(url);
                document.body.removeChild(a);
            })
            .catch(err => console.error("Export failed", err));
    };

    return (
        <div className="space-y-6">
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-bold text-white flex items-center gap-2">
                        <ShieldCheck className="h-6 w-6 text-emerald-400" />
                        Compliance & Export
                    </h1>
                    <p className="text-slate-400 mt-1">ISO 27001 continuous compliance reporting and asset inventory exports.</p>
                </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                {/* Export Card */}
                <div className="rounded-xl border border-slate-800 bg-slate-900 p-6 flex flex-col justify-between">
                    <div>
                        <h3 className="font-semibold text-white mb-2 flex items-center gap-2">
                            <Download className="h-5 w-5 text-indigo-400" />
                            Generate Auditor Report
                        </h3>
                        <p className="text-sm text-slate-400 mb-6 line-clamp-3">
                            Exports a complete inventory of all documented software assets mapped to
                            <strong className="text-slate-300 font-medium"> ISO 27001 Annex A.8.3 (Information Asset Management)</strong>.
                            The report includes ownership tracing, criticality assessments, and operational status.
                        </p>
                    </div>

                    <div className="flex gap-4">
                        <button
                            onClick={() => handleExport("json")}
                            className="flex-1 inline-flex justify-center items-center gap-2 rounded-lg bg-slate-800 border border-slate-700 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-slate-700 hover:border-slate-600"
                        >
                            <FileJson className="h-4 w-4 text-emerald-400" />
                            Export JSON
                        </button>
                        <button
                            onClick={() => handleExport("csv")}
                            className="flex-1 inline-flex justify-center items-center gap-2 rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-indigo-500"
                        >
                            <FileSpreadsheet className="h-4 w-4 text-emerald-400" />
                            Export CSV
                        </button>
                    </div>
                </div>

                {/* Summary Matrix */}
                <div className="rounded-xl border border-slate-800 bg-slate-900 p-6">
                    <h3 className="font-semibold text-white mb-4">Inventory Matrix (Criticality × Status)</h3>

                    {isLoading ? (
                        <div className="h-32 flex items-center justify-center text-slate-500 text-sm">Loading matrix...</div>
                    ) : !summary || summary.length === 0 ? (
                        <div className="h-32 flex items-center justify-center text-slate-500 text-sm">No asset data available.</div>
                    ) : (
                        <div className="overflow-hidden rounded-lg border border-slate-800">
                            <table className="w-full text-left text-sm">
                                <thead className="bg-slate-800/50 text-slate-400">
                                    <tr>
                                        <th className="px-4 py-2 font-medium">Criticality</th>
                                        <th className="px-4 py-2 font-medium">Status</th>
                                        <th className="px-4 py-2 font-medium text-right">Asset Volume</th>
                                    </tr>
                                </thead>
                                <tbody className="divide-y divide-slate-800/50">
                                    {summary.map((row, i) => (
                                        <tr key={i} className="hover:bg-slate-800/20">
                                            <td className="px-4 py-3 capitalize text-slate-300">
                                                <span className={`inline-flex items-center gap-1.5 ${row.criticality === 'critical' ? 'text-red-400 font-medium' :
                                                        row.criticality === 'high' ? 'text-orange-400' :
                                                            row.criticality === 'medium' ? 'text-amber-400' : 'text-emerald-400'
                                                    }`}>
                                                    <span className="h-1.5 w-1.5 rounded-full bg-current"></span>
                                                    {row.criticality}
                                                </span>
                                            </td>
                                            <td className="px-4 py-3 text-slate-400">{row.status.replace(/_/g, ' ')}</td>
                                            <td className="px-4 py-3 text-right font-mono text-white">{row.count}</td>
                                        </tr>
                                    ))}
                                </tbody>
                            </table>
                        </div>
                    )}
                </div>
            </div>
        </div>
    );
}
