const AMOUNT = String.raw`(?:[¼½¾⅓⅔⅛⅜⅝⅞]|\d+(?:[./]\d+)?(?:\s+\d+\/\d+)?)`;
const UNIT = String.raw`(?:cups?|tbsps?|tablespoons?|tsps?|teaspoons?|fl\.?\s*oz|oz|ounces?|lbs?|pounds?|grams?|g|kg|ml|milliliters?|liters?|litres?|l|cloves?|heads?|stalks?|sprigs?|slices?|cans?|packages?|pkgs?|bunches?|handfuls?|pinch(?:es)?|dashes?|pieces?|pcs?)`;
const INGREDIENT_RE = new RegExp(`^(${AMOUNT})\\s*(?:(${UNIT})\\b[,.]?\\s*)?(.*)$`, 'i');

export function parseIngredient(line) {
  line = line.trim();
  if (!line) return { name: '', quantity: '' };
  const m = line.match(INGREDIENT_RE);
  if (!m) return { name: line, quantity: '' };
  const amount = m[1]?.trim() ?? '';
  const unit = m[2]?.trim() ?? '';
  const name = m[3]?.trim() || line;
  return { quantity: [amount, unit].filter(Boolean).join(' '), name };
}
