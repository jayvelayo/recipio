import { useParams } from "react-router";
import { useQuery } from '@tanstack/react-query';
import { getRecipes } from "./recipe_apis";
import { Link } from "react-router";

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
  const { data, isLoading, error } = useQuery({
      queryKey: ['recipes'],
      queryFn: getRecipes,
  });

  if (isLoading) return <p>Loading...</p>
  if (error) return <p>Error: {error.message}</p>

  let filtered = data.filter(item => item.id == id)
  const recipe = filtered[0];

  const listIngredients = recipe.ingredients.map((item, index) => 
    <li key={index}>{item.quantity} {item.name}</li>
  )

  const listSteps = recipe.instructions.map((step, index) =>
      <li key={index}>{step}</li>
  )

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