
import { createBrowserRouter, Outlet } from "react-router";
import { AddRecipeForm, RecipeRowList, ViewRecipe} from "./Recipes";
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
                    {index: true, Component: RecipeRowList},
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