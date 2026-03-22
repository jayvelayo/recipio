import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router';
import { getMealPlans, getGroceryList, deleteMealPlan } from './mealplan_apis';
import LoadingPage from '/src/pages/common/LoadingPage';

function MealplanRow({ plan }) {
  const [showIngredients, setShowIngredients] = React.useState(false);
  const queryClient = useQueryClient();

  const { data: ingredients, isLoading: loadingIngredients } = useQuery({
    queryKey: ['grocery', plan.id],
    queryFn: () => getGroceryList(plan.id),
    enabled: showIngredients,
  });

  const deleteMutation = useMutation({
    mutationFn: () => deleteMealPlan(plan.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['mealplans'] });
    },
  });

  const names = plan.recipe_names?.length ? plan.recipe_names.join(', ') : 'No recipes';

  return (
    <div className="item">
      <div className="content">
        <span className="header">Meal plan #{plan.id}</span>
        <div className="description">{names}</div>
        <div className="extra">
          <button
            className="ui button small"
            onClick={() => setShowIngredients(!showIngredients)}
          >
            {showIngredients ? 'Hide' : 'View'} Ingredients
          </button>
          <Link
            to="/grocery/add"
            className="ui button small blue"
          >
            Create Grocery List
          </Link>
          <button
            className="ui button small red"
            onClick={() => {
              if (confirm('Are you sure you want to delete this meal plan?')) {
                deleteMutation.mutate();
              }
            }}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </button>
        </div>
        {showIngredients && (
          <div className="ui list" style={{ marginTop: '10px' }}>
            {loadingIngredients ? (
              <div>Loading ingredients...</div>
            ) : ingredients?.length ? (
              ingredients.map((ing, idx) => (
                <div key={idx} className="item">
                  {ing}
                </div>
              ))
            ) : (
              <div>No ingredients found.</div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}

export function MealplanList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['mealplans'],
    queryFn: getMealPlans,
  });

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
