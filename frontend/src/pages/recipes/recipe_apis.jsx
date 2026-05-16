export function getRecipes() {
  return fetch('/recipes')
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function getRecipeId(id) {
  return fetch(`/recipes/${id}`)
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
  };
  return fetch('/recipes', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  }).then(res => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function parseRecipe(rawRecipeText) {
  return fetch('/parse-recipe', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
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
  };
  return fetch(`/recipes/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  }).then(res => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}

export function deleteRecipe(id) {
  return fetch(`/recipes/${id}`, { method: 'DELETE' })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
    });
}
