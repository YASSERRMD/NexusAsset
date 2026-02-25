"use client";

import { useCallback } from "react";
import { useRouter } from "next/navigation";
import { useAuthStore } from "@/store/authStore";
import apiClient from "@/lib/api";
import type { Role } from "@/types";

export interface UseAuthReturn {
    user: ReturnType<typeof useAuthStore>["user"];
    role: Role | null;
    isAdmin: boolean;
    isContributor: boolean;
    isReader: boolean;
    isAuthenticated: boolean;
    isHydrated: boolean;
    logout: () => Promise<void>;
    hasRole: (...roles: Role[]) => boolean;
}

/**
 * useAuth - central auth hook.
 *
 * Exposes user, role flags, and logout function.
 * Wrap pages or components with role checks using:
 *   const { isAdmin } = useAuth();
 */
export function useAuth(): UseAuthReturn {
    const { user, clearAuth, isHydrated } = useAuthStore();
    const router = useRouter();
    const role = user?.role ?? null;

    const hasRole = useCallback(
        (...roles: Role[]) => {
            if (!role) return false;
            return roles.includes(role);
        },
        [role]
    );

    const logout = useCallback(async () => {
        try {
            await apiClient.post("/auth/logout");
        } catch {
            // Silently fail — we still clear local state
        }
        clearAuth();
        router.push("/login");
    }, [clearAuth, router]);

    return {
        user,
        role,
        isAdmin: role === "admin",
        isContributor: role === "contributor",
        isReader: role === "reader",
        isAuthenticated: !!user,
        isHydrated,
        logout,
        hasRole,
    };
}
