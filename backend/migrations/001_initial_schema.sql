-- 001_initial_schema.sql - Initial database schema
-- Creates tables for content, providers, and related metadata
-- This migration creates the foundation for the search engine database

-- Migration history table
-- Tracks which migrations have been applied to prevent duplicate execution
CREATE TABLE IF NOT EXISTS migration_history (
    id INT AUTO_INCREMENT PRIMARY KEY,
    migration_name VARCHAR(255) NOT NULL UNIQUE,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_migration_name (migration_name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Providers table
-- Stores information about content providers (JSON, XML sources)
-- This allows us to manage multiple providers and their rate limits
CREATE TABLE IF NOT EXISTS providers (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE COMMENT 'Provider identifier (e.g., provider1, provider2)',
    url VARCHAR(500) NOT NULL COMMENT 'API endpoint URL',
    format ENUM('json', 'xml') NOT NULL COMMENT 'Data format type',
    rate_limit_per_minute INT DEFAULT 60 COMMENT 'Maximum requests per minute',
    last_fetched_at TIMESTAMP NULL COMMENT 'Last successful data fetch timestamp',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name),
    INDEX idx_last_fetched (last_fetched_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Contents table
-- Stores standardized content from all providers
-- This is the main table for search operations
-- We store both video and article metrics in the same table for simplicity
CREATE TABLE IF NOT EXISTS contents (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    provider_id INT NOT NULL COMMENT 'Reference to the provider',
    external_id VARCHAR(255) NOT NULL COMMENT 'Original ID from provider',
    title VARCHAR(500) NOT NULL COMMENT 'Content title/headline',
    type ENUM('video', 'article') NOT NULL COMMENT 'Content type',
    
    -- Video-specific metrics (NULL for articles)
    views INT DEFAULT 0 COMMENT 'Number of views (for videos)',
    likes INT DEFAULT 0 COMMENT 'Number of likes (for videos)',
    duration_seconds INT NULL COMMENT 'Duration in seconds (for videos)',
    
    -- Article-specific metrics (NULL for videos)
    reading_time INT NULL COMMENT 'Reading time in minutes (for articles)',
    reactions INT DEFAULT 0 COMMENT 'Number of reactions (for articles)',
    comments INT DEFAULT 0 COMMENT 'Number of comments (for articles)',
    
    -- Common fields
    published_at TIMESTAMP NOT NULL COMMENT 'Publication date from provider',
    score DECIMAL(10, 4) DEFAULT 0.0000 COMMENT 'Calculated relevance/popularity score',
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    -- Foreign key and indexes
    FOREIGN KEY (provider_id) REFERENCES providers(id) ON DELETE CASCADE,
    UNIQUE KEY uk_provider_external (provider_id, external_id),
    INDEX idx_type (type),
    INDEX idx_published_at (published_at),
    INDEX idx_score (score DESC),
    INDEX idx_title (title(255)),
    FULLTEXT INDEX ft_title (title) COMMENT 'Full-text search on title'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Content tags table
-- Stores tags/categories for content (many-to-many relationship)
-- Tags come from providers and are used for filtering and search
CREATE TABLE IF NOT EXISTS content_tags (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    content_id BIGINT NOT NULL COMMENT 'Reference to content',
    tag VARCHAR(100) NOT NULL COMMENT 'Tag name',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Foreign key and indexes
    FOREIGN KEY (content_id) REFERENCES contents(id) ON DELETE CASCADE,
    UNIQUE KEY uk_content_tag (content_id, tag),
    INDEX idx_tag (tag),
    INDEX idx_content_id (content_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

