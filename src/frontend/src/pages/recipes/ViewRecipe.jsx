import { useParams } from "react-router";
import { useQuery } from '@tanstack/react-query';
import { getRecipeId } from "./recipe_apis";
import { Link } from "react-router";
import LoadingPage from '/src/pages/common/LoadingPage'
import { FiArrowLeft } from 'react-icons/fi';

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

export function ViewRecipe() {
  const param = useParams()
  const id = param.uid
  const { data, isLoading, isError, error } = useQuery({
      queryKey: ['recipes', id],
      queryFn: () => getRecipeId(id),
  });
  
  if (isLoading) return <LoadingPage />
  if (isError) return <p className="text-red-600">Error: {error.message}</p>

  const recipe = data;
  const ingredients = recipe.ingredients ?? [];
  const steps = recipe.steps ?? recipe.instructions ?? [];

  const listIngredients = ingredients.map((item, index) => (
    <li key={index} className="text-gray-700 py-2">
      {typeof item === 'string' ? item : `${item.quantity ?? ''} ${item.name ?? ''}`.trim()}
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
        <h1 className="text-4xl font-bold text-gray-900">{recipe.name}</h1>
      </div>

      {/* Tags */}
      {listTags && (
        <div className="mb-8">
          {listTags}
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Ingredients */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6">
            <h3 className="text-xl font-bold text-gray-900 mb-4">Ingredients</h3>
            <ul className="space-y-0">
              {listIngredients}
            </ul>
          </div>
        </div>

        {/* Instructions */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-6">
            <h3 className="text-xl font-bold text-gray-900 mb-4">Instructions</h3>
            <ol className="space-y-0">
              {listSteps}
            </ol>
          </div>
        </div>
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