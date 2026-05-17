function authHeaders() {
  const session = JSON.parse(localStorage.getItem('session'));
  return session?.token ? { Authorization: `Bearer ${session.token}` } : {};
}

export function getMealPlans() {
  return fetch('/meal-plans', { headers: authHeaders() })
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function createMealPlan(recipeIds) {
  return fetch('/meal-plans', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', ...authHeaders() },
    body: JSON.stringify({ recipes: recipeIds }),
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function getGroceryList(mealPlanId) {
  return fetch(`/grocery-list/${mealPlanId}`, { headers: authHeaders() })
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    })
    .then((data) => data.ingredients || []);
}

export function deleteMealPlan(mealPlanId) {
  return fetch(`/meal-plans/${mealPlanId}`, { method: 'DELETE', headers: authHeaders() })
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
    });
}
