// Mirrors backend/internal/recipes/categories.go.
// Order here drives section order in the grocery list UI.
export const CATEGORY_ORDER = [
  'Baking Items',
  'Bread & Cereals',
  'Canned Food & Soups',
  'Dairy, Eggs & Cheese',
  'Deli',
  'Meat',
  'Sauces & Condiments',
  'Seafood',
  'Snack & Candy',
  'Spices & Seasonings',
  'Fruits',
  'Vegetable',
  'Frozen Food',
  'Uncategorized',
];

// groupByCategory takes a list of items (each with a `.category` field) and
// returns [{ category, items: [{item, originalIndex}, ...] }, ...] in CATEGORY_ORDER,
// skipping categories with no items. originalIndex is the position in the input
// array — callers that mutate by index (check/delete) need the un-grouped index.
export function groupByCategory(items) {
  const buckets = new Map();
  items.forEach((item, originalIndex) => {
    const cat = item.category || 'Uncategorized';
    if (!buckets.has(cat)) buckets.set(cat, []);
    buckets.get(cat).push({ item, originalIndex });
  });
  return CATEGORY_ORDER
    .filter(cat => buckets.has(cat))
    .map(cat => ({ category: cat, items: buckets.get(cat) }));
}
