"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { Monitor, Eye, EyeOff, AlertCircle } from "lucide-react";
import { useAuthStore } from "@/store/authStore";
import type { AuthTokens, ApiResponse } from "@/types";

const loginSchema = z.object({
    email: z.string().email("Enter a valid email address"),
    password: z.string().min(1, "Password is required"),
});
type LoginForm = z.infer<typeof loginSchema>;

export default function LoginPage() {
    const router = useRouter();
    const { setAuth } = useAuthStore();
    const [showPassword, setShowPassword] = useState(false);
    const [serverError, setServerError] = useState<string | null>(null);

    const {
        register,
        handleSubmit,
        formState: { errors, isSubmitting },
    } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) });

    const onSubmit = async (data: LoginForm) => {
        setServerError(null);
        try {
            const res = await fetch("/api/auth/login", {
                method: "POST",
                headers: { "Content-Type": "application/json" },
                body: JSON.stringify(data),
            });
            const json: ApiResponse<AuthTokens> = await res.json();

            if (!res.ok || !json.success) {
                setServerError(json.error || "Login failed");
                return;
            }

            setAuth(
                {
                    user_id: json.data.user_id,
                    username: json.data.username,
                    email: json.data.email,
                    role: json.data.role,
                },
                json.data.access_token
            );

            router.push("/");
        } catch {
            setServerError("Unable to connect to server. Please try again.");
        }
    };

    return (
        <div className="min-h-screen bg-slate-950 flex items-center justify-center p-4">
            <div className="w-full max-w-md space-y-8">
                {/* Logo */}
                <div className="text-center">
                    <div className="mx-auto flex h-14 w-14 items-center justify-center rounded-2xl bg-indigo-600">
                        <Monitor className="h-7 w-7 text-white" />
                    </div>
                    <h1 className="mt-4 text-2xl font-bold text-white">NexusAsset</h1>
                    <p className="mt-1 text-sm text-slate-400">
                        IT Catalog Management System
                    </p>
                </div>

                {/* Form card */}
                <div className="rounded-2xl border border-slate-800 bg-slate-900 p-8">
                    <h2 className="text-lg font-semibold text-white mb-6">Sign in to your account</h2>

                    {serverError && (
                        <div className="mb-4 flex items-start gap-2 rounded-lg border border-red-900 bg-red-950/50 p-3 text-sm text-red-400">
                            <AlertCircle className="h-4 w-4 mt-0.5 shrink-0" />
                            {serverError}
                        </div>
                    )}

                    <form onSubmit={handleSubmit(onSubmit)} className="space-y-5">
                        {/* Email */}
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1.5" htmlFor="email">
                                Email address
                            </label>
                            <input
                                id="email"
                                type="email"
                                autoComplete="email"
                                placeholder="admin@nexus.local"
                                {...register("email")}
                                className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3.5 py-2.5 text-sm text-white placeholder-slate-500 outline-none transition focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                            />
                            {errors.email && (
                                <p className="mt-1 text-xs text-red-400">{errors.email.message}</p>
                            )}
                        </div>

                        {/* Password */}
                        <div>
                            <label className="block text-sm font-medium text-slate-300 mb-1.5" htmlFor="password">
                                Password
                            </label>
                            <div className="relative">
                                <input
                                    id="password"
                                    type={showPassword ? "text" : "password"}
                                    autoComplete="current-password"
                                    placeholder="••••••••"
                                    {...register("password")}
                                    className="w-full rounded-lg border border-slate-700 bg-slate-800 px-3.5 py-2.5 pr-10 text-sm text-white placeholder-slate-500 outline-none transition focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500"
                                />
                                <button
                                    type="button"
                                    onClick={() => setShowPassword((v) => !v)}
                                    className="absolute inset-y-0 right-3 flex items-center text-slate-500 hover:text-slate-300 transition-colors"
                                    aria-label={showPassword ? "Hide password" : "Show password"}
                                >
                                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                                </button>
                            </div>
                            {errors.password && (
                                <p className="mt-1 text-xs text-red-400">{errors.password.message}</p>
                            )}
                        </div>

                        {/* Submit */}
                        <button
                            id="login-submit"
                            type="submit"
                            disabled={isSubmitting}
                            className="w-full rounded-lg bg-indigo-600 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-indigo-500 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 focus:ring-offset-slate-900 disabled:opacity-60 disabled:cursor-not-allowed"
                        >
                            {isSubmitting ? "Signing in…" : "Sign in"}
                        </button>
                    </form>
                </div>

                <p className="text-center text-xs text-slate-600">
                    Default admin: admin@nexus.local / Admin1234!
                </p>
            </div>
        </div>
    );
}
