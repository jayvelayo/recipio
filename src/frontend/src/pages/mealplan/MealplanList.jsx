import { useQuery } from '@tanstack/react-query';
import { Link } from 'react-router';
import { getMealPlans } from './mealplan_apis';
import LoadingPage from '/src/pages/common/LoadingPage';

function debugLog(location, message, data, hypothesisId) {
  const payload = { sessionId: '95486f', location, message, data, timestamp: Date.now(), hypothesisId };
  fetch('http://127.0.0.1:7895/ingest/ec1257c3-4d82-4824-880b-7f61561359be', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json', 'X-Debug-Session-Id': '95486f' },
    body: JSON.stringify(payload),
  }).catch(() => {});
}

function MealplanRow({ plan }) {
  const names = plan.recipe_names?.length ? plan.recipe_names.join(', ') : 'No recipes';
  return (
    <div className="item">
      <div className="content">
        <span className="header">Meal plan #{plan.id}</span>
        <div className="description">{names}</div>
      </div>
    </div>
  );
}

export function MealplanList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['mealplans'],
    queryFn: getMealPlans,
  });
  // #region agent log
  debugLog('MealplanList.jsx:MealplanList', 'render state', {
    isLoading,
    errorMessage: error?.message ?? null,
    dataLength: data != null && Array.isArray(data) ? data.length : (data != null ? 'not-array' : null),
  }, 'H5');
  // #endregion

  if (isLoading) return <LoadingPage />;
  if (error) return <p>Error: {error.message}</p>;

  return (
    <>
      <h2>Meal plans</h2>
      <div className="ui relaxed divided list">
        {data?.length ? (
          data.map((plan) => <MealplanRow plan={plan} key={plan.id} />)
        ) : (
          <p>No meal plans yet. Create one to get started.</p>
        )}
      </div>
      <Link to="/mealplan/add" className="ui button primary">
        Add meal plan
      </Link>
    </>
  );
}
