import { API_BASE } from '../../apiConfig';


export function getMealPlans() {
  const url = `${API_BASE}/meal-plans`;
  return fetch(url, { mode: 'cors' })
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    })
    .then((data) => {
      return data;
    })
    .catch((e) => {
      throw e;
    });
}

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

export function deleteMealPlan(mealPlanId) {
  return fetch(`${API_BASE}/meal-plans/${mealPlanId}`, {
    method: 'DELETE',
    mode: 'cors',
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}
