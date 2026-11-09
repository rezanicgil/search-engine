#!/usr/bin/env bash
set -euo pipefail
# Use only the disposable database created by the integration workflow.
[[ "${DB_NAME}" == search_test ]]
MYSQL_PWD="$DB_PASSWORD" mysql -h "$DB_HOST" -u "$DB_USER" "$DB_NAME" <<'SQL'
INSERT INTO providers (name,url,format) VALUES ('search-fixture','https://example.invalid','json');
SET @provider = LAST_INSERT_ID();
INSERT INTO contents (provider_id, external_id, title, type, published_at, score) VALUES
(@provider,'1','Orchard Guide','article',NOW(),10),
(@provider,'2','Orchard Tour','video',NOW(),30),
(@provider,'3','Go Recipes','article',NOW(),20);
SQL

search() { curl -fsS "http://localhost:8080/api/v1/search?$1"; }
search 'query=Orchard&per_page=1&page=1' | jq -e '.data | .total == 2 and .total_pages == 2 and [.results[].title] == ["Orchard Tour"]'
search 'query=Orchard&per_page=1&page=2' | jq -e '.data | .total == 2 and [.results[].title] == ["Orchard Guide"]'
search 'query=Orchard&type=article' | jq -e '.data | .total == 1 and [.results[].title] == ["Orchard Guide"]'
search 'query=Go' | jq -e '.data | .total == 1 and [.results[].title] == ["Go Recipes"]'
