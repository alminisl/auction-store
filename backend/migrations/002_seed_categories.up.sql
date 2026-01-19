-- Seed initial categories
INSERT INTO categories (id, name, slug, description) VALUES
    (uuid_generate_v4(), 'Electronics', 'electronics', 'Computers, phones, cameras, and other electronic devices'),
    (uuid_generate_v4(), 'Fashion', 'fashion', 'Clothing, shoes, accessories, and jewelry'),
    (uuid_generate_v4(), 'Home & Garden', 'home-garden', 'Furniture, decor, tools, and gardening supplies'),
    (uuid_generate_v4(), 'Sports & Outdoors', 'sports-outdoors', 'Sports equipment, outdoor gear, and fitness items'),
    (uuid_generate_v4(), 'Collectibles & Art', 'collectibles-art', 'Antiques, art, coins, stamps, and memorabilia'),
    (uuid_generate_v4(), 'Toys & Hobbies', 'toys-hobbies', 'Toys, games, models, and hobby supplies'),
    (uuid_generate_v4(), 'Motors', 'motors', 'Cars, motorcycles, parts, and accessories'),
    (uuid_generate_v4(), 'Books & Media', 'books-media', 'Books, movies, music, and video games'),
    (uuid_generate_v4(), 'Health & Beauty', 'health-beauty', 'Personal care, cosmetics, and health products'),
    (uuid_generate_v4(), 'Business & Industrial', 'business-industrial', 'Office equipment, industrial supplies, and commercial items');

-- Create an admin user (password: Admin123!)
INSERT INTO users (id, email, username, password_hash, role, email_verified) VALUES
    (uuid_generate_v4(), 'admin@auction.com', 'admin', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/X.VNYbPYkCk6QdDXu', 'admin', true);
