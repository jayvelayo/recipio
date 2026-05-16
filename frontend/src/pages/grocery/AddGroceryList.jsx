import React, { useState } from 'react';
import { useNavigate } from 'react-router';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Link } from 'react-router';
import { getMealPlans, getGroceryList } from '../mealplan/mealplan_apis';
import { createGroceryList } from './grocery_apis';
import LoadingPage from '/src/pages/common/LoadingPage';
import { FiArrowLeft, FiPlus, FiTrash2 } from 'react-icons/fi';

export function AddGroceryList() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [creationType, setCreationType] = useState('manual');
  const [selectedMealPlan, setSelectedMealPlan] = useState('');
  const [manualItems, setManualItems] = useState([{ name: '', quantity: '' }]);

  const { data: mealPlans, isLoading: loadingMealPlans } = useQuery({
    queryKey: ['mealplans'],
    queryFn: getMealPlans,
    enabled: creationType === 'from-mealplan',
  });

  const { data: mealPlanIngredients, isLoading: loadingIngredients } = useQuery({
    queryKey: ['grocery', selectedMealPlan],
    queryFn: () => getGroceryList(selectedMealPlan),
    enabled: creationType === 'from-mealplan' && selectedMealPlan !== '',
  });

  const createMutation = useMutation({
    mutationFn: (data) => createGroceryList(data.name, data.items, data.mealPlanId),
    onSuccess: (result) => {
      navigate(`/grocery/view/${result.id}`);
    },
  });

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!name.trim()) return;

    let items = [];
    if (creationType === 'manual') {
      items = manualItems
        .filter(item => item.name.trim())
        .map(item => ({
          name: item.name.trim(),
          quantity: item.quantity.trim(),
          checked: false,
        }));
    } else if (creationType === 'from-mealplan' && mealPlanIngredients) {
      items = mealPlanIngredients.map(ing => {
        const parts = ing.split(' ');
        if (parts.length > 1) {
          const quantity = parts.slice(0, -1).join(' ');
          const itemName = parts[parts.length - 1];
          return { name: itemName, quantity, checked: false };
        }
        return { name: ing, quantity: '', checked: false };
      });
    }

    createMutation.mutate({
      name: name.trim(),
      items,
      mealPlanId: creationType === 'from-mealplan' ? selectedMealPlan : null,
    });
  };

  const addManualItem = () => {
    setManualItems([...manualItems, { name: '', quantity: '' }]);
  };

  const updateManualItem = (index, field, value) => {
    const newItems = [...manualItems];
    newItems[index][field] = value;
    setManualItems(newItems);
  };

  const removeManualItem = (index) => {
    setManualItems(manualItems.filter((_, i) => i !== index));
  };

  if (loadingMealPlans && creationType === 'from-mealplan') return <LoadingPage />;

  return (
    <div className="max-w-2xl mx-auto">
      <div className="mb-6 flex items-center gap-4">
        <button 
          onClick={() => navigate("/grocery")} 
          className="p-2 hover:bg-gray-100 rounded-lg transition"
          title="Back to grocery lists"
        >
          <FiArrowLeft size={24} />
        </button>
        <h1 className="text-3xl font-bold text-gray-900">Create Grocery List</h1>
      </div>

      <div className="bg-white rounded-lg border border-gray-200 shadow-sm p-8">
        <form onSubmit={handleSubmit} className="space-y-6">
          {/* List Name */}
          <div>
            <label htmlFor="listName" className="block text-sm font-medium text-gray-900 mb-2">
              List Name
            </label>
            <input
              id="listName"
              type="text"
              value={name}
              onChange={(e) => setName(e.target.value)}
              placeholder="e.g., Weekly Groceries"
              className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
              required
            />
          </div>

          {/* Creation Type Tabs */}
          <div>
            <label className="block text-sm font-medium text-gray-900 mb-3">How to add items?</label>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => setCreationType('manual')}
                className={`px-4 py-2 rounded-lg font-medium transition ${
                  creationType === 'manual'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                Manual Entry
              </button>
              <button
                type="button"
                onClick={() => setCreationType('from-mealplan')}
                className={`px-4 py-2 rounded-lg font-medium transition ${
                  creationType === 'from-mealplan'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
                }`}
              >
                From Meal Plan
              </button>
            </div>
          </div>

          {/* Manual Entry */}
          {creationType === 'manual' && (
            <div>
              <label className="block text-sm font-medium text-gray-900 mb-3">Items</label>
              <div className="space-y-3">
                {manualItems.map((item, index) => (
                  <div key={index} className="flex gap-3">
                    <input
                      type="text"
                      placeholder="Quantity (optional)"
                      value={item.quantity}
                      onChange={(e) => updateManualItem(index, 'quantity', e.target.value)}
                      className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                    />
                    <input
                      type="text"
                      placeholder="Item name"
                      value={item.name}
                      onChange={(e) => updateManualItem(index, 'name', e.target.value)}
                      className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                      required
                    />
                    <button
                      type="button"
                      onClick={() => removeManualItem(index)}
                      disabled={manualItems.length === 1}
                      className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
                      title="Remove item"
                    >
                      <FiTrash2 size={20} />
                    </button>
                  </div>
                ))}
              </div>
              <button
                type="button"
                onClick={addManualItem}
                className="mt-3 inline-flex items-center gap-2 px-4 py-2 text-indigo-600 border border-indigo-600 rounded-lg hover:bg-indigo-50 transition font-medium"
              >
                <FiPlus size={18} />
                Add Item
              </button>
            </div>
          )}

          {/* From Meal Plan */}
          {creationType === 'from-mealplan' && (
            <div>
              <label htmlFor="mealPlan" className="block text-sm font-medium text-gray-900 mb-3">
                Select Meal Plan
              </label>
              <select
                id="mealPlan"
                value={selectedMealPlan}
                onChange={(e) => setSelectedMealPlan(e.target.value)}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
                required
              >
                <option value="">Choose a meal plan...</option>
                {mealPlans?.map(plan => (
                  <option key={plan.id} value={plan.id}>
                    Meal Plan #{plan.id} - {plan.recipe_names?.join(', ') || 'No recipes'}
                  </option>
                ))}
              </select>

              {selectedMealPlan && loadingIngredients && (
                <p className="mt-3 text-gray-500">Loading ingredients...</p>
              )}

              {selectedMealPlan && mealPlanIngredients && (
                <div className="mt-4 p-4 bg-gray-50 rounded-lg">
                  <p className="text-sm font-medium text-gray-900 mb-3">Items that will be added:</p>
                  <ul className="space-y-2">
                    {mealPlanIngredients.map((ing, index) => (
                      <li key={index} className="text-sm text-gray-700">• {ing}</li>
                    ))}
                  </ul>
                </div>
              )}
            </div>
          )}

          {/* Form Actions */}
          <div className="flex gap-3 pt-6 border-t border-gray-200">
            <button
              type="submit"
              disabled={createMutation.isPending}
              className="flex-1 bg-indigo-600 text-white font-medium py-2 rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition"
            >
              {createMutation.isPending ? 'Creating...' : 'Create List'}
            </button>
            <button
              type="button"
              onClick={() => navigate('/grocery')}
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