import { useState } from "react";
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createRecipe } from "./recipe_apis";
import { useNavigate } from "react-router";
import { FiArrowLeft } from 'react-icons/fi';

export function AddRecipeForm() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createRecipe,
    onSuccess: () => {
      queryClient.invalidateQueries(['recipes']);
      navigate("/recipe");
    },
    onError: () => {
      alert("Failed to insert recipe")
    }
  });

  const blankRecipe = {
    name: "",
    ingredients: [],
    instructions: [],
  }
  const [recipe, setRecipe] = useState(blankRecipe);
  
  const handleFormChange = (e) => {
    if (e.target.name == "recipeName") {
      setRecipe({...recipe, name: e.target.value})
    }
    if (e.target.name == "ingredientsList") {
      setRecipe({...recipe, ingredients: e.target.value.split(/\r?\n/)});
    }
    if (e.target.name == "instructions") {
      setRecipe({...recipe, instructions: e.target.value.split(/\r?\n/)});
    }
  }
  
  const addRecipeSubmitHandler = (e) => {
    e.preventDefault();
    mutation.mutate(recipe);
  }

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6 flex items-center gap-4">
        <button 
          onClick={() => navigate("/recipe")} 
          className="p-2 hover:bg-gray-100 rounded-lg transition"
          title="Back to recipes"
        >
          <FiArrowLeft size={24} />
        </button>
        <h1 className="text-3xl font-bold text-gray-900">Create New Recipe</h1>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-8">
        <form onSubmit={addRecipeSubmitHandler} className="space-y-6">
          {/* Recipe Name */}
          <div>
            <label htmlFor="recipeName" className="block text-sm font-medium text-gray-900 mb-2">
              Recipe Name
            </label>
            <input 
              id="recipeName"
              type="text"
              placeholder="e.g., Chocolate Chip Cookies"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
              name="recipeName"
              value={recipe.name}
              onChange={handleFormChange}
              required
            />
          </div>

          {/* Ingredients */}
          <div>
            <label htmlFor="ingredientsList" className="block text-sm font-medium text-gray-900 mb-2">
              Ingredients
            </label>
            <p className="text-xs text-gray-500 mb-2">One ingredient per line</p>
            <textarea
              id="ingredientsList"
              placeholder="2 cups flour&#10;1 cup sugar&#10;2 eggs"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition font-mono text-sm"
              name="ingredientsList"
              onChange={handleFormChange}
              value={recipe.ingredients.join('\r\n')}
              rows="6"
              required
            />
          </div>

          {/* Instructions */}
          <div>
            <label htmlFor="instructions" className="block text-sm font-medium text-gray-900 mb-2">
              Instructions
            </label>
            <p className="text-xs text-gray-500 mb-2">One step per line</p>
            <textarea
              id="instructions"
              placeholder="Preheat oven to 350°F&#10;Mix ingredients together&#10;Bake for 12 minutes"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition font-mono text-sm"
              name="instructions"
              onChange={handleFormChange}
              value={recipe.instructions.join('\r\n')}
              rows="6"
              required
            />
          </div>

          {/* Form Actions */}
          <div className="flex gap-3 pt-4">
            <button 
              type="submit" 
              disabled={mutation.isPending}
              className="flex-1 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition"
            >
              {mutation.isPending ? 'Saving...' : 'Save Recipe'}
            </button>
            <button 
              onClick={() => navigate("/recipe")} 
              type="button"
              className="flex-1 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
