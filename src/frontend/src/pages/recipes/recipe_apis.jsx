export function getRecipes() {
  return fetch('http://localhost:4002/v1/recipe', { mode: "cors"})
    .then(res => res.json());
}

export function createRecipe(newRecipe) {
  return fetch('http://localhost:4002/v1/recipe', { 
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    mode: "cors",
    body: JSON.stringify(newRecipe),
  }).then(res => res.json());
}

export function deleteRecipe(id) {
  return fetch(`http://localhost:4002/v1/recipe/${id}`, { 
    method: 'DELETE',
    mode: "cors"
  }).then( res => {
    if (!res.ok) {
      throw new Error(`http error status: ${res.status}`)
    }
    return;
  });
}

export function getRecipeId(id) {
  return fetch(`http://localhost:4002/v1/recipe/${id}`, { 
    method: 'GET',
    mode: "cors"
  }).then( res => {
    if (!res.ok) {
      throw new Error(`http error status: ${res.status}`)
    }
    return res.json();
  });
}
