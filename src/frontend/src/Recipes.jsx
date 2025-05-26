import { useState } from "react";
import { useNavigate, useParams } from "react-router";

const mockRecipes = [
  {
    uid: 1,
    name: "French toast",
    tags: [
      "breakfast",
      "easy",
    ],
    ingredientList: [
      "2 pieces eggs",
      "2 pieces bread",
      "50 mL milk",
    ],
    instructions: [
      "Beats eggs until smooth",
      "Add milk to the egg mixture",
      "Dip the bread into the milk-egg mixture",
      "Cook on a non-stick pan with butter until desired texture"
    ]
  },
  {
    uid: 2,
    name: "Pancakes",
    tags: [
      "breakfast",
      "easy",
    ],
    ingredientList: [
      "200 grams flour",
      "300 mL milk",
      "1 pieces egg",
      "2 tablespoons sugar",
      "1 teaspoons baking powder",
    ],
    instructions: [
      "Mix all dry ingredients in a bowl",
      "Add milk and egg, then whisk until smooth",
      "Heat a lightly oiled pan over medium heat",
      "Pour batter onto the pan and cook until bubbles form, then flip and cook until golden"
    ]
  },
  {
    uid: 3,
    name: "Scrambled Eggs",
    ingredientList: [
      "3 pieces eggs",
      "30 mL milk",
      "1 tablespoons butter",
      "0.5 teaspoons salt",
    ],
    instructions: [
      "Crack the eggs into a bowl and add milk and salt",
      "Whisk the mixture until well combined",
      "Melt butter in a pan over medium heat",
      "Pour in the egg mixture and stir gently until just set"
    ]
  },
  {
    uid: 4,
    name: "Grilled Cheese Sandwich",
    tags: [
      "easy",
    ],
    ingredientList: [
      "2 pieces bread",
      "2 pieces cheese slices",
      "1 tablespoons butter",
    ],
    instructions: [
      "Butter one side of each bread slice",
      "Place cheese between the unbuttered sides of the bread",
      "Cook in a pan over medium heat until golden on both sides and cheese is melted"
    ]
  },
  {
    uid: 5,
    name: "Banana Smoothie",
    tags: [
      "snack",
    ],
    ingredientList: [
      "1 pieces banana",
      "200 mL milk",
      "100 grams yogurt",
      "1 tablespoons honey",
    ],
    instructions: [
      "Peel and slice the banana",
      "Add banana, milk, yogurt, and honey to a blender",
      "Blend until smooth",
      "Serve chilled"
    ]
  }
];

function RecipeGridPreview({ recipe }) {

  return (
    <div>
      <img src="#" className="recipePreviewImage" height={200} width={200} alt={recipe.name}/>
      <h2 className="recipePreviewName">{recipe.name}</h2>
    </div>
  )
}

export default function RecipeGridList() {

    let recipes = mockRecipes;
    return (
      <div className="recipePreviewGrid">
      { recipes.map((recipe) => (
        <RecipeGridPreview recipe={recipe} key={recipe.uid} />
      ))}
      </div>
    )
}

function displayTagsToString(tags) {
  if (typeof tags === 'undefined') {
    return "None"
  }
  return tags.join(", ")
}

function RecipeRowPreview({recipe}) {
  return (
    <div className="recipePreviewRowItem item">
      <a className="header" href={`/recipe/view/${recipe.uid}`}>{recipe.name}</a>
      Tags: {displayTagsToString(recipe.tags)}
    </div>
  )
}

export function RecipeRowList() {
  let recipes = mockRecipes;
  return (
    <>
    <h2>Available Recipes</h2>
    <div className="recipePreviewRow ui celled relaxed selection list large animated">
      { recipes.map( (recipe) => (
        <RecipeRowPreview recipe={recipe} key={recipe.uid} />
      ))}
    </div>
    <form action="/recipe/add/">
      <button className="ui button primary" type="submit">Create new recipe</button>
    </form>
      </>
  )
}

export function AddRecipeForm() {
  const navigate = useNavigate();
  const blankRecipe = {
    name: "",
    ingredientList: [],
    instructions: [],
  }
  const [recipe, setRecipe] = useState(blankRecipe);
  const handleFormChange = (e) => {
    if (e.target.name == "recipeName") {
      setRecipe({...recipe, name: e.target.value})
    }
    if (e.target.name == "ingredientsList") {
      setRecipe({...recipe, ingredientList: e.target.value.split(/\r?\n/)});
    }
    if (e.target.name == "instructions") {
      setRecipe({...recipe, instructions: e.target.value.split(/\r?\n/)});
    }
  }
  const addRecipeSubmitHandler = (e) => {
    e.preventDefault();
    console.log(recipe)
    setRecipe(blankRecipe)
  }

  return (
    <>
      <h2>Add Recipe</h2>
      <form className="recipeFormName ui form large" onSubmit={addRecipeSubmitHandler}>
        <div className="field">
        <label>Recipe name:</label>
        <input 
          type="text"
          placeholder="Recipe Name"
          className="recipeNameForm"
          name="recipeName"
          value={recipe.name}
          onChange={handleFormChange}
          required
        ></input>
        </div>
        <div className="field">
          <label>Ingredients</label>
          <textarea
            placeholder="Ingredients"
            name="ingredientsList"
            onChange={handleFormChange}
            value={recipe.ingredientList.join('\r\n')}
            required
            ></textarea>
        </div>
        <div className="field">
          <label>Instructions</label>
          <textarea
            placeholder="Instructions"
            name="instructions"
            onChange={handleFormChange}
            value={recipe.instructions.join('\r\n')}
            required
          ></textarea>
        </div>
        <button className="ui button primary" type="submit">Save</button>
        <button onClick={() => navigate("/recipe")} className="ui button negative" type="button">Cancel</button>
      </form>
    </>
  )
}

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
  let params = useParams()
  let filtered = mockRecipes.filter(item => item.uid == params.uid)
  const recipe = filtered[0];

  const listIngredients = recipe.ingredientList.map((item) => 
    <div className="item">
      {item}
    </div>
  )

  const listSteps = recipe.instructions.map((step) =>
    <div className="item">
      {step}
    </div>
  )

  const listTags = recipe.tags ? recipe.tags.map((tag) => (
    <a className={"ui tag label " + getClassFromTag(tag)} key={tag}>{tag}</a> 
  )) : null;

  return (
    <div className="recipeView">
      <h2 className="ui header">{recipe.name}</h2>
      <h3 className="ingredients ui horizontal divider header">Ingredients</h3>
      <div className="ui list big bulleted">
        {listIngredients}
      </div>
      <h3 className="instructions ui horizontal divider header">Instructions</h3>
      <div className="ui list big ordered">
        {listSteps}
      </div>
      {listTags}
    </div>
  );
}