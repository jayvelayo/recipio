import { useState } from "react";
import { useParams } from "react-router";
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getRecipeId, updateRecipe } from "./recipe_apis";
import { Link } from "react-router";
import LoadingPage from '/src/pages/common/LoadingPage'
import { FiArrowLeft, FiEdit2, FiCheck, FiX } from 'react-icons/fi';
import { parseIngredient } from '../../utils/parseIngredient';
import { motion } from 'framer-motion';
import { toast } from 'sonner';

function getTagColor(tag) {
  const tagColorMap = {
    "easy": "bg-green-100 text-green-800",
    "breakfast": "bg-yellow-100 text-yellow-800",
    "lunch": "bg-blue-100 text-blue-800",
    "dinner": "bg-purple-100 text-purple-800",
    "dessert": "bg-pink-100 text-pink-800",
  }
  return tagColorMap[tag] || "bg-gray-100 text-gray-800"
}

function ingredientsToText(ingredients) {
  return ingredients.map(ing => [ing.quantity, ing.name].filter(Boolean).join(' ')).join('\n');
}

export function ViewRecipe() {
  const param = useParams()
  const id = param.uid
  const queryClient = useQueryClient();

  const { data, isLoading, isError, error } = useQuery({
      queryKey: ['recipes', id],
      queryFn: () => getRecipeId(id),
  });

  const [isEditing, setIsEditing] = useState(false);
  const [editName, setEditName] = useState('');
  const [editIngredients, setEditIngredients] = useState('');
  const [editInstructions, setEditInstructions] = useState('');

  const mutation = useMutation({
    mutationFn: updateRecipe,
    onSuccess: () => {
      queryClient.invalidateQueries(['recipes', id]);
      queryClient.invalidateQueries(['recipes']);
      setIsEditing(false);
      toast.success('Recipe saved');
    },
    onError: () => {
      toast.error('Failed to save recipe');
    },
  });

  if (isLoading) return <LoadingPage />
  if (isError) return <p className="text-red-600">Error: {error.message}</p>

  const recipe = data;
  const ingredients = recipe.ingredients ?? [];
  const steps = recipe.instructions ?? [];

  const handleStartEdit = () => {
    setEditName(recipe.name);
    setEditIngredients(ingredientsToText(ingredients));
    setEditInstructions(steps.join('\n'));
    setIsEditing(true);
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
  };

  const handleSaveEdit = (e) => {
    e.preventDefault();
    mutation.mutate({
      id,
      recipe: {
        name: editName,
        ingredients: editIngredients.split(/\r?\n/).filter(Boolean).map(parseIngredient),
        instructions: editInstructions.split(/\r?\n/).filter(Boolean),
      },
    });
  };

  if (isEditing) {
    return (
      <div className="max-w-2xl mx-auto">
        <div className="mb-6 flex items-center gap-4">
          <button
            onClick={handleCancelEdit}
            className="p-2 hover:bg-gray-100 rounded-lg transition"
            title="Cancel editing"
          >
            <FiArrowLeft size={24} />
          </button>
          <h1 className="text-3xl font-bold text-gray-900">Edit Recipe</h1>
        </div>

        <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-8">
          <form onSubmit={handleSaveEdit} className="space-y-6">
            <div>
              <label className="block text-sm font-medium text-gray-900 mb-2">
                Recipe Name
              </label>
              <input
                type="text"
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                value={editName}
                onChange={e => setEditName(e.target.value)}
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-900 mb-2">
                Ingredients
              </label>
              <p className="text-xs text-gray-500 mb-2">One ingredient per line</p>
              <textarea
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition font-mono text-sm"
                value={editIngredients}
                onChange={e => setEditIngredients(e.target.value)}
                rows="8"
                required
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-900 mb-2">
                Instructions
              </label>
              <p className="text-xs text-gray-500 mb-2">One step per line</p>
              <textarea
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition font-mono text-sm"
                value={editInstructions}
                onChange={e => setEditInstructions(e.target.value)}
                rows="8"
                required
              />
            </div>

            <div className="flex gap-3 pt-4">
              <button
                type="submit"
                disabled={mutation.isPending}
                className="flex-1 inline-flex items-center justify-center gap-2 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition"
              >
                <FiCheck size={18} />
                {mutation.isPending ? 'Saving...' : 'Save Changes'}
              </button>
              <button
                type="button"
                onClick={handleCancelEdit}
                className="flex-1 inline-flex items-center justify-center gap-2 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition"
              >
                <FiX size={18} />
                Cancel
              </button>
            </div>
          </form>
        </div>
      </div>
    );
  }

  const listIngredients = ingredients.map((item, index) => (
    <li key={index} className="flex items-baseline gap-2 py-2 text-gray-700">
      {item.quantity && (
        <span className="italic text-gray-400 shrink-0">{item.quantity}</span>
      )}
      <span>{item.name}</span>
    </li>
  ));

  const listSteps = steps.map((step, index) => (
    <li key={index} className="text-gray-700 py-3">
      <span className="font-semibold text-indigo-600 mr-3">{index + 1}.</span>
      {step}
    </li>
  ));

  const listTags = recipe.tags ? recipe.tags.map((tag) => (
    <span
      key={tag}
      className={`inline-block px-3 py-1 rounded-full text-sm font-medium ${getTagColor(tag)} mr-2 mb-2`}
    >
      {tag}
    </span>
  )) : null;

  return (
    <div className="max-w-4xl mx-auto">
      <div className="mb-6 flex items-center gap-4">
        <Link
          to="/recipe"
          className="p-2 hover:bg-gray-100 rounded-lg transition"
          title="Back to recipes"
        >
          <FiArrowLeft size={24} />
        </Link>
        <h1 className="text-4xl font-bold text-gray-900 flex-1">{recipe.name}</h1>
        <button
          onClick={handleStartEdit}
          className="p-2 hover:bg-gray-100 rounded-lg transition"
          title="Edit recipe"
        >
          <FiEdit2 size={22} />
        </button>
      </div>

      {/* Tags */}
      {listTags && (
        <div className="mb-8">
          {listTags}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Ingredients */}
        <motion.div
          className="lg:col-span-1"
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.25, delay: 0.05 }}
        >
          <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6">
            <h3 className="text-xl font-bold text-gray-900 mb-4">Ingredients</h3>
            <ul className="space-y-0">
              {listIngredients}
            </ul>
          </div>
        </motion.div>

        {/* Instructions */}
        <motion.div
          className="lg:col-span-2"
          initial={{ opacity: 0, y: 12 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.25, delay: 0.12 }}
        >
          <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6">
            <h3 className="text-xl font-bold text-gray-900 mb-4">Instructions</h3>
            <ol className="space-y-0">
              {listSteps}
            </ol>
          </div>
        </motion.div>
      </div>

      {/* Back Button */}
      <div className="mt-8">
        <Link
          to="/recipe"
          className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition"
        >
          <FiArrowLeft size={18} />
          Back to Recipes
        </Link>
      </div>
    </div>
  );
}
