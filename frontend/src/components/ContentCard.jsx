// ContentCard.jsx - Individual content item card
// Displays content information in a card format

import { formatDate, formatNumber, formatScore, truncateText } from '../utils/helpers';
import { CONTENT_TYPES } from '../utils/constants';

const ContentCard = ({ content }) => {
  const isVideo = content.type === CONTENT_TYPES.VIDEO;
  const isArticle = content.type === CONTENT_TYPES.ARTICLE;

  return (
    <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-200 p-6 border border-gray-200">
      {/* Header */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex-1">
          <h3 className="text-lg font-semibold text-gray-900 mb-1 line-clamp-2">
            {content.title}
          </h3>
          <div className="flex items-center gap-2 text-sm text-gray-500">
            <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded-full text-xs font-medium">
              {content.type}
            </span>
            {content.provider && (
              <span className="px-2 py-1 bg-gray-100 text-gray-700 rounded-full text-xs">
                {content.provider.name}
              </span>
            )}
          </div>
        </div>
        <div className="ml-4 text-right">
          <div className="text-2xl font-bold text-blue-600">
            {formatScore(content.score)}
          </div>
          <div className="text-xs text-gray-500">Score</div>
        </div>
      </div>

      {/* Description */}
      {content.description && (
        <p className="text-sm text-gray-600 mb-4 line-clamp-3">
          {truncateText(content.description, 150)}
        </p>
      )}

      {/* Metrics */}
      <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4 pt-4 border-t border-gray-200">
        {isVideo && (
          <>
            {content.views !== undefined && content.views !== null && (
              <div>
                <div className="text-xs text-gray-500 mb-1">Views</div>
                <div className="text-sm font-semibold text-gray-900">
                  {formatNumber(content.views)}
                </div>
              </div>
            )}
            {content.likes !== undefined && content.likes !== null && (
              <div>
                <div className="text-xs text-gray-500 mb-1">Likes</div>
                <div className="text-sm font-semibold text-gray-900">
                  {formatNumber(content.likes)}
                </div>
              </div>
            )}
            {content.duration_seconds !== null && content.duration_seconds !== undefined && (
              <div>
                <div className="text-xs text-gray-500 mb-1">Duration</div>
                <div className="text-sm font-semibold text-gray-900">
                  {Math.floor((content.duration_seconds || 0) / 60)}:{(content.duration_seconds || 0) % 60 < 10 ? '0' : ''}{(content.duration_seconds || 0) % 60}
                </div>
              </div>
            )}
          </>
        )}
        {isArticle && (
          <>
            {content.reactions !== undefined && content.reactions !== null && (
              <div>
                <div className="text-xs text-gray-500 mb-1">Reactions</div>
                <div className="text-sm font-semibold text-gray-900">
                  {formatNumber(content.reactions)}
                </div>
              </div>
            )}
            {content.comments !== undefined && content.comments !== null && (
              <div>
                <div className="text-xs text-gray-500 mb-1">Comments</div>
                <div className="text-sm font-semibold text-gray-900">
                  {formatNumber(content.comments)}
                </div>
              </div>
            )}
            {content.reading_time !== null && content.reading_time !== undefined && (
              <div>
                <div className="text-xs text-gray-500 mb-1">Reading Time</div>
                <div className="text-sm font-semibold text-gray-900">
                  {content.reading_time} min
                </div>
              </div>
            )}
          </>
        )}
      </div>

      {/* Footer */}
      <div className="flex items-center justify-between text-xs text-gray-500 pt-3 border-t border-gray-100">
        <div className="flex items-center gap-4">
          <span>Published: {formatDate(content.published_at)}</span>
          {content.tags && content.tags.length > 0 && (
            <div className="flex items-center gap-1">
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M7 7h.01M7 3h5c.512 0 1.024.195 1.414.586l7 7a2 2 0 010 2.828l-7 7a2 2 0 01-2.828 0l-7-7A1.994 1.994 0 013 12V7a4 4 0 014-4z" />
              </svg>
              <span>{content.tags.length} tags</span>
            </div>
          )}
        </div>
        <div className="flex items-center gap-3">
          {content.url && (
            <a
              href={content.url}
              target="_blank"
              rel="noopener noreferrer"
              className="text-blue-600 hover:text-blue-800 font-medium text-sm"
            >
              View →
            </a>
          )}
          <a
            href={`/content/${content.id}`}
            className="text-gray-600 hover:text-gray-900 font-medium text-sm"
          >
            Details →
          </a>
        </div>
      </div>
    </div>
  );
};

export default ContentCard;
