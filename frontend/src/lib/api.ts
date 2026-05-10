/**
 * API Client for Digital Papyrus Backend
 * Handles all HTTP communication with the Go/Gin REST API.
 */

const API_BASE_URL = import.meta.env.PUBLIC_API_URL || 'https://api.digitalpapyrus.web.id';

// ─── Types ───────────────────────────────────────────────────────────

export interface Book {
  id: string;
  title: string;
  author: string;
  isbn: string;
  badge: string;
  ggkey: string;
  qrcbn: string;
  price: number;
  rating: number;
  review_count: number;
  description: string;
  synopsis: string;
  image_url: string;
  category_id: string;
  category_name?: string;
  category_slug?: string;
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

export interface CoreService {
  id: string;
  title: string;
  description: string;
  icon: string;
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
  phone_number?: string;
  address?: string;
  province?: string;
  city?: string;
  zip_code?: string;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: string;
  invoice: string;
  user_id: string;
  user_name: string;
  items?: OrderItem[];
  notes: string;
  total_qty: number;
  total_weight: number;
  total_price: number;
  payment_type: string;
  status: string;
  shipping_name: string;
  shipping_service: string;
  shipping_price: number;
  created_at: string;
  updated_at: string;
}

export interface OrderItem {
  id: string;
  order_id: string;
  service_id?: string;
  book_id?: string;
  item_name: string;
  qty: number;
  total_price: number;
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
  badge?: string;
  search?: string;
  min_price?: number;
  max_price?: number;
  max_rating?: number;
  sort?: 'newest' | 'popular';
}

export interface OrderFilter {
  page?: number;
  per_page?: number;
  status?: string;
  search?: string;
}

export interface CreateOrderPayload {
  invoice?: string;
  user_id: string;
  items?: string;
  notes?: string;
  total_qty: number;
  total_weight?: number;
  total_price: number;
  payment_type?: string;
  status: string;
  shipping_name?: string;
  shipping_service?: string;
  shipping_price?: number;
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

/** Build URL-safe category slugs (e.g. "Seni & Desain" -> "seni-desain"). */
export function slugifyCategory(value: string): string {
  const slug = value
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '');

  return slug || 'category';
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

export async function register(payload: any): Promise<LoginResult> {
  const res = await apiFetch<LoginResult>('/api/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(payload),
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

export async function sendOTP(email: string): Promise<void> {
  await apiFetch('/api/v1/auth/send-otp', {
    method: 'POST',
    body: JSON.stringify({ email }),
  });
}

export async function verifyOTP(email: string, code: string): Promise<void> {
  await apiFetch('/api/v1/auth/verify-otp', {
    method: 'POST',
    body: JSON.stringify({ email, code }),
  });
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
  if (filter.badge) params.set('badge', filter.badge);
  if (filter.search) params.set('search', filter.search);
  if (filter.min_price) params.set('min_price', String(filter.min_price));
  if (filter.max_price) params.set('max_price', String(filter.max_price));
  if (filter.max_rating) params.set('max_rating', String(filter.max_rating));
  if (filter.sort) params.set('sort', filter.sort);

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

export async function getOrders(filter: OrderFilter = {}): Promise<{ orders: Order[]; meta: PaginationMeta }> {
  const params = new URLSearchParams();
  if (filter.page) params.set('page', String(filter.page));
  if (filter.per_page) params.set('per_page', String(filter.per_page));
  if (filter.status) params.set('status', filter.status);
  if (filter.search) params.set('search', filter.search);

  const query = params.toString();
  const res = await apiFetch<Order[]>(`/api/v1/orders${query ? '?' + query : ''}`);
  return {
    orders: res.data || [],
    meta: res.meta || { page: 1, per_page: 10, total: 0, total_pages: 0 },
  };
}

export async function createOrder(payload: CreateOrderPayload): Promise<Order> {
  const res = await apiFetch<Order>('/api/v1/orders', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function getOrder(id: string): Promise<Order> {
  const res = await apiFetch<Order>(`/api/v1/orders/${id}`);
  return res.data;
}

export async function deleteOrder(id: string): Promise<void> {
  await apiFetch(`/api/v1/orders/${id}`, { method: 'DELETE' });
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

export async function getServices(activeOnly = true): Promise<Service[]> {
  const res = await apiFetch<Service[]>(`/api/v1/services?active_only=${activeOnly}`);
  return res.data || [];
}

export async function getService(id: string): Promise<Service> {
  const res = await apiFetch<Service>(`/api/v1/services/${id}`);
  return res.data;
}

export async function createService(payload: Omit<Service, 'id' | 'created_at' | 'updated_at'>): Promise<Service> {
  const res = await apiFetch<Service>('/api/v1/services', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function updateService(id: string, payload: Partial<Omit<Service, 'id' | 'created_at' | 'updated_at'>>): Promise<Service> {
  const res = await apiFetch<Service>(`/api/v1/services/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function deleteService(id: string): Promise<void> {
  await apiFetch(`/api/v1/services/${id}`, { method: 'DELETE' });
}

// ─── Core Services (Layanan Kami) API ────────────────────────────────

export async function getCoreServices(): Promise<CoreService[]> {
  const res = await apiFetch<CoreService[]>('/api/v1/core-services');
  return res.data;
}

export async function createCoreService(data: Partial<CoreService>): Promise<CoreService> {
  const res = await apiFetch<CoreService>('/api/v1/core-services', {
    method: 'POST',
    body: JSON.stringify(data),
  });
  return res.data;
}

export async function updateCoreService(id: string, data: Partial<CoreService>): Promise<CoreService> {
  const res = await apiFetch<CoreService>(`/api/v1/core-services/${id}`, {
    method: 'PUT',
    body: JSON.stringify(data),
  });
  return res.data;
}

export async function deleteCoreService(id: string): Promise<void> {
  await apiFetch(`/api/v1/core-services/${id}`, { method: 'DELETE' });
}

// --- Categories API ------------------------------------------

export interface Category {
  id: string;
  name: string;
  slug?: string;
  description?: string;
  created_at: string;
  updated_at: string;
}

export async function getCategories(): Promise<Category[]> {
  const res = await apiFetch<Category[]>('/api/v1/categories');
  return res.data || [];
}

export async function getCategory(id: string): Promise<Category> {
  const res = await apiFetch<Category>(`/api/v1/categories/${id}`);
  return res.data;
}

export async function createCategory(payload: Pick<Category, 'name'>): Promise<Category> {
  const res = await apiFetch<Category>('/api/v1/categories', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function updateCategory(id: string, payload: Partial<Pick<Category, 'name'>>): Promise<Category> {
  const res = await apiFetch<Category>(`/api/v1/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function deleteCategory(id: string): Promise<void> {
  await apiFetch(`/api/v1/categories/${id}`, { method: 'DELETE' });
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

// --- Users API ----------------------------------------------

export async function getUsers(): Promise<User[]> {
  const res = await apiFetch<User[]>('/api/v1/users');
  return res.data || [];
}

export async function createUser(payload: Partial<User>): Promise<User> {
  const res = await apiFetch<User>('/api/v1/users', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function updateUser(id: string, payload: Partial<User>): Promise<User> {
  const res = await apiFetch<User>(`/api/v1/users/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function deleteUser(id: string): Promise<void> {
  await apiFetch(`/api/v1/users/${id}`, { method: 'DELETE' });
}

// --- Reviews API ----------------------------------------------

export interface Review {
  id: string;
  user_id: string;
  order_id: string;
  invoice?: string;
  service_id: string[];
  book_id: string[];
  details: Record<string, string>;
  rating: Record<string, number>;
  created_at: string;
  updated_at: string;
}

export async function getReviews(page = 1, limit = 10): Promise<{ reviews: Review[]; meta: PaginationMeta }> {
  const res = await apiFetch<Review[]>(`/api/v1/reviews?page=${page}&per_page=${limit}`);
  return {
    reviews: res.data || [],
    meta: res.meta || { page, per_page: limit, total: 0, total_pages: 0 },
  };
}

export async function getReview(id: string): Promise<Review> {
  const res = await apiFetch<Review>(`/api/v1/reviews/${id}`);
  return res.data;
}

export async function createReview(payload: Partial<Review>): Promise<Review> {
  const res = await apiFetch<Review>('/api/v1/reviews', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function updateReview(id: string, payload: Partial<Review>): Promise<Review> {
  const res = await apiFetch<Review>(`/api/v1/reviews/${id}`, {
    method: 'PUT',
    body: JSON.stringify(payload),
  });
  return res.data;
}

export async function deleteReview(id: string): Promise<void> {
  await apiFetch(`/api/v1/reviews/${id}`, { method: 'DELETE' });
}
