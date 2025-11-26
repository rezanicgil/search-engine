// searchSlice.js - Redux slice for search functionality
// Manages search state, results, pagination, and filters

import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { searchContent } from '../../services/api';
import { DEFAULT_PAGE_SIZE, SORT_OPTIONS, SORT_ORDERS } from '../../utils/constants';

// Initial state
const initialState = {
  query: '',
  results: [],
  loading: false,
  error: null,
  pagination: null,
  filters: {
    type: null,
    provider_id: null,
    start_date: null,
    end_date: null,
    sort_by: SORT_OPTIONS.SCORE,
    sort_order: SORT_ORDERS.DESC,
    page: 1,
    per_page: DEFAULT_PAGE_SIZE,
  },
};

// Async thunk for searching content
export const performSearch = createAsyncThunk(
  'search/performSearch',
  async (params, { rejectWithValue }) => {
    try {
      // Clean up params - remove null/empty values
      const cleanParams = { ...params };
      Object.keys(cleanParams).forEach((key) => {
        if (cleanParams[key] === null || cleanParams[key] === '') {
          delete cleanParams[key];
        }
      });

      // Allow empty query to fetch all content
      // If query is empty, remove it from params to fetch all
      if (!cleanParams.query || !cleanParams.query.trim()) {
        delete cleanParams.query;
      }

      const response = await searchContent(cleanParams);
      return {
        results: response.results || [],
        total: response.total || 0,
        page: response.page || 1,
        per_page: response.per_page || DEFAULT_PAGE_SIZE,
        total_pages: response.total_pages || 0,
      };
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.message ||
        error.message ||
        'Failed to search content'
      );
    }
  }
);

// Search slice
const searchSlice = createSlice({
  name: 'search',
  initialState,
  reducers: {
    setQuery: (state, action) => {
      state.query = action.payload;
      state.filters.page = 1; // Reset to first page on new query
    },
    setFilters: (state, action) => {
      state.filters = { ...state.filters, ...action.payload, page: 1 };
    },
    setPage: (state, action) => {
      state.filters.page = action.payload;
    },
    resetSearch: (state) => {
      state.query = '';
      state.results = [];
      state.pagination = null;
      state.error = null;
      state.filters = initialState.filters;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Search pending
      .addCase(performSearch.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      // Search fulfilled
      .addCase(performSearch.fulfilled, (state, action) => {
        state.loading = false;
        state.results = action.payload.results;
        state.pagination = {
          total: action.payload.total,
          page: action.payload.page,
          per_page: action.payload.per_page,
          total_pages: action.payload.total_pages,
        };
      })
      // Search rejected
      .addCase(performSearch.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload;
        state.results = [];
        state.pagination = null;
      });
  },
});

export const { setQuery, setFilters, setPage, resetSearch, clearError } = searchSlice.actions;
export default searchSlice.reducer;

