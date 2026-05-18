import { useState, useEffect } from 'react';

const KEY = 'recipio_recently_viewed';
const EVENT = 'recipio_recently_viewed_updated';
const MAX = 10;

function read() {
  try { return JSON.parse(localStorage.getItem(KEY)) ?? []; }
  catch { return []; }
}

export function addRecentlyViewed({ id, name }) {
  const next = [{ id, name }, ...read().filter(r => r.id !== id)].slice(0, MAX);
  localStorage.setItem(KEY, JSON.stringify(next));
  window.dispatchEvent(new Event(EVENT));
}

export function useRecentlyViewed() {
  const [items, setItems] = useState(read);
  useEffect(() => {
    const handler = () => setItems(read());
    window.addEventListener(EVENT, handler);
    return () => window.removeEventListener(EVENT, handler);
  }, []);
  return items;
}
