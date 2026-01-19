import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Heart, Clock, Trash2, Loader2, TrendingUp, TrendingDown } from 'lucide-react';
import { usersApi, bidsApi } from '../api';
import { useAuthStore } from '../store';
import { useCountdown } from '../hooks';
import { formatCurrency } from '../utils';
import { cn } from '../utils/cn';
import { Button } from '../components/common';
import type { Auction, Bid, WatchlistItem } from '../types';

interface WatchlistCardProps {
  item: WatchlistItem;
  userBids: Bid[];
  onRemove: (auctionId: string) => void;
  isRemoving: boolean;
}

function WatchlistCard({ item, userBids, onRemove, isRemoving }: WatchlistCardProps) {
  const { t } = useTranslation();
  const { user } = useAuthStore();
  const auction = item.auction;

  if (!auction) return null;

  const countdown = useCountdown(auction.end_time);
  const imageUrl = auction.images?.[0]?.url || '/placeholder-auction.svg';
  const isEndingSoon = countdown.total > 0 && countdown.total < 3600;

  // Find user's highest bid on this auction
  const userBidsOnAuction = userBids.filter((bid) => bid.auction_id === auction.id);
  const userHighestBid = userBidsOnAuction.length > 0
    ? userBidsOnAuction.reduce((max, bid) =>
        parseFloat(bid.amount) > parseFloat(max.amount) ? bid : max
      )
    : null;

  // Check if user is winning (use tolerance for floating point comparison)
  const userBidAmount = userHighestBid ? parseFloat(userHighestBid.amount) : 0;
  const currentPrice = parseFloat(auction.current_price);
  const bidDifference = currentPrice - userBidAmount;

  // User is winning if their bid is equal to or greater than current price (within 1 cent tolerance)
  const isWinning = auction.status === 'active' && userHighestBid && bidDifference <= 0.01;
  const isOutbid = auction.status === 'active' && userHighestBid && bidDifference > 0.01;
  const hasEnded = auction.status !== 'active' || countdown.isExpired;
  const didWin = hasEnded && auction.winner_id === user?.id;

  return (
    <div className="bg-card rounded-lg border border-border overflow-hidden">
      <div className="flex flex-col sm:flex-row">
        {/* Image */}
        <Link to={`/auctions/${auction.id}`} className="sm:w-48 flex-shrink-0">
          <div className="aspect-square sm:h-48 bg-muted overflow-hidden">
            <img
              src={imageUrl}
              alt={auction.title}
              className="w-full h-full object-cover hover:scale-105 transition-transform duration-300"
              onError={(e) => {
                (e.target as HTMLImageElement).src = '/placeholder-auction.svg';
              }}
            />
          </div>
        </Link>

        {/* Content */}
        <div className="flex-1 p-4">
          <div className="flex justify-between items-start gap-4">
            <div className="flex-1">
              <Link
                to={`/auctions/${auction.id}`}
                className="font-medium text-foreground hover:text-primary transition-colors line-clamp-2"
              >
                {auction.title}
              </Link>

              {/* Status Badge */}
              <div className="mt-2 flex flex-wrap gap-2">
                {isWinning && (
                  <span className="inline-flex items-center gap-1 bg-green-500/10 text-green-600 text-xs font-medium px-2 py-1 rounded">
                    <TrendingUp className="h-3 w-3" />
                    {t('watchlist.winning')}
                  </span>
                )}
                {isOutbid && (
                  <span className="inline-flex items-center gap-1 bg-destructive/10 text-destructive text-xs font-medium px-2 py-1 rounded">
                    <TrendingDown className="h-3 w-3" />
                    {t('watchlist.outbid')}
                  </span>
                )}
                {didWin && (
                  <span className="inline-flex items-center gap-1 bg-green-500/10 text-green-600 text-xs font-medium px-2 py-1 rounded">
                    {t('watchlist.won')}
                  </span>
                )}
                {hasEnded && !didWin && (
                  <span className="inline-flex items-center gap-1 bg-muted text-muted-foreground text-xs font-medium px-2 py-1 rounded">
                    {t('auction.ended')}
                  </span>
                )}
                {isEndingSoon && !hasEnded && (
                  <span className="bg-destructive text-destructive-foreground text-xs font-medium px-2 py-1 rounded">
                    {t('auction.endingSoon')}
                  </span>
                )}
              </div>
            </div>

            {/* Remove button */}
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onRemove(auction.id)}
              disabled={isRemoving}
              className="text-muted-foreground hover:text-destructive"
            >
              <Trash2 className="h-4 w-4" />
            </Button>
          </div>

          {/* Price Info */}
          <div className="mt-4 grid grid-cols-2 sm:grid-cols-3 gap-4">
            {/* Current Price */}
            <div>
              <p className="text-xs text-muted-foreground">{t('auction.currentPrice')}</p>
              <p className="font-bold text-lg">{formatCurrency(auction.current_price, auction.currency || 'USD')}</p>
              <p className="text-xs text-muted-foreground">
                {auction.bid_count} {t('auction.bids')}
              </p>
            </div>

            {/* Your Bid */}
            {userHighestBid && (
              <div>
                <p className="text-xs text-muted-foreground">{t('watchlist.yourBid')}</p>
                <p className={cn(
                  'font-bold text-lg',
                  isWinning ? 'text-green-600' : isOutbid ? 'text-destructive' : ''
                )}>
                  {formatCurrency(userHighestBid.amount, auction.currency || 'USD')}
                </p>
                {isOutbid && bidDifference > 0.01 && (
                  <p className="text-xs text-destructive">
                    {t('watchlist.outbidBy', {
                      amount: formatCurrency(bidDifference.toFixed(2), auction.currency || 'USD')
                    })}
                  </p>
                )}
              </div>
            )}

            {/* Time Left */}
            <div>
              <p className="text-xs text-muted-foreground">{t('auction.timeLeft')}</p>
              <p className={cn(
                'font-medium flex items-center gap-1',
                isEndingSoon && !hasEnded && 'text-destructive'
              )}>
                <Clock className="h-4 w-4" />
                {countdown.formatted}
              </p>
            </div>
          </div>

          {/* Action Buttons */}
          {!hasEnded && (
            <div className="mt-4 flex gap-2">
              <Link to={`/auctions/${auction.id}`} className="flex-1">
                <Button variant={isOutbid ? 'default' : 'outline'} className="w-full">
                  {isOutbid ? t('watchlist.bidAgain') : t('watchlist.viewAuction')}
                </Button>
              </Link>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

export default function Watchlist() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { user } = useAuthStore();

  // Fetch watchlist
  const { data: watchlistData, isLoading: isLoadingWatchlist } = useQuery({
    queryKey: ['watchlist'],
    queryFn: () => usersApi.getWatchlist({ limit: 50 }),
    enabled: !!user?.id,
  });

  // Fetch user's bids to show bid status
  const { data: bidsData } = useQuery({
    queryKey: ['my-bids'],
    queryFn: () => bidsApi.getMyBids({ limit: 100 }),
    enabled: !!user?.id,
  });

  const watchlistItems = watchlistData?.data || [];
  const userBids = bidsData?.data || [];

  // Remove from watchlist mutation
  const removeMutation = useMutation({
    mutationFn: (auctionId: string) => usersApi.removeFromWatchlist(auctionId),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['watchlist'] });
    },
  });

  const isLoading = isLoadingWatchlist;

  return (
    <div className="container-custom py-8">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-2xl font-bold flex items-center gap-2">
          <Heart className="h-6 w-6 text-primary" />
          {t('watchlist.title')}
        </h1>
        <p className="text-muted-foreground mt-1">{t('watchlist.subtitle')}</p>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      )}

      {/* Empty State */}
      {!isLoading && watchlistItems.length === 0 && (
        <div className="text-center py-12">
          <Heart className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h3 className="text-lg font-medium mb-2">{t('watchlist.empty')}</h3>
          <p className="text-muted-foreground mb-6">{t('watchlist.emptyDesc')}</p>
          <Link to="/auctions">
            <Button>{t('watchlist.browseAuctions')}</Button>
          </Link>
        </div>
      )}

      {/* Watchlist Items */}
      {!isLoading && watchlistItems.length > 0 && (
        <div className="space-y-4">
          {watchlistItems.map((item) => (
            <WatchlistCard
              key={item.id}
              item={item}
              userBids={userBids}
              onRemove={(auctionId) => removeMutation.mutate(auctionId)}
              isRemoving={removeMutation.isPending}
            />
          ))}
        </div>
      )}
    </div>
  );
}
