import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router';
import { getRecipes } from '../recipes/recipe_apis';
import { createMealPlan } from './mealplan_apis';
import LoadingPage from '/src/pages/common/LoadingPage';

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
  if (recipesError) return <p>Error loading recipes: {recipesError.message}</p>;

  const sortedRecipes = [...(recipes || [])].sort((a, b) =>
    (a.name || '').localeCompare(b.name || '')
  );

  return (
    <>
      <h2>Add meal plan</h2>
      <p>Select the recipes you want in this meal plan.</p>
      <form className="ui form" onSubmit={handleSubmit}>
        <div className="ui relaxed divided list">
          {sortedRecipes.map((recipe) => {
            const idStr = String(recipe.id);
            const checked = selectedIds.has(idStr);
            return (
              <div className="item" key={recipe.id}>
                <div className="ui checkbox">
                  <input
                    type="checkbox"
                    id={`recipe-${recipe.id}`}
                    checked={checked}
                    onChange={() => toggleRecipe(recipe.id)}
                  />
                  <label htmlFor={`recipe-${recipe.id}`}>{recipe.name}</label>
                </div>
              </div>
            );
          })}
        </div>
        {sortedRecipes.length === 0 && (
          <p>No recipes available. Add recipes first from the Recipes page.</p>
        )}
        <div className="ui segment">
          <button
            type="submit"
            className="ui button primary"
            disabled={selectedIds.size === 0 || mutation.isPending}
          >
            {mutation.isPending ? 'Creating…' : 'Create meal plan'}
          </button>
          <button
            type="button"
            className="ui button"
            onClick={() => navigate('/mealplan')}
          >
            Cancel
          </button>
        </div>
      </form>
    </>
  );
}
