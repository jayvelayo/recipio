const API_BASE = 'http://localhost:4002';

// Design API: GET /recipes returns array of { id, name, ingredients, steps }
export function getRecipes() {
  return fetch(`${API_BASE}/recipes`, { mode: "cors" })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

// Design API: GET /recipes/{id} returns single { id, name, ingredients, steps }
export function getRecipeId(id) {
  return fetch(`${API_BASE}/recipes/${id}`, { mode: "cors" })
    .then(res => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

// Design API: POST /recipes body { name, ingredients, steps }, response { id, message }
export function createRecipe(newRecipe) {
  const body = {
    name: newRecipe.name,
    ingredients: Array.isArray(newRecipe.ingredients) ? newRecipe.ingredients : [],
    steps: Array.isArray(newRecipe.instructions) ? newRecipe.instructions : (Array.isArray(newRecipe.steps) ? newRecipe.steps : []),
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

// DELETE not in design API; keep using v1 for now
export function deleteRecipe(id) {
  return fetch(`${API_BASE}/v1/recipe/${id}`, {
    method: 'DELETE',
    mode: "cors"
  }).then(res => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}
