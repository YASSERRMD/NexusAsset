"use client";

import { useAuth } from "@/hooks/useAuth";
import type { Role } from "@/types";

interface CanAccessProps {
    roles: Role[];
    children: React.ReactNode;
    fallback?: React.ReactNode;
}

/**
 * CanAccess conditionally renders children based on the current user's role.
 *
 * Usage:
 *   <CanAccess roles={["admin"]}>
 *     <DeleteButton />
 *   </CanAccess>
 *
 *   <CanAccess roles={["admin", "contributor"]} fallback={<span>Read-only</span>}>
 *     <EditButton />
 *   </CanAccess>
 */
export function CanAccess({ roles, children, fallback = null }: CanAccessProps) {
    const { hasRole, isHydrated } = useAuth();

    // While store is rehydrating from localStorage, render nothing
    if (!isHydrated) return null;

    if (hasRole(...roles)) {
        return <>{children}</>;
    }

    return <>{fallback}</>;
}
