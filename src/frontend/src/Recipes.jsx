import { useState } from "react";
import { useNavigate, useParams, Form } from "react-router";
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';

const mockRecipes = [
  {
    id: 2,
    name: "French toast",
    tags: [
      "breakfast",
      "easy",
    ],
    ingredients: [
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
    id: 1,
    name: "Pancakes",
    tags: [
      "breakfast",
      "easy",
    ],
    ingredients: [
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
    ingredients: [
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
    ingredients: [
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
    ingredients: [
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

function getRecipes() {
  return fetch('http://localhost:4002/v1/recipe', { mode: "cors"})
    .then(res => res.json());
}

function createRecipe(newRecipe) {
  return fetch('http://localhost:4002/v1/recipe', { 
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    mode: "cors",
    body: JSON.stringify(newRecipe),
  }).then(res => res.json());
}

function displayTagsToString(tags) {
  if (typeof tags === 'undefined') {
    return "None"
  }
  return tags.join(", ")
}

export function RecipeRowPreview({recipe}) {
  return (
    <div className="recipePreviewRowItem item">
      <a className="header" href={`/recipe/view/${recipe.id}`}>{recipe.name}</a>
      Tags: {displayTagsToString(recipe.tags)}
    </div>
  )
}

export function RecipeList() {
  const { data, isLoading, error } = useQuery({
      queryKey: ['recipes'],
      queryFn: getRecipes,
  });

  if (isLoading) return <p>Loading...</p>;
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

export function AddRecipeForm() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createRecipe,
    onSuccess: () => {
      // Invalidate recipes list to refetch
      queryClient.invalidateQueries(['recipes']);
    },
  });

  const blankRecipe = {
    name: "",
    ingredients: [],
    instructions: [],
  }
  const [recipe, setRecipe] = useState(blankRecipe);
  const handleFormChange = (e) => {
    if (e.target.name == "recipeName") {
      setRecipe({...recipe, name: e.target.value})
    }
    if (e.target.name == "ingredientsList") {
      setRecipe({...recipe, ingredients: e.target.value.split(/\r?\n/)});
    }
    if (e.target.name == "instructions") {
      setRecipe({...recipe, instructions: e.target.value.split(/\r?\n/)});
    }
  }
  const addRecipeSubmitHandler = (e) => {
    e.preventDefault();
    console.log(recipe);
    mutation.mutate(recipe);
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
            value={recipe.ingredients.join('\r\n')}
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
        <button className="ui button primary" type="submit" disabled={mutation.isPending}>
          {mutation.isPending ? 'Saving..' : 'Save'}
        </button>
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
  let filtered = mockRecipes.filter(item => item.id == params.id)
  const recipe = filtered[0];

  const listIngredients = recipe.ingredients.map((item) => 
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