// TypeScript interfaces matching backend Go structs

export type Role = "admin" | "contributor" | "reader";

export interface User {
    user_id: string;
    username: string;
    email: string;
    role: Role;
}

export interface UserRecord {
    id: string;
    person_id?: string;
    username: string;
    email: string;
    role: Role;
    is_active: boolean;
    last_login_at?: string;
    created_at: string;
    updated_at: string;
}

export interface AuthTokens {
    access_token: string;
    refresh_token: string;
    role: Role;
    user_id: string;
    username: string;
    email: string;
    expires_at: string;
}

export interface ApiResponse<T> {
    success: boolean;
    data: T;
    error: string;
    meta?: {
        page: number;
        limit: number;
        total: number;
    };
}

export interface LoginRequest {
    email: string;
    password: string;
}

// Lookup table types
export interface SoftwareCategory {
    id: string;
    name: string;
    description?: string;
    color?: string;
    is_active: boolean;
}

export interface SoftwareType {
    id: string;
    name: string;
    description?: string;
    icon?: string;
    is_active: boolean;
}

export interface Environment {
    id: string;
    name: string;
    display_name: string;
    color?: string;
    order_index: number;
    is_active: boolean;
}

export interface RepoPlatform {
    id: string;
    name: string;
    display_name: string;
    icon?: string;
    is_active: boolean;
}

export interface ResponsibilityRole {
    id: string;
    name: string;
    description?: string;
    is_active: boolean;
}

export interface TechCategory {
    id: string;
    name: string;
}

// People
export interface Person {
    id: string;
    full_name: string;
    email: string;
    phone?: string;
    title?: string;
    department?: string;
    team_id?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
    deleted_at?: string;
}

export interface Team {
    id: string;
    name: string;
    department: string;
    team_email?: string;
    manager_id?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

// Server
export interface Server {
    id: string;
    name: string;
    hostname?: string;
    ip_address?: string;
    server_type: string;
    os?: string;
    os_version?: string;
    cpu_cores?: number;
    ram_gb?: number;
    disk_gb?: number;
    datacenter?: string;
    region?: string;
    cloud_provider?: string;
    environment_id?: string;
    owner_team_id?: string;
    managed_by?: string;
    status: string;
    tags?: string[];
    notes?: string;
    created_at: string;
    updated_at: string;
}

// Vendor
export interface Vendor {
    id: string;
    name: string;
    country?: string;
    vendor_type?: string;
    website?: string;
    primary_contact_name?: string;
    primary_contact_email?: string;
    primary_contact_phone?: string;
    support_email?: string;
    support_phone?: string;
    notes?: string;
    tags?: string[];
    status: string;
    created_at: string;
    updated_at: string;
}

export interface VendorContract {
    id: string;
    vendor_id: string;
    contract_ref?: string;
    scope?: string;
    start_date?: string;
    end_date?: string;
    sla_terms?: string;
    auto_renew: boolean;
    value_amount?: number;
    value_currency: string;
    document_url?: string;
    notes?: string;
    created_at: string;
}

// Software
export interface Software {
    id: string;
    name: string;
    display_name: string;
    description?: string;
    version?: string;
    software_kind: "inhouse" | "vendor";
    category_id?: string;
    type_id?: string;
    criticality: "critical" | "high" | "medium" | "low";
    status: string;
    architecture?: string;
    owner_team_id?: string;
    tags?: string[];
    notes?: string;
    created_at: string;
    updated_at: string;
}

// Database instance
export interface DatabaseInstance {
    id: string;
    name: string;
    db_type: string;
    version?: string;
    host_server_id?: string;
    environment_id?: string;
    port?: number;
    size_gb?: number;
    owner_team_id?: string;
    classification?: string;
    backup_policy?: string;
    status: string;
    notes?: string;
    created_at: string;
    updated_at: string;
}

// Integration
export interface IntegrationExposed {
    id: string;
    software_id: string;
    name: string;
    description?: string;
    protocol: string;
    endpoint_url?: string;
    port?: number;
    auth_method?: string;
    version?: string;
    spec_url?: string;
    sla_uptime_percent?: number;
    owner_team_id?: string;
    status: string;
    deprecation_date?: string;
    notes?: string;
    created_at: string;
    updated_at: string;
}

export interface IntegrationConsumed {
    id: string;
    software_id: string;
    name: string;
    description?: string;
    protocol: string;
    vendor_id?: string;
    endpoint_url?: string;
    auth_method?: string;
    dependency_criticality?: string;
    fallback_strategy?: string;
    status: string;
    notes?: string;
    created_at: string;
    updated_at: string;
}

// Audit log
export interface AuditLog {
    id: string;
    user_id?: string;
    entity_type: string;
    entity_id: string;
    action: string;
    changes?: Record<string, { old: unknown; new: unknown }>;
    ip_address?: string;
    user_agent?: string;
    created_at: string;
}
