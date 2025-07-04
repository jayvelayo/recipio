import { useQuery } from '@tanstack/react-query';
import { getRecipes } from "./recipe_apis";
import { Form } from "react-router";

function displayTagsToString(tags) {
  if (typeof tags === 'undefined') {
    return "None"
  }
  return tags.join(", ")
}

function RecipeRowPreview({recipe}) {
  return (
    <div className="recipePreviewRowItem item ui grid">
      <div className='fourteen wide column'>
        <a className="header" href={`/recipe/view/${recipe.id}`} state={{ recipe }}>{recipe.name}</a>
        Tags: {displayTagsToString(recipe.tags)}
      </div>
      <div className='column'>
        <i className='recipeTrashButton link large aligned trash red icon'></i>
      </div>
    </div>
  )
}

export function RecipeList() {
  const { data, isLoading, error } = useQuery({
      queryKey: ['recipes'],
      queryFn: getRecipes,
  });

  if (isLoading) return <p>Loading...</p>
  if (error) return <p>Error: {error.message}</p>
  return (
    <>
    <h2>Available Recipes</h2>
    <div className="recipePreviewRow ui celled relaxed selection list large animated">
      { data.map( (recipe) => (
        <RecipeRowPreview recipe={recipe} key={recipe.id} />
      ))}
    </div>
    <Form action="/recipe/add/">
      <button className="ui button primary" type="submit">Create new recipe</button>
    </Form>
      </>
  )
}
