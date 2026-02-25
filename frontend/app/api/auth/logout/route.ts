import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

/**
 * POST /api/auth/logout
 * Clears auth cookies and proxies logout to the backend.
 */
export async function POST(req: NextRequest) {
    const refreshToken = req.cookies.get("nexus_refresh_token")?.value;

    // Best-effort call to backend
    try {
        await fetch(`${BACKEND_URL}/api/v1/auth/logout`, {
            method: "POST",
            headers: {
                Cookie: `nexus_refresh_token=${refreshToken}`,
            },
        });
    } catch {
        // ignore
    }

    const response = NextResponse.json({ success: true });

    // Clear all auth cookies
    response.cookies.set("nexus_access_token", "", { maxAge: 0, path: "/" });
    response.cookies.set("nexus_refresh_token", "", {
        maxAge: 0,
        path: "/api/auth/refresh",
    });

    return response;
}
