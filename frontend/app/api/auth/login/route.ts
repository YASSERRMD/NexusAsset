import { NextRequest, NextResponse } from "next/server";

const BACKEND_URL = process.env.BACKEND_URL || "http://localhost:8080";

/**
 * POST /api/auth/login
 *
 * Proxies the login request to the Go backend and sets httpOnly cookies
 * on the Next.js domain so they cannot be accessed by JavaScript.
 */
export async function POST(req: NextRequest) {
    const body = await req.json();

    const backendRes = await fetch(`${BACKEND_URL}/api/v1/auth/login`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
    });

    const data = await backendRes.json();

    if (!backendRes.ok || !data.success) {
        return NextResponse.json(data, { status: backendRes.status });
    }

    const response = NextResponse.json(data, { status: 200 });

    const { access_token, refresh_token, expires_at } = data.data;

    // Set httpOnly cookies on the Next.js domain
    response.cookies.set("nexus_access_token", access_token, {
        httpOnly: true,
        sameSite: "lax",
        path: "/",
        expires: new Date(expires_at),
        secure: process.env.NODE_ENV === "production",
    });

    response.cookies.set("nexus_refresh_token", refresh_token, {
        httpOnly: true,
        sameSite: "lax",
        path: "/api/auth/refresh",
        maxAge: 7 * 24 * 60 * 60, // 7 days
        secure: process.env.NODE_ENV === "production",
    });

    return response;
}
