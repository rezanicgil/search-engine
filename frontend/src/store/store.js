// store.js - Redux store configuration
// Centralized state management for the application

import { configureStore } from '@reduxjs/toolkit';
import searchReducer from './slices/searchSlice';
import providersReducer from './slices/providersSlice';
import statsReducer from './slices/statsSlice';
import contentReducer from './slices/contentSlice';

export const store = configureStore({
  reducer: {
    search: searchReducer,
    providers: providersReducer,
    stats: statsReducer,
    content: contentReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      serializableCheck: {
        // Ignore these action types
        ignoredActions: ['search/setResults'],
        // Ignore these field paths in all actions
        ignoredActionPaths: ['meta.arg', 'payload.timestamp'],
        // Ignore these paths in the state
        ignoredPaths: ['search.results'],
      },
    }),
});

// Type exports removed - JavaScript project
// If using TypeScript, uncomment:
// export type RootState = ReturnType<typeof store.getState>;
// export type AppDispatch = typeof store.dispatch;

