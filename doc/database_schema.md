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
```