import { renderHook } from "@testing-library/react";
import { useAuthStore } from "@/store/authStore";

// Reset store between tests
afterEach(() => {
    useAuthStore.setState({ user: null, accessToken: null, isHydrated: false });
});

describe("authStore", () => {
    it("starts with null user and token", () => {
        const { result } = renderHook(() => useAuthStore());
        expect(result.current.user).toBeNull();
        expect(result.current.accessToken).toBeNull();
    });

    it("setAuth stores user and token", () => {
        const { result } = renderHook(() => useAuthStore());
        const mockUser = {
            user_id: "u1",
            username: "admin",
            email: "admin@nexus.local",
            role: "admin" as const,
        };

        result.current.setAuth(mockUser, "tok_test");
        expect(useAuthStore.getState().user?.username).toBe("admin");
        expect(useAuthStore.getState().user?.role).toBe("admin");
        expect(useAuthStore.getState().accessToken).toBe("tok_test");
    });

    it("clearAuth resets user and token", () => {
        const store = useAuthStore.getState();
        store.setAuth(
            { user_id: "u1", username: "a", email: "a@b.com", role: "reader" },
            "tok"
        );
        store.clearAuth();
        expect(useAuthStore.getState().user).toBeNull();
        expect(useAuthStore.getState().accessToken).toBeNull();
    });
});
