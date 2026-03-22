const API_BASE = 'http://localhost:4002';

function debugLog(location, message, data, hypothesisId) {
  const payload = { sessionId: '95486f', location, message, data, timestamp: Date.now(), hypothesisId };
  fetch('http://127.0.0.1:7895/ingest/ec1257c3-4d82-4824-880b-7f61561359be', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Debug-Session-Id': '95486f' },
    body: JSON.stringify(payload),
  }).catch(() => {});
}

/** GET /meal-plans: returns array of { id, recipe_names } */
export function getMealPlans() {
  const url = `${API_BASE}/meal-plans`;
  // #region agent log
  debugLog('mealplan_apis.jsx:getMealPlans', 'fetch start', { url }, 'H2');
  // #endregion
  return fetch(url, { mode: 'cors' })
    .then((res) => {
      // #region agent log
      debugLog('mealplan_apis.jsx:getMealPlans', 'fetch response', { status: res.status, ok: res.ok }, 'H1,H3');
      // #endregion
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      return res.json();
    })
    .then((data) => {
      // #region agent log
      debugLog('mealplan_apis.jsx:getMealPlans', 'parse ok', { dataLength: Array.isArray(data) ? data.length : 'not-array' }, 'H5');
      // #endregion
      return data;
    })
    .catch((e) => {
      // #region agent log
      debugLog('mealplan_apis.jsx:getMealPlans', 'fetch/reject', { name: e?.name, message: e?.message }, 'H1,H2,H3,H5');
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
