import { useParams } from "react-router";
import { useQuery } from '@tanstack/react-query';
import { getRecipeId } from "./recipe_apis";
import { Link } from "react-router";
import LoadingPage from '/src/pages/common/LoadingPage'

function getClassFromTag(tag) {
  const tagColourMap = {
    "easy": ["green"],
    "breakfast": ["yellow"]
  }
  if (tagColourMap[tag] == undefined) {
    return ""
  }
  return tagColourMap[tag].join(" ")
}

export function ViewRecipe() {
  const param = useParams()
  const id = param.uid
  const { data, isLoading, isError, error } = useQuery({
      queryKey: ['recipes', id],
      queryFn: () => getRecipeId(id),
  });
  if (isLoading) return <LoadingPage />
  if (isError) return <p>Error: {error.message}</p>

  // Design API returns single recipe { id, name, ingredients[], steps[] }
  const recipe = data;
  const ingredients = recipe.ingredients ?? [];
  const steps = recipe.steps ?? recipe.instructions ?? [];

  const listIngredients = ingredients.map((item, index) => (
    <li key={index}>{typeof item === 'string' ? item : `${item.quantity ?? ''} ${item.name ?? ''}`.trim()}</li>
  ));
  const listSteps = steps.map((step, index) => <li key={index}>{step}</li>);

  const listTags = recipe.tags ? recipe.tags.map((tag) => (
    <a className={"ui tag label " + getClassFromTag(tag)} key={tag}>{tag}</a> 
  )) : null;
  return (
    <div className="recipeView">
      <h2 className="ui header">{recipe.name}</h2>
      <h3 className="ingredients ui horizontal divider header">Ingredients</h3>
      <div className="ui list big bulleted">
        <ul>{listIngredients}</ul>
      </div>
      <h3 className="instructions ui horizontal divider header">Instructions</h3>
      <div className="ui list big ordered">
        <ol>{listSteps}</ol>
      </div>
      {listTags}
      <Link className="ui button primary" to="/recipe">Back </Link>
    </div>
  );
}