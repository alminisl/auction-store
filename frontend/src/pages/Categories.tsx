import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Loader2, Grid3X3 } from 'lucide-react';
import { auctionsApi } from '../api';
import type { Category } from '../types';

// Placeholder images for categories (can be replaced with actual images later)
const categoryImages: Record<string, string> = {
  'pokemon': '/categories/pokemon.svg',
  'magic-the-gathering': '/categories/magic-the-gathering.svg',
  'yugioh': '/categories/yugioh.svg',
  'one-piece': '/categories/one-piece.svg',
  'sports-cards': '/categories/sports-cards.svg',
  'trading-cards': '/categories/trading-cards.svg',
};

// Default placeholder for categories without specific images
const defaultCategoryImage = '/categories/trading-cards.svg';

function getCategoryImage(category: Category): string {
  // First check if category has its own image
  if (category.image_url) {
    return category.image_url;
  }
  // Then check our placeholder map by slug
  if (category.slug && categoryImages[category.slug]) {
    return categoryImages[category.slug];
  }
  // Finally use default
  return defaultCategoryImage;
}

interface CategoryCardProps {
  category: Category;
}

function CategoryCard({ category }: CategoryCardProps) {
  const { t } = useTranslation();
  const imageUrl = getCategoryImage(category);

  return (
    <Link
      to={`/auctions?category_id=${category.id}`}
      className="group block rounded-lg overflow-hidden bg-background border border-border hover:shadow-md hover:border-primary/30 transition-all duration-200"
    >
      {/* Image */}
      <div className="aspect-[3/2] bg-muted overflow-hidden">
        <img
          src={imageUrl}
          alt={category.name}
          className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          onError={(e) => {
            (e.target as HTMLImageElement).src = defaultCategoryImage;
          }}
        />
      </div>

      {/* Content */}
      <div className="p-3">
        <h3 className="font-medium text-sm text-foreground group-hover:text-primary transition-colors">
          {category.name}
        </h3>
        {category.auction_count !== undefined && (
          <p className="text-xs text-muted-foreground mt-1">
            {category.auction_count} {t('categories.auctions')}
          </p>
        )}
      </div>
    </Link>
  );
}

export default function Categories() {
  const { t } = useTranslation();

  const { data: categoriesData, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => auctionsApi.getCategories(),
  });

  const categories = categoriesData?.data || [];

  // Separate parent categories and subcategories
  const parentCategories = categories.filter(c => !c.parent_id);
  const getSubcategories = (parentId: string) => categories.filter(c => c.parent_id === parentId);

  return (
    <div className="container-custom py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold flex items-center gap-3">
          <Grid3X3 className="h-8 w-8 text-primary" />
          {t('categories.title')}
        </h1>
        <p className="text-muted-foreground mt-2">
          {t('categories.subtitle')}
        </p>
      </div>

      {/* Loading */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {/* Categories Grid */}
      {!isLoading && categories.length > 0 && (
        <div className="space-y-12">
          {/* If there are parent categories, show them with their subcategories */}
          {parentCategories.length > 0 ? (
            parentCategories.map(parent => {
              const subcategories = getSubcategories(parent.id);
              return (
                <div key={parent.id}>
                  {/* Parent Category Header */}
                  <div className="flex items-center justify-between mb-6">
                    <h2 className="text-xl font-semibold">{parent.name}</h2>
                    <Link
                      to={`/auctions?category_id=${parent.id}`}
                      className="text-sm text-primary hover:underline"
                    >
                      {t('categories.viewAll')}
                    </Link>
                  </div>

                  {/* Subcategories Grid */}
                  {subcategories.length > 0 ? (
                    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                      {subcategories.map(category => (
                        <CategoryCard key={category.id} category={category} />
                      ))}
                    </div>
                  ) : (
                    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
                      <CategoryCard category={parent} />
                    </div>
                  )}
                </div>
              );
            })
          ) : (
            /* If no parent/child structure, just show all categories */
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
              {categories.map(category => (
                <CategoryCard key={category.id} category={category} />
              ))}
            </div>
          )}
        </div>
      )}

      {/* Empty State */}
      {!isLoading && categories.length === 0 && (
        <div className="text-center py-12">
          <Grid3X3 className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium mb-2">{t('categories.noCategories')}</h3>
          <p className="text-muted-foreground">
            {t('categories.noCategoriesDesc')}
          </p>
        </div>
      )}
    </div>
  );
}
