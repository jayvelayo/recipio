const API_BASE = 'http://localhost:4002';


/** GET /meal-plans: returns array of { id, recipe_names } */
export function getMealPlans() {
  const url = `${API_BASE}/meal-plans`;
  // #region agent log
  // #endregion
  return fetch(url, { mode: 'cors' })
    .then((res) => {
      // #region agent log
      // #endregion
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    })
    .then((data) => {
      // #region agent log
      // #endregion
      return data;
    })
    .catch((e) => {
      // #region agent log
      // #endregion
      throw e;
    });
}

/** POST /meal-plans: body { recipes: string[] } (recipe IDs), returns { id, message } */
export function createMealPlan(recipeIds) {
  return fetch(`${API_BASE}/meal-plans`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    mode: 'cors',
    body: JSON.stringify({ recipes: recipeIds }),
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

/** GET /grocery-list/{meal_plan_id}: returns { ingredients: string[] } */
export function getGroceryList(mealPlanId) {
  const url = `${API_BASE}/grocery-list/${mealPlanId}`;
  return fetch(url, { mode: 'cors' })
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    })
    .then((data) => {
      return data.ingredients || [];
    });
}

/** DELETE /meal-plans/{meal_plan_id}: deletes the meal plan */
export function deleteMealPlan(mealPlanId) {
  return fetch(`${API_BASE}/meal-plans/${mealPlanId}`, {
    method: 'DELETE',
    mode: 'cors',
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}
