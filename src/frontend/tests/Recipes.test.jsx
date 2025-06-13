import '@testing-library/jest-dom'
import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { RecipeRowPreview, RecipeList, RecipeGridPreview, RecipeGridList, ViewRecipe } from "../src/Recipes";

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
  }
];

describe("RecipeRowPreview", () => {
  it("renders the recipe name as a link", () => {
    const recipe = mockRecipes[0];
    render(<RecipeRowPreview recipe={recipe} />);
    const link = screen.getByRole("link", { name: recipe.name });
    expect(link).toBeInTheDocument();
    expect(link).toHaveAttribute("href", `/recipe/view/${recipe.uid}`);
  });

  it("renders tags as a string", () => {
    const recipe = mockRecipes[0];
    render(<RecipeRowPreview recipe={recipe} />);
    expect(screen.getByText(/Tags:/)).toHaveTextContent("Tags: breakfast, easy");
  });

  it("renders 'None' if tags are missing", () => {
    const recipe = { ...mockRecipes[0], tags: undefined };
    render(<RecipeRowPreview recipe={recipe} />);
    expect(screen.getByText(/Tags:/)).toHaveTextContent("Tags: None");
  });
});

describe("RecipeList", () => {
  it("renders all recipe names in the list", () => {
    render(<RecipeList />);
    mockRecipes.forEach(recipe => {
      expect(screen.getByRole("link", { name: recipe.name })).toBeInTheDocument();
    });
  });
});

/*
unused functions for now

describe("RecipeGridPreview", () => {
  it("renders the recipe name and image", () => {
    const recipe = mockRecipes[0];
    render(<RecipeGridPreview recipe={recipe} />);
    expect(screen.getByText(recipe.name)).toBeInTheDocument();
    expect(screen.getByAltText(recipe.name)).toBeInTheDocument();
  });
});

describe("RecipeGridList", () => {
  it("renders all recipes in grid", () => {
    render(<RecipeGridList />);
    mockRecipes.forEach(recipe => {
      expect(screen.getByText(recipe.name)).toBeInTheDocument();
    });
  });
});
*/