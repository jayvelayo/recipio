# Design Doc

## Intro
Personal project for fun. I'm tired of saving multiple recipes in multiple places (youtube, instagram, etc.)

## Objective
 - To easily compile a list of recipes found across the web into a single application
 - To create a meal plan that can be picked based on the available recipes
 - To create a grocery list based on the meal plan

# User stories

- Main server to process requests
- As a User, I should be able to :
    1. create a recipe
        each recipe can be split into multiple parts
    2. view the recipe
    3. Create a list of recipes based on the available recipes for the week
    4. Generate a grocery list based on the list from step 3
- Database to store the recipes
- Datebase to store a list of recipes
- Service to generate grocery list

## System Architecture
The system will consist of the following components:
- **Frontend**: A web-based interface for users to interact with the application.
- **Backend**: A server to handle API requests and business logic.
- **Database**: A storage solution to persist recipes, meal plans, and grocery lists.
- **Services**: A dedicated service for generating grocery lists based on meal plans.

### Architecture Diagram
```
[Frontend] <--> [Backend Server] <--> [Database]
                              |
                              v
                      [Grocery List Service]
```

## Technology Stack
- **Frontend**: React.js (or any preferred framework)
- **Backend**: Go (Golang)
- **Database**: PostgreSQL with GORM ORM
- **Services**: Golang for grocery list generation

## Backend Implementation with Go
The backend will be implemented using native Go (Golang) for its performance and scalability. Below are the updated details:

### Framework
- **Native Go**: The backend will use Go's standard `net/http` package for handling HTTP requests and routing.

### Database Integration
- **GORM**: An ORM library for Go to interact with PostgreSQL.

### API Endpoints
The API endpoints remain the same as previously defined but will be implemented using Go's `net/http` package.

### Service Layer
- The service layer will handle business logic, such as generating grocery lists, and will be implemented as separate Go packages for modularity.

### Testing
- Use Go's built-in `testing` package for unit and integration tests.

### Deployment
- The backend will be containerized using Docker for easy deployment and scalability.

## API Design


## Database Schema
### Tables
1. **Recipes**
   - `id` (Primary Key)
   - `name` (String)
   - `ingredients` (JSON Array)
   - `steps` (JSON Array)

2. **MealPlans**
   - `id` (Primary Key)
   - `recipes` (JSON Array of Recipe IDs)

3. **GroceryLists**
   - `id` (Primary Key)
   - `meal_plan_id` (Foreign Key)
   - `ingredients` (JSON Array)

## Service Details
### Grocery List Generation
- Input: Meal plan ID
- Process:
  1. Fetch the recipes associated with the meal plan.
  2. Extract and aggregate the ingredients from all recipes.
  3. Return a consolidated list of ingredients.
- Output: JSON object containing the grocery list.