import { useState } from "react";
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { createRecipe } from "./recipe_apis";

export function AddRecipeForm() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const mutation = useMutation({
    mutationFn: createRecipe,
    onSuccess: (data, variable, context) => {
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
