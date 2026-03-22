import React from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { Link } from 'react-router';
import { getGroceryLists, deleteGroceryList } from './grocery_apis';
import LoadingPage from '/src/pages/common/LoadingPage';

function GroceryListRow({ list }) {
  const queryClient = useQueryClient();

  const deleteMutation = useMutation({
    mutationFn: () => deleteGroceryList(list.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['grocery-lists'] });
    },
  });

  const checkedCount = list.items?.filter(item => item.checked).length || 0;
  const totalCount = list.items?.length || 0;

  return (
    <div className="item">
      <div className="content">
        <span className="header">{list.name}</span>
        <div className="description">
          {totalCount} items ({checkedCount} checked)
          {list.mealPlanID && <span> - From Meal Plan #{list.mealPlanID}</span>}
        </div>
        <div className="extra">
          <Link
            to={`/grocery/view/${list.id}`}
            className="ui button small blue"
          >
            View
          </Link>
          <button
            className="ui button small red"
            onClick={() => {
              if (confirm('Are you sure you want to delete this grocery list?')) {
                deleteMutation.mutate();
              }
            }}
            disabled={deleteMutation.isPending}
          >
            {deleteMutation.isPending ? 'Deleting...' : 'Delete'}
          </button>
        </div>
      </div>
    </div>
  );
}

export function GroceryListList() {
  const { data, isLoading, error } = useQuery({
    queryKey: ['grocery-lists'],
    queryFn: getGroceryLists,
  });

  if (isLoading) return <LoadingPage />;
  if (error) return <p>Error: {error.message}</p>;

  return (
    <>
      <h2>Grocery Lists</h2>
      <div className="ui relaxed divided list">
        {data?.length ? (
          data.map((list) => <GroceryListRow list={list} key={list.id} />)
        ) : (
          <p>No grocery lists yet. Create one to get started.</p>
        )}
      </div>
      <Link to="/grocery/add" className="ui button primary">
        Add Grocery List
      </Link>
    </>
  );
}