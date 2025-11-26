// providersSlice.js - Redux slice for providers
// Manages provider list state

import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { getProviders } from '../../services/api';

// Initial state
const initialState = {
  providers: [],
  loading: false,
  error: null,
};

// Async thunk for fetching providers
export const fetchProviders = createAsyncThunk(
  'providers/fetchProviders',
  async (_, { rejectWithValue }) => {
    try {
      const providers = await getProviders();
      return providers;
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.message ||
        error.message ||
        'Failed to fetch providers'
      );
    }
  }
);

// Providers slice
const providersSlice = createSlice({
  name: 'providers',
  initialState,
  reducers: {
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch providers pending
      .addCase(fetchProviders.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      // Fetch providers fulfilled
      .addCase(fetchProviders.fulfilled, (state, action) => {
        state.loading = false;
        state.providers = action.payload;
      })
      // Fetch providers rejected
      .addCase(fetchProviders.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload;
      });
  },
});

export const { clearError } = providersSlice.actions;
export default providersSlice.reducer;

