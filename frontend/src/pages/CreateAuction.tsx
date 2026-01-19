import { useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useQuery } from '@tanstack/react-query';
import { Upload, X, ImagePlus } from 'lucide-react';
import { Button, Input } from '../components/common';
import { auctionsApi } from '../api';
import { CURRENCIES, type SupportedCurrency } from '../utils/formatters';
import type { AuctionCondition, Category, CardCondition, GeneralCondition } from '../types';
import { TRADING_CARD_CATEGORY_SLUGS } from '../types';

const DURATION_OPTIONS = [
  { value: 1, labelKey: 'auction.duration1Day' },
  { value: 3, labelKey: 'auction.duration3Days' },
  { value: 5, labelKey: 'auction.duration5Days' },
  { value: 7, labelKey: 'auction.duration7Days' },
  { value: 14, labelKey: 'auction.duration14Days' },
];

export default function CreateAuction() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const fileInputRef = useRef<HTMLInputElement>(null);

  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState('');

  // Form state
  const [title, setTitle] = useState('');
  const [description, setDescription] = useState('');
  const [categoryId, setCategoryId] = useState('');
  const [condition, setCondition] = useState<AuctionCondition | ''>('');
  const [currency, setCurrency] = useState<SupportedCurrency>('USD');
  const [startingPrice, setStartingPrice] = useState('');
  const [buyNowPrice, setBuyNowPrice] = useState('');
  const [duration, setDuration] = useState(7);
  const [images, setImages] = useState<File[]>([]);
  const [imagePreviews, setImagePreviews] = useState<string[]>([]);

  // Fetch categories
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => auctionsApi.getCategories(),
  });

  const categories = categoriesData?.data || [];

  // Determine if selected category is a trading card category
  const selectedCategory = categories.find((cat: Category) => cat.id === categoryId);
  const isTradingCardCategory = selectedCategory &&
    TRADING_CARD_CATEGORY_SLUGS.some(slug =>
      selectedCategory.slug?.toLowerCase().includes(slug) ||
      selectedCategory.name?.toLowerCase().includes('card') ||
      selectedCategory.name?.toLowerCase().includes('tcg')
    );

  // Card conditions (CardMarket style)
  const cardConditions: { value: CardCondition; label: string }[] = [
    { value: 'mint', label: t('auction.conditionMint') },
    { value: 'near_mint', label: t('auction.conditionNearMint') },
    { value: 'excellent', label: t('auction.conditionExcellent') },
    { value: 'good', label: t('auction.conditionGood') },
    { value: 'played', label: t('auction.conditionPlayed') },
  ];

  // General conditions (for other items)
  const generalConditions: { value: GeneralCondition; label: string }[] = [
    { value: 'new', label: t('auction.conditionNew') },
    { value: 'like_new', label: t('auction.conditionLikeNew') },
    { value: 'very_good', label: t('auction.conditionVeryGood') },
    { value: 'good', label: t('auction.conditionGoodGeneral') },
    { value: 'acceptable', label: t('auction.conditionAcceptable') },
  ];

  const conditions = isTradingCardCategory ? cardConditions : generalConditions;

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length + images.length > 10) {
      setError(t('auction.maxImagesError'));
      return;
    }

    const validFiles = files.filter(file => {
      if (!file.type.startsWith('image/')) return false;
      if (file.size > 10 * 1024 * 1024) return false;
      return true;
    });

    setImages(prev => [...prev, ...validFiles]);

    validFiles.forEach(file => {
      const reader = new FileReader();
      reader.onload = (e) => {
        setImagePreviews(prev => [...prev, e.target?.result as string]);
      };
      reader.readAsDataURL(file);
    });
  };

  const removeImage = (index: number) => {
    setImages(prev => prev.filter((_, i) => i !== index));
    setImagePreviews(prev => prev.filter((_, i) => i !== index));
  };

  // Calculate bid increment based on starting price
  const calculateBidIncrement = (price: number): string => {
    if (price < 10) return '0.50';
    if (price < 50) return '1.00';
    if (price < 100) return '2.00';
    if (price < 500) return '5.00';
    if (price < 1000) return '10.00';
    return '25.00';
  };

  const handleSubmit = async (publish: boolean) => {
    setError('');
    setIsSubmitting(true);

    try {
      // Calculate times
      const now = new Date();
      const startTime = now.toISOString();
      const endTime = new Date(now.getTime() + duration * 24 * 60 * 60 * 1000).toISOString();

      // Calculate bid increment automatically
      const priceNum = parseFloat(startingPrice) || 0;
      const bidIncrement = calculateBidIncrement(priceNum);

      // Create auction
      const auctionData = {
        title,
        description: description || undefined,
        category_id: categoryId || undefined,
        condition: condition || undefined,
        currency,
        starting_price: startingPrice,
        buy_now_price: buyNowPrice || undefined,
        bid_increment: bidIncrement,
        start_time: startTime,
        end_time: endTime,
      };

      const response = await auctionsApi.create(auctionData);

      if (!response.success || !response.data) {
        throw new Error(response.message || 'Failed to create auction');
      }

      const auctionId = response.data.id;

      // Upload images
      for (const image of images) {
        try {
          await auctionsApi.uploadImage(auctionId, image);
        } catch (imgErr) {
          console.error('Failed to upload image:', imgErr);
        }
      }

      // Publish if requested
      if (publish) {
        await auctionsApi.publish(auctionId);
      }

      navigate(`/auctions/${auctionId}`);
    } catch (err: any) {
      setError(err.message || 'Failed to create auction');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="container-custom py-8">
      <div className="max-w-2xl mx-auto">
        <div className="mb-8">
          <h1 className="text-2xl font-bold">{t('auction.create')}</h1>
          <p className="text-muted-foreground mt-1">{t('auction.createSubtitle')}</p>
        </div>

        <div className="space-y-6">
          {/* Images */}
          <div>
            <label className="block text-sm font-medium mb-2">{t('auction.images')}</label>
            <div className="border-2 border-dashed rounded-lg p-6">
              {imagePreviews.length > 0 ? (
                <div className="grid grid-cols-3 gap-4 mb-4">
                  {imagePreviews.map((preview, index) => (
                    <div key={index} className="relative aspect-square">
                      <img
                        src={preview}
                        alt={`Preview ${index + 1}`}
                        className="w-full h-full object-cover rounded-lg"
                      />
                      <button
                        onClick={() => removeImage(index)}
                        className="absolute -top-2 -right-2 bg-destructive text-destructive-foreground rounded-full p-1"
                      >
                        <X className="h-4 w-4" />
                      </button>
                    </div>
                  ))}
                  {images.length < 10 && (
                    <button
                      onClick={() => fileInputRef.current?.click()}
                      className="aspect-square border-2 border-dashed rounded-lg flex flex-col items-center justify-center hover:bg-accent transition-colors"
                    >
                      <ImagePlus className="h-8 w-8 text-muted-foreground" />
                    </button>
                  )}
                </div>
              ) : (
                <div
                  onClick={() => fileInputRef.current?.click()}
                  className="cursor-pointer text-center"
                >
                  <Upload className="h-10 w-10 mx-auto text-muted-foreground mb-2" />
                  <p className="text-sm text-muted-foreground">{t('auction.dropImages')}</p>
                  <p className="text-xs text-muted-foreground mt-1">{t('auction.maxImages')}</p>
                </div>
              )}
              <input
                ref={fileInputRef}
                type="file"
                accept="image/*"
                multiple
                onChange={handleImageSelect}
                className="hidden"
              />
            </div>
          </div>

          {/* Title */}
          <Input
            label={`${t('auction.title')} *`}
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder={t('auction.titlePlaceholder')}
            required
          />

          {/* Description */}
          <div>
            <label className="block text-sm font-medium mb-2">{t('auction.description')}</label>
            <textarea
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              placeholder={t('auction.descriptionPlaceholder')}
              rows={4}
              className="w-full rounded-lg border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            />
          </div>

          {/* Category & Condition */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium mb-2">{t('auction.category')}</label>
              <select
                value={categoryId}
                onChange={(e) => { setCategoryId(e.target.value); setCondition(''); }}
                className="w-full h-10 rounded-lg border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="">{t('auction.selectCategory')}</option>
                {categories.map((cat: Category) => (
                  <option key={cat.id} value={cat.id}>{cat.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium mb-2">{t('auction.condition')}</label>
              <select
                value={condition}
                onChange={(e) => setCondition(e.target.value as AuctionCondition)}
                className="w-full h-10 rounded-lg border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
              >
                <option value="">{t('auction.selectCondition')}</option>
                {conditions.map((cond) => (
                  <option key={cond.value} value={cond.value}>{cond.label}</option>
                ))}
              </select>
            </div>
          </div>

          {/* Currency */}
          <div>
            <label className="block text-sm font-medium mb-2">{t('auction.currency')}</label>
            <select
              value={currency}
              onChange={(e) => setCurrency(e.target.value as SupportedCurrency)}
              className="w-full h-10 rounded-lg border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              {CURRENCIES.map((curr) => (
                <option key={curr.value} value={curr.value}>
                  {curr.label}
                </option>
              ))}
            </select>
          </div>

          {/* Pricing */}
          <div className="grid grid-cols-2 gap-4">
            <Input
              label={`${t('auction.startingPrice')} *`}
              type="number"
              step="0.01"
              min="0.01"
              value={startingPrice}
              onChange={(e) => setStartingPrice(e.target.value)}
              placeholder={t('auction.startingPricePlaceholder')}
              required
            />
            <div>
              <Input
                label={t('auction.buyNowPrice')}
                type="number"
                step="0.01"
                min="0"
                value={buyNowPrice}
                onChange={(e) => setBuyNowPrice(e.target.value)}
                placeholder={t('auction.startingPricePlaceholder')}
              />
              <p className="text-xs text-muted-foreground mt-1">{t('auction.buyNowPriceHelp')}</p>
            </div>
          </div>

          {/* Duration */}
          <div>
            <label className="block text-sm font-medium mb-2">{t('auction.duration')}</label>
            <select
              value={duration}
              onChange={(e) => setDuration(Number(e.target.value))}
              className="w-full h-10 rounded-lg border border-input bg-background px-3 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
            >
              {DURATION_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {t(opt.labelKey)}
                </option>
              ))}
            </select>
            <p className="text-xs text-muted-foreground mt-1">{t('auction.durationHelp')}</p>
          </div>

          {error && (
            <div className="p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
              {error}
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-4 pt-4">
            <Button
              variant="outline"
              onClick={() => handleSubmit(false)}
              disabled={isSubmitting || !title || !startingPrice}
              className="flex-1"
            >
              {isSubmitting ? t('auction.saving') : t('auction.saveDraft')}
            </Button>
            <Button
              onClick={() => handleSubmit(true)}
              disabled={isSubmitting || !title || !startingPrice}
              className="flex-1"
            >
              {isSubmitting ? t('auction.publishing') : t('auction.publishNow')}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
