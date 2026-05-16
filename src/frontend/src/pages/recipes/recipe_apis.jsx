import { API_BASE } from '../../apiConfig';

export function getRecipes() {
  return fetch(`${API_BASE}/recipes`, { mode: "cors" })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function getRecipeId(id) {
  return fetch(`${API_BASE}/recipes/${id}`, { mode: "cors" })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function createRecipe(newRecipe) {
  const body = {
    name: newRecipe.name,
    ingredients: Array.isArray(newRecipe.ingredients) ? newRecipe.ingredients : [],
    steps: Array.isArray(newRecipe.steps) ? newRecipe.steps : [],
  };
  return fetch(`${API_BASE}/recipes`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    mode: "cors",
    body: JSON.stringify(body),
  }).then(res => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function parseRecipe(rawRecipeText) {
  return fetch(`${API_BASE}/parse-recipe`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    mode: "cors",
    body: JSON.stringify({ raw_recipe_text: rawRecipeText }),
  }).then(res => {
    if (res.status === 429) throw new Error('RATE_LIMIT');
    if (res.status === 504) throw new Error('TIMEOUT');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function deleteRecipe(id) {
  return fetch(`${API_BASE}/recipes/${id}`, {
    method: 'DELETE',
    mode: "cors"
  }).then(res => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}
