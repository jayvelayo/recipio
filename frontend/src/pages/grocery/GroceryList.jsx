import React, { useState, useEffect } from 'react';
import { useParams, Link } from 'react-router';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { getGroceryListById, updateGroceryList } from './grocery_apis';
import LoadingPage from '/src/pages/common/LoadingPage';
import { FiArrowLeft, FiTrash2, FiCheck, FiPlus } from 'react-icons/fi';
import { motion, AnimatePresence } from 'framer-motion';
import { toast } from 'sonner';

export function GroceryList() {
  const { id } = useParams();
  const [filter, setFilter] = useState('all');
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

    const name = newItemName.trim();
    const newItem = {
      name,
      quantity: newItemQuantity.trim(),
      checked: false,
    };

    const newItems = [...(groceryList?.items || []), newItem];
    updateMutation.mutate(newItems, {
      onSuccess: () => toast.success(`"${name}" added`),
    });

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

  const checkedCount = groceryList?.items?.filter(item => item.checked).length || 0;
  const totalCount = groceryList?.items?.length || 0;
  const progressPercent = totalCount > 0 ? Math.round((checkedCount / totalCount) * 100) : 0;

  if (isLoading) return <LoadingPage />;
  if (error) return <p className="text-red-600">Error: {error.message}</p>;

  return (
    <div className="max-w-2xl mx-auto">
      {/* Header */}
      <div className="mb-6 flex items-center justify-between">
        <div className="flex items-center gap-4">
          <Link 
            to="/grocery"
            className="p-2 hover:bg-gray-100 rounded-lg transition"
            title="Back to grocery lists"
          >
            <FiArrowLeft size={24} />
          </Link>
          <h1 className="text-3xl font-bold text-gray-900">{groceryList?.name || 'Grocery List'}</h1>
        </div>
      </div>

      {/* Progress Bar */}
      <div className="mb-6 bg-white rounded-lg border border-gray-200 shadow-sm p-4">
        <div className="flex justify-between items-center mb-2">
          <p className="text-sm font-medium text-gray-900">Progress</p>
          <p className="text-sm text-gray-500">{checkedCount} of {totalCount}</p>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
          <div
            className="bg-indigo-600 h-3 transition-all duration-300"
            style={{ width: `${progressPercent}%` }}
          />
        </div>
        <p className="text-xs text-gray-500 mt-2">{progressPercent}% complete</p>
      </div>

      {/* Add Item Form */}
      <div className="mb-6 bg-white rounded-lg border border-gray-200 shadow-sm p-4">
        <form onSubmit={handleAddItem} className="flex gap-3">
          <input
            type="text"
            placeholder="Quantity (e.g., 2 lbs)"
            value={newItemQuantity}
            onChange={(e) => setNewItemQuantity(e.target.value)}
            className="flex-shrink px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition w-24"
          />
          <input
            type="text"
            placeholder="Item name"
            value={newItemName}
            onChange={(e) => setNewItemName(e.target.value)}
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-transparent outline-none transition"
            required
          />
          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 disabled:bg-indigo-400 disabled:cursor-not-allowed transition flex-shrink-0"
          >
            <FiPlus size={18} />
            Add
          </button>
        </form>
      </div>

      {/* Filter Buttons */}
      <div className="mb-6 flex gap-2">
        {['all', 'unchecked', 'checked'].map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-4 py-2 rounded-lg font-medium transition capitalize ${
              filter === f
                ? 'bg-indigo-600 text-white'
                : 'bg-gray-100 text-gray-700 hover:bg-gray-200'
            }`}
          >
            {f === 'unchecked' ? 'To Do' : f === 'checked' ? 'Done' : 'All'}
          </button>
        ))}
      </div>

      {/* Items List */}
      {filteredItems?.length ? (
        <div className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden">
          <AnimatePresence initial={false}>
            {filteredItems.map((item) => {
              const originalIndex = groceryList.items.indexOf(item);
              return (
                <motion.div
                  key={originalIndex}
                  layout
                  initial={{ opacity: 0, height: 0 }}
                  animate={{ opacity: 1, height: 'auto' }}
                  exit={{ opacity: 0, height: 0, overflow: 'hidden' }}
                  transition={{ duration: 0.18 }}
                  className="flex items-center gap-4 p-4 border-b border-gray-200 last:border-b-0 hover:bg-gray-50 transition-colors"
                >
                  <input
                    type="checkbox"
                    checked={item.checked}
                    onChange={() => handleCheck(originalIndex)}
                    disabled={updateMutation.isPending}
                    className="w-5 h-5 text-indigo-600 rounded cursor-pointer"
                  />
                  <div className="flex-1 min-w-0">
                    <label
                      className={`text-sm font-medium cursor-pointer block transition-all duration-200 ${
                        item.checked
                          ? 'line-through text-gray-400'
                          : 'text-gray-900'
                      }`}
                    >
                      {item.quantity ? `${item.quantity} ${item.name}` : item.name}
                    </label>
                  </div>
                  <button
                    className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition flex-shrink-0 disabled:opacity-50 disabled:cursor-not-allowed"
                    onClick={() => handleDeleteItem(originalIndex)}
                    disabled={updateMutation.isPending}
                    title="Delete item"
                  >
                    <FiTrash2 size={18} />
                  </button>
                </motion.div>
              );
            })}
          </AnimatePresence>
        </div>
      ) : (
        <div className="text-center py-8 bg-white rounded-lg border border-gray-200">
          <p className="text-gray-500">No items to show.</p>
        </div>
      )}
    </div>
  );
}