// Dashboard.jsx - Main search dashboard page
// Uses Redux for state management

import { useEffect } from 'react';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import SearchBar from '../components/SearchBar';
import FilterBar from '../components/FilterBar';
import ContentList from '../components/ContentList';
import StatsPanel from '../components/StatsPanel';
import {
  setQuery,
  setFilters,
  setPage,
  performSearch,
} from '../store/slices/searchSlice';
import { fetchProviders } from '../store/slices/providersSlice';
import { fetchStats, toggleStats } from '../store/slices/statsSlice';

const Dashboard = () => {
  const dispatch = useAppDispatch();

  // Redux state
  const { query, results, loading, error, pagination, filters } = useAppSelector(
    (state) => state.search
  );
  const { providers } = useAppSelector((state) => state.providers);
  const { stats, showStats } = useAppSelector((state) => state.stats);

  // Load providers and stats on mount
  useEffect(() => {
    dispatch(fetchProviders());
    dispatch(fetchStats());
  }, [dispatch]);

  // Perform search when query or filters change
  // On initial load or when filters change, fetch all content if no query
  useEffect(() => {
    const searchParams = {
      ...filters,
    };
    
    // Add query only if it's not empty
    if (query.trim()) {
      searchParams.query = query.trim();
    }
    
    dispatch(performSearch(searchParams));
  }, [dispatch, query, filters]);

  const handleSearch = (newQuery) => {
    dispatch(setQuery(newQuery));
  };

  const handleFilterChange = (newFilters) => {
    dispatch(setFilters(newFilters));
  };

  const handlePageChange = (newPage) => {
    dispatch(setPage(newPage));
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-3xl font-bold text-gray-900 mb-2">
                Search Engine Dashboard
              </h1>
              <p className="text-gray-600">
                Search and discover content from multiple providers
              </p>
            </div>
            {stats && (
              <button
                onClick={() => dispatch(toggleStats())}
                className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm font-medium"
              >
                {showStats ? 'Hide' : 'Show'} Stats
              </button>
            )}
          </div>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Stats Panel */}
        {showStats && stats && (
          <div className="mb-6">
            <StatsPanel stats={stats} />
          </div>
        )}

        {/* Search Bar */}
        <div className="mb-6">
          <SearchBar onSearch={handleSearch} initialQuery={query} />
        </div>

        {/* Filters */}
        <FilterBar
          filters={filters}
          onFilterChange={handleFilterChange}
          providers={providers}
        />

        {/* Results */}
        <ContentList
          results={results}
          loading={loading}
          error={error}
          pagination={pagination}
          onPageChange={handlePageChange}
        />
      </main>
    </div>
  );
};

export default Dashboard;
