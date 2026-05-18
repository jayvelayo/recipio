import { useState, useEffect, useRef, useCallback } from "react";
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createRecipe, parseRecipe, parseRecipeURL } from "./recipe_apis";
import { useNavigate, useBlocker } from "react-router";
import { FiArrowLeft } from 'react-icons/fi';
import { parseIngredient } from '../../utils/parseIngredient';
import { toast } from 'sonner';
import { TagInput } from '../common/TagInput';

export function AddRecipeForm() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const skipBlockRef = useRef(false);

  const mutation = useMutation({
    mutationFn: createRecipe,
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries(['recipes']);
      toast.success(`"${variables.name}" added`);
      skipBlockRef.current = true;
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
    tags: [],
    source_url: "",
  }
  const [recipe, setRecipe] = useState(blankRecipe);
  const [mode, setMode] = useState('manual'); // 'manual' | 'ai' | 'url'
  const [rawRecipeText, setRawRecipeText] = useState("");
  const [recipeURL, setRecipeURL] = useState("");
  const [preview, setPreview] = useState(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [previewError, setPreviewError] = useState(null);
  const [progress, setProgress] = useState(0);

  const isDirty =
    recipe.name.trim() !== "" ||
    recipe.ingredients.some(i => (typeof i === 'string' ? i : i.name ?? '').trim()) ||
    recipe.instructions.some(Boolean) ||
    recipe.tags.length > 0 ||
    recipe.source_url.trim() !== "" ||
    rawRecipeText.trim() !== "" ||
    recipeURL.trim() !== "";

  const isDirtyRef = useRef(isDirty);
  isDirtyRef.current = isDirty;

  const blockerFn = useCallback(
    () => isDirtyRef.current && !skipBlockRef.current,
    []
  );
  const blocker = useBlocker(blockerFn);

  useEffect(() => {
    const onBeforeUnload = (e) => {
      if (isDirtyRef.current && !skipBlockRef.current) {
        e.preventDefault();
        e.returnValue = '';
      }
    };
    window.addEventListener('beforeunload', onBeforeUnload);
    return () => window.removeEventListener('beforeunload', onBeforeUnload);
  }, []);

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
      ingredients: recipe.ingredients.filter(Boolean).map(ing =>
        typeof ing === 'string' ? parseIngredient(ing) : ing
      ),
      instructions: recipe.instructions.filter(Boolean),
      tags: recipe.tags,
      source_url: recipe.source_url.trim(),
    });
  };

  const applyParsedRecipe = (parsedRecipe, sourceURL = "") => {
    const normalized = {
      name: parsedRecipe.name || "",
      ingredients: Array.isArray(parsedRecipe.ingredients) ? parsedRecipe.ingredients : [],
      instructions: Array.isArray(parsedRecipe.instructions) ? parsedRecipe.instructions : [],
      tags: Array.isArray(parsedRecipe.tags) ? parsedRecipe.tags : [],
      source_url: sourceURL,
    };
    setRecipe(normalized);
    setPreview(normalized);
  };

  const handlePreviewRecipe = async () => {
    setPreviewLoading(true);
    setPreviewError(null);
    setPreview(null);
    try {
      applyParsedRecipe(await parseRecipe(rawRecipeText));
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
  };

  const handlePreviewFromURL = async () => {
    setPreviewLoading(true);
    setPreviewError(null);
    setPreview(null);
    try {
      applyParsedRecipe(await parseRecipeURL(recipeURL), recipeURL);
    } catch (error) {
      if (error.message === 'TIMEOUT') {
        setPreviewError('Request timed out. The page may be too slow or large.');
      } else if (error.message === 'RATE_LIMIT') {
        setPreviewError('Too many requests. Please wait a moment before trying again.');
      } else if (error.message === 'FETCH_FAILED') {
        setPreviewError('Could not fetch the page. The site may be blocking requests or the URL is invalid.');
      } else {
        setPreviewError('Failed to extract recipe from URL. Try copying the recipe text and using the AI tab instead.');
      }
    } finally {
      setPreviewLoading(false);
    }
  };

  const handleUsePreview = () => {
    setMode('manual');
    setRawRecipeText("");
    setRecipeURL("");
    setPreview(null);
  };

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
        {/* Mode Toggle */}
        <div className="mb-6 flex gap-2">
          {[['manual', 'Manual'], ['ai', 'Paste Text (AI)'], ['url', 'From URL']].map(([m, label]) => (
            <button
              key={m}
              onClick={() => {
                if (m !== mode && isDirty) {
                  if (!window.confirm('You have unsaved changes. Switch tabs and lose your work?')) return;
                }
                setMode(m);
                setPreview(null);
                setPreviewError(null);
                if (m !== 'manual') setRecipe(blankRecipe);
              }}
              type="button"
              className={`flex-1 font-medium py-2 rounded-lg transition text-sm ${
                mode === m
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-200 text-gray-900 hover:bg-gray-300'
              }`}
            >
              {label}
            </button>
          ))}
        </div>

        {/* AI Mode: paste text */}
        {mode === 'ai' && !preview && (
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
                placeholder="Paste your recipe here..."
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
                  <div className="bg-blue-500 h-2 rounded-full transition-all duration-300 ease-out" style={{ width: `${progress}%` }} />
                </div>
              )}
            </div>
          </div>
        )}

        {/* URL Mode: paste URL */}
        {mode === 'url' && !preview && (
          <div className="space-y-6">
            <div>
              <label htmlFor="recipeURL" className="block text-sm font-medium text-gray-900 mb-2">
                Recipe Page URL
              </label>
              <p className="text-xs text-gray-500 mb-2">
                Paste the URL of a recipe page. The server will fetch it and extract the recipe text for AI parsing.
              </p>
              <input
                id="recipeURL"
                type="url"
                placeholder="https://example.com/recipe/chocolate-chip-cookies"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                value={recipeURL}
                onChange={(e) => setRecipeURL(e.target.value)}
              />
            </div>
            <div className="space-y-2">
              <button
                onClick={handlePreviewFromURL}
                disabled={previewLoading || !recipeURL.trim()}
                type="button"
                className="w-full bg-blue-600 text-white font-medium py-2 rounded-lg hover:bg-blue-700 disabled:bg-blue-400 disabled:cursor-not-allowed transition"
              >
                {previewLoading ? 'Fetching & Parsing...' : 'Fetch & Parse Recipe'}
              </button>
              {previewLoading && (
                <div className="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
                  <div className="bg-blue-500 h-2 rounded-full transition-all duration-300 ease-out" style={{ width: `${progress}%` }} />
                </div>
              )}
            </div>
          </div>
        )}

        {/* Shared preview area for AI and URL modes */}
        {(mode === 'ai' || mode === 'url') && (
          <div className="space-y-4">
            {previewError && (
              <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <p className="text-sm text-red-700">{previewError}</p>
              </div>
            )}
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
                    {preview.ingredients?.length > 0 ? (
                      preview.ingredients.map((ing, idx) => (
                        <li key={idx} className="flex items-baseline gap-2 text-sm">
                          {ing.quantity && <span className="italic text-gray-400 shrink-0">{ing.quantity}</span>}
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
                    {preview.instructions?.length > 0 ? (
                      preview.instructions.map((step, idx) => (
                        <li key={idx} className="text-gray-900 text-sm">{idx + 1}. {step}</li>
                      ))
                    ) : (
                      <p className="text-gray-500 text-sm">(No instructions)</p>
                    )}
                  </ol>
                </div>
                <div className="flex gap-3 pt-4">
                  <button onClick={handleUsePreview} type="button" className="flex-1 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 transition">
                    Use This Recipe
                  </button>
                  <button onClick={() => { setPreview(null); setPreviewError(null); }} type="button" className="flex-1 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition">
                    Back to Edit
                  </button>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Manual Mode Form */}
        {mode === 'manual' && (
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
                value={recipe.ingredients.map(ing =>
                  typeof ing === 'string' ? ing : [ing.quantity, ing.name].filter(Boolean).join(' ')
                ).join('\r\n')}
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

            {/* Tags */}
            <div>
              <label className="block text-sm font-medium text-gray-900 mb-2">
                Tags <span className="text-gray-400 font-normal">(optional)</span>
              </label>
              <TagInput tags={recipe.tags} onChange={tags => setRecipe({ ...recipe, tags })} />
            </div>

            {/* Source URL */}
            <div>
              <label htmlFor="sourceURL" className="block text-sm font-medium text-gray-900 mb-2">
                Source Link <span className="text-gray-400 font-normal">(optional)</span>
              </label>
              <input
                id="sourceURL"
                type="url"
                placeholder="https://example.com/recipe/..."
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                value={recipe.source_url}
                onChange={(e) => setRecipe({ ...recipe, source_url: e.target.value })}
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
      {blocker.state === 'blocked' && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-sm mx-4 shadow-xl">
            <h2 className="text-lg font-semibold text-gray-900 mb-2">Discard changes?</h2>
            <p className="text-sm text-gray-600 mb-6">You have unsaved changes. If you leave, your work will be lost.</p>
            <div className="flex gap-3">
              <button
                onClick={() => blocker.proceed()}
                className="flex-1 bg-red-600 text-white font-medium py-2 rounded-lg hover:bg-red-700 transition"
              >
                Leave anyway
              </button>
              <button
                onClick={() => blocker.reset()}
                className="flex-1 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition"
              >
                Stay
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
