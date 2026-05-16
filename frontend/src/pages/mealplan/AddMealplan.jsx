import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router';
import { getRecipes } from '../recipes/recipe_apis';
import { createMealPlan } from './mealplan_apis';
import LoadingPage from '/src/pages/common/LoadingPage';
import { FiArrowLeft, FiCheck } from 'react-icons/fi';

export function AddMealplan() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [selectedIds, setSelectedIds] = useState(new Set());

  const { data: recipes, isLoading: recipesLoading, error: recipesError } = useQuery({
    queryKey: ['recipes'],
    queryFn: getRecipes,
  });

  const mutation = useMutation({
    mutationFn: createMealPlan,
    onSuccess: () => {
      queryClient.invalidateQueries(['mealplans']);
      navigate('/mealplan');
    },
    onError: (err) => {
      alert(`Failed to create meal plan: ${err.message}`);
    },
  });

  const toggleRecipe = (id) => {
    const idStr = String(id);
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(idStr)) next.delete(idStr);
      else next.add(idStr);
      return next;
    });
  };

  const handleSubmit = (e) => {
    e.preventDefault();
    const ids = Array.from(selectedIds);
    if (ids.length === 0) {
      alert('Select at least one recipe.');
      return;
    }
    mutation.mutate(ids);
  };

  if (recipesLoading) return <LoadingPage />;
  if (recipesError) return <p className="text-red-600">Error loading recipes: {recipesError.message}</p>;

  const sortedRecipes = [...(recipes || [])].sort((a, b) =>
    (a.name || '').localeCompare(b.name || '')
  );

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6 flex items-center gap-4">
        <button 
          onClick={() => navigate("/mealplan")} 
          className="p-2 hover:bg-gray-100 rounded-lg transition"
          title="Back to meal plans"
        >
          <FiArrowLeft size={24} />
        </button>
        <h1 className="text-3xl font-bold text-gray-900">Create Meal Plan</h1>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-8">
        <p className="text-gray-600 mb-6">Select the recipes you want in this meal plan.</p>

        <form onSubmit={handleSubmit} className="space-y-6">
          {/* Recipe List */}
          <div className="space-y-3">
            {sortedRecipes.length > 0 ? (
              sortedRecipes.map((recipe) => {
                const idStr = String(recipe.id);
                const checked = selectedIds.has(idStr);
                return (
                  <label 
                    key={recipe.id} 
                    className="flex items-center p-4 border border-gray-200 rounded-lg cursor-pointer hover:bg-gray-50 transition"
                  >
                    <input
                      type="checkbox"
                      id={`recipe-${recipe.id}`}
                      checked={checked}
                      onChange={() => toggleRecipe(recipe.id)}
                      className="w-5 h-5 text-indigo-600 rounded cursor-pointer"
                    />
                    <div className="ml-3 flex-1">
                      <p className="font-medium text-gray-900">{recipe.name}</p>
                    </div>
                    {checked && <FiCheck className="text-indigo-600" size={20} />}
                  </label>
                );
              })
            ) : (
              <p className="text-gray-500 text-center py-8">No recipes available. Add recipes first from the Recipes page.</p>
            )}
          </div>

          {/* Form Actions */}
          <div className="flex gap-3 pt-6 border-t border-gray-200">
            <button
              type="submit"
              disabled={selectedIds.size === 0 || mutation.isPending}
              className="flex-1 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition"
            >
              {mutation.isPending ? 'Creating...' : 'Create Meal Plan'}
            </button>
            <button
              type="button"
              onClick={() => navigate('/mealplan')}
              className="flex-1 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
