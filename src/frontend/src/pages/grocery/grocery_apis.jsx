import { API_BASE } from '../../apiConfig';

export function getGroceryLists() {
  const url = `${API_BASE}/grocery-lists`;
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

export function getGroceryListById(id) {
  const url = `${API_BASE}/grocery-lists/${id}`;
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

export function createGroceryList(name, items, mealPlanId = null) {
  return fetch(`${API_BASE}/grocery-lists`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    mode: 'cors',
    body: JSON.stringify({ name, items, meal_plan_id: mealPlanId }),
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return res.json();
  });
}

export function updateGroceryList(id, items) {
  return fetch(`${API_BASE}/grocery-lists/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    mode: 'cors',
    body: JSON.stringify(items),
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}

export function deleteGroceryList(id) {
  return fetch(`${API_BASE}/grocery-lists/${id}`, {
    method: 'DELETE',
    mode: 'cors',
  }).then((res) => {
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
  });
}