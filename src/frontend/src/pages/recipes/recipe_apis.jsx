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
