# Overview
The backend server will act as the engine for processing user requests and database management.


## Architecture

### Endpoints
The backend will have serve the following APIs:

1. **POST /recipes**: Create a new recipe.
   - Request Body: `{ "name": "Recipe Name", "ingredients": [...], "steps": [...] }`
   - Response: `{ "id": "recipe_id", "message": "Recipe created successfully" }`

2. **GET /recipes/{id}**: Retrieve a specific recipe.
   - Response: `{ "id": "recipe_id", "name": "Recipe Name", "ingredients": [...], "steps": [...] }`

3. **GET /recipes**: Retrieves all  recipe.
   - Response: `{ "id": "recipe_id", "name": "Recipe Name", "ingredients": [...], "steps": [...] }`

4. **POST /meal-plans**: Create a meal plan for the week.
   - Request Body: `{ "recipes": ["recipe_id1", "recipe_id2"] }`
   - Response: `{ "id": "meal_plan_id", "message": "Meal plan created successfully" }`

5. **GET /grocery-list/{meal_plan_id}**: Generate a grocery list based on a meal plan.
   - Response: `{ "ingredients": [...] }`