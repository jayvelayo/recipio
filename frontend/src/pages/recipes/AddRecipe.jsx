import { useState, useEffect } from "react";
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createRecipe, parseRecipe } from "./recipe_apis";
import { useNavigate } from "react-router";
import { FiArrowLeft } from 'react-icons/fi';
import { parseIngredient } from '../../utils/parseIngredient';
import { toast } from 'sonner';

export function AddRecipeForm() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createRecipe,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries(['recipes']);
      toast.success(`"${variables.name}" added`);
      navigate("/recipe");
    },
    onError: () => {
      toast.error('Failed to save recipe');
    }
  });

  const blankRecipe = {
    name: "",
    ingredients: [],
    instructions: [],
  }
  const [recipe, setRecipe] = useState(blankRecipe);
  const [useAI, setUseAI] = useState(false);
  const [rawRecipeText, setRawRecipeText] = useState("");
  const [preview, setPreview] = useState(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState(null);
  const [progress, setProgress] = useState(0);

  useEffect(() => {
    if (!previewLoading) {
      setProgress(0);
      return;
    }
    setProgress(0);
    const interval = setInterval(() => {
      setProgress(prev => prev >= 90 ? 90 : prev + 1);
    }, 333);
    return () => clearInterval(interval);
  }, [previewLoading]);
  
  const sanitize = (text) => text.replace(/[^a-zA-Z0-9\s.,\-/()°'":À-ɏ]/g, '');

const handleFormChange = (e) => {
    if (e.target.name == "recipeName") {
      setRecipe({...recipe, name: sanitize(e.target.value)})
    }
    if (e.target.name == "ingredientsList") {
      setRecipe({...recipe, ingredients: sanitize(e.target.value).split(/\r?\n/)});
    }
    if (e.target.name == "instructions") {
      setRecipe({...recipe, instructions: sanitize(e.target.value).split(/\r?\n/)});
    }
  }
  
  const addRecipeSubmitHandler = (e) => {
    e.preventDefault();
    mutation.mutate({
      ...recipe,
      ingredients: recipe.ingredients.filter(Boolean).map(parseIngredient),
      instructions: recipe.instructions.filter(Boolean),
    });
  };

  const handlePreviewRecipe = async () => {
    setPreviewLoading(true);
    setPreviewError(null);
    setPreview(null);
    
    try {
      const parsedRecipe = await parseRecipe(rawRecipeText);
      // Normalize the recipe structure
      const normalizedRecipe = {
        name: parsedRecipe.name || "",
        ingredients: Array.isArray(parsedRecipe.ingredients) ? parsedRecipe.ingredients : [],
        instructions: Array.isArray(parsedRecipe.instructions) ? parsedRecipe.instructions : [],
      };
      setRecipe(normalizedRecipe);
      setPreview(normalizedRecipe);
    } catch (error) {
      if (error.message === 'TIMEOUT') {
        setPreviewError('Parsing timed out. Try reducing the amount of text and try again.');
      } else if (error.message === 'RATE_LIMIT') {
        setPreviewError('Too many requests. Please wait a moment before trying again.');
      } else {
        setPreviewError('Failed to parse recipe. Please check your input and try again.');
      }
    } finally {
      setPreviewLoading(false);
    }
  }

  const handleUsePreview = () => {
    setUseAI(false);
    setRawRecipeText("");
    setPreview(null);
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
        {/* AI Mode Toggle Button */}
        <div className="mb-6 flex gap-3">
          <button
            onClick={() => {
              setUseAI(false);
              setRawRecipeText("");
              setPreview(null);
              setPreviewError(null);
            }}
            type="button"
            className={`flex-1 font-medium py-2 rounded-lg transition ${
              !useAI
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-900 hover:bg-gray-300'
            }`}
          >
            Manual Entry
          </button>
          <button
            onClick={() => {
              setUseAI(true);
              setRecipe(blankRecipe);
            }}
            type="button"
            className={`flex-1 font-medium py-2 rounded-lg transition ${
              useAI
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-200 text-gray-900 hover:bg-gray-300'
            }`}
          >
            Add Recipe using AI?
          </button>
        </div>

        {/* AI Mode */}
        {useAI ? (
          <div className="space-y-6">
            <div>
              <label htmlFor="rawRecipeText" className="block text-sm font-medium text-gray-900 mb-2">
                Paste Your Recipe
              </label>
              <p className="text-xs text-gray-500 mb-2">
                Paste the full recipe text. Our AI will parse it into ingredients and instructions.
              </p>
              <textarea
                id="rawRecipeText"
                placeholder="Paste your recipe here... (e.g., '2 cups flour, 1 cup sugar, 2 eggs. Mix ingredients together. Bake at 350°F for 20 minutes.')"
                className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition font-mono text-sm"
                value={rawRecipeText}
                onChange={(e) => setRawRecipeText(sanitize(e.target.value).slice(0, 2000))}
                rows="12"
                maxLength={2000}
              />
              <p className={`text-xs mt-1 text-right ${rawRecipeText.length >= 2000 ? 'text-red-500' : 'text-gray-400'}`}>
                {rawRecipeText.length} / 2000
              </p>
            </div>

            {/* Preview Error */}
            {previewError && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <p className="text-sm text-red-700">{previewError}</p>
              </div>
            )}

            {/* Preview */}
            {preview && (
              <div className="bg-gray-50 border border-gray-200 rounded-lg p-6 space-y-4">
                <h3 className="font-semibold text-gray-900">Recipe Preview</h3>

                <div>
                  <p className="text-sm font-medium text-gray-700 mb-2">Recipe Name</p>
                  <p className="text-gray-900 font-medium">{preview.name || "(No name)"}</p>
                </div>

                <div>
                  <p className="text-sm font-medium text-gray-700 mb-2">Ingredients</p>
                  <ul className="space-y-1">
                    {preview.ingredients && preview.ingredients.length > 0 ? (
                      preview.ingredients.map((ing, idx) => (
                        <li key={idx} className="flex items-baseline gap-2 text-sm">
                          {ing.quantity && (
                            <span className="italic text-gray-400 shrink-0">{ing.quantity}</span>
                          )}
                          <span className="text-gray-900">{ing.name}</span>
                        </li>
                      ))
                    ) : (
                      <p className="text-gray-500 text-sm">(No ingredients)</p>
                    )}
                  </ul>
                </div>

                <div>
                  <p className="text-sm font-medium text-gray-700 mb-2">Instructions</p>
                  <ol className="space-y-1">
                    {preview.instructions && preview.instructions.length > 0 ? (
                      preview.instructions.map((step, idx) => (
                        <li key={idx} className="text-gray-900 text-sm">{idx + 1}. {step}</li>
                      ))
                    ) : (
                      <p className="text-gray-500 text-sm">(No instructions)</p>
                    )}
                  </ol>
                </div>

                <div className="flex gap-3 pt-4">
                  <button
                    onClick={handleUsePreview}
                    type="button"
                    className="flex-1 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 transition"
                  >
                    Use This Recipe
                  </button>
                  <button
                    onClick={() => {
                      setPreview(null);
                      setPreviewError(null);
                    }}
                    type="button"
                    className="flex-1 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition"
                  >
                    Back to Edit
                  </button>
                </div>
              </div>
            )}

            {/* Preview Button */}
            {!preview && (
              <div className="space-y-2">
                <button
                  onClick={handlePreviewRecipe}
                  disabled={previewLoading || !rawRecipeText.trim()}
                  type="button"
                  className="w-full bg-blue-600 text-white font-medium py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition"
                >
                  {previewLoading ? 'Parsing Recipe...' : 'Preview Recipe'}
                </button>
                {previewLoading && (
                  <div className="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
                    <div
                      className="bg-blue-500 h-2 rounded-full transition-all duration-300 ease-out"
                      style={{ width: `${progress}%` }}
                    />
                  </div>
                )}
              </div>
            )}
          </div>
        ) : (
          /* Manual Mode Form */
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
              <label htmlFor="steps" className="block text-sm font-medium text-gray-900 mb-2">
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
        )}
      </div>
    </div>
  )
}
