import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router';
import { getGroceryLists, deleteGroceryList } from './grocery_apis';
import { SkeletonList } from '/src/pages/common/LoadingPage';
import { FiEye, FiTrash2, FiPlus, FiCheck } from 'react-icons/fi';
import { motion } from 'framer-motion';
import { toast } from 'sonner';

const containerVariants = {
  hidden: {},
  show: { transition: { staggerChildren: 0.06 } },
};

const rowVariants = {
  hidden: { opacity: 0, y: 10 },
  show: { opacity: 1, y: 0, transition: { duration: 0.22 } },
};

function GroceryListRow({ list }) {
  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: () => deleteGroceryList(list.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['grocery-lists'] });
      toast.success('Grocery list deleted');
    },
    onError: () => {
      toast.error('Failed to delete grocery list');
    },
  });

  const checkedCount = list.items?.filter(item => item.checked).length || 0;
  const totalCount = list.items?.length || 0;
  const progressPercent = totalCount > 0 ? Math.round((checkedCount / totalCount) * 100) : 0;

  return (
    <div className="border-b border-gray-200 last:border-b-0 p-4 hover:bg-gray-50 transition">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1">
          <h3 className="font-semibold text-gray-900">{list.name}</h3>
          <div className="text-sm text-gray-600 mt-2">
            <div className="flex items-center gap-2 mb-2">
              <FiCheck size={16} className="text-green-600" />
              <span>{checkedCount} of {totalCount} items checked</span>
            </div>
            {list.mealPlanID && (
              <p className="text-xs text-gray-500">From Meal Plan #{list.mealPlanID}</p>
            )}
          </div>

          {/* Progress bar */}
          <div className="mt-3 w-full bg-gray-200 rounded-full h-2 overflow-hidden">
            <div
              className="bg-indigo-600 h-2 transition-all duration-300"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
          <p className="text-xs text-gray-500 mt-1">{progressPercent}% complete</p>
        </div>
      </div>

      <div className="flex flex-wrap gap-2 mt-4">
        <Link
          to={`/grocery/view/${list.id}`}
          className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 rounded-lg transition"
        >
          <FiEye size={16} />
          View
        </Link>
        <button
          className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-red-600 hover:bg-red-50 rounded-lg transition disabled:opacity-50 disabled:cursor-not-allowed"
          onClick={() => {
            if (confirm('Are you sure you want to delete this grocery list?')) {
              deleteMutation.mutate();
            }
          }}
          disabled={deleteMutation.isPending}
        >
          <FiTrash2 size={16} />
          {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
        </button>
      </div>
    </div>
  );
}

export function GroceryListList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['grocery-lists'],
    queryFn: getGroceryLists,
  });

  if (isLoading) return <SkeletonList />;
  if (error) return <p className="text-red-600">Error: {error.message}</p>;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold text-gray-900">Grocery Lists</h1>
        <Link to="/grocery/add" className="inline-flex items-center gap-2 px-4 py-2 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition">
          <FiPlus size={20} />
          New List
        </Link>
      </div>

      {data?.length ? (
        <motion.div
          className="bg-white rounded-lg border border-gray-200 shadow-sm overflow-hidden"
          variants={containerVariants}
          initial="hidden"
          animate="show"
        >
          {data.map((list) => (
            <motion.div key={list.id} variants={rowVariants}>
              <GroceryListRow list={list} />
            </motion.div>
          ))}
        </motion.div>
      ) : (
        <div className="text-center py-12 bg-white rounded-lg border border-gray-200">
          <p className="text-gray-500 mb-4">No grocery lists yet. Create your first list!</p>
        </div>
      )}
    </div>
  );
}
