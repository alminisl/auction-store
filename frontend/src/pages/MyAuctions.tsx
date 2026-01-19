import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Plus, Loader2 } from 'lucide-react';
import { auctionsApi } from '../api';
import { useAuthStore } from '../store';
import { AuctionCard } from '../components/auction';
import { Button } from '../components/common';
import type { AuctionStatus } from '../types';

const STATUS_TABS: { value: AuctionStatus | 'all'; labelKey: string }[] = [
  { value: 'all', labelKey: 'myAuctions.all' },
  { value: 'active', labelKey: 'myAuctions.active' },
  { value: 'draft', labelKey: 'myAuctions.draft' },
  { value: 'completed', labelKey: 'myAuctions.completed' },
  { value: 'unsold', labelKey: 'myAuctions.unsold' },
];

export default function MyAuctions() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const [statusFilter, setStatusFilter] = useState<AuctionStatus | 'all'>('all');
  const [page, setPage] = useState(1);

  const queryParams = {
    page,
    limit: 12,
    seller_id: user?.id,
    ...(statusFilter !== 'all' && { status: statusFilter }),
  };

  const { data: auctionsData, isLoading, isFetching } = useQuery({
    queryKey: ['my-auctions', queryParams],
    queryFn: () => auctionsApi.list(queryParams),
    enabled: !!user?.id,
  });

  const auctions = auctionsData?.data || [];
  const totalPages = auctionsData?.meta?.total_pages || 1;

  const handleStatusChange = (status: AuctionStatus | 'all') => {
    setStatusFilter(status);
    setPage(1);
  };

  const handleLoadMore = () => {
    setPage((prev) => prev + 1);
  };

  return (
    <div className="container-custom py-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
        <div>
          <h1 className="text-2xl font-bold">{t('myAuctions.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('myAuctions.subtitle')}</p>
        </div>
        <Link to="/auctions/create">
          <Button>
            <Plus className="h-4 w-4 mr-2" />
            {t('myAuctions.createNew')}
          </Button>
        </Link>
      </div>

      {/* Status Tabs */}
      <div className="flex gap-2 mb-8 overflow-x-auto pb-2">
        {STATUS_TABS.map((tab) => (
          <button
            key={tab.value}
            onClick={() => handleStatusChange(tab.value)}
            className={`px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors ${
              statusFilter === tab.value
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted hover:bg-muted/80 text-foreground'
            }`}
          >
            {t(tab.labelKey)}
          </button>
        ))}
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {/* Empty State */}
      {!isLoading && auctions.length === 0 && (
        <div className="text-center py-12">
          <p className="text-muted-foreground mb-4">{t('myAuctions.noAuctions')}</p>
          <Link to="/auctions/create">
            <Button>
              <Plus className="h-4 w-4 mr-2" />
              {t('myAuctions.createFirst')}
            </Button>
          </Link>
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
