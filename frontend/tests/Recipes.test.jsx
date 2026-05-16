import '@testing-library/jest-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'

vi.mock('../src/pages/recipes/recipe_apis', () => ({
  getRecipes: vi.fn(),
  getRecipeId: vi.fn(),
  createRecipe: vi.fn(),
  updateRecipe: vi.fn(),
  deleteRecipe: vi.fn(),
  parseRecipe: vi.fn(),
}))

import { RecipeList } from '../src/pages/recipes/RecipeList'
import { ViewRecipe } from '../src/pages/recipes/ViewRecipe'
import { AddRecipeForm } from '../src/pages/recipes/AddRecipe'
import * as apis from '../src/pages/recipes/recipe_apis'

const mockRecipes = [
  {
    id: '1',
    name: 'French Toast',
    tags: ['breakfast', 'easy'],
    ingredients: [
      { quantity: '2', name: 'eggs' },
      { quantity: '2 slices', name: 'bread' },
    ],
    instructions: ['Beat eggs until smooth', 'Dip bread in egg mix', 'Cook on pan'],
  },
  {
    id: '2',
    name: 'Pancakes',
    tags: ['breakfast'],
    ingredients: [
      { quantity: '1 cup', name: 'flour' },
      { quantity: '1', name: 'egg' },
    ],
    instructions: ['Mix ingredients', 'Pour on heated pan', 'Flip when bubbles form'],
  },
]

function makeQueryClient() {
  return new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
}

function renderWithRouter(ui, { route = '/', path = '/' } = {}) {
  const queryClient = makeQueryClient()
  const router = createMemoryRouter(
    [{ path, element: <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider> }],
    { initialEntries: [route] }
  )
  return render(<RouterProvider router={router} />)
}

// ─── RecipeList ────────────────────────────────────────────────────────────────

describe('RecipeList', () => {
  beforeEach(() => vi.clearAllMocks())

  it('shows loading state while fetching', () => {
    apis.getRecipes.mockReturnValue(new Promise(() => {}))
    renderWithRouter(<RecipeList />)
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('shows empty state when there are no recipes', async () => {
    apis.getRecipes.mockResolvedValue([])
    renderWithRouter(<RecipeList />)
    await waitFor(() =>
      expect(screen.getByText(/no recipes found yet/i)).toBeInTheDocument()
    )
  })

  it('renders recipe names as links', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    renderWithRouter(<RecipeList />)
    await waitFor(() => {
      expect(screen.getByRole('link', { name: /french toast/i })).toBeInTheDocument()
      expect(screen.getByRole('link', { name: /pancakes/i })).toBeInTheDocument()
    })
  })

  it('recipe links point to the correct view URL', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    renderWithRouter(<RecipeList />)
    await waitFor(() => {
      expect(screen.getByRole('link', { name: /french toast/i }))
        .toHaveAttribute('href', '/recipe/view/1')
    })
  })

  it('renders recipe tags', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    renderWithRouter(<RecipeList />)
    await waitFor(() =>
      expect(screen.getAllByText(/breakfast/i).length).toBeGreaterThan(0)
    )
  })

  it('filters recipes by search query', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    renderWithRouter(<RecipeList />)
    await waitFor(() =>
      expect(screen.getByRole('link', { name: /french toast/i })).toBeInTheDocument()
    )
    fireEvent.change(screen.getByPlaceholderText(/search recipes/i), {
      target: { value: 'pancake' },
    })
    await waitFor(() => {
      expect(screen.queryByRole('link', { name: /french toast/i })).not.toBeInTheDocument()
      expect(screen.getByRole('link', { name: /pancakes/i })).toBeInTheDocument()
    })
  })

  it('shows a New Recipe button', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    renderWithRouter(<RecipeList />)
    await waitFor(() =>
      expect(screen.getByRole('button', { name: /new recipe/i })).toBeInTheDocument()
    )
  })

  it('shows a delete button for each recipe', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    renderWithRouter(<RecipeList />)
    await waitFor(() =>
      expect(screen.getAllByTitle('Delete recipe')).toHaveLength(mockRecipes.length)
    )
  })

  it('asks for confirmation before deleting', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    apis.deleteRecipe.mockResolvedValue(undefined)
    vi.spyOn(window, 'confirm').mockReturnValue(false)

    renderWithRouter(<RecipeList />)
    await waitFor(() => screen.getAllByTitle('Delete recipe'))
    fireEvent.click(screen.getAllByTitle('Delete recipe')[0])

    expect(window.confirm).toHaveBeenCalledWith('Are you sure you want to delete this recipe?')
    expect(apis.deleteRecipe).not.toHaveBeenCalled()
  })

  it('calls deleteRecipe when deletion is confirmed', async () => {
    apis.getRecipes.mockResolvedValue(mockRecipes)
    apis.deleteRecipe.mockResolvedValue(undefined)
    vi.spyOn(window, 'confirm').mockReturnValue(true)

    renderWithRouter(<RecipeList />)
    await waitFor(() => screen.getAllByTitle('Delete recipe'))
    fireEvent.click(screen.getAllByTitle('Delete recipe')[0])

    await waitFor(() => expect(apis.deleteRecipe).toHaveBeenCalled())
    expect(apis.deleteRecipe.mock.calls[0][0]).toBe(mockRecipes[0].id)
  })
})

// ─── ViewRecipe ────────────────────────────────────────────────────────────────

describe('ViewRecipe', () => {
  beforeEach(() => vi.clearAllMocks())

  function renderViewRecipe(id = '1') {
    const queryClient = makeQueryClient()
    const router = createMemoryRouter(
      [
        {
          path: '/recipe/view/:uid',
          element: (
            <QueryClientProvider client={queryClient}>
              <ViewRecipe />
            </QueryClientProvider>
          ),
        },
        { path: '/recipe', element: <div>Recipe List</div> },
      ],
      { initialEntries: [`/recipe/view/${id}`] }
    )
    return render(<RouterProvider router={router} />)
  }

  it('shows loading state while fetching', () => {
    apis.getRecipeId.mockReturnValue(new Promise(() => {}))
    renderViewRecipe()
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('renders the recipe name', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    renderViewRecipe()
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: /french toast/i })).toBeInTheDocument()
    )
  })

  it('renders ingredient quantities and names', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    renderViewRecipe()
    await waitFor(() => {
      expect(screen.getByText('eggs')).toBeInTheDocument()
      expect(screen.getByText('bread')).toBeInTheDocument()
      expect(screen.getByText('2')).toBeInTheDocument()
    })
  })

  it('renders instructions', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    renderViewRecipe()
    await waitFor(() => {
      expect(screen.getByText('Beat eggs until smooth')).toBeInTheDocument()
      expect(screen.getByText('Cook on pan')).toBeInTheDocument()
    })
  })

  it('renders tags', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    renderViewRecipe()
    await waitFor(() => {
      expect(screen.getByText('breakfast')).toBeInTheDocument()
      expect(screen.getByText('easy')).toBeInTheDocument()
    })
  })

  it('does not render tags section when recipe has no tags', async () => {
    apis.getRecipeId.mockResolvedValue({ ...mockRecipes[0], tags: null })
    renderViewRecipe()
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: /french toast/i })).toBeInTheDocument()
    )
    expect(screen.queryByText('breakfast')).not.toBeInTheDocument()
  })

  it('shows edit form when edit button is clicked', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    renderViewRecipe()
    await waitFor(() => screen.getByTitle('Edit recipe'))
    fireEvent.click(screen.getByTitle('Edit recipe'))
    expect(screen.getByRole('heading', { name: /edit recipe/i })).toBeInTheDocument()
  })

  it('pre-populates the name field in edit mode', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    renderViewRecipe()
    await waitFor(() => screen.getByTitle('Edit recipe'))
    fireEvent.click(screen.getByTitle('Edit recipe'))
    expect(screen.getByDisplayValue('French Toast')).toBeInTheDocument()
  })

  it('returns to view mode when cancel is clicked', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    renderViewRecipe()
    await waitFor(() => screen.getByTitle('Edit recipe'))
    fireEvent.click(screen.getByTitle('Edit recipe'))
    fireEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: /french toast/i })).toBeInTheDocument()
    )
  })

  it('calls updateRecipe when save is submitted', async () => {
    apis.getRecipeId.mockResolvedValue(mockRecipes[0])
    apis.updateRecipe.mockResolvedValue(undefined)
    renderViewRecipe()
    await waitFor(() => screen.getByTitle('Edit recipe'))
    fireEvent.click(screen.getByTitle('Edit recipe'))
    const form = screen.getByRole('button', { name: /save changes/i }).closest('form')
    fireEvent.submit(form)
    await waitFor(() => expect(apis.updateRecipe).toHaveBeenCalled())
    expect(apis.updateRecipe.mock.calls[0][0]).toMatchObject({ id: '1' })
  })
})

// ─── AddRecipeForm ─────────────────────────────────────────────────────────────

describe('AddRecipeForm', () => {
  beforeEach(() => vi.clearAllMocks())

  function renderAddRecipe() {
    const queryClient = makeQueryClient()
    const router = createMemoryRouter(
      [
        {
          path: '/recipe/add',
          element: (
            <QueryClientProvider client={queryClient}>
              <AddRecipeForm />
            </QueryClientProvider>
          ),
        },
        { path: '/recipe', element: <div>Recipe List</div> },
      ],
      { initialEntries: ['/recipe/add'] }
    )
    return render(<RouterProvider router={router} />)
  }

  it('shows the Create New Recipe heading', () => {
    renderAddRecipe()
    expect(screen.getByRole('heading', { name: /create new recipe/i })).toBeInTheDocument()
  })

  it('shows manual entry mode by default', () => {
    renderAddRecipe()
    expect(screen.getByLabelText(/recipe name/i)).toBeInTheDocument()
    expect(screen.getByLabelText(/ingredients/i)).toBeInTheDocument()
  })

  it('has Manual Entry and AI toggle buttons', () => {
    renderAddRecipe()
    expect(screen.getByRole('button', { name: /manual entry/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /add recipe using ai/i })).toBeInTheDocument()
  })

  it('accepts input in the recipe name field', () => {
    renderAddRecipe()
    fireEvent.change(screen.getByLabelText(/recipe name/i), {
      target: { name: 'recipeName', value: 'Scrambled Eggs' },
    })
    expect(screen.getByDisplayValue('Scrambled Eggs')).toBeInTheDocument()
  })

  it('accepts input in the ingredients field', () => {
    renderAddRecipe()
    fireEvent.change(screen.getByLabelText(/ingredients/i), {
      target: { name: 'ingredientsList', value: '2 eggs\n1 tbsp butter' },
    })
    expect(screen.getByLabelText(/ingredients/i).value).toContain('2 eggs')
  })

  it('switches to AI mode when the AI button is clicked', () => {
    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    expect(screen.getByLabelText(/paste your recipe/i)).toBeInTheDocument()
  })

  it('disables Preview button when AI textarea is empty', () => {
    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    expect(screen.getByRole('button', { name: /preview recipe/i })).toBeDisabled()
  })

  it('enables Preview button after text is entered', () => {
    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    fireEvent.change(screen.getByLabelText(/paste your recipe/i), {
      target: { value: 'Some recipe text here' },
    })
    expect(screen.getByRole('button', { name: /preview recipe/i })).not.toBeDisabled()
  })

  it('shows a character counter in AI mode', () => {
    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    expect(screen.getByText(/0 \/ 2000/)).toBeInTheDocument()
  })

  it('calls parseRecipe and shows the preview panel on success', async () => {
    apis.parseRecipe.mockResolvedValue({
      name: 'French Toast',
      ingredients: [{ quantity: '2', name: 'eggs' }],
      instructions: ['Beat eggs', 'Cook on pan'],
    })

    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    fireEvent.change(screen.getByLabelText(/paste your recipe/i), {
      target: { value: 'French toast recipe text' },
    })
    fireEvent.click(screen.getByRole('button', { name: /preview recipe/i }))

    await waitFor(() => {
      expect(screen.getByText('Recipe Preview')).toBeInTheDocument()
      expect(screen.getByText('French Toast')).toBeInTheDocument()
      expect(screen.getByText('eggs')).toBeInTheDocument()
      expect(screen.getByText(/1\. Beat eggs/)).toBeInTheDocument()
    })
  })

  it('shows a timeout error message when parsing times out', async () => {
    apis.parseRecipe.mockRejectedValue(new Error('TIMEOUT'))

    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    fireEvent.change(screen.getByLabelText(/paste your recipe/i), {
      target: { value: 'Some recipe text' },
    })
    fireEvent.click(screen.getByRole('button', { name: /preview recipe/i }))

    await waitFor(() =>
      expect(screen.getByText(/parsing timed out/i)).toBeInTheDocument()
    )
  })

  it('shows a rate limit error message on 429 response', async () => {
    apis.parseRecipe.mockRejectedValue(new Error('RATE_LIMIT'))

    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    fireEvent.change(screen.getByLabelText(/paste your recipe/i), {
      target: { value: 'Some recipe text' },
    })
    fireEvent.click(screen.getByRole('button', { name: /preview recipe/i }))

    await waitFor(() =>
      expect(screen.getByText(/too many requests/i)).toBeInTheDocument()
    )
  })

  it('"Use This Recipe" switches to manual mode with the parsed data pre-filled', async () => {
    apis.parseRecipe.mockResolvedValue({
      name: 'French Toast',
      ingredients: [{ quantity: '2', name: 'eggs' }],
      instructions: ['Beat eggs'],
    })

    renderAddRecipe()
    fireEvent.click(screen.getByRole('button', { name: /add recipe using ai/i }))
    fireEvent.change(screen.getByLabelText(/paste your recipe/i), {
      target: { value: 'French toast recipe text' },
    })
    fireEvent.click(screen.getByRole('button', { name: /preview recipe/i }))
    await waitFor(() => screen.getByText('Recipe Preview'))

    fireEvent.click(screen.getByRole('button', { name: /use this recipe/i }))

    await waitFor(() =>
      expect(screen.getByDisplayValue('French Toast')).toBeInTheDocument()
    )
  })

  it('calls createRecipe when the manual form is submitted', async () => {
    apis.createRecipe.mockResolvedValue({ id: '99', name: 'Test Recipe' })
    const { container } = renderAddRecipe()

    fireEvent.change(screen.getByLabelText(/recipe name/i), {
      target: { name: 'recipeName', value: 'Test Recipe' },
    })
    fireEvent.change(screen.getByLabelText(/ingredients/i), {
      target: { name: 'ingredientsList', value: '2 eggs' },
    })
    fireEvent.change(container.querySelector('[name="instructions"]'), {
      target: { name: 'instructions', value: 'Beat eggs' },
    })
    fireEvent.click(screen.getByRole('button', { name: /save recipe/i }))

    await waitFor(() => expect(apis.createRecipe).toHaveBeenCalled())
    expect(apis.createRecipe.mock.calls[0][0]).toMatchObject({ name: 'Test Recipe' })
  })
})
