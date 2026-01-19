import { formatDistanceToNow, format, differenceInSeconds } from 'date-fns';

export type SupportedCurrency = 'USD' | 'EUR' | 'BAM';

export const CURRENCIES: { value: SupportedCurrency; label: string; symbol: string }[] = [
  { value: 'USD', label: 'USD ($)', symbol: '$' },
  { value: 'EUR', label: 'EUR (€)', symbol: '€' },
  { value: 'BAM', label: 'KM (BAM)', symbol: 'KM' },
];

export function formatCurrency(amount: string | number, currency: SupportedCurrency = 'USD'): string {
  const numericAmount = typeof amount === 'string' ? parseFloat(amount) : amount;

  // BAM (Bosnian Convertible Mark) needs custom formatting since Intl may not support it well
  if (currency === 'BAM') {
    return `${numericAmount.toFixed(2)} KM`;
  }

  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency,
  }).format(numericAmount);
}

export function formatDate(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return format(d, 'MMM d, yyyy');
}

export function formatDateTime(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return format(d, 'MMM d, yyyy h:mm a');
}

export function formatTimeAgo(date: string | Date): string {
  const d = typeof date === 'string' ? new Date(date) : date;
  return formatDistanceToNow(d, { addSuffix: true });
}

export function formatCountdown(endTime: string | Date): string {
  const end = typeof endTime === 'string' ? new Date(endTime) : endTime;
  const now = new Date();
  const seconds = differenceInSeconds(end, now);

  if (seconds <= 0) {
    return 'Ended';
  }

  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (days > 0) {
    return `${days}d ${hours}h`;
  }
  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  }
  if (minutes > 0) {
    return `${minutes}m ${secs}s`;
  }
  return `${secs}s`;
}

export function getTimeLeft(endTime: string | Date): {
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
  total: number;
} {
  const end = typeof endTime === 'string' ? new Date(endTime) : endTime;
  const now = new Date();
  const total = Math.max(0, differenceInSeconds(end, now));

  return {
    days: Math.floor(total / 86400),
    hours: Math.floor((total % 86400) / 3600),
    minutes: Math.floor((total % 3600) / 60),
    seconds: total % 60,
    total,
  };
}

export function truncateText(text: string, maxLength: number): string {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength - 3) + '...';
}

export function slugify(text: string): string {
  return text
    .toLowerCase()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_-]+/g, '-')
    .replace(/^-+|-+$/g, '');
}
