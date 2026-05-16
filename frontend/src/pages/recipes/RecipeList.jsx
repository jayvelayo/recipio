import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteRecipe, getRecipes } from "./recipe_apis";
import { Form } from "react-router";
import { SkeletonList } from '/src/pages/common/LoadingPage';
import { useEffect, useState } from 'react';
import { FiTrash2, FiSearch, FiPlus } from 'react-icons/fi';
import { motion } from 'framer-motion';
import { toast } from 'sonner';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const rowVariants = {
  hidden: { opacity: 0, y: 8 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2 } },
};

function displayTagsToString(tags) {
  if (typeof tags === 'undefined') {
    return "None"
  }
  return tags.join(", ")
}

function RecipeDeleteIcon({ id }) {
  const queryClient = useQueryClient();
  const [isDeleting, setIsDeleting] = useState(false);

  const mutation = useMutation({
    mutationFn: deleteRecipe,
    onSuccess: () => {
      queryClient.invalidateQueries(['recipes']);
      setIsDeleting(false);
      toast.success('Recipe deleted');
    },
    onError: (error) => {
      toast.error(`Failed to delete: ${error.message}`);
      setIsDeleting(false);
    }
  });

  const handleClick = () => {
    if (window.confirm('Are you sure you want to delete this recipe?')) {
      setIsDeleting(true);
      mutation.mutate(id);
    }
  }

  return (
    <button
      onClick={handleClick}
      disabled={isDeleting}
      className="p-2 text-red-500 hover:bg-red-50 rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
      title="Delete recipe"
    >
      <FiTrash2 size={18} />
    </button>
  )
}

function RecipeRowPreview({recipe}) {
  return (
    <div className="flex items-center justify-between p-4 border-b border-gray-200 hover:bg-gray-50 transition">
      <a
        href={`/recipe/view/${recipe.id}`}
        className="flex-1 cursor-pointer"
      >
        <h3 className="font-semibold text-gray-900 hover:text-indigo-600">{recipe.name}</h3>
        <p className="text-sm text-gray-500 mt-1">Tags: {displayTagsToString(recipe.tags)}</p>
      </a>
      <RecipeDeleteIcon id={recipe.id}/>
    </div>
  )
}

function RecipeListRows({recipes}) {
  const [searchQuery, setSearchQuery] = useState('');
  const [recipeItems, setRecipeItems] = useState(recipes);

  useEffect(() => {
    const results = recipes?.filter(item =>
      item.name.toLowerCase().includes(searchQuery.toLowerCase())
    );
    setRecipeItems(results);
  }, [searchQuery, recipes]);

  if (recipes === undefined || recipes.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">No recipes found yet. Create your first recipe!</p>
      </div>
    )
  }

  const sorted = recipeItems?.sort((a, b) => a.name.localeCompare(b.name));

  return (
    <>
      {/* Search Bar */}
      <div className="mb-6">
        <div className="relative">
          <FiSearch className="absolute left-3 top-3 text-gray-400" size={20} />
          <input
            type="text"
            placeholder="Search recipes..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
          />
        </div>
      </div>

      {/* Recipe List */}
      <motion.div
        className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden"
        variants={containerVariants}
        initial="hidden"
        animate="show"
      >
        {sorted?.map(recipe => (
          <motion.div key={recipe.id} variants={rowVariants}>
            <RecipeRowPreview recipe={recipe} />
          </motion.div>
        ))}
      </motion.div>
    </>
  )
}

export function RecipeList() {
  const { data, isLoading, error } = useQuery({
      queryKey: ['recipes'],
      queryFn: getRecipes,
  });

  if (isLoading) return <SkeletonList />
  if (error) return <p className="text-red-600">Error: {error.message}</p>

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Recipes</h1>
        <Form action="/recipe/add/" className="inline">
          <button
            className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition"
            type="submit"
          >
            <FiPlus size={20} />
            New Recipe
          </button>
        </Form>
      </div>
      <RecipeListRows recipes={data}/>
    </div>
  )
}
