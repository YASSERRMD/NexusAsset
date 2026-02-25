"use client";

import { create } from "zustand";
import { persist, createJSONStorage } from "zustand/middleware";
import type { Role, User } from "@/types";

interface AuthState {
    user: User | null;
    accessToken: string | null;
    setAuth: (user: User, accessToken: string) => void;
    clearAuth: () => void;
    isHydrated: boolean;
    setHydrated: () => void;
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set) => ({
            user: null,
            accessToken: null,
            isHydrated: false,
            setAuth: (user: User, accessToken: string) => {
                if (typeof window !== "undefined") {
                    localStorage.setItem("nexus_access_token", accessToken);
                }
                set({ user, accessToken });
            },
            clearAuth: () => {
                if (typeof window !== "undefined") {
                    localStorage.removeItem("nexus_access_token");
                }
                set({ user: null, accessToken: null });
            },
            setHydrated: () => set({ isHydrated: true }),
        }),
        {
            name: "nexus-auth",
            storage: createJSONStorage(() =>
                typeof window !== "undefined" ? localStorage : ({} as Storage)
            ),
            partialize: (state) => ({ user: state.user, accessToken: state.accessToken }),
            onRehydrateStorage: () => (state) => {
                state?.setHydrated();
            },
        }
    )
);

export const useRole = (): Role | null => {
    return useAuthStore((s) => s.user?.role ?? null);
};
