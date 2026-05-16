import '@testing-library/jest-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'

vi.mock('../src/pages/mealplan/mealplan_apis', () => ({
  getMealPlans: vi.fn(),
  getGroceryList: vi.fn(),
  deleteMealPlan: vi.fn(),
  createMealPlan: vi.fn(),
}))

import { MealplanList } from '../src/pages/mealplan/MealplanList'
import * as apis from '../src/pages/mealplan/mealplan_apis'

const mockPlans = [
  { id: 1, recipe_names: ['French Toast', 'Pancakes'] },
  { id: 2, recipe_names: ['Spaghetti Bolognese'] },
]

function makeQueryClient() {
  return new QueryClient({ defaultOptions: { queries: { retry: false } } })
}

function renderWithRouter(ui) {
  const queryClient = makeQueryClient()
  const router = createMemoryRouter(
    [{ path: '/', element: <QueryClientProvider client={queryClient}>{ui}</QueryClientProvider> }],
    { initialEntries: ['/'] }
  )
  return render(<RouterProvider router={router} />)
}

describe('MealplanList', () => {
  beforeEach(() => vi.clearAllMocks())

  it('shows loading state while fetching', () => {
    apis.getMealPlans.mockReturnValue(new Promise(() => {}))
    renderWithRouter(<MealplanList />)
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('shows empty state when there are no meal plans', async () => {
    apis.getMealPlans.mockResolvedValue([])
    renderWithRouter(<MealplanList />)
    await waitFor(() =>
      expect(screen.getByText(/no meal plans yet/i)).toBeInTheDocument()
    )
  })

  it('renders meal plan IDs and recipe names', async () => {
    apis.getMealPlans.mockResolvedValue(mockPlans)
    renderWithRouter(<MealplanList />)
    await waitFor(() => {
      expect(screen.getByText('Meal Plan #1')).toBeInTheDocument()
      expect(screen.getByText('Meal Plan #2')).toBeInTheDocument()
      expect(screen.getByText('French Toast, Pancakes')).toBeInTheDocument()
      expect(screen.getByText('Spaghetti Bolognese')).toBeInTheDocument()
    })
  })

  it('shows "No recipes" when a plan has no recipe names', async () => {
    apis.getMealPlans.mockResolvedValue([{ id: 3, recipe_names: [] }])
    renderWithRouter(<MealplanList />)
    await waitFor(() =>
      expect(screen.getByText('No recipes')).toBeInTheDocument()
    )
  })

  it('shows a New Meal Plan link', async () => {
    apis.getMealPlans.mockResolvedValue(mockPlans)
    renderWithRouter(<MealplanList />)
    await waitFor(() =>
      expect(screen.getByRole('link', { name: /new meal plan/i })).toBeInTheDocument()
    )
  })

  it('fetches and shows ingredients when the expand button is clicked', async () => {
    apis.getMealPlans.mockResolvedValue([mockPlans[0]])
    apis.getGroceryList.mockResolvedValue(['2 eggs', '2 slices bread'])
    renderWithRouter(<MealplanList />)
    await waitFor(() => screen.getByText('Meal Plan #1'))

    fireEvent.click(screen.getByTitle('Show ingredients'))
    await waitFor(() => {
      expect(screen.getByText('• 2 eggs')).toBeInTheDocument()
      expect(screen.getByText('• 2 slices bread')).toBeInTheDocument()
    })
    expect(apis.getGroceryList).toHaveBeenCalledWith(1)
  })

  it('asks for confirmation before deleting a meal plan', async () => {
    apis.getMealPlans.mockResolvedValue([mockPlans[0]])
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    renderWithRouter(<MealplanList />)
    await waitFor(() => screen.getByText('Meal Plan #1'))

    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    expect(window.confirm).toHaveBeenCalledWith('Are you sure you want to delete this meal plan?')
    expect(apis.deleteMealPlan).not.toHaveBeenCalled()
  })

  it('calls deleteMealPlan when deletion is confirmed', async () => {
    apis.getMealPlans.mockResolvedValue([mockPlans[0]])
    apis.deleteMealPlan.mockResolvedValue(undefined)
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    renderWithRouter(<MealplanList />)
    await waitFor(() => screen.getByText('Meal Plan #1'))

    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    await waitFor(() => expect(apis.deleteMealPlan).toHaveBeenCalledWith(1))
  })
})
