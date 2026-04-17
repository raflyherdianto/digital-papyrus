/**
 * API Client for Digital Papyrus Backend
 * Handles all HTTP communication with the Go/Gin REST API.
 */

const API_BASE_URL = import.meta.env.PUBLIC_API_URL || 'http://localhost:8080';

// ─── Types ───────────────────────────────────────────────────────────

export interface Book {
  id: string;
  title: string;
  author: string;
  isbn: string;
  price: number;
  rating: number;
  review_count: number;
  description: string;
  synopsis: string;
  image_url: string;
  category_id: string;
  category_name?: string;
  status: 'draft' | 'published' | 'archived';
  stock: number;
  publisher: string;
  publication_date: string;
  pages: number;
  format: string;
  language: string;
  dimensions: string;
  weight: string;
  created_at: string;
  updated_at: string;
}

export interface Service {
  id: string;
  title: string;
  description: string;
  icon: string;
  tier: 'basic' | 'silver' | 'gold' | 'platinum';
  price: number;
  price_label: string;
  features: string; // JSON array string
  is_featured: boolean;
  badge: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface User {
  id: string;
  email: string;
  name: string;
  role: 'superadmin' | 'author' | 'customer';
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface APIResponse<T> {
  success: boolean;
  message: string;
  data: T;
  meta?: PaginationMeta;
  error?: { code: string; details?: Record<string, string> };
}

export interface LoginResult {
  token: string;
  user: User;
}

export interface BookFilter {
  page?: number;
  per_page?: number;
  status?: string;
  category_id?: string;
  search?: string;
}

// ─── Helper Functions ────────────────────────────────────────────────

/** Format price integer (Rupiah) to display string */
export function formatRupiah(amount: number): string {
  if (amount <= 0) return '--';
  return 'Rp ' + amount.toLocaleString('id-ID');
}

/** Parse features JSON string to array */
export function parseFeatures(features: string): string[] {
  try {
    return JSON.parse(features);
  } catch {
    return [];
  }
}

// ─── Core Fetch Wrapper ──────────────────────────────────────────────

async function apiFetch<T>(
  endpoint: string,
  options: RequestInit = {}
): Promise<APIResponse<T>> {
  const url = `${API_BASE_URL}${endpoint}`;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options.headers as Record<string, string> || {}),
  };

  // Attach JWT token if available
  const token = typeof localStorage !== 'undefined' ? localStorage.getItem('dp_token') : null;
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(url, { ...options, headers });
  const data: APIResponse<T> = await response.json();

  if (!response.ok || !data.success) {
    throw new APIError(data.message || 'Request failed', response.status, data.error);
  }

  return data;
}

export class APIError extends Error {
  status: number;
  error?: { code: string; details?: Record<string, string> };

  constructor(message: string, status: number, error?: any) {
    super(message);
    this.name = 'APIError';
    this.status = status;
    this.error = error;
  }
}

// ─── Auth API ────────────────────────────────────────────────────────

export async function login(email: string, password: string): Promise<LoginResult> {
  const res = await apiFetch<LoginResult>('/api/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password }),
  });
  return res.data;
}

export async function getCurrentUser(): Promise<User> {
  const res = await apiFetch<User>('/api/v1/auth/me');
  return res.data;
}

export async function logout(): Promise<void> {
  await apiFetch('/api/v1/auth/logout', { method: 'POST' });
}

// ─── Auth Helpers ────────────────────────────────────────────────────

export function setToken(token: string): void {
  localStorage.setItem('dp_token', token);
}

export function getToken(): string | null {
  return typeof localStorage !== 'undefined' ? localStorage.getItem('dp_token') : null;
}

export function removeToken(): void {
  localStorage.removeItem('dp_token');
}

export function isAuthenticated(): boolean {
  return !!getToken();
}

// ─── Books API ───────────────────────────────────────────────────────

export async function getBooks(filter: BookFilter = {}): Promise<{ books: Book[]; meta: PaginationMeta }> {
  const params = new URLSearchParams();
  if (filter.page) params.set('page', String(filter.page));
  if (filter.per_page) params.set('per_page', String(filter.per_page));
  if (filter.status) params.set('status', filter.status);
  if (filter.category_id) params.set('category_id', filter.category_id);
  if (filter.search) params.set('search', filter.search);

  const query = params.toString();
  const res = await apiFetch<Book[]>(`/api/v1/books${query ? '?' + query : ''}`);
  return {
    books: res.data || [],
    meta: res.meta || { page: 1, per_page: 12, total: 0, total_pages: 0 },
  };
}

export async function getBook(id: string): Promise<Book> {
  const res = await apiFetch<Book>(`/api/v1/books/${id}`);
  return res.data;
}

export async function createBook(book: Partial<Book>): Promise<Book> {
  const res = await apiFetch<Book>('/api/v1/books', {
    method: 'POST',
    body: JSON.stringify(book),
  });
  return res.data;
}

export async function updateBook(id: string, book: Partial<Book>): Promise<Book> {
  const res = await apiFetch<Book>(`/api/v1/books/${id}`, {
    method: 'PUT',
    body: JSON.stringify(book),
  });
  return res.data;
}

export async function deleteBook(id: string): Promise<void> {
  await apiFetch(`/api/v1/books/${id}`, { method: 'DELETE' });
}

// ─── Services API ────────────────────────────────────────────────────

export async function getServices(): Promise<Service[]> {
  const res = await apiFetch<Service[]>('/api/v1/services');
  return res.data || [];
}

export async function getService(id: string): Promise<Service> {
  const res = await apiFetch<Service>(`/api/v1/services/${id}`);
  return res.data;
}

// --- Categories API ------------------------------------------

export interface Category {
  id: string;
  name: string;
  slug: string;
  description: string;
  created_at: string;
  updated_at: string;
}

export async function getCategories(): Promise<Category[]> {
  const res = await apiFetch<Category[]>('/api/v1/categories');
  return res.data || [];
}

// --- Upload API ----------------------------------------------

export interface UploadResult {
  url: string;
  filename: string;
  size: number;
}

export async function uploadImage(file: File): Promise<UploadResult> {
  const formData = new FormData();
  formData.append('image', file);
  
  const token = getToken();
  const headers: Record<string, string> = {};
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}/api/v1/upload`, {
    method: 'POST',
    body: formData,
    headers,
  });

  const data = await response.json();
  if (!response.ok || !data.success) {
    throw new Error(data.message || 'Failed to upload image');
  }

  return data.data;
}
