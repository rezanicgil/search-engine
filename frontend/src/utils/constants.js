// constants.js - Application constants
// Centralizes configuration values and constants

export const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1';

export const CONTENT_TYPES = {
  VIDEO: 'video',
  ARTICLE: 'article',
};

export const SORT_OPTIONS = {
  SCORE: 'score',
  PUBLISHED_AT: 'published_at',
  TITLE: 'title',
};

export const SORT_ORDERS = {
  ASC: 'asc',
  DESC: 'desc',
};

export const DEFAULT_PAGE_SIZE = 10;
export const MAX_PAGE_SIZE = 100;
