import { useState, useRef, useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Upload, X, ImagePlus, Loader2 } from 'lucide-react';
import { Button, Input } from '../components/common';
import { auctionsApi } from '../api';
import { CURRENCIES, type SupportedCurrency } from '../utils/formatters';
import type { AuctionCondition, Category, CardCondition, GeneralCondition, AuctionImage } from '../types';
import { TRADING_CARD_CATEGORY_SLUGS } from '../types';

export default function EditAuction() {
  const { t } = useTranslation();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
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
  const [existingImages, setExistingImages] = useState<AuctionImage[]>([]);
  const [newImages, setNewImages] = useState<File[]>([]);
  const [newImagePreviews, setNewImagePreviews] = useState<string[]>([]);

  // Fetch auction data
  const { data: auctionData, isLoading: isLoadingAuction } = useQuery({
    queryKey: ['auction', id],
    queryFn: () => auctionsApi.getById(id!),
    enabled: !!id,
  });

  // Fetch categories
  const { data: categoriesData } = useQuery({
    queryKey: ['categories'],
    queryFn: () => auctionsApi.getCategories(),
  });

  const categories = categoriesData?.data || [];
  const auction = auctionData?.data;

  // Populate form when auction data loads
  useEffect(() => {
    if (auction) {
      setTitle(auction.title || '');
      setDescription(auction.description || '');
      setCategoryId(auction.category_id || '');
      setCondition((auction.condition as AuctionCondition) || '');
      setCurrency((auction.currency as SupportedCurrency) || 'USD');
      setStartingPrice(auction.starting_price || '');
      setBuyNowPrice(auction.buy_now_price || '');
      setExistingImages(auction.images || []);
    }
  }, [auction]);

  // Owner can always edit their auctions
  const canEdit = true;

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
    const totalImages = existingImages.length + newImages.length + files.length;
    if (totalImages > 10) {
      setError(t('auction.maxImagesError'));
      return;
    }

    const validFiles = files.filter(file => {
      if (!file.type.startsWith('image/')) return false;
      if (file.size > 10 * 1024 * 1024) return false;
      return true;
    });

    setNewImages(prev => [...prev, ...validFiles]);

    validFiles.forEach(file => {
      const reader = new FileReader();
      reader.onload = (e) => {
        setNewImagePreviews(prev => [...prev, e.target?.result as string]);
      };
      reader.readAsDataURL(file);
    });
  };

  const removeNewImage = (index: number) => {
    setNewImages(prev => prev.filter((_, i) => i !== index));
    setNewImagePreviews(prev => prev.filter((_, i) => i !== index));
  };

  const deleteExistingImageMutation = useMutation({
    mutationFn: (imageId: string) => auctionsApi.deleteImage(id!, imageId),
    onSuccess: (_, imageId) => {
      setExistingImages(prev => prev.filter(img => img.id !== imageId));
    },
  });

  const removeExistingImage = (imageId: string) => {
    deleteExistingImageMutation.mutate(imageId);
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
      // Calculate bid increment automatically
      const priceNum = parseFloat(startingPrice) || 0;
      const bidIncrement = calculateBidIncrement(priceNum);

      // Update auction
      const updateData = {
        title,
        description: description || undefined,
        category_id: categoryId || undefined,
        condition: condition || undefined,
        starting_price: startingPrice,
        buy_now_price: buyNowPrice || undefined,
        bid_increment: bidIncrement,
      };

      await auctionsApi.update(id!, updateData);

      // Upload new images
      for (const image of newImages) {
        try {
          await auctionsApi.uploadImage(id!, image);
        } catch (imgErr) {
          console.error('Failed to upload image:', imgErr);
        }
      }

      // Publish if requested
      if (publish) {
        await auctionsApi.publish(id!);
      }

      queryClient.invalidateQueries({ queryKey: ['auction', id] });
      queryClient.invalidateQueries({ queryKey: ['my-auctions'] });

      navigate(`/auctions/${id}`);
    } catch (err: any) {
      setError(err.message || 'Failed to update auction');
    } finally {
      setIsSubmitting(false);
    }
  };

  if (isLoadingAuction) {
    return (
      <div className="container-custom py-8 flex justify-center">
        <Loader2 className="h-8 w-8 animate-spin text-primary" />
      </div>
    );
  }

  if (!auction) {
    return (
      <div className="container-custom py-8">
        <p className="text-muted-foreground">{t('auctions.notFound')}</p>
      </div>
    );
  }

  if (!canEdit) {
    return (
      <div className="container-custom py-8">
        <p className="text-muted-foreground">{t('myAuctions.cannotEdit')}</p>
        <Button className="mt-4" onClick={() => navigate(`/auctions/${id}`)}>
          {t('auctions.backToAuctions')}
        </Button>
      </div>
    );
  }

  const totalImages = existingImages.length + newImages.length;

  return (
    <div className="container-custom py-8">
      <div className="max-w-2xl mx-auto">
        <div className="mb-8">
          <h1 className="text-2xl font-bold">{t('myAuctions.editTitle')}</h1>
          <p className="text-muted-foreground mt-1">{t('myAuctions.editSubtitle')}</p>
        </div>

        <div className="space-y-6">
          {/* Images */}
          <div>
            <label className="block text-sm font-medium mb-2">{t('auction.images')}</label>
            <div className="border-2 border-dashed rounded-lg p-6">
              {(existingImages.length > 0 || newImagePreviews.length > 0) ? (
                <div className="grid grid-cols-3 gap-4 mb-4">
                  {/* Existing images */}
                  {existingImages.map((img) => (
                    <div key={img.id} className="relative aspect-square">
                      <img
                        src={img.url}
                        alt="Auction image"
                        className="w-full h-full object-cover rounded-lg"
                      />
                      <button
                        onClick={() => removeExistingImage(img.id)}
                        disabled={deleteExistingImageMutation.isPending}
                        className="absolute -top-2 -right-2 bg-destructive text-destructive-foreground rounded-full p-1"
                      >
                        <X className="h-4 w-4" />
                      </button>
                    </div>
                  ))}
                  {/* New image previews */}
                  {newImagePreviews.map((preview, index) => (
                    <div key={`new-${index}`} className="relative aspect-square">
                      <img
                        src={preview}
                        alt={`New preview ${index + 1}`}
                        className="w-full h-full object-cover rounded-lg"
                      />
                      <button
                        onClick={() => removeNewImage(index)}
                        className="absolute -top-2 -right-2 bg-destructive text-destructive-foreground rounded-full p-1"
                      >
                        <X className="h-4 w-4" />
                      </button>
                    </div>
                  ))}
                  {totalImages < 10 && (
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

          {error && (
            <div className="p-3 rounded-lg bg-destructive/10 text-destructive text-sm">
              {error}
            </div>
          )}

          {/* Actions */}
          <div className="flex gap-4 pt-4">
            <Button
              variant="outline"
              onClick={() => navigate('/my-auctions')}
              disabled={isSubmitting}
              className="flex-1"
            >
              {t('common.cancel')}
            </Button>
            <Button
              onClick={() => handleSubmit(false)}
              disabled={isSubmitting || !title || !startingPrice}
              className="flex-1"
            >
              {isSubmitting ? t('auction.saving') : t('common.save')}
            </Button>
            {auction?.status === 'draft' && (
              <Button
                onClick={() => handleSubmit(true)}
                disabled={isSubmitting || !title || !startingPrice}
                className="flex-1"
              >
                {isSubmitting ? t('auction.publishing') : t('auction.publishNow')}
              </Button>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
