
import { createBrowserRouter, Outlet } from "react-router";
import { AddRecipeForm } from "./pages/recipes/AddRecipe";
import { RecipeList } from "./pages/recipes/RecipeList";
import { ViewRecipe } from "./pages/recipes/ViewRecipe";
import { HomePage, Layout } from "./App";

export const sidebarLinks = [
  {label: "Home", dst: "/"},
  {label: "Recipes", dst: "/recipe"},
  {label: "Mealplan", dst: "/mealplan"},
]

function ErrorPage() {
    return <h1>This is not the site that you're looking for.</h1>
}

export const router = createBrowserRouter([
    { 
        Component: Layout,
        children: [
            {
                path: "/",
                Component: HomePage
            },
            { 
                path: "/recipe", 
                children: [
                    {index: true, Component: RecipeList},
                    {path: "add", Component: AddRecipeForm},
                    {
                        path: "view/:uid", 
                        Component: ViewRecipe
                    }
                ]
            },
            {
                path: "/mealplan",
                Component: ErrorPage,
            }
        ]
    },
])