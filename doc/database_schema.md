# Recipio Database Schema Design Document

## Overview

This document describes the initial database schema design for **Recipio**, a recipe management application. The schema is designed to efficiently store and retrieve recipes, including their names, ingredients, and instructions.

## Schema Details

### Table: `recipes`

| Column       | Type     | Constraints      | Description                        |
|--------------|----------|-----------------|------------------------------------|
| id           | integer  | PRIMARY KEY     | Unique identifier for each recipe  |
| name         | text     | NOT NULL        | Name of the recipe                 |
| instruction  | text     | NOT NULL        | Cooking instructions (as text)     |

### Table: `ingredients`

| Column      | Type     | Constraints                          | Description                              |
|-------------|----------|--------------------------------------|------------------------------------------|
| id          | integer  | PRIMARY KEY                          | Unique identifier for each ingredient    |
| recipe_id   | integer  | NOT NULL, FOREIGN KEY references recipes(id) | Links ingredient to a recipe      |
| name        | text     | NOT NULL                             | Name of the ingredient                   |
| quantity    | text     |                                      | Quantity or measurement (optional)       |

### Table: `meal_plan`

| Column  | Type    | Constraints | Description                        |
|---------|---------|-------------|------------------------------------|
| id      | integer | PRIMARY KEY | Unique identifier for each meal plan |
| user_id | integer | NOT NULL    | Owner of the meal plan             |

### Table: `meal_plan_recipes`

Junction table linking meal plans to recipes (many-to-many).

| Column       | Type    | Constraints                                  | Description                        |
|--------------|---------|----------------------------------------------|------------------------------------|
| meal_plan_id | integer | NOT NULL, FOREIGN KEY references meal_plan(id) | Links to a meal plan             |
| recipe_id    | integer | NOT NULL, FOREIGN KEY references recipes(id)   | Links to a recipe                |

Primary key: `(meal_plan_id, recipe_id)`

### Table: `grocery_lists`

| Column       | Type    | Constraints                                  | Description                              |
|--------------|---------|----------------------------------------------|------------------------------------------|
| id           | integer | PRIMARY KEY                                  | Unique identifier for each grocery list  |
| name         | text    | NOT NULL                                     | Name of the grocery list                 |
| meal_plan_id | integer | FOREIGN KEY references meal_plan(id)         | Associated meal plan (optional)          |

### Table: `grocery_list_items`

| Column          | Type    | Constraints                                        | Description                               |
|-----------------|---------|----------------------------------------------------|-------------------------------------------|
| id              | integer | PRIMARY KEY                                        | Unique identifier for each item           |
| grocery_list_id | integer | NOT NULL, FOREIGN KEY references grocery_lists(id) | Links item to its grocery list            |
| name            | text    | NOT NULL                                           | Name of the item                          |
| quantity        | text    |                                                    | Quantity or measurement (optional)        |
| checked         | boolean | DEFAULT FALSE                                      | Whether the item has been checked off     |

## AuthN Schema

### Table: `users`

| Column  | Type      | Constraints        | Description          |
|---------|-----------|--------------------|----------------------|
| id      | text      | PRIMARY KEY        | UUID, unique user ID |
| email   | text      | NOT NULL, UNIQUE   | Associated email     |
| name    | text      |                    | Display name         |
| created | timestamp | NOT NULL           | Creation date        |

### Table: `sessions`

Token must be SHA-256 hashed before storage.

| Column  | Type      | Constraints                        | Description              |
|---------|-----------|------------------------------------|--------------------------|
| token   | text      | PRIMARY KEY                        | SHA-256 hashed token     |
| user_id | text      | NOT NULL, FOREIGN KEY references users(id) | Associated user    |
| expires | timestamp | NOT NULL                           | Expiration time          |

### Table: `credentials`

| Column   | Type | Constraints                              | Description  |
|----------|------|------------------------------------------|--------------|
| user_id  | text | NOT NULL, FOREIGN KEY references users(id) | Associated user |
| password | text | NOT NULL                                 | Salted hash  |

### Table: `oauth`

Composite primary key on `(provider, sub)` since `sub` is only unique per provider. `UNIQUE(user_id, provider)` prevents 
a user from linking more than one account per provider.

| Column   | Type | Constraints                              | Description                                 |
|----------|------|------------------------------------------|---------------------------------------------|
| user_id  | text | NOT NULL, FOREIGN KEY references users(id), UNIQUE with provider | Associated user |
| provider | text | NOT NULL                                 | OAuth provider name (e.g. "google")         |
| sub      | text | NOT NULL                                 | Subject claim — unique ID within provider   |

Primary key: `(provider, sub)`
Unique constraint: `(user_id, provider)`

## Example Table Definitions (SQL)

```sql
CREATE TABLE recipes (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    instruction TEXT NOT NULL
);

CREATE TABLE ingredients (
    id INTEGER PRIMARY KEY,
    recipe_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    quantity TEXT,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

CREATE TABLE meal_plan (
    id INTEGER PRIMARY KEY,
    user_id INTEGER NOT NULL
);

CREATE TABLE meal_plan_recipes (
    meal_plan_id INTEGER NOT NULL,
    recipe_id INTEGER NOT NULL,
    PRIMARY KEY (meal_plan_id, recipe_id),
    FOREIGN KEY (meal_plan_id) REFERENCES meal_plan(id),
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

CREATE TABLE grocery_lists (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    meal_plan_id INTEGER,
    FOREIGN KEY (meal_plan_id) REFERENCES meal_plan(id)
);

CREATE TABLE grocery_list_items (
    id INTEGER PRIMARY KEY,
    grocery_list_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    quantity TEXT,
    checked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (grocery_list_id) REFERENCES grocery_lists(id)
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT,
    created TIMESTAMP NOT NULL
);

CREATE TABLE sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE credentials (
    user_id TEXT NOT NULL,
    password TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE oauth (
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    sub TEXT NOT NULL,
    PRIMARY KEY (provider, sub),
    UNIQUE (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

