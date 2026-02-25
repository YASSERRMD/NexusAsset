import apiClient from "@/lib/api";
import type { ApiResponse, Person, Team, Server, SoftwareCategory, SoftwareType, Environment, ResponsibilityRole, RepoPlatform, TechCategory } from "@/types";

// ─── People ───────────────────────────────────────────────────────────────────

export const personsApi = {
    list: (params?: Record<string, string | number>) =>
        apiClient.get<ApiResponse<Person[]>>("/persons", { params }).then(r => r.data),
    getById: (id: string) =>
        apiClient.get<ApiResponse<Person>>(`/persons/${id}`).then(r => r.data),
    create: (data: Partial<Person>) =>
        apiClient.post<ApiResponse<Person>>("/persons", data).then(r => r.data),
    update: (id: string, data: Partial<Person>) =>
        apiClient.put<ApiResponse<Person>>(`/persons/${id}`, data).then(r => r.data),
    delete: (id: string) =>
        apiClient.delete(`/persons/${id}`).then(r => r.data),
};

export const teamsApi = {
    list: (params?: Record<string, string | number>) =>
        apiClient.get<ApiResponse<Team[]>>("/teams", { params }).then(r => r.data),
    getById: (id: string) =>
        apiClient.get<ApiResponse<Team>>(`/teams/${id}`).then(r => r.data),
    create: (data: Partial<Team>) =>
        apiClient.post<ApiResponse<Team>>("/teams", data).then(r => r.data),
    update: (id: string, data: Partial<Team>) =>
        apiClient.put<ApiResponse<Team>>(`/teams/${id}`, data).then(r => r.data),
    delete: (id: string) =>
        apiClient.delete(`/teams/${id}`).then(r => r.data),
};

// ─── Servers ──────────────────────────────────────────────────────────────────

export const serversApi = {
    list: (params?: Record<string, string | number>) =>
        apiClient.get<ApiResponse<Server[]>>("/servers", { params }).then(r => r.data),
    getById: (id: string) =>
        apiClient.get<ApiResponse<Server>>(`/servers/${id}`).then(r => r.data),
    create: (data: Partial<Server>) =>
        apiClient.post<ApiResponse<Server>>("/servers", data).then(r => r.data),
    update: (id: string, data: Partial<Server>) =>
        apiClient.put<ApiResponse<Server>>(`/servers/${id}`, data).then(r => r.data),
    delete: (id: string) =>
        apiClient.delete(`/servers/${id}`).then(r => r.data),
};

// ─── Lookups ──────────────────────────────────────────────────────────────────

type LookupTable =
    | "software-categories"
    | "software-types"
    | "environments"
    | "responsibility-roles"
    | "repo-platforms"
    | "tech-categories";

type LookupRow = SoftwareCategory | SoftwareType | Environment | ResponsibilityRole | RepoPlatform | TechCategory;

export const lookupsApi = {
    list: (table: LookupTable) =>
        apiClient.get<ApiResponse<LookupRow[]>>(`/lookups/${table}`).then(r => r.data),
    create: (table: LookupTable, data: { name: string; description?: string; color?: string; icon?: string }) =>
        apiClient.post<ApiResponse<{ id: string; name: string }>>(`/lookups/${table}`, data).then(r => r.data),
    update: (table: LookupTable, id: string, data: { name: string }) =>
        apiClient.put<ApiResponse<{ id: string; name: string }>>(`/lookups/${table}/${id}`, data).then(r => r.data),
    patchActive: (table: LookupTable, id: string, is_active: boolean) =>
        apiClient.patch<ApiResponse<{ is_active: boolean }>>(`/lookups/${table}/${id}/active`, { is_active }).then(r => r.data),
};
