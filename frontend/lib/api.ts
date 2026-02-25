import axios, { AxiosError, InternalAxiosRequestConfig } from "axios";

// Backend URL - proxied through Next.js rewrites in production
const BACKEND_BASE = process.env.NEXT_PUBLIC_BACKEND_URL || "http://localhost:8080";

export const apiClient = axios.create({
    baseURL: `${BACKEND_BASE}/api/v1`,
    headers: { "Content-Type": "application/json" },
    withCredentials: true, // send cookies (httpOnly tokens)
    timeout: 30_000,
});

// ── Request interceptor ───────────────────────────────────────────────────────
// Attach access token from localStorage as a fallback for non-cookie clients.
apiClient.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    if (typeof window !== "undefined") {
        const token = localStorage.getItem("nexus_access_token");
        if (token && !config.headers.Authorization) {
            config.headers.Authorization = `Bearer ${token}`;
        }
    }
    return config;
});

// ── Response interceptor ─────────────────────────────────────────────────────
// On 401: attempt a single token refresh then retry the original request.
let isRefreshing = false;
let failedQueue: Array<{
    resolve: (token: string) => void;
    reject: (err: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null = null) {
    failedQueue.forEach((prom) => {
        if (error) prom.reject(error);
        else prom.resolve(token!);
    });
    failedQueue = [];
}

apiClient.interceptors.response.use(
    (response) => response,
    async (error: AxiosError) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & {
            _retry?: boolean;
        };

        if (error.response?.status === 401 && !originalRequest._retry) {
            if (isRefreshing) {
                return new Promise((resolve, reject) => {
                    failedQueue.push({ resolve, reject });
                }).then((token) => {
                    originalRequest.headers.Authorization = `Bearer ${token}`;
                    return apiClient(originalRequest);
                });
            }

            originalRequest._retry = true;
            isRefreshing = true;

            try {
                const { data } = await axios.post(
                    `${BACKEND_BASE}/api/v1/auth/refresh`,
                    {},
                    { withCredentials: true }
                );
                const newToken: string = data?.data?.access_token;
                if (newToken && typeof window !== "undefined") {
                    localStorage.setItem("nexus_access_token", newToken);
                }
                processQueue(null, newToken);
                originalRequest.headers.Authorization = `Bearer ${newToken}`;
                return apiClient(originalRequest);
            } catch (err) {
                processQueue(err);
                if (typeof window !== "undefined") {
                    localStorage.removeItem("nexus_access_token");
                    window.location.href = "/login";
                }
                return Promise.reject(err);
            } finally {
                isRefreshing = false;
            }
        }

        return Promise.reject(error);
    }
);

export default apiClient;
