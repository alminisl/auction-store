import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { Clock, Gavel, Pencil, Trash2 } from 'lucide-react';
import { useCountdown } from '../../hooks';
import { formatCurrency } from '../../utils';
import { cn } from '../../utils/cn';
import type { Auction } from '../../types';

interface AuctionCardProps {
  auction: Auction;
  className?: string;
  showActions?: boolean;
  onEdit?: (auction: Auction) => void;
  onDelete?: (auction: Auction) => void;
}

export default function AuctionCard({ auction, className, showActions, onEdit, onDelete }: AuctionCardProps) {
  const { t } = useTranslation();
  const countdown = useCountdown(auction.end_time);

  const imageUrl = auction.images?.[0]?.url || '/placeholder-auction.svg';
  const isEndingSoon = countdown.total > 0 && countdown.total < 3600; // Less than 1 hour
  const hasBuyNow = !!auction.buy_now_price && parseFloat(auction.buy_now_price) > 0;
  const isDraft = auction.status === 'draft';
  const canModify = isDraft && showActions;

  return (
    <div className={cn('relative', className)}>
      {/* Action Buttons */}
      {canModify && (
        <div className="absolute top-2 right-2 z-10 flex gap-1">
          {onEdit && (
            <button
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onEdit(auction);
              }}
              className="p-2 bg-background/90 hover:bg-background rounded-md border border-border shadow-sm transition-colors"
              title={t('common.edit')}
            >
              <Pencil className="h-4 w-4 text-muted-foreground hover:text-primary" />
            </button>
          )}
          {onDelete && (
            <button
              onClick={(e) => {
                e.preventDefault();
                e.stopPropagation();
                onDelete(auction);
              }}
              className="p-2 bg-background/90 hover:bg-destructive/10 rounded-md border border-border shadow-sm transition-colors"
              title={t('common.delete')}
            >
              <Trash2 className="h-4 w-4 text-muted-foreground hover:text-destructive" />
            </button>
          )}
        </div>
      )}

      <Link
        to={`/auctions/${auction.id}`}
        className="group block bg-background rounded-lg border border-border overflow-hidden hover:shadow-lg transition-all duration-200"
      >
      {/* Image */}
      <div className="relative aspect-square bg-muted overflow-hidden">
        <img
          src={imageUrl}
          alt={auction.title}
          className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
          onError={(e) => {
            (e.target as HTMLImageElement).src = '/placeholder-auction.svg';
          }}
        />
        {/* Status badges */}
        <div className="absolute top-3 left-3 flex flex-col gap-1.5">
          {isDraft && (
            <span className="bg-amber-500 text-white text-xs font-medium px-2 py-1 rounded-md">
              {t('myAuctions.draft')}
            </span>
          )}
          {isEndingSoon && !countdown.isExpired && !isDraft && (
            <span className="bg-destructive text-destructive-foreground text-xs font-medium px-2 py-1 rounded-md">
              {t('auction.endingSoon')}
            </span>
          )}
          {countdown.isExpired && !isDraft && (
            <span className="bg-muted-foreground/80 text-white text-xs font-medium px-2 py-1 rounded-md">
              {t('auction.ended')}
            </span>
          )}
          {hasBuyNow && !countdown.isExpired && !isDraft && (
            <span className="bg-primary text-primary-foreground text-xs font-medium px-2 py-1 rounded-md">
              {t('auction.buyNow')}
            </span>
          )}
        </div>
      </div>

      {/* Content */}
      <div className="p-4">
        {/* Title */}
        <h3 className="font-medium text-foreground line-clamp-2 group-hover:text-primary transition-colors min-h-[2.5rem]">
          {auction.title}
        </h3>

        {/* Price */}
        <div className="mt-3">
          <p className="text-xl font-bold text-foreground">
            {formatCurrency(auction.current_price, auction.currency || 'USD')}
          </p>
        </div>

        {/* Bid count and time */}
        <div className="mt-3 flex items-center justify-between text-sm text-muted-foreground">
          <div className="flex items-center gap-1.5">
            <Gavel className="h-4 w-4" />
            <span>
              {auction.bid_count} {t('auction.bids')}
            </span>
          </div>
          <div className="flex items-center gap-1.5">
            <Clock className="h-4 w-4" />
            <span className={cn(isEndingSoon && !countdown.isExpired && 'text-destructive font-medium')}>
              {countdown.formatted}
            </span>
          </div>
        </div>
      </div>
    </Link>
    </div>
  );
}
