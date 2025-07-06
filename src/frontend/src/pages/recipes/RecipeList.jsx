import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { deleteRecipe, getRecipes } from "./recipe_apis";
import { Form } from "react-router";
import LoadingPage from '/src/pages/common/LoadingPage'
import { useEffect, useState } from 'react';

function displayTagsToString(tags) {
  if (typeof tags === 'undefined') {
    return "None"
  }
  return tags.join(", ")
}

function RecipeDeleteIcon({ id }) {
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: deleteRecipe,
    onSuccess: () => {
      // Invalidate recipes list to refetch
      queryClient.invalidateQueries(['recipes']);
    },
    onError: (error) => { alert(`Failed to delete: ${error}` ) }
  });

  const handleClick = () => {
    mutation.mutate(id)
  }
  if (id === undefined) {
    id = 0
  }
  return (
    <div>
      <i className='recipeTrashButton link large aligned trash red icon' onClick={handleClick}></i>
    </div>
  )
}

function RecipeRowPreview({recipe}) {
  return (
    <div className="item">
      <div className='right floated content'>
        <RecipeDeleteIcon id={recipe.id}/>
      </div>
      <a className="header" href={`/recipe/view/${recipe.id}`} state={{ recipe }}>{recipe.name}</a>
      Tags: {displayTagsToString(recipe.tags)}
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

  if (recipes === undefined) {
    return <>No recipes found?</>
  }
  return (
    <>
      <div class="ui icon input">
        <input type="text" placeholder="Search..." value={searchQuery} onChange={(e) => setSearchQuery(e.target.value)}/>
        <i class="search link icon" />
      </div>
      <div className="ui middle large aligned divided list">
      { recipeItems?.sort(
          (a, b) => a.name.localeCompare(b.name))
        .map( (recipe) => (
          <RecipeRowPreview recipe={recipe} key={recipe.id}/>
      ))}
      </div>
    </>
  )
}

export function RecipeList() {
  const { data, isLoading, error } = useQuery({
      queryKey: ['recipes'],
      queryFn: getRecipes,
  });

  if (isLoading) return <LoadingPage />
  if (error) return <p>Error: {error.message}</p>
  return (
    <>
    <RecipeListRows recipes={data}/>
    <Form action="/recipe/add/">
      <button className="ui button primary" type="submit">Create new recipe</button>
    </Form>
    </>
  )
}
