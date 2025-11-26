// StatsPanel.jsx - Statistics display component
// Shows system statistics including content counts, provider info, etc.

import { formatNumber } from '../utils/helpers';

const StatsPanel = ({ stats }) => {
  if (!stats) return null;

  return (
    <div className="bg-white rounded-lg shadow-md border border-gray-200 p-6">
      <h2 className="text-xl font-semibold text-gray-900 mb-4">System Statistics</h2>
      
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6">
        {/* Total Content */}
        <div className="bg-blue-50 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Total Content</div>
          <div className="text-2xl font-bold text-blue-600">
            {formatNumber(stats.total_content || 0)}
          </div>
        </div>

        {/* Videos */}
        <div className="bg-purple-50 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Videos</div>
          <div className="text-2xl font-bold text-purple-600">
            {formatNumber(stats.videos || 0)}
          </div>
        </div>

        {/* Articles */}
        <div className="bg-green-50 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Articles</div>
          <div className="text-2xl font-bold text-green-600">
            {formatNumber(stats.articles || 0)}
          </div>
        </div>

        {/* Average Score */}
        <div className="bg-yellow-50 rounded-lg p-4">
          <div className="text-sm text-gray-600 mb-1">Avg Score</div>
          <div className="text-2xl font-bold text-yellow-600">
            {(stats.average_score || 0).toFixed(2)}
          </div>
        </div>
      </div>

      {/* Providers Info */}
      {stats.providers && (
        <div className="mb-4">
          <h3 className="text-sm font-medium text-gray-700 mb-2">Providers</h3>
          <div className="text-lg font-semibold text-gray-900">
            {stats.providers.total || 0} active provider{stats.providers.total !== 1 ? 's' : ''}
          </div>
        </div>
      )}

      {/* Content by Provider */}
      {stats.by_provider && stats.by_provider.length > 0 && (
        <div>
          <h3 className="text-sm font-medium text-gray-700 mb-2">Content by Provider</h3>
          <div className="space-y-2">
            {stats.by_provider.map((item) => (
              <div key={item.provider_id} className="flex items-center justify-between text-sm">
                <span className="text-gray-600">Provider {item.provider_id}</span>
                <span className="font-semibold text-gray-900">{formatNumber(item.count)} items</span>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Total Tags */}
      {stats.total_tags !== undefined && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          <div className="text-sm text-gray-600">Total Tags</div>
          <div className="text-lg font-semibold text-gray-900">
            {formatNumber(stats.total_tags || 0)}
          </div>
        </div>
      )}
    </div>
  );
};

export default StatsPanel;

