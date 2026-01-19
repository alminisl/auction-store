import { useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { ArrowLeft, Clock, User, Eye, Gavel, Loader2 } from 'lucide-react';
import { auctionsApi, bidsApi, usersApi } from '../api';
import { useCountdown } from '../hooks';
import { useAuthStore } from '../store';
import { formatCurrency, formatTimeAgo } from '../utils';
import { cn } from '../utils/cn';
import { Button, Input } from '../components/common';
import type { Bid } from '../types';
import { TRADING_CARD_CATEGORY_SLUGS } from '../types';

export default function AuctionDetail() {
  const { id } = useParams<{ id: string }>();
  const { t } = useTranslation();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { isAuthenticated, user } = useAuthStore();

  const [selectedImageIndex, setSelectedImageIndex] = useState(0);
  const [bidAmount, setBidAmount] = useState('');
  const [error, setError] = useState('');
  const [success, setSuccess] = useState('');

  // Fetch auction
  const { data: auctionData, isLoading: isLoadingAuction } = useQuery({
    queryKey: ['auction', id],
    queryFn: () => auctionsApi.getById(id!),
    enabled: !!id,
  });

  // Fetch bids
  const { data: bidsData } = useQuery({
    queryKey: ['auction-bids', id],
    queryFn: () => bidsApi.getBidsByAuction(id!, { limit: 10 }),
    enabled: !!id,
  });

  const auction = auctionData?.data;
  const bids = bidsData?.data?.data || [];
  const countdown = useCountdown(auction?.end_time || new Date());

  const minimumBid = auction
    ? parseFloat(auction.current_price) + parseFloat(auction.bid_increment)
    : 0;

  // Place bid mutation
  const placeBidMutation = useMutation({
    mutationFn: (amount: string) => bidsApi.placeBid(id!, { amount }),
    onSuccess: async () => {
      setSuccess(t('auction.bidPlaced'));
      setBidAmount('');
      setError('');
      queryClient.invalidateQueries({ queryKey: ['auction', id] });
      queryClient.invalidateQueries({ queryKey: ['auction-bids', id] });

      // Auto-add to watchlist when placing a bid
      try {
        await usersApi.addToWatchlist(id!);
        queryClient.invalidateQueries({ queryKey: ['watchlist'] });
      } catch {
        // Ignore if already in watchlist
      }

      setTimeout(() => setSuccess(''), 3000);
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || 'Failed to place bid');
      setSuccess('');
    },
  });

  // Buy now mutation
  const buyNowMutation = useMutation({
    mutationFn: () => bidsApi.buyNow(id!),
    onSuccess: () => {
      setSuccess(t('auction.purchased'));
      setError('');
      queryClient.invalidateQueries({ queryKey: ['auction', id] });
      setTimeout(() => setSuccess(''), 3000);
    },
    onError: (err: any) => {
      setError(err.response?.data?.message || 'Failed to complete purchase');
      setSuccess('');
    },
  });

  const handlePlaceBid = (e: React.FormEvent) => {
    e.preventDefault();
    setError('');

    if (!isAuthenticated) {
      navigate('/login');
      return;
    }

    const amount = parseFloat(bidAmount);
    if (isNaN(amount) || amount < minimumBid) {
      setError(`${t('auction.minimumBid')}: ${formatCurrency(minimumBid, auction?.currency || 'USD')}`);
      return;
    }

    placeBidMutation.mutate(bidAmount);
  };

  const handleBuyNow = () => {
    if (!isAuthenticated) {
      navigate('/login');
      return;
    }
    buyNowMutation.mutate();
  };

  if (isLoadingAuction) {
    return (
      <div className="container-custom py-8">
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-primary" />
        </div>
      </div>
    );
  }

  if (!auction) {
    return (
      <div className="container-custom py-8">
        <div className="text-center py-12">
          <p className="text-muted-foreground">{t('auctions.notFound')}</p>
          <Link to="/auctions" className="text-primary hover:underline mt-4 inline-block">
            {t('auctions.backToAuctions')}
          </Link>
        </div>
      </div>
    );
  }

  const images = auction.images || [];
  const selectedImage = images[selectedImageIndex]?.url || '/placeholder-auction.svg';
  const hasBuyNow = !!auction.buy_now_price && parseFloat(auction.buy_now_price) > 0;
  const isOwner = user?.id === auction.seller_id;
  const isActive = auction.status === 'active' && !countdown.isExpired;
  const isEndingSoon = countdown.total > 0 && countdown.total < 3600;

  // Determine if this is a trading card category
  const isTradingCardCategory = auction.category &&
    TRADING_CARD_CATEGORY_SLUGS.some(slug =>
      auction.category?.slug?.toLowerCase().includes(slug) ||
      auction.category?.name?.toLowerCase().includes('card') ||
      auction.category?.name?.toLowerCase().includes('tcg')
    );

  // Card conditions (CardMarket style)
  const cardConditionLabels: Record<string, string> = {
    mint: t('auction.conditionMint'),
    near_mint: t('auction.conditionNearMint'),
    excellent: t('auction.conditionExcellent'),
    good: t('auction.conditionGood'),
    played: t('auction.conditionPlayed'),
  };

  // General conditions
  const generalConditionLabels: Record<string, string> = {
    new: t('auction.conditionNew'),
    like_new: t('auction.conditionLikeNew'),
    very_good: t('auction.conditionVeryGood'),
    good: t('auction.conditionGoodGeneral'),
    acceptable: t('auction.conditionAcceptable'),
  };

  const conditionLabels = isTradingCardCategory ? cardConditionLabels : generalConditionLabels;

  return (
    <div className="container-custom py-8">
      {/* Back link */}
      <Link
        to="/auctions"
        className="inline-flex items-center gap-2 text-muted-foreground hover:text-foreground mb-6"
      >
        <ArrowLeft className="h-4 w-4" />
        {t('auctions.backToAuctions')}
      </Link>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
        {/* Image Gallery */}
        <div>
          <div className="aspect-square bg-muted rounded-lg overflow-hidden mb-4">
            <img
              src={selectedImage}
              alt={auction.title}
              className="w-full h-full object-contain"
              onError={(e) => {
                (e.target as HTMLImageElement).src = '/placeholder-auction.svg';
              }}
            />
          </div>
          {images.length > 1 && (
            <div className="flex gap-2 overflow-x-auto pb-2">
              {images.map((image, index) => (
                <button
                  key={image.id}
                  onClick={() => setSelectedImageIndex(index)}
                  className={cn(
                    'flex-shrink-0 w-20 h-20 rounded-lg overflow-hidden border-2 transition-colors',
                    selectedImageIndex === index
                      ? 'border-primary'
                      : 'border-transparent hover:border-muted-foreground'
                  )}
                >
                  <img
                    src={image.url}
                    alt={`${auction.title} ${index + 1}`}
                    className="w-full h-full object-cover"
                  />
                </button>
              ))}
            </div>
          )}
        </div>

        {/* Auction Info */}
        <div>
          <h1 className="text-2xl font-bold mb-4">{auction.title}</h1>

          {/* Price and Status */}
          <div className="bg-muted/50 rounded-lg p-6 mb-6">
            <div className="mb-4">
              <p className="text-sm text-muted-foreground">{t('auction.currentPrice')}</p>
              <p className="text-3xl font-bold">{formatCurrency(auction.current_price, auction.currency || 'USD')}</p>
              <p className="text-sm text-muted-foreground">
                ({auction.bid_count} {t('auction.bids')})
              </p>
            </div>

            {/* Time remaining */}
            <div className="flex items-center gap-2 mb-6">
              <Clock className={cn('h-5 w-5', isEndingSoon ? 'text-destructive' : 'text-muted-foreground')} />
              <span className={cn('font-medium', isEndingSoon && 'text-destructive')}>
                {t('auction.timeLeft')}: {countdown.formatted}
              </span>
            </div>

            {/* Bidding form */}
            {isActive && !isOwner && (
              <>
                <form onSubmit={handlePlaceBid} className="mb-4">
                  <label className="block text-sm font-medium mb-2">{t('auction.yourBid')}</label>
                  <div className="flex gap-2">
                    <Input
                      type="number"
                      step="0.01"
                      min={minimumBid}
                      value={bidAmount}
                      onChange={(e) => setBidAmount(e.target.value)}
                      placeholder={formatCurrency(minimumBid, auction.currency || 'USD')}
                      className="flex-1"
                    />
                    <Button
                      type="submit"
                      disabled={placeBidMutation.isPending}
                      isLoading={placeBidMutation.isPending}
                    >
                      {t('auction.placeBid')}
                    </Button>
                  </div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {t('auction.minimumBid')}: {formatCurrency(minimumBid, auction.currency || 'USD')}
                  </p>
                </form>

                {/* Buy Now */}
                {hasBuyNow && (
                  <>
                    <div className="relative my-6">
                      <div className="absolute inset-0 flex items-center">
                        <div className="w-full border-t border-border" />
                      </div>
                      <div className="relative flex justify-center text-xs uppercase">
                        <span className="bg-muted/50 px-2 text-muted-foreground">
                          {t('auctions.or')}
                        </span>
                      </div>
                    </div>

                    <div>
                      <p className="text-sm text-muted-foreground mb-2">{t('auction.buyNow')}</p>
                      <p className="text-xl font-bold mb-3">
                        {formatCurrency(auction.buy_now_price!, auction.currency || 'USD')}
                      </p>
                      <Button
                        onClick={handleBuyNow}
                        variant="secondary"
                        className="w-full"
                        disabled={buyNowMutation.isPending}
                        isLoading={buyNowMutation.isPending}
                      >
                        {t('auction.buyNow')}
                      </Button>
                    </div>
                  </>
                )}
              </>
            )}

            {/* Ended state */}
            {!isActive && (
              <div className="text-center py-4">
                <p className="text-muted-foreground font-medium">
                  {auction.status === 'completed' ? t('auction.sold') : t('auction.ended')}
                </p>
              </div>
            )}

            {/* Owner view */}
            {isOwner && (
              <div className="text-center py-4">
                <p className="text-muted-foreground">{t('auctions.yourAuction')}</p>
              </div>
            )}

            {/* Error/Success messages */}
            {error && (
              <div className="mt-4 p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
                {error}
              </div>
            )}
            {success && (
              <div className="mt-4 p-3 rounded-lg bg-green-500/10 text-green-600 text-sm">
                {success}
              </div>
            )}
          </div>

          {/* Seller and Condition */}
          <div className="space-y-3 mb-6">
            <div className="flex items-center gap-2">
              <User className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">{t('auction.seller')}:</span>
              <span className="text-sm font-medium">@{auction.seller?.username || 'Unknown'}</span>
            </div>
            {auction.condition && (
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">{t('auction.condition')}:</span>
                <span className="text-sm font-medium">{conditionLabels[auction.condition]}</span>
              </div>
            )}
            {auction.category && (
              <div className="flex items-center gap-2">
                <span className="text-sm text-muted-foreground">{t('auction.category')}:</span>
                <span className="text-sm font-medium">{auction.category.name}</span>
              </div>
            )}
            <div className="flex items-center gap-2">
              <Eye className="h-4 w-4 text-muted-foreground" />
              <span className="text-sm text-muted-foreground">
                {auction.views_count} {t('auction.views')}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Description */}
      {auction.description && (
        <div className="mt-8">
          <h2 className="text-lg font-semibold mb-4">{t('auction.description')}</h2>
          <div className="prose prose-sm max-w-none">
            <p className="whitespace-pre-wrap text-muted-foreground">{auction.description}</p>
          </div>
        </div>
      )}

      {/* Bid History */}
      {bids.length > 0 && (
        <div className="mt-8">
          <h2 className="text-lg font-semibold mb-4">{t('auctions.bidHistory')}</h2>
          <div className="bg-muted/30 rounded-lg overflow-hidden">
            <div className="divide-y divide-border">
              {bids.map((bid: Bid) => (
                <div key={bid.id} className="flex items-center justify-between p-4">
                  <div className="flex items-center gap-3">
                    <Gavel className="h-4 w-4 text-muted-foreground" />
                    <span className="font-medium">@{bid.bidder?.username || 'Unknown'}</span>
                  </div>
                  <div className="text-right">
                    <p className="font-medium">{formatCurrency(bid.amount, auction.currency || 'USD')}</p>
                    <p className="text-xs text-muted-foreground">{formatTimeAgo(bid.created_at)}</p>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
