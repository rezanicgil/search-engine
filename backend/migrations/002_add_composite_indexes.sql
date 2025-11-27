-- 002_add_composite_indexes.sql - Add composite indexes for better query performance
-- These indexes optimize COUNT queries and filtered searches on large datasets
-- Note: MySQL doesn't support IF NOT EXISTS for CREATE INDEX, so we check existence first

-- Composite index for provider_id + type filtering
-- Speeds up queries like: WHERE provider_id = ? AND type = ?
SET @index_exists = (SELECT COUNT(*) FROM information_schema.statistics 
    WHERE table_schema = DATABASE() 
    AND table_name = 'contents' 
    AND index_name = 'idx_provider_type');
SET @sql = IF(@index_exists = 0, 
    'CREATE INDEX idx_provider_type ON contents(provider_id, type)', 
    'SELECT ''Index idx_provider_type already exists''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Composite index for provider_id + published_at filtering
-- Speeds up queries like: WHERE provider_id = ? AND published_at >= ? AND published_at <= ?
SET @index_exists = (SELECT COUNT(*) FROM information_schema.statistics 
    WHERE table_schema = DATABASE() 
    AND table_name = 'contents' 
    AND index_name = 'idx_provider_published');
SET @sql = IF(@index_exists = 0, 
    'CREATE INDEX idx_provider_published ON contents(provider_id, published_at)', 
    'SELECT ''Index idx_provider_published already exists''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Composite index for type + published_at filtering
-- Speeds up queries like: WHERE type = ? AND published_at >= ? AND published_at <= ?
SET @index_exists = (SELECT COUNT(*) FROM information_schema.statistics 
    WHERE table_schema = DATABASE() 
    AND table_name = 'contents' 
    AND index_name = 'idx_type_published');
SET @sql = IF(@index_exists = 0, 
    'CREATE INDEX idx_type_published ON contents(type, published_at)', 
    'SELECT ''Index idx_type_published already exists''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

-- Composite index for provider_id + type + published_at (covers most common filter combinations)
-- This is a covering index that can satisfy many queries without accessing the table
SET @index_exists = (SELECT COUNT(*) FROM information_schema.statistics 
    WHERE table_schema = DATABASE() 
    AND table_name = 'contents' 
    AND index_name = 'idx_provider_type_published');
SET @sql = IF(@index_exists = 0, 
    'CREATE INDEX idx_provider_type_published ON contents(provider_id, type, published_at)', 
    'SELECT ''Index idx_provider_type_published already exists''');
PREPARE stmt FROM @sql;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

