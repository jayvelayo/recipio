import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getGroceryListById, updateGroceryList } from './grocery_apis';
import LoadingPage from '/src/pages/common/LoadingPage';

export function GroceryList() {
  const { id } = useParams();
  const [filter, setFilter] = useState('all'); // 'all', 'checked', 'unchecked'
  const [newItemName, setNewItemName] = useState('');
  const [newItemQuantity, setNewItemQuantity] = useState('');
  const queryClient = useQueryClient();

  const { data: groceryList, isLoading, error } = useQuery({
    queryKey: ['grocery-list', id],
    queryFn: () => getGroceryListById(id),
  });

  const updateMutation = useMutation({
    mutationFn: (items) => updateGroceryList(id, items),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['grocery-list', id] });
    },
  });

  const handleCheck = (index) => {
    if (!groceryList) return;
    const newItems = [...groceryList.items];
    newItems[index] = { ...newItems[index], checked: !newItems[index].checked };
    updateMutation.mutate(newItems);
  };

  const handleAddItem = (e) => {
    e.preventDefault();
    if (!newItemName.trim()) return;

    const newItem = {
      name: newItemName.trim(),
      quantity: newItemQuantity.trim(),
      checked: false,
    };

    const newItems = [...(groceryList?.items || []), newItem];
    updateMutation.mutate(newItems);

    // Clear form
    setNewItemName('');
    setNewItemQuantity('');
  };

  const handleDeleteItem = (index) => {
    if (!groceryList) return;
    const newItems = groceryList.items.filter((_, i) => i !== index);
    updateMutation.mutate(newItems);
  };

  const filteredItems = groceryList?.items?.filter((item) => {
    if (filter === 'checked') return item.checked;
    if (filter === 'unchecked') return !item.checked;
    return true;
  });

  if (isLoading) return <LoadingPage />;
  if (error) return <p>Error: {error.message}</p>;

  return (
    <div>
      <h2>{groceryList?.name || 'Grocery List'}</h2>
      <Link to="/grocery" className="ui button">Back to Grocery Lists</Link>

      {/* Add new item form */}
      <form className="ui form" onSubmit={handleAddItem} style={{ marginTop: '20px' }}>
        <div className="inline fields">
          <div className="field">
            <input
              type="text"
              placeholder="Quantity (optional)"
              value={newItemQuantity}
              onChange={(e) => setNewItemQuantity(e.target.value)}
              style={{ width: '120px' }}
            />
          </div>
          <div className="field">
            <input
              type="text"
              placeholder="Item name"
              value={newItemName}
              onChange={(e) => setNewItemName(e.target.value)}
              required
              style={{ width: '200px' }}
            />
          </div>
          <div className="field">
            <button
              type="submit"
              className="ui button primary small"
              disabled={updateMutation.isPending}
            >
              Add Item
            </button>
          </div>
        </div>
      </form>

      <div style={{ marginTop: '20px' }}>
        <div className="ui buttons">
          <button
            className={`ui button ${filter === 'all' ? 'active' : ''}`}
            onClick={() => setFilter('all')}
          >
            All
          </button>
          <button
            className={`ui button ${filter === 'unchecked' ? 'active' : ''}`}
            onClick={() => setFilter('unchecked')}
          >
            Unchecked
          </button>
          <button
            className={`ui button ${filter === 'checked' ? 'active' : ''}`}
            onClick={() => setFilter('checked')}
          >
            Checked
          </button>
        </div>
      </div>

      <div className="ui relaxed divided list" style={{ marginTop: '20px' }}>
        {filteredItems?.length ? (
          filteredItems.map((item, index) => {
            const originalIndex = groceryList.items.indexOf(item);
            return (
              <div key={originalIndex} className="item" style={{ display: 'flex', alignItems: 'center' }}>
                <div className="ui checkbox" style={{ flex: 1 }}>
                  <input
                    type="checkbox"
                    checked={item.checked}
                    onChange={() => handleCheck(originalIndex)}
                    disabled={updateMutation.isPending}
                  />
                  <label style={{
                    textDecoration: item.checked ? 'line-through' : 'none',
                    color: item.checked ? '#888' : 'inherit'
                  }}>
                    {item.quantity ? `${item.quantity} ${item.name}` : item.name}
                  </label>
                </div>
                <button
                  className="ui button tiny red"
                  onClick={() => handleDeleteItem(originalIndex)}
                  disabled={updateMutation.isPending}
                  style={{ marginLeft: '10px' }}
                >
                  Delete
                </button>
              </div>
            );
          })
        ) : (
          <p>No items to show.</p>
        )}
      </div>
    </div>
  );
}