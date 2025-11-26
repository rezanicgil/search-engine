// contentSlice.js - Redux slice for content detail
// Manages individual content item state

import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { getContentById } from '../../services/api';

// Initial state
const initialState = {
  currentContent: null,
  loading: false,
  error: null,
};

// Async thunk for fetching content by ID
export const fetchContentById = createAsyncThunk(
  'content/fetchContentById',
  async (id, { rejectWithValue }) => {
    try {
      const content = await getContentById(id);
      return content;
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.message ||
        error.message ||
        'Failed to fetch content'
      );
    }
  }
);

// Content slice
const contentSlice = createSlice({
  name: 'content',
  initialState,
  reducers: {
    clearContent: (state) => {
      state.currentContent = null;
      state.error = null;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch content pending
      .addCase(fetchContentById.pending, (state) => {
        state.loading = true;
        state.error = null;
        state.currentContent = null;
      })
      // Fetch content fulfilled
      .addCase(fetchContentById.fulfilled, (state, action) => {
        state.loading = false;
        state.currentContent = action.payload;
      })
      // Fetch content rejected
      .addCase(fetchContentById.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload;
        state.currentContent = null;
      });
  },
});

export const { clearContent, clearError } = contentSlice.actions;
export default contentSlice.reducer;

