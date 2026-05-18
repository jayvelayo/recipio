import { useState } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteRecipe, getRecipes, updateRecipe } from "./recipe_apis";
import { getMealPlans, createMealPlan, addRecipesToMealPlan } from "../mealplan/mealplan_apis";
import { Form } from "react-router";
import { SkeletonList } from '/src/pages/common/LoadingPage';
import { FiTrash2, FiSearch, FiPlus, FiX, FiCalendar, FiTag } from 'react-icons/fi';
import { motion, AnimatePresence } from 'framer-motion';
import { toast } from 'sonner';
import { TagInput } from '../common/TagInput';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.05 } },
};

const rowVariants = {
  hidden: { opacity: 0, y: 8 },
  show: { opacity: 1, y: 0, transition: { duration: 0.2 } },
};

function displayTagsToString(tags) {
  if (!tags?.length) return "None";
  return tags.join(", ");
}

function AddToMealPlanModal({ selectedIds, onClose, onSuccess }) {
  const queryClient = useQueryClient();
  const [selectedPlanId, setSelectedPlanId] = useState(null);

  const { data: plans, isLoading } = useQuery({
    queryKey: ['meal-plans'],
    queryFn: getMealPlans,
  });

  const addMutation = useMutation({
    mutationFn: ({ planId, ids }) => addRecipesToMealPlan(planId, ids),
    onSuccess: () => {
      queryClient.invalidateQueries(['meal-plans']);
      toast.success(`Added ${selectedIds.length} recipe${selectedIds.length > 1 ? 's' : ''} to meal plan`);
      onSuccess();
    },
    onError: () => toast.error('Failed to add to meal plan'),
  });

  const handleConfirm = () => {
    if (!selectedPlanId) return;
    addMutation.mutate({ planId: selectedPlanId, ids: Array.from(selectedIds) });
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-sm w-full mx-4 shadow-xl">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Add to meal plan</h2>
          <button onClick={onClose} className="p-1 hover:bg-gray-100 rounded-lg transition">
            <FiX size={20} />
          </button>
        </div>
        {isLoading ? (
          <p className="text-gray-500 text-sm py-4 text-center">Loading meal plans…</p>
        ) : !plans?.length ? (
          <p className="text-gray-500 text-sm py-4 text-center">No meal plans yet. Use "New meal plan" instead.</p>
        ) : (
          <ul className="space-y-2 mb-4 max-h-60 overflow-y-auto">
            {plans.map(plan => (
              <li key={plan.id}>
                <button
                  onClick={() => setSelectedPlanId(plan.id)}
                  className={`w-full text-left px-3 py-2 rounded-lg border transition text-sm ${
                    selectedPlanId === plan.id
                      ? 'border-indigo-500 bg-indigo-50 text-indigo-700'
                      : 'border-gray-200 hover:border-gray-300 text-gray-700'
                  }`}
                >
                  <span className="font-medium">Meal plan #{plan.id}</span>
                  {plan.recipe_names?.length > 0 && (
                    <span className="text-gray-400 ml-2">· {plan.recipe_names.join(', ')}</span>
                  )}
                </button>
              </li>
            ))}
          </ul>
        )}
        <div className="flex gap-3">
          <button
            onClick={handleConfirm}
            disabled={!selectedPlanId || addMutation.isPending}
            className="flex-1 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition text-sm"
          >
            {addMutation.isPending ? 'Adding…' : 'Add to plan'}
          </button>
          <button
            onClick={onClose}
            className="flex-1 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition text-sm"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}

function AddTagsModal({ selectedIds, recipes, onClose, onSuccess }) {
  const queryClient = useQueryClient();
  const [newTags, setNewTags] = useState([]);

  const updateMutation = useMutation({
    mutationFn: ({ id, recipe }) => updateRecipe({ id, recipe }),
    onError: () => toast.error('Failed to update some recipes'),
  });

  const handleConfirm = async () => {
    if (!newTags.length) return;
    const targets = recipes.filter(r => selectedIds.has(r.id));
    await Promise.all(targets.map(r =>
      updateMutation.mutateAsync({
        id: r.id,
        recipe: {
          name: r.name,
          ingredients: r.ingredients ?? [],
          instructions: r.instructions ?? [],
          tags: [...new Set([...(r.tags ?? []), ...newTags])],
        },
      })
    ));
    queryClient.invalidateQueries(['recipes']);
    toast.success(`Tags added to ${targets.length} recipe${targets.length > 1 ? 's' : ''}`);
    onSuccess();
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-6 max-w-sm w-full mx-4 shadow-xl">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-lg font-semibold text-gray-900">Add tags</h2>
          <button onClick={onClose} className="p-1 hover:bg-gray-100 rounded-lg transition">
            <FiX size={20} />
          </button>
        </div>
        <p className="text-sm text-gray-500 mb-3">
          These tags will be added to {selectedIds.size} selected recipe{selectedIds.size > 1 ? 's' : ''}.
        </p>
        <TagInput tags={newTags} onChange={setNewTags} />
        <div className="flex gap-3 mt-4">
          <button
            onClick={handleConfirm}
            disabled={!newTags.length || updateMutation.isPending}
            className="flex-1 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition text-sm"
          >
            {updateMutation.isPending ? 'Saving…' : 'Add tags'}
          </button>
          <button
            onClick={onClose}
            className="flex-1 bg-gray-200 text-gray-900 font-medium py-2 rounded-lg hover:bg-gray-300 transition text-sm"
          >
            Cancel
          </button>
        </div>
      </div>
    </div>
  );
}

function RecipeRowPreview({ recipe, selected, onToggle }) {
  return (
    <div className={`flex items-center gap-3 p-4 border-b border-gray-200 hover:bg-gray-50 transition ${selected ? 'bg-indigo-50 hover:bg-indigo-50' : ''}`}>
      <input
        type="checkbox"
        checked={selected}
        onChange={onToggle}
        onClick={e => e.stopPropagation()}
        className="h-4 w-4 rounded border-gray-300 text-indigo-600 focus:ring-indigo-500 cursor-pointer shrink-0"
      />
      <a href={`/recipe/view/${recipe.id}`} className="flex-1 min-w-0 cursor-pointer">
        <h3 className="font-semibold text-gray-900 hover:text-indigo-600 truncate">{recipe.name}</h3>
        <p className="text-sm text-gray-500 mt-0.5">Tags: {displayTagsToString(recipe.tags)}</p>
      </a>
    </div>
  );
}

function BulkActionBar({ count, onNewMealPlan, onAddToMealPlan, onAddTags, onDelete, onClear, isDeleting }) {
  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0, y: -8 }}
        animate={{ opacity: 1, y: 0 }}
        exit={{ opacity: 0, y: -8 }}
        transition={{ duration: 0.15 }}
        className="flex items-center gap-2 flex-wrap bg-indigo-50 border border-indigo-200 rounded-lg px-4 py-2 mb-4"
      >
        <span className="text-sm font-medium text-indigo-700 mr-1">{count} selected</span>
        <button
          onClick={onNewMealPlan}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-indigo-600 text-white text-sm font-medium rounded-lg hover:bg-indigo-700 transition"
        >
          <FiCalendar size={14} />
          New meal plan
        </button>
        <button
          onClick={onAddToMealPlan}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white border border-indigo-300 text-indigo-700 text-sm font-medium rounded-lg hover:bg-indigo-50 transition"
        >
          <FiCalendar size={14} />
          Add to meal plan
        </button>
        <button
          onClick={onAddTags}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white border border-indigo-300 text-indigo-700 text-sm font-medium rounded-lg hover:bg-indigo-50 transition"
        >
          <FiTag size={14} />
          Add tags
        </button>
        <button
          onClick={onDelete}
          disabled={isDeleting}
          className="inline-flex items-center gap-1.5 px-3 py-1.5 bg-white border border-red-300 text-red-600 text-sm font-medium rounded-lg hover:bg-red-50 transition disabled:opacity-50"
        >
          <FiTrash2 size={14} />
          {isDeleting ? 'Deleting…' : 'Delete'}
        </button>
        <button
          onClick={onClear}
          className="ml-auto p-1.5 text-gray-400 hover:text-gray-600 hover:bg-white rounded-lg transition"
          title="Clear selection"
        >
          <FiX size={16} />
        </button>
      </motion.div>
    </AnimatePresence>
  );
}

function RecipeListRows({ recipes }) {
  const queryClient = useQueryClient();
  const [searchQuery, setSearchQuery] = useState('');
  const [selectedTags, setSelectedTags] = useState([]);
  const [selectedIds, setSelectedIds] = useState(new Set());
  const [modal, setModal] = useState(null); // null | 'addToPlan' | 'addTags'

  const allTags = [...new Set(recipes?.flatMap(r => r.tags ?? []))].sort();

  const toggleTag = (tag) => {
    setSelectedTags(prev =>
      prev.includes(tag) ? prev.filter(t => t !== tag) : [...prev, tag]
    );
  };

  const filtered = recipes?.filter(item => {
    const matchesSearch = item.name.toLowerCase().includes(searchQuery.toLowerCase());
    const matchesTags = selectedTags.length === 0 || selectedTags.every(t => item.tags?.includes(t));
    return matchesSearch && matchesTags;
  });

  const sorted = filtered?.sort((a, b) => a.name.localeCompare(b.name));

  const toggleSelect = (id) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      next.has(id) ? next.delete(id) : next.add(id);
      return next;
    });
  };

  const clearSelection = () => setSelectedIds(new Set());

  const deleteMutation = useMutation({
    mutationFn: (ids) => Promise.all(ids.map(id => deleteRecipe(id))),
    onSuccess: (_, ids) => {
      queryClient.invalidateQueries(['recipes']);
      toast.success(`Deleted ${ids.length} recipe${ids.length > 1 ? 's' : ''}`);
      clearSelection();
    },
    onError: () => toast.error('Failed to delete some recipes'),
  });

  const createMealPlanMutation = useMutation({
    mutationFn: (ids) => createMealPlan(ids),
    onSuccess: () => {
      queryClient.invalidateQueries(['meal-plans']);
      toast.success('Meal plan created');
      clearSelection();
    },
    onError: () => toast.error('Failed to create meal plan'),
  });

  const handleDelete = () => {
    const ids = Array.from(selectedIds);
    if (!window.confirm(`Delete ${ids.length} recipe${ids.length > 1 ? 's' : ''}? This cannot be undone.`)) return;
    deleteMutation.mutate(ids);
  };

  const handleNewMealPlan = () => {
    createMealPlanMutation.mutate(Array.from(selectedIds));
  };

  if (!recipes || recipes.length === 0) {
    return (
      <div className="text-center py-12">
        <p className="text-gray-500 mb-4">No recipes found yet. Create your first recipe!</p>
      </div>
    );
  }

  return (
    <>
      {selectedIds.size > 0 && (
        <BulkActionBar
          count={selectedIds.size}
          onNewMealPlan={handleNewMealPlan}
          onAddToMealPlan={() => setModal('addToPlan')}
          onAddTags={() => setModal('addTags')}
          onDelete={handleDelete}
          onClear={clearSelection}
          isDeleting={deleteMutation.isPending}
        />
      )}

      {/* Search Bar */}
      <div className="mb-4">
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

      {/* Tag Filters */}
      {allTags.length > 0 && (
        <div className="flex flex-wrap gap-2 mb-6">
          {allTags.map(tag => (
            <button
              key={tag}
              onClick={() => toggleTag(tag)}
              className={`px-3 py-1 rounded-full text-sm font-medium transition ${
                selectedTags.includes(tag)
                  ? 'bg-indigo-600 text-white'
                  : 'bg-gray-100 text-gray-600 hover:bg-gray-200'
              }`}
            >
              {tag}
            </button>
          ))}
          {selectedTags.length > 0 && (
            <button
              onClick={() => setSelectedTags([])}
              className="px-3 py-1 rounded-full text-sm font-medium text-gray-400 hover:text-gray-600 transition"
            >
              Clear
            </button>
          )}
        </div>
      )}

      {/* Recipe List */}
      <motion.div
        className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden"
        variants={containerVariants}
        initial="hidden"
        animate="show"
      >
        {sorted?.length === 0 ? (
          <p className="text-gray-500 text-center py-8">No recipes match your filters.</p>
        ) : (
          sorted?.map(recipe => (
            <motion.div key={recipe.id} variants={rowVariants}>
              <RecipeRowPreview
                recipe={recipe}
                selected={selectedIds.has(recipe.id)}
                onToggle={() => toggleSelect(recipe.id)}
              />
            </motion.div>
          ))
        )}
      </motion.div>

      {modal === 'addToPlan' && (
        <AddToMealPlanModal
          selectedIds={Array.from(selectedIds)}
          onClose={() => setModal(null)}
          onSuccess={() => { setModal(null); clearSelection(); }}
        />
      )}
      {modal === 'addTags' && (
        <AddTagsModal
          selectedIds={selectedIds}
          recipes={recipes}
          onClose={() => setModal(null)}
          onSuccess={() => { setModal(null); clearSelection(); }}
        />
      )}
    </>
  );
}

export function RecipeList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['recipes'],
    queryFn: getRecipes,
  });

  if (isLoading) return <SkeletonList />;
  if (error) return <p className="text-red-600">Error: {error.message}</p>;

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
      <RecipeListRows recipes={data} />
    </div>
  );
}
