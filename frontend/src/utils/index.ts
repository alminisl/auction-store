export * from './formatters';
export * from './cn';

export const AUCTION_CONDITIONS = [
  { value: 'new', label: 'New' },
  { value: 'like_new', label: 'Like New' },
  { value: 'good', label: 'Good' },
  { value: 'fair', label: 'Fair' },
  { value: 'poor', label: 'Poor' },
] as const;

export const AUCTION_STATUSES = [
  { value: 'draft', label: 'Draft' },
  { value: 'active', label: 'Active' },
  { value: 'completed', label: 'Completed' },
  { value: 'cancelled', label: 'Cancelled' },
  { value: 'unsold', label: 'Unsold' },
] as const;

export function getConditionLabel(condition: string): string {
  const found = AUCTION_CONDITIONS.find((c) => c.value === condition);
  return found?.label || condition;
}

export function getStatusLabel(status: string): string {
  const found = AUCTION_STATUSES.find((s) => s.value === status);
  return found?.label || status;
}

export function getStatusColor(status: string): string {
  switch (status) {
    case 'active':
      return 'text-green-600 bg-green-100';
    case 'completed':
      return 'text-blue-600 bg-blue-100';
    case 'cancelled':
      return 'text-red-600 bg-red-100';
    case 'unsold':
      return 'text-yellow-600 bg-yellow-100';
    case 'draft':
    default:
      return 'text-gray-600 bg-gray-100';
  }
}
