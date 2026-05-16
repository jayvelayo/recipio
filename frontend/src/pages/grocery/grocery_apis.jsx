export function getGroceryLists() {
  return fetch('/grocery-lists')
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function getGroceryListById(id) {
  return fetch(`/grocery-lists/${id}`)
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    });
}

export function createGroceryList(name, items, mealPlanId = null) {
  return fetch('/grocery-lists', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, items, meal_plan_id: mealPlanId }),
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function updateGroceryList(id, items) {
  return fetch(`/grocery-lists/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(items),
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}

export function deleteGroceryList(id) {
  return fetch(`/grocery-lists/${id}`, { method: 'DELETE' })
    .then((res) => {
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
    });
}
