// api.js - API service layer
// Handles all HTTP requests to the backend API
// Centralizes API communication logic

import axios from 'axios';
import { API_BASE_URL } from '../utils/constants';

// Create axios instance with base configuration
const apiClient = axios.create({
  baseURL: API_BASE_URL,
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor for logging (optional)
apiClient.interceptors.request.use(
  (config) => {
    // You can add auth tokens here if needed
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    // Handle common errors
    if (error.response) {
      // Server responded with error status
      // Backend returns: { error: { code, message, details? }, trace_id }
      const errorData = error.response.data;
      if (errorData?.error) {
        // Extract error message from new format
        error.response.data.message = errorData.error.message || errorData.error.code || 'An error occurred';
        error.response.data.code = errorData.error.code;
      }
      console.error('API Error:', error.response.data);
    } else if (error.request) {
      // Request made but no response received
      console.error('Network Error:', error.message);
    } else {
      // Something else happened
      console.error('Error:', error.message);
    }
    return Promise.reject(error);
  }
);

/**
 * Search for content
 * @param {Object} params - Search parameters
 * @param {string} params.query - Search keyword (required)
 * @param {string} [params.type] - Content type filter (video/article)
 * @param {number} [params.provider_id] - Provider ID filter
 * @param {string} [params.start_date] - Start date (YYYY-MM-DD)
 * @param {string} [params.end_date] - End date (YYYY-MM-DD)
 * @param {number} [params.page] - Page number (default: 1)
 * @param {number} [params.per_page] - Items per page (default: 10)
 * @param {string} [params.sort_by] - Sort field (score/published_at/title)
 * @param {string} [params.sort_order] - Sort order (asc/desc)
 * @returns {Promise} API response with search results
 */
export const searchContent = async (params) => {
  try {
    const response = await apiClient.get('/search', { params });
    // Backend returns: { data: { results, total, ... }, trace_id }
    return response.data.data || response.data;
  } catch (error) {
    throw error;
  }
};

/**
 * Get all providers
 * @returns {Promise} List of providers
 */
export const getProviders = async () => {
  try {
    const response = await apiClient.get('/providers');
    return response.data.data || [];
  } catch (error) {
    throw error;
  }
};

/**
 * Get content by ID
 * @param {number} id - Content ID
 * @returns {Promise} Content details
 */
export const getContentById = async (id) => {
  try {
    const response = await apiClient.get(`/content/${id}`);
    return response.data.data;
  } catch (error) {
    throw error;
  }
};

/**
 * Get system statistics
 * @returns {Promise} System statistics
 */
export const getStats = async () => {
  try {
    const response = await apiClient.get('/stats');
    return response.data.data;
  } catch (error) {
    throw error;
  }
};

/**
 * Health check endpoint
 * @returns {Promise} Health status
 */
export const healthCheck = async () => {
  try {
    const response = await axios.get(`${API_BASE_URL.replace('/api/v1', '')}/health`);
    return response.data;
  } catch (error) {
    throw error;
  }
};

export default apiClient;
