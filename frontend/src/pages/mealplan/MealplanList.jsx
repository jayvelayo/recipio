import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router';
import { getMealPlans, getGroceryList, deleteMealPlan } from './mealplan_apis';
import LoadingPage from '/src/pages/common/LoadingPage';
import { FiChevronDown, FiChevronUp, FiTrash2, FiShoppingCart, FiPlus } from 'react-icons/fi';

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
    <div className="border-b border-gray-200 last:border-b-0">
      <div className="p-4 hover:bg-gray-50 transition">
        <div className="flex items-start justify-between gap-4">
          <div className="flex-1">
            <h3 className="font-semibold text-gray-900">Meal Plan #{plan.id}</h3>
            <p className="text-sm text-gray-600 mt-1">{names}</p>
          </div>
          <button
            onClick={() => setShowIngredients(!showIngredients)}
            className="p-2 hover:bg-gray-200 rounded transition flex-shrink-0"
            title={showIngredients ? 'Hide ingredients' : 'Show ingredients'}
          >
            {showIngredients ? <FiChevronUp size={20} /> : <FiChevronDown size={20} />}
          </button>
        </div>

        {showIngredients && (
          <div className="mt-4 p-4 bg-gray-50 rounded-lg space-y-2">
            {loadingIngredients ? (
              <p className="text-sm text-gray-500">Loading ingredients...</p>
            ) : ingredients?.length ? (
              <>
                <p className="text-sm font-medium text-gray-900 mb-2">Ingredients needed:</p>
                {ingredients.map((ing, idx) => (
                  <div key={idx} className="text-sm text-gray-700 py-1">
                    • {ing}
                  </div>
                ))}
              </>
            ) : (
              <p className="text-sm text-gray-500">No ingredients found.</p>
            )}
          </div>
        )}

        <div className="flex flex-wrap gap-2 mt-4">
          <Link
            to="/grocery/add"
            className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition"
          >
            <FiShoppingCart size={16} />
            Create Grocery List
          </Link>
          <button
            className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-red-600 hover:bg-red-50 rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
            onClick={() => {
              if (confirm('Are you sure you want to delete this meal plan?')) {
                deleteMutation.mutate();
              }
            }}
            disabled={deleteMutation.isPending}
          >
            <FiTrash2 size={16} />
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </button>
        </div>
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
  if (error) return <p className="text-red-600">Error: {error.message}</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Meal Plans</h1>
        <Link to="/mealplan/add" className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition">
          <FiPlus size={20} />
          New Meal Plan
        </Link>
      </div>

      {data?.length ? (
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
          {data.map((plan) => <MealplanRow plan={plan} key={plan.id} />)}
        </div>
      ) : (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <p className="text-gray-500 mb-4">No meal plans yet. Create your first meal plan!</p>
        </div>
      )}
    </div>
  );
}
