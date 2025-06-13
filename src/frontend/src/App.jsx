import './App.css'
import { RouterProvider, NavLink, Outlet } from 'react-router'
import { router, sidebarLinks} from './routes.jsx'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

const queryClient = new QueryClient();

function App() {
    return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
    )
}

export function Layout() {
  return (
    <>
      <HeaderBar />
      <div className="container">
        <SidebarNavigation />
        <div className="maincontent">
          <Outlet />
        </div>
      </div>
    </>
  )
}

const user = {
  name: "blanker"
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
  return (
    <div className="userinfo">
      <b>{user.name}</b>
      <a href="#">Profile</a>
      <a href="#">Log out</a>
    </div>
  )
}


function createNavigationLink({label, dst}, index) {
  return (
    <NavLink key={index} to={dst}>{label}</NavLink>
  )
}

function SidebarNavigation() {

  return (
    <nav className="sidebar">
       {sidebarLinks.map(createNavigationLink)}
    </nav>
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
