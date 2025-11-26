// statsSlice.js - Redux slice for statistics
// Manages system statistics state

import { createSlice, createAsyncThunk } from '@reduxjs/toolkit';
import { getStats } from '../../services/api';

// Initial state
const initialState = {
  stats: null,
  loading: false,
  error: null,
  showStats: false,
};

// Async thunk for fetching stats
export const fetchStats = createAsyncThunk(
  'stats/fetchStats',
  async (_, { rejectWithValue }) => {
    try {
      const stats = await getStats();
      return stats;
    } catch (error) {
      return rejectWithValue(
        error.response?.data?.message ||
        error.message ||
        'Failed to fetch statistics'
      );
    }
  }
);

// Stats slice
const statsSlice = createSlice({
  name: 'stats',
  initialState,
  reducers: {
    toggleStats: (state) => {
      state.showStats = !state.showStats;
    },
    setShowStats: (state, action) => {
      state.showStats = action.payload;
    },
    clearError: (state) => {
      state.error = null;
    },
  },
  extraReducers: (builder) => {
    builder
      // Fetch stats pending
      .addCase(fetchStats.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      // Fetch stats fulfilled
      .addCase(fetchStats.fulfilled, (state, action) => {
        state.loading = false;
        state.stats = action.payload;
      })
      // Fetch stats rejected
      .addCase(fetchStats.rejected, (state, action) => {
        state.loading = false;
        state.error = action.payload;
      });
  },
});

export const { toggleStats, setShowStats, clearError } = statsSlice.actions;
export default statsSlice.reducer;

