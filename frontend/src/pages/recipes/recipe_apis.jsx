function authHeaders() {
  const session = JSON.parse(localStorage.getItem('session'));
  return session?.token ? { Authorization: `Bearer ${session.token}` } : {};
}

export function getRecipes() {
  return fetch('/recipes', { headers: authHeaders() })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function getRecipeId(id) {
  return fetch(`/recipes/${id}`, { headers: authHeaders() })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function createRecipe(newRecipe) {
  const body = {
    name: newRecipe.name,
    ingredients: Array.isArray(newRecipe.ingredients) ? newRecipe.ingredients : [],
    instructions: Array.isArray(newRecipe.instructions) ? newRecipe.instructions : [],
    tags: Array.isArray(newRecipe.tags) ? newRecipe.tags : [],
    source_url: newRecipe.source_url ?? '',
  };
  return fetch('/recipes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(body),
  }).then(res => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function parseRecipeURL(url) {
  return fetch('/parse-recipe-url', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ url }),
  }).then(res => {
    if (res.status === 429) throw new Error('RATE_LIMIT');
    if (res.status === 504) throw new Error('TIMEOUT');
    if (res.status === 502) throw new Error('FETCH_FAILED');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function parseRecipe(rawRecipeText) {
  return fetch('/parse-recipe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ raw_recipe_text: rawRecipeText }),
  }).then(res => {
    if (res.status === 429) throw new Error('RATE_LIMIT');
    if (res.status === 504) throw new Error('TIMEOUT');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function updateRecipe({ id, recipe }) {
  const body = {
    name: recipe.name,
    ingredients: Array.isArray(recipe.ingredients) ? recipe.ingredients : [],
    instructions: Array.isArray(recipe.instructions) ? recipe.instructions : [],
    tags: Array.isArray(recipe.tags) ? recipe.tags : [],
    source_url: recipe.source_url ?? '',
    is_favourite: recipe.is_favourite ?? false,
  };
  return fetch(`/recipes/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify(body),
  }).then(res => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}

export function deleteRecipe(id) {
  return fetch(`/recipes/${id}`, { method: 'DELETE', headers: authHeaders() })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
    });
}
