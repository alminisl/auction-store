import { useState } from 'react';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Package, Loader2, ShoppingBag } from 'lucide-react';
import { bidsApi } from '../api';
import { useAuthStore } from '../store';
import { AuctionCard } from '../components/auction';
import { Button } from '../components/common';
import type { Auction } from '../types';

type PurchaseFilter = 'all' | 'won' | 'bought';

export default function MyPurchases() {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const [filter, setFilter] = useState<PurchaseFilter>('all');
  const [page, setPage] = useState(1);

  const { data: bidsData, isLoading } = useQuery({
    queryKey: ['my-bids', page],
    queryFn: () => bidsApi.getMyBids({ page, limit: 50 }),
    enabled: !!user?.id,
  });

  const bids = bidsData?.data || [];

  // Extract unique auctions where user is the winner
  const wonAuctions: Auction[] = [];
  const seenAuctionIds = new Set<string>();

  bids.forEach((bid) => {
    if (bid.auction && !seenAuctionIds.has(bid.auction.id)) {
      // Check if auction is completed and user is the winner
      if (bid.auction.status === 'completed' && bid.auction.winner_id === user?.id) {
        seenAuctionIds.add(bid.auction.id);
        wonAuctions.push(bid.auction);
      }
    }
  });

  // Filter based on how they won (bid vs buy now)
  const filteredAuctions = wonAuctions.filter((auction) => {
    if (filter === 'all') return true;
    // If current_price equals buy_now_price, it was likely a Buy Now purchase
    const wasBuyNow = auction.buy_now_price && auction.current_price === auction.buy_now_price;
    if (filter === 'bought') return wasBuyNow;
    if (filter === 'won') return !wasBuyNow;
    return true;
  });

  const handleLoadMore = () => {
    setPage((prev) => prev + 1);
  };

  return (
    <div className="container-custom py-8">
      {/* Header */}
      <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
        <div>
          <h1 className="text-2xl font-bold">{t('purchases.title')}</h1>
          <p className="text-muted-foreground mt-1">{t('purchases.subtitle')}</p>
        </div>
        <Link to="/auctions">
          <Button variant="outline">
            <ShoppingBag className="h-4 w-4 mr-2" />
            {t('purchases.browseMore')}
          </Button>
        </Link>
      </div>

      {/* Filter Tabs */}
      <div className="flex gap-2 mb-8 overflow-x-auto pb-2">
        {[
          { value: 'all' as PurchaseFilter, label: t('purchases.all') },
          { value: 'won' as PurchaseFilter, label: t('purchases.wonBidding') },
          { value: 'bought' as PurchaseFilter, label: t('purchases.boughtNow') },
        ].map((tab) => (
          <button
            key={tab.value}
            onClick={() => setFilter(tab.value)}
            className={`px-4 py-2 rounded-lg text-sm font-medium whitespace-nowrap transition-colors ${
              filter === tab.value
                ? 'bg-primary text-primary-foreground'
                : 'bg-muted hover:bg-muted/80 text-foreground'
            }`}
          >
            {tab.label}
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
      {!isLoading && filteredAuctions.length === 0 && (
        <div className="text-center py-12">
          <Package className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium mb-2">{t('purchases.noPurchases')}</h3>
          <p className="text-muted-foreground mb-6">{t('purchases.noPurchasesDesc')}</p>
          <Link to="/auctions">
            <Button>
              <ShoppingBag className="h-4 w-4 mr-2" />
              {t('purchases.startShopping')}
            </Button>
          </Link>
        </div>
      )}

      {/* Purchases Grid */}
      {!isLoading && filteredAuctions.length > 0 && (
        <>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {filteredAuctions.map((auction) => (
              <AuctionCard key={auction.id} auction={auction} />
            ))}
          </div>

          {/* Load More Button - if we have more pages */}
          {bidsData?.meta?.total_pages && page < bidsData.meta.total_pages && (
            <div className="mt-8 text-center">
              <Button variant="outline" onClick={handleLoadMore}>
                {t('auctions.loadMore')}
              </Button>
            </div>
          )}
        </>
      )}
    </div>
  );
}
