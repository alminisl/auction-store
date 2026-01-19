-- Remove seeded data
DELETE FROM users WHERE email = 'admin@auction.com';
DELETE FROM categories WHERE slug IN (
    'electronics', 'fashion', 'home-garden', 'sports-outdoors',
    'collectibles-art', 'toys-hobbies', 'motors', 'books-media',
    'health-beauty', 'business-industrial'
);
