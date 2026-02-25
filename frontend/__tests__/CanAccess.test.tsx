import React from "react";
import { render, screen } from "@testing-library/react";
import { CanAccess } from "@/components/layout/CanAccess";
import { useAuthStore } from "@/store/authStore";

// Mock useAuth hook  
jest.mock("@/hooks/useAuth", () => ({
    useAuth: jest.fn(),
}));

import { useAuth } from "@/hooks/useAuth";
const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

// Mock router
jest.mock("next/navigation", () => ({
    useRouter: () => ({ push: jest.fn(), replace: jest.fn() }),
    usePathname: () => "/",
}));

afterEach(() => {
    jest.clearAllMocks();
    useAuthStore.setState({ user: null, accessToken: null, isHydrated: true });
});

describe("CanAccess", () => {
    it("renders children when user has matching role", () => {
        mockUseAuth.mockReturnValue({
            user: { user_id: "u1", username: "admin", email: "a@b.com", role: "admin" },
            role: "admin",
            isAdmin: true,
            isContributor: false,
            isReader: false,
            isAuthenticated: true,
            isHydrated: true,
            logout: jest.fn(),
            hasRole: (...roles) => roles.includes("admin"),
        });

        render(
            <CanAccess roles={["admin"]}>
                <button>Delete</button>
            </CanAccess>
        );

        expect(screen.getByRole("button", { name: "Delete" })).toBeInTheDocument();
    });

    it("renders fallback when user lacks required role", () => {
        mockUseAuth.mockReturnValue({
            user: { user_id: "u2", username: "reader", email: "r@b.com", role: "reader" },
            role: "reader",
            isAdmin: false,
            isContributor: false,
            isReader: true,
            isAuthenticated: true,
            isHydrated: true,
            logout: jest.fn(),
            hasRole: (...roles) => roles.includes("reader"),
        });

        render(
            <CanAccess roles={["admin"]} fallback={<span>No access</span>}>
                <button>Delete</button>
            </CanAccess>
        );

        expect(screen.queryByRole("button")).toBeNull();
        expect(screen.getByText("No access")).toBeInTheDocument();
    });

    it("renders nothing while store is not hydrated", () => {
        mockUseAuth.mockReturnValue({
            user: null,
            role: null,
            isAdmin: false,
            isContributor: false,
            isReader: false,
            isAuthenticated: false,
            isHydrated: false,
            logout: jest.fn(),
            hasRole: () => false,
        });

        const { container } = render(
            <CanAccess roles={["admin"]}>
                <button>Delete</button>
            </CanAccess>
        );

        expect(container.firstChild).toBeNull();
    });
});
