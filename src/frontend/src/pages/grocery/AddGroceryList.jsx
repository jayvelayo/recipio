import React, { useState } from 'react';
import { useNavigate } from 'react-router';
import { useQuery, useMutation } from '@tanstack/react-query';
import { Link } from 'react-router';
import { getMealPlans, getGroceryList } from '../mealplan/mealplan_apis';
import { createGroceryList } from './grocery_apis';
import LoadingPage from '/src/pages/common/LoadingPage';

export function AddGroceryList() {
  const navigate = useNavigate();
  const [name, setName] = useState('');
  const [creationType, setCreationType] = useState('manual'); // 'manual' or 'from-mealplan'
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
        // Parse "quantity name" format
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
    <div>
      <h2>Create Grocery List</h2>
      <Link to="/grocery" className="ui button">Back to Grocery Lists</Link>

      <form className="ui form" onSubmit={handleSubmit} style={{ marginTop: '20px' }}>
        <div className="field">
          <label>List Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Enter grocery list name"
            required
          />
        </div>

        <div className="field">
          <label>Creation Type</label>
          <div className="ui buttons">
            <button
              type="button"
              className={`ui button ${creationType === 'manual' ? 'active' : ''}`}
              onClick={() => setCreationType('manual')}
            >
              Manual Entry
            </button>
            <button
              type="button"
              className={`ui button ${creationType === 'from-mealplan' ? 'active' : ''}`}
              onClick={() => setCreationType('from-mealplan')}
            >
              From Meal Plan
            </button>
          </div>
        </div>

        {creationType === 'manual' && (
          <div className="field">
            <label>Items</label>
            {manualItems.map((item, index) => (
              <div key={index} className="ui grid" style={{ marginBottom: '10px' }}>
                <div className="eight wide column">
                  <input
                    type="text"
                    placeholder="Quantity (optional)"
                    value={item.quantity}
                    onChange={(e) => updateManualItem(index, 'quantity', e.target.value)}
                  />
                </div>
                <div className="six wide column">
                  <input
                    type="text"
                    placeholder="Item name"
                    value={item.name}
                    onChange={(e) => updateManualItem(index, 'name', e.target.value)}
                    required
                  />
                </div>
                <div className="two wide column">
                  <button
                    type="button"
                    className="ui button red small"
                    onClick={() => removeManualItem(index)}
                    disabled={manualItems.length === 1}
                  >
                    Remove
                  </button>
                </div>
              </div>
            ))}
            <button type="button" className="ui button small" onClick={addManualItem}>
              Add Item
            </button>
          </div>
        )}

        {creationType === 'from-mealplan' && (
          <div className="field">
            <label>Select Meal Plan</label>
            <select
              value={selectedMealPlan}
              onChange={(e) => setSelectedMealPlan(e.target.value)}
              required
            >
              <option value="">Choose a meal plan...</option>
              {mealPlans?.map(plan => (
                <option key={plan.id} value={plan.id}>
                  Meal Plan #{plan.id} - {plan.recipe_names?.join(', ') || 'No recipes'}
                </option>
              ))}
            </select>
            {selectedMealPlan && loadingIngredients && <p>Loading ingredients...</p>}
            {selectedMealPlan && mealPlanIngredients && (
              <div className="ui list" style={{ marginTop: '10px' }}>
                <p>Items that will be added:</p>
                {mealPlanIngredients.map((ing, index) => (
                  <div key={index} className="item">{ing}</div>
                ))}
              </div>
            )}
          </div>
        )}

        <button
          type="submit"
          className="ui button primary"
          disabled={createMutation.isPending}
        >
          {createMutation.isPending ? 'Creating...' : 'Create Grocery List'}
        </button>
      </form>
    </div>
  );
}