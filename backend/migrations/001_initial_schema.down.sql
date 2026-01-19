-- Drop triggers
DROP TRIGGER IF EXISTS update_auctions_updated_at ON auctions;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS reported_listings;
DROP TABLE IF EXISTS ratings;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS watchlist;
DROP TABLE IF EXISTS bids;
DROP TABLE IF EXISTS auction_images;
DROP TABLE IF EXISTS auctions;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS oauth_accounts;
DROP TABLE IF EXISTS users;

-- Drop extension
DROP EXTENSION IF EXISTS "uuid-ossp";
