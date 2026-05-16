import '@testing-library/jest-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createMemoryRouter, RouterProvider } from 'react-router'

vi.mock('framer-motion', async (importOriginal) => {
  const mod = await importOriginal()
  return { ...mod, AnimatePresence: ({ children }) => children }
})

vi.mock('../src/pages/grocery/grocery_apis', () => ({
  getGroceryLists: vi.fn(),
  getGroceryListById: vi.fn(),
  updateGroceryList: vi.fn(),
  deleteGroceryList: vi.fn(),
  createGroceryList: vi.fn(),
}))

import { GroceryListList } from '../src/pages/grocery/GroceryListList'
import { GroceryList } from '../src/pages/grocery/GroceryList'
import * as apis from '../src/pages/grocery/grocery_apis'

const mockLists = [
  {
    id: 1,
    name: 'Weekly Shop',
    items: [
      { name: 'eggs', quantity: '2', checked: false },
      { name: 'milk', quantity: '1L', checked: true },
    ],
  },
  {
    id: 2,
    name: 'Party Supplies',
    items: [],
  },
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

// ─── GroceryListList ───────────────────────────────────────────────────────────

describe('GroceryListList', () => {
  beforeEach(() => vi.clearAllMocks())

  it('shows loading state while fetching', () => {
    apis.getGroceryLists.mockReturnValue(new Promise(() => {}))
    renderWithRouter(<GroceryListList />)
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('shows empty state when there are no lists', async () => {
    apis.getGroceryLists.mockResolvedValue([])
    renderWithRouter(<GroceryListList />)
    await waitFor(() =>
      expect(screen.getByText(/no grocery lists yet/i)).toBeInTheDocument()
    )
  })

  it('renders grocery list names', async () => {
    apis.getGroceryLists.mockResolvedValue(mockLists)
    renderWithRouter(<GroceryListList />)
    await waitFor(() => {
      expect(screen.getByText('Weekly Shop')).toBeInTheDocument()
      expect(screen.getByText('Party Supplies')).toBeInTheDocument()
    })
  })

  it('shows checked item count for each list', async () => {
    apis.getGroceryLists.mockResolvedValue([mockLists[0]])
    renderWithRouter(<GroceryListList />)
    await waitFor(() =>
      expect(screen.getByText('1 of 2 items checked')).toBeInTheDocument()
    )
  })

  it('view links point to the correct URL', async () => {
    apis.getGroceryLists.mockResolvedValue(mockLists)
    renderWithRouter(<GroceryListList />)
    await waitFor(() => {
      const viewLinks = screen.getAllByRole('link', { name: /view/i })
      expect(viewLinks[0]).toHaveAttribute('href', '/grocery/view/1')
      expect(viewLinks[1]).toHaveAttribute('href', '/grocery/view/2')
    })
  })

  it('asks for confirmation before deleting a list', async () => {
    apis.getGroceryLists.mockResolvedValue([mockLists[0]])
    vi.spyOn(window, 'confirm').mockReturnValue(false)
    renderWithRouter(<GroceryListList />)
    await waitFor(() => screen.getByText('Weekly Shop'))

    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    expect(window.confirm).toHaveBeenCalledWith('Are you sure you want to delete this grocery list?')
    expect(apis.deleteGroceryList).not.toHaveBeenCalled()
  })

  it('calls deleteGroceryList when deletion is confirmed', async () => {
    apis.getGroceryLists.mockResolvedValue([mockLists[0]])
    apis.deleteGroceryList.mockResolvedValue(undefined)
    vi.spyOn(window, 'confirm').mockReturnValue(true)
    renderWithRouter(<GroceryListList />)
    await waitFor(() => screen.getByText('Weekly Shop'))

    fireEvent.click(screen.getByRole('button', { name: /delete/i }))
    await waitFor(() => expect(apis.deleteGroceryList).toHaveBeenCalledWith(1))
  })
})

// ─── GroceryList ───────────────────────────────────────────────────────────────

describe('GroceryList', () => {
  beforeEach(() => vi.clearAllMocks())

  function renderGroceryList(id = '1') {
    const queryClient = makeQueryClient()
    const router = createMemoryRouter(
      [
        {
          path: '/grocery/view/:id',
          element: (
            <QueryClientProvider client={queryClient}>
              <GroceryList />
            </QueryClientProvider>
          ),
        },
      ],
      { initialEntries: [`/grocery/view/${id}`] }
    )
    return render(<RouterProvider router={router} />)
  }

  it('shows loading state while fetching', () => {
    apis.getGroceryListById.mockReturnValue(new Promise(() => {}))
    renderGroceryList()
    expect(screen.getByText(/loading/i)).toBeInTheDocument()
  })

  it('renders the list name as a heading', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    renderGroceryList()
    await waitFor(() =>
      expect(screen.getByRole('heading', { name: 'Weekly Shop' })).toBeInTheDocument()
    )
  })

  it('renders all items with quantity and name', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    renderGroceryList()
    await waitFor(() => {
      expect(screen.getByText('2 eggs')).toBeInTheDocument()
      expect(screen.getByText('1L milk')).toBeInTheDocument()
    })
  })

  it('shows progress count and percentage', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    renderGroceryList()
    await waitFor(() => {
      expect(screen.getByText('1 of 2')).toBeInTheDocument()
      expect(screen.getByText('50% complete')).toBeInTheDocument()
    })
  })

  it('renders checked items with line-through styling', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    renderGroceryList()
    await waitFor(() =>
      expect(screen.getByText('1L milk')).toHaveClass('line-through')
    )
  })

  it('calls updateGroceryList with the item toggled when a checkbox is clicked', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    apis.updateGroceryList.mockResolvedValue(undefined)
    renderGroceryList()
    await waitFor(() => screen.getByText('2 eggs'))

    const checkboxes = screen.getAllByRole('checkbox')
    fireEvent.click(checkboxes[0])
    await waitFor(() => expect(apis.updateGroceryList).toHaveBeenCalled())
    const updatedItems = apis.updateGroceryList.mock.calls[0][1]
    expect(updatedItems[0].checked).toBe(true)
  })

  it('calls updateGroceryList with the item removed when delete is clicked', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    apis.updateGroceryList.mockResolvedValue(undefined)
    renderGroceryList()
    await waitFor(() => screen.getByText('2 eggs'))

    fireEvent.click(screen.getAllByTitle('Delete item')[0])
    await waitFor(() => expect(apis.updateGroceryList).toHaveBeenCalled())
    const updatedItems = apis.updateGroceryList.mock.calls[0][1]
    expect(updatedItems).toHaveLength(1)
    expect(updatedItems[0].name).toBe('milk')
  })

  it('calls updateGroceryList with the new item appended when the add form is submitted', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    apis.updateGroceryList.mockResolvedValue(undefined)
    renderGroceryList()
    await waitFor(() => screen.getByText('2 eggs'))

    fireEvent.change(screen.getByPlaceholderText('Item name'), { target: { value: 'butter' } })
    fireEvent.change(screen.getByPlaceholderText(/quantity/i), { target: { value: '250g' } })
    fireEvent.click(screen.getByRole('button', { name: /^add$/i }))

    await waitFor(() => expect(apis.updateGroceryList).toHaveBeenCalled())
    const updatedItems = apis.updateGroceryList.mock.calls[0][1]
    expect(updatedItems).toContainEqual({ name: 'butter', quantity: '250g', checked: false })
  })

  it('shows only unchecked items when To Do tab is clicked', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    renderGroceryList()
    await waitFor(() => screen.getByText('2 eggs'))

    fireEvent.click(screen.getByRole('button', { name: 'To Do' }))
    expect(screen.getByText('2 eggs')).toBeInTheDocument()
    expect(screen.queryByText('1L milk')).not.toBeInTheDocument()
  })

  it('shows only checked items when Done tab is clicked', async () => {
    apis.getGroceryListById.mockResolvedValue(mockLists[0])
    renderGroceryList()
    await waitFor(() => screen.getByText('2 eggs'))

    fireEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(screen.queryByText('2 eggs')).not.toBeInTheDocument()
    expect(screen.getByText('1L milk')).toBeInTheDocument()
  })
})
