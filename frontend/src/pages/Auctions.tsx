import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Search, Filter, Loader2 } from 'lucide-react';
import { auctionsApi } from '../api';
import { AuctionCard } from '../components/auction';
import { Button } from '../components/common';
import type { Category, AuctionListParams } from '../types';

export default function Auctions() {
  const { t } = useTranslation();
  const [search, setSearch] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [page, setPage] = useState(1);
  const [showFilters, setShowFilters] = useState(false);

  const queryParams: AuctionListParams = {
    page,
    limit: 12,
    status: 'active',
    ...(search && { search }),
    ...(categoryId && { category_id: categoryId }),
    sort_by: 'end_time',
    sort_order: 'asc',
  };

  const { data: auctionsData, isLoading, isFetching } = useQuery({
    queryKey: ['auctions', queryParams],
    queryFn: () => auctionsApi.list(queryParams),
  });

  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => auctionsApi.getCategories(),
  });

  const auctions = auctionsData?.data || [];
  const totalPages = auctionsData?.meta?.total_pages || 1;
  const categories = categoriesData?.data || [];

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    setPage(1);
  };

  const handleCategoryChange = (newCategoryId: string) => {
    setCategoryId(newCategoryId);
    setPage(1);
  };

  const handleLoadMore = () => {
    setPage((prev) => prev + 1);
  };

  return (
    <div className="container-custom py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold">{t('auctions.browseTitle')}</h1>
        <p className="text-muted-foreground mt-1">{t('auctions.browseSubtitle')}</p>
      </div>

      {/* Search and Filter Bar */}
      <div className="flex flex-col sm:flex-row gap-4 mb-8">
        <form onSubmit={handleSearch} className="flex-1 relative">
          <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder={t('common.searchPlaceholder')}
            className="w-full h-10 pl-10 pr-4 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-ring"
          />
        </form>
        <Button
          variant="outline"
          onClick={() => setShowFilters(!showFilters)}
          className="sm:w-auto"
        >
          <Filter className="h-4 w-4 mr-2" />
          {t('auctions.filter')}
        </Button>
      </div>

      {/* Filters Panel */}
      {showFilters && (
        <div className="mb-8 p-4 bg-muted/50 rounded-lg">
          <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">{t('auction.category')}</label>
              <select
                value={categoryId}
                onChange={(e) => handleCategoryChange(e.target.value)}
                className="w-full h-10 rounded-lg border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="">{t('auctions.allCategories')}</option>
                {categories.map((cat: Category) => (
                  <option key={cat.id} value={cat.id}>
                    {cat.name}
                  </option>
                ))}
              </select>
            </div>
          </div>
        </div>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {/* Empty State */}
      {!isLoading && auctions.length === 0 && (
        <div className="text-center py-12">
          <p className="text-muted-foreground">{t('auctions.noAuctions')}</p>
        </div>
      )}

      {/* Auctions Grid */}
      {!isLoading && auctions.length > 0 && (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {auctions.map((auction) => (
              <AuctionCard key={auction.id} auction={auction} />
            ))}
          </div>

          {/* Load More Button */}
          {page < totalPages && (
            <div className="mt-8 text-center">
              <Button
                variant="outline"
                onClick={handleLoadMore}
                disabled={isFetching}
                isLoading={isFetching}
              >
                {t('auctions.loadMore')}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
