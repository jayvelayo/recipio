import { describe, it, expect } from 'vitest';
import { parseIngredient } from '../src/utils/parseIngredient';

describe('parseIngredient', () => {
  it('parses amount + unit + multi-word name', () => {
    expect(parseIngredient('2 cups all-purpose flour')).toEqual({ quantity: '2 cups', name: 'all-purpose flour' });
  });

  it('parses fractional amount', () => {
    expect(parseIngredient('1/2 tsp salt')).toEqual({ quantity: '1/2 tsp', name: 'salt' });
  });

  it('parses unicode fraction', () => {
    expect(parseIngredient('½ cup butter')).toEqual({ quantity: '½ cup', name: 'butter' });
  });

  it('parses decimal amount with kg', () => {
    expect(parseIngredient('1.5 kg chicken breast')).toEqual({ quantity: '1.5 kg', name: 'chicken breast' });
  });

  it('parses amount without a known unit (count)', () => {
    expect(parseIngredient('3 large eggs')).toEqual({ quantity: '3', name: 'large eggs' });
  });

  it('parses no-amount ingredient', () => {
    expect(parseIngredient('vanilla extract')).toEqual({ quantity: '', name: 'vanilla extract' });
  });

  it('parses "salt to taste" as no quantity', () => {
    expect(parseIngredient('salt to taste')).toEqual({ quantity: '', name: 'salt to taste' });
  });

  it('parses amount + can', () => {
    expect(parseIngredient('1 can diced tomatoes')).toEqual({ quantity: '1 can', name: 'diced tomatoes' });
  });

  it('parses tablespoon abbreviation', () => {
    expect(parseIngredient('2 tbsp olive oil')).toEqual({ quantity: '2 tbsp', name: 'olive oil' });
  });

  it('parses cloves', () => {
    expect(parseIngredient('4 cloves garlic')).toEqual({ quantity: '4 cloves', name: 'garlic' });
  });

  it('parses pinch', () => {
    expect(parseIngredient('1 pinch cayenne')).toEqual({ quantity: '1 pinch', name: 'cayenne' });
  });

  it('parses compact notation without space before unit (100g)', () => {
    expect(parseIngredient('100g dark chocolate')).toEqual({ quantity: '100 g', name: 'dark chocolate' });
  });

  it('does not confuse "l" in "large" with liters', () => {
    const result = parseIngredient('3 large eggs');
    expect(result.name).toBe('large eggs');
    expect(result.quantity).toBe('3');
  });

  it('returns empty name and quantity for empty string', () => {
    expect(parseIngredient('')).toEqual({ quantity: '', name: '' });
  });

  it('returns full line as name when no amount is present', () => {
    expect(parseIngredient('bay leaf')).toEqual({ quantity: '', name: 'bay leaf' });
  });
});
