// ContentDetail.jsx - Content detail page
// Uses Redux for state management

import { useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { useAppDispatch, useAppSelector } from '../store/hooks';
import { fetchContentById, clearContent } from '../store/slices/contentSlice';
import { formatDate, formatNumber, formatScore } from '../utils/helpers';
import { CONTENT_TYPES } from '../utils/constants';

const ContentDetail = () => {
  const { id } = useParams();
  const navigate = useNavigate();
  const dispatch = useAppDispatch();

  const { currentContent, loading, error } = useAppSelector(
    (state) => state.content
  );

  useEffect(() => {
    if (id) {
      dispatch(fetchContentById(id));
    }

    // Clear content when component unmounts
    return () => {
      dispatch(clearContent());
    };
  }, [dispatch, id]);

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading content...</p>
        </div>
      </div>
    );
  }

  if (error || !currentContent) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-2">Content Not Found</h2>
          <p className="text-gray-600 mb-4">{error || 'The requested content could not be found.'}</p>
          <button
            onClick={() => navigate('/')}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
          >
            Back to Search
          </button>
        </div>
      </div>
    );
  }

  const content = currentContent;
  const isVideo = content.type === CONTENT_TYPES.VIDEO;
  const isArticle = content.type === CONTENT_TYPES.ARTICLE;

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <button
            onClick={() => navigate('/')}
            className="text-blue-600 hover:text-blue-800 font-medium mb-4"
          >
            ← Back to Search
          </button>
        </div>
      </header>

      {/* Main Content */}
      <main className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-white rounded-lg shadow-md p-8">
          {/* Title and Type */}
          <div className="mb-6">
            <div className="flex items-center gap-3 mb-4">
              <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium">
                {content.type}
              </span>
              {content.provider && (
                <span className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm">
                  {content.provider.name}
                </span>
              )}
              <span className="ml-auto text-2xl font-bold text-blue-600">
                Score: {formatScore(content.score)}
              </span>
            </div>
            <h1 className="text-3xl font-bold text-gray-900 mb-2">
              {content.title}
            </h1>
            <p className="text-gray-500">
              Published: {formatDate(content.published_at)}
            </p>
          </div>

          {/* Metrics */}
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-6 p-4 bg-gray-50 rounded-lg">
            {isVideo && (
              <>
                {content.views !== undefined && content.views !== null && (
                  <div>
                    <div className="text-xs text-gray-500 mb-1">Views</div>
                    <div className="text-lg font-semibold text-gray-900">
                      {formatNumber(content.views)}
                    </div>
                  </div>
                )}
                {content.likes !== undefined && content.likes !== null && (
                  <div>
                    <div className="text-xs text-gray-500 mb-1">Likes</div>
                    <div className="text-lg font-semibold text-gray-900">
                      {formatNumber(content.likes)}
                    </div>
                  </div>
                )}
                {content.duration_seconds !== null && content.duration_seconds !== undefined && (
                  <div>
                    <div className="text-xs text-gray-500 mb-1">Duration</div>
                    <div className="text-lg font-semibold text-gray-900">
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
                    <div className="text-lg font-semibold text-gray-900">
                      {formatNumber(content.reactions)}
                    </div>
                  </div>
                )}
                {content.comments !== undefined && content.comments !== null && (
                  <div>
                    <div className="text-xs text-gray-500 mb-1">Comments</div>
                    <div className="text-lg font-semibold text-gray-900">
                      {formatNumber(content.comments)}
                    </div>
                  </div>
                )}
                {content.reading_time !== null && content.reading_time !== undefined && (
                  <div>
                    <div className="text-xs text-gray-500 mb-1">Reading Time</div>
                    <div className="text-lg font-semibold text-gray-900">
                      {content.reading_time} min
                    </div>
                  </div>
                )}
              </>
            )}
          </div>

          {/* Tags */}
          {content.tags && content.tags.length > 0 && (
            <div className="mb-6">
              <h3 className="text-sm font-medium text-gray-700 mb-2">Tags</h3>
              <div className="flex flex-wrap gap-2">
                {content.tags.map((tag, index) => (
                  <span
                    key={index}
                    className="px-3 py-1 bg-gray-100 text-gray-700 rounded-full text-sm"
                  >
                    {tag}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Metadata */}
          <div className="pt-6 border-t border-gray-200">
            <div className="grid grid-cols-2 gap-4 text-sm">
              <div>
                <span className="text-gray-500">Content ID:</span>
                <span className="ml-2 font-medium text-gray-900">{content.id}</span>
              </div>
              <div>
                <span className="text-gray-500">External ID:</span>
                <span className="ml-2 font-medium text-gray-900">{content.external_id}</span>
              </div>
              <div>
                <span className="text-gray-500">Provider ID:</span>
                <span className="ml-2 font-medium text-gray-900">{content.provider_id}</span>
              </div>
              <div>
                <span className="text-gray-500">Created:</span>
                <span className="ml-2 font-medium text-gray-900">{formatDate(content.created_at)}</span>
              </div>
            </div>
          </div>

          {/* External Link */}
          {content.url && (
            <div className="mt-6 pt-6 border-t border-gray-200">
              <a
                href={content.url}
                target="_blank"
                rel="noopener noreferrer"
                className="inline-flex items-center px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
              >
                View Original Content
                <svg className="ml-2 w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
                </svg>
              </a>
            </div>
          )}
        </div>
      </main>
    </div>
  );
};

export default ContentDetail;
