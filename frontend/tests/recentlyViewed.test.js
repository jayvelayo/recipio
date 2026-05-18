import { describe, it, expect, beforeEach } from 'vitest';
import { addRecentlyViewed, useRecentlyViewed } from '../src/utils/recentlyViewed';
import { renderHook } from '@testing-library/react';

const KEY = 'recipio_recently_viewed';

beforeEach(() => {
  localStorage.clear();
});

describe('addRecentlyViewed', () => {
  it('stores a recipe in localStorage', () => {
    addRecentlyViewed({ id: '1', name: 'Pad Thai' });
    const stored = JSON.parse(localStorage.getItem(KEY));
    expect(stored).toEqual([{ id: '1', name: 'Pad Thai' }]);
  });

  it('prepends new entries so most recent is first', () => {
    addRecentlyViewed({ id: '1', name: 'Pad Thai' });
    addRecentlyViewed({ id: '2', name: 'Sinigang' });
    const stored = JSON.parse(localStorage.getItem(KEY));
    expect(stored[0]).toEqual({ id: '2', name: 'Sinigang' });
    expect(stored[1]).toEqual({ id: '1', name: 'Pad Thai' });
  });

  it('deduplicates by id and moves the entry to the front', () => {
    addRecentlyViewed({ id: '1', name: 'Pad Thai' });
    addRecentlyViewed({ id: '2', name: 'Sinigang' });
    addRecentlyViewed({ id: '1', name: 'Pad Thai' });
    const stored = JSON.parse(localStorage.getItem(KEY));
    expect(stored).toHaveLength(2);
    expect(stored[0].id).toBe('1');
  });

  it('caps the list at 10 entries', () => {
    for (let i = 0; i < 12; i++) {
      addRecentlyViewed({ id: String(i), name: `Recipe ${i}` });
    }
    const stored = JSON.parse(localStorage.getItem(KEY));
    expect(stored).toHaveLength(10);
    expect(stored[0].id).toBe('11');
  });
});

describe('useRecentlyViewed', () => {
  it('returns an empty array when localStorage is empty', () => {
    const { result } = renderHook(() => useRecentlyViewed());
    expect(result.current).toEqual([]);
  });

  it('returns existing entries from localStorage on mount', () => {
    localStorage.setItem(KEY, JSON.stringify([
      { id: '1', name: 'Pad Thai' },
      { id: '2', name: 'Sinigang' },
    ]));
    const { result } = renderHook(() => useRecentlyViewed());
    expect(result.current).toHaveLength(2);
    expect(result.current[0].id).toBe('1');
  });

  it('returns empty array when localStorage contains invalid JSON', () => {
    localStorage.setItem(KEY, 'not-json');
    const { result } = renderHook(() => useRecentlyViewed());
    expect(result.current).toEqual([]);
  });
});
