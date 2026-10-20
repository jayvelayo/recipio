import '@testing-library/jest-dom'
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'

import { AuthProvider, useAuth } from '../src/pages/common/AuthContext'

function Probe() {
  const { isAuthenticated, sessionExpired } = useAuth()
  return (
    <>
      <div>{isAuthenticated ? 'logged-in' : 'logged-out'}</div>
      {sessionExpired && <div>expired</div>}
      <button onClick={() => fetch('/recipes')}>recipes</button>
      <button onClick={() => fetch('/auth/login', { method: 'POST' })}>login</button>
    </>
  )
}

describe('AuthProvider fetch interceptor', () => {
  let originalFetch
  let status
  let mockFetch

  beforeEach(() => {
    originalFetch = window.fetch
    status = 200
    mockFetch = vi.fn(() => Promise.resolve(new Response('', { status })))
    window.fetch = mockFetch
    localStorage.setItem('session', JSON.stringify({ token: 't', email: 'a@b.c', name: 'A' }))
  })

  afterEach(() => {
    window.fetch = originalFetch
    localStorage.clear()
  })

  const renderProbe = () =>
    render(
      <AuthProvider>
        <Probe />
      </AuthProvider>
    )

  it('logs out on 401 from an API call', async () => {
    status = 401
    renderProbe()
    expect(screen.getByText('logged-in')).toBeInTheDocument()

    screen.getByText('recipes').click()

    await waitFor(() => expect(screen.getByText('logged-out')).toBeInTheDocument())
    expect(localStorage.getItem('session')).toBeNull()
    expect(screen.getByText('expired')).toBeInTheDocument()
  })

  it('stays logged in on other error statuses', async () => {
    status = 400
    renderProbe()

    screen.getByText('recipes').click()

    await waitFor(() => expect(mockFetch).toHaveBeenCalled())
    expect(screen.getByText('logged-in')).toBeInTheDocument()
  })

  it('does not log out when /auth/login returns 401', async () => {
    status = 401
    renderProbe()

    screen.getByText('login').click()

    await waitFor(() => expect(mockFetch).toHaveBeenCalled())
    expect(screen.getByText('logged-in')).toBeInTheDocument()
  })

  it('restores the original fetch on unmount', () => {
    const wrapped = window.fetch
    const { unmount } = renderProbe()
    expect(window.fetch).not.toBe(wrapped)
    unmount()
    expect(window.fetch).toBe(wrapped)
  })
})
