import './App.css'
import { RouterProvider, NavLink, Outlet } from 'react-router'
import { router, sidebarLinks} from './routes.jsx'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';
import Login from '/src/pages/common/Login';
import { AuthProvider, useAuth } from '/src/pages/common/AuthContext';

const queryClient = new QueryClient();

function App() {
    return (
    <AuthProvider>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
      </QueryClientProvider>
    </AuthProvider>
    )
}

export function Layout() {
  const { isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return <Login/>
  }
  return (
    <>
      <HeaderBar />
      <div className='ui large top attached menu pointing'>
          <SidebarNavigation/>
      </div>
      <div className="maincontent">
        <Outlet />
      </div>
    </>
  )
}

function HeaderBar() {
  return (
    <span className="headerBar">
      <h1>Recipio</h1>
      <div className='userbox'>
        <UserInfoBox />
      </div>
    </span>
  )
}

function UserInfoBox() {
  const { user, logout } = useAuth();
  return (
    <div className="userinfo">
      <b>{user.username}</b>
      <a href="#" onClick={logout}>Log out</a>
    </div>
  )
}


function createNavigationLink({label, dst}, index) {
  return (
    <NavLink className="item" key={index} to={dst}>{label}</NavLink>
  )
}

function SidebarNavigation() {

  return (
    <>
       {sidebarLinks.map(createNavigationLink)}
    </>
  )
}

export function HomePage() {
    return (
        <>
        <h1>Hello world!</h1>
        <p>This app is still under construction. Stay tune for more!</p>
        </>
    )
}

export default App
