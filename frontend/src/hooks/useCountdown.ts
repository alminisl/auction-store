import { useState, useEffect, useCallback } from 'react';
import { getTimeLeft } from '../utils';

interface CountdownResult {
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
  total: number;
  isExpired: boolean;
  formatted: string;
}

export function useCountdown(endTime: string | Date): CountdownResult {
  const calculateTimeLeft = useCallback(() => {
    const timeLeft = getTimeLeft(endTime);
    const isExpired = timeLeft.total <= 0;

    let formatted = '';
    if (isExpired) {
      formatted = 'Ended';
    } else if (timeLeft.days > 0) {
      formatted = `${timeLeft.days}d ${timeLeft.hours}h ${timeLeft.minutes}m`;
    } else if (timeLeft.hours > 0) {
      formatted = `${timeLeft.hours}h ${timeLeft.minutes}m ${timeLeft.seconds}s`;
    } else if (timeLeft.minutes > 0) {
      formatted = `${timeLeft.minutes}m ${timeLeft.seconds}s`;
    } else {
      formatted = `${timeLeft.seconds}s`;
    }

    return {
      ...timeLeft,
      isExpired,
      formatted,
    };
  }, [endTime]);

  const [countdown, setCountdown] = useState<CountdownResult>(calculateTimeLeft);

  // Immediately recalculate when endTime changes
  useEffect(() => {
    setCountdown(calculateTimeLeft());
  }, [endTime, calculateTimeLeft]);

  useEffect(() => {
    const timer = setInterval(() => {
      const newCountdown = calculateTimeLeft();
      setCountdown(newCountdown);

      if (newCountdown.isExpired) {
        clearInterval(timer);
      }
    }, 1000);

    return () => clearInterval(timer);
  }, [calculateTimeLeft]);

  return countdown;
}

export default useCountdown;
