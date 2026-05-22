# Recipio Database Schema Design Document

## Overview

This document describes the database schema for **Recipio**, a recipe management application. The authoritative schema definition is `backend/internal/sqlite_db/schema.tmpl` — keep this document in sync when making schema changes.

## Schema Details

### Table: `recipes`

| Column       | Type     | Constraints                                  | Description                        |
|--------------|----------|----------------------------------------------|------------------------------------|
| id           | text     | PRIMARY KEY                                  | UUID, unique identifier            |
| user_id      | text     | FOREIGN KEY references users(id)             | Owner of the recipe                |
| name         | text     | NOT NULL                                     | Name of the recipe                 |
| instruction  | text     | NOT NULL                                     | Cooking instructions (as text)     |
| source_url   | text     |                                              | Original source link (optional)    |
| is_favourite | boolean  | NOT NULL, DEFAULT FALSE                      | Whether the recipe is favourited   |
| created_at   | text     |                                              | RFC3339 timestamp of creation      |

### Table: `ingredients`

| Column      | Type     | Constraints                                  | Description                              |
|-------------|----------|----------------------------------------------|------------------------------------------|
| id          | integer  | PRIMARY KEY                                  | Unique identifier for each ingredient    |
| recipe_id   | text     | NOT NULL, FOREIGN KEY references recipes(id) | Links ingredient to a recipe             |
| name        | text     | NOT NULL                                     | Name of the ingredient                   |
| quantity    | text     |                                              | Quantity or measurement (optional)       |

### Table: `recipe_tags`

| Column    | Type | Constraints                                  | Description          |
|-----------|------|----------------------------------------------|----------------------|
| recipe_id | text | NOT NULL, FOREIGN KEY references recipes(id) | Links tag to recipe  |
| tag       | text | NOT NULL                                     | Tag value            |

Primary key: `(recipe_id, tag)`

### Table: `meal_plan`

| Column  | Type    | Constraints                              | Description                        |
|---------|---------|------------------------------------------|------------------------------------|
| id      | integer | PRIMARY KEY                              | Unique identifier for each meal plan |
| user_id | text    | NOT NULL, FOREIGN KEY references users(id) | Owner of the meal plan             |

### Table: `meal_plan_recipes`

Junction table linking meal plans to recipes (many-to-many).

| Column       | Type    | Constraints                                    | Description                        |
|--------------|---------|------------------------------------------------|------------------------------------|
| meal_plan_id | integer | NOT NULL, FOREIGN KEY references meal_plan(id) | Links to a meal plan               |
| recipe_id    | text    | NOT NULL, FOREIGN KEY references recipes(id)   | Links to a recipe                  |

Primary key: `(meal_plan_id, recipe_id)`

### Table: `grocery_lists`

| Column       | Type    | Constraints                                  | Description                              |
|--------------|---------|----------------------------------------------|------------------------------------------|
| id           | integer | PRIMARY KEY                                  | Unique identifier for each grocery list  |
| user_id      | text    | NOT NULL, FOREIGN KEY references users(id)   | Owner of the grocery list                |
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

### Table: `ingredient_categories`

Lookup table seeded at startup from `internal/recipes/categories.go` (kept in sync with `doc/grocerylist_categories.md`).

| Column | Type    | Constraints       | Description                                   |
|--------|---------|-------------------|-----------------------------------------------|
| id     | integer | PRIMARY KEY       | 1-indexed position in `internal/recipes/categories.go` (seeded with explicit ids) |
| name   | text    | NOT NULL, UNIQUE  | Category display name (e.g. "Meat", "Fruits")                                     |

### Table: `ingredient_category_map`

Cache of categorization results keyed by a normalized ingredient name (lowercase, whitespace collapsed to underscores). Populated lazily when grocery lists are read; can be rebuilt with `cmd/recategorize`.

| Column          | Type    | Constraints                                                  | Description                              |
|-----------------|---------|--------------------------------------------------------------|------------------------------------------|
| normalized_name | text    | PRIMARY KEY                                                  | Lowercase + underscores key              |
| category_id     | integer | NOT NULL, FOREIGN KEY references ingredient_categories(id)   | Resolved category for this ingredient    |

## AuthN Schema

### Table: `users`

| Column         | Type      | Constraints              | Description                            |
|----------------|-----------|--------------------------|----------------------------------------|
| id             | text      | PRIMARY KEY              | UUID, unique user ID                   |
| email          | text      | NOT NULL, UNIQUE         | Associated email                       |
| name           | text      |                          | Display name                           |
| created        | timestamp | NOT NULL                 | Creation date                          |
| email_verified | boolean   | NOT NULL, DEFAULT FALSE  | Whether the user's email is verified   |

### Table: `email_verifications`

| Column  | Type      | Constraints                              | Description                      |
|---------|-----------|------------------------------------------|----------------------------------|
| token   | text      | PRIMARY KEY                              | SHA-256 hashed verification token |
| user_id | text      | NOT NULL, FOREIGN KEY references users(id) | Associated user                |
| expires | timestamp | NOT NULL                                 | Expiration time (24 hours)       |

### Table: `sessions`

Token must be SHA-256 hashed before storage.

| Column  | Type      | Constraints                              | Description              |
|---------|-----------|------------------------------------------|--------------------------|
| token   | text      | PRIMARY KEY                              | SHA-256 hashed token     |
| user_id | text      | NOT NULL, FOREIGN KEY references users(id) | Associated user        |
| expires | timestamp | NOT NULL                                 | Expiration time          |

### Table: `credentials`

| Column   | Type | Constraints                              | Description        |
|----------|------|------------------------------------------|--------------------|
| user_id  | text | NOT NULL, FOREIGN KEY references users(id) | Associated user  |
| password | text | NOT NULL                                 | bcrypt salted hash |

### Table: `oauth`

Composite primary key on `(provider, sub)` since `sub` is only unique per provider. `UNIQUE(user_id, provider)` prevents a user from linking more than one account per provider.

| Column   | Type | Constraints                                           | Description                                 |
|----------|------|-------------------------------------------------------|---------------------------------------------|
| user_id  | text | NOT NULL, FOREIGN KEY references users(id), UNIQUE with provider | Associated user                  |
| provider | text | NOT NULL                                              | OAuth provider name (e.g. "google")         |
| sub      | text | NOT NULL                                              | Subject claim — unique ID within provider   |

Primary key: `(provider, sub)`
Unique constraint: `(user_id, provider)`

## Table Definitions (SQL)

```sql
CREATE TABLE IF NOT EXISTS recipes (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    name TEXT NOT NULL,
    instruction TEXT NOT NULL,
    source_url TEXT,
    is_favourite BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TEXT,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS ingredients (
    id INTEGER PRIMARY KEY,
    recipe_id TEXT NOT NULL,
    name TEXT NOT NULL,
    quantity TEXT,
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

CREATE TABLE IF NOT EXISTS recipe_tags (
    recipe_id TEXT NOT NULL,
    tag TEXT NOT NULL,
    PRIMARY KEY (recipe_id, tag),
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

CREATE TABLE IF NOT EXISTS meal_plan (
    id INTEGER PRIMARY KEY,
    user_id TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS meal_plan_recipes (
    meal_plan_id INTEGER NOT NULL,
    recipe_id TEXT NOT NULL,
    PRIMARY KEY (meal_plan_id, recipe_id),
    FOREIGN KEY (meal_plan_id) REFERENCES meal_plan(id),
    FOREIGN KEY (recipe_id) REFERENCES recipes(id)
);

CREATE TABLE IF NOT EXISTS grocery_lists (
    id INTEGER PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    meal_plan_id INTEGER,
    FOREIGN KEY (user_id) REFERENCES users(id),
    FOREIGN KEY (meal_plan_id) REFERENCES meal_plan(id)
);

CREATE TABLE IF NOT EXISTS grocery_list_items (
    id INTEGER PRIMARY KEY,
    grocery_list_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    quantity TEXT,
    checked BOOLEAN DEFAULT FALSE,
    FOREIGN KEY (grocery_list_id) REFERENCES grocery_lists(id)
);

CREATE TABLE IF NOT EXISTS ingredient_categories (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS ingredient_category_map (
    normalized_name TEXT PRIMARY KEY,
    category_id INTEGER NOT NULL,
    FOREIGN KEY (category_id) REFERENCES ingredient_categories(id)
);

CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    email TEXT NOT NULL UNIQUE,
    name TEXT,
    created TIMESTAMP NOT NULL,
    email_verified BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS email_verifications (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS sessions (
    token TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    expires TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS credentials (
    user_id TEXT NOT NULL,
    password TEXT NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS oauth (
    user_id TEXT NOT NULL,
    provider TEXT NOT NULL,
    sub TEXT NOT NULL,
    PRIMARY KEY (provider, sub),
    UNIQUE (user_id, provider),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```
