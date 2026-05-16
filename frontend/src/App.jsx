import './App.css'
import { RouterProvider, NavLink, Outlet, useLocation } from 'react-router'
import { router, sidebarLinks} from './routes.jsx'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { useState } from 'react';
import Login from '/src/pages/common/Login';
import { AuthProvider, useAuth } from '/src/pages/common/AuthContext';
import { FiLogOut, FiMenu, FiX } from 'react-icons/fi';
import { motion, AnimatePresence } from 'framer-motion';
import { Toaster } from 'sonner';

const queryClient = new QueryClient();

function App() {
    return (
    <AuthProvider>
      <QueryClientProvider client={queryClient}>
        <RouterProvider router={router} />
        <Toaster richColors position="bottom-right" />
      </QueryClientProvider>
    </AuthProvider>
    )
}

export function Layout() {
  const { isAuthenticated } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const location = useLocation();

  if (!isAuthenticated) {
    return <Login/>
  }

  return (
    <div className="flex flex-col h-screen bg-gray-50">
      <HeaderBar onMenuClick={() => setSidebarOpen(!sidebarOpen)} />

      <div className="flex flex-1 overflow-hidden">
        <SidebarNavigation isOpen={sidebarOpen} onClose={() => setSidebarOpen(false)} />
        <div className="flex-1 relative overflow-hidden">
          <AnimatePresence mode="sync">
            <motion.div
              key={location.pathname}
              className="absolute inset-0 overflow-auto"
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.18 }}
            >
              <div className="maincontent">
                <Outlet />
              </div>
            </motion.div>
          </AnimatePresence>
        </div>
      </div>
    </div>
  )
}

function HeaderBar({ onMenuClick }) {
  return (
    <header className="bg-white border-b border-gray-200 shadow-sm">
      <div className="px-6 py-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <button 
            onClick={onMenuClick}
            className="lg:hidden p-2 hover:bg-gray-100 rounded-lg transition"
          >
            <FiMenu size={24} />
          </button>
          <h1 className="text-3xl font-bold text-indigo-600">Recipio</h1>
        </div>
        <UserInfoBox />
      </div>
    </header>
  )
}

function UserInfoBox() {
  const { user, logout } = useAuth();
  return (
    <div className="flex items-center gap-4">
      <div className="text-right">
        <p className="text-sm font-medium text-gray-900">{user?.username}</p>
      </div>
      <button
        onClick={logout}
        className="inline-flex items-center gap-2 px-3 py-2 text-sm font-medium text-gray-700 hover:text-red-600 hover:bg-red-50 rounded-lg transition"
        title="Log out"
      >
        <FiLogOut size={18} />
      </button>
    </div>
  )
}


function createNavigationLink({label, dst}, index) {
  return (
    <NavLink 
      className={({ isActive }) => `
        block px-4 py-3 text-sm font-medium transition-colors rounded-lg
        ${isActive 
          ? 'bg-indigo-50 text-indigo-600 border-l-4 border-indigo-600' 
          : 'text-gray-700 hover:bg-gray-100'
        }
      `} 
      key={index} 
      to={dst}
    >
      {label}
    </NavLink>
  )
}

function SidebarNavigation({ isOpen, onClose }) {
  return (
    <>
      {/* Mobile sidebar overlay */}
      {isOpen && (
        <div 
          className="fixed inset-0 bg-black bg-opacity-50 lg:hidden z-30"
          onClick={onClose}
        />
      )}
      
      {/* Sidebar */}
      <nav className={`
        fixed lg:static inset-y-0 left-0 z-40 w-64 bg-white border-r border-gray-200
        transform transition-transform duration-300 ease-in-out
        ${isOpen ? 'translate-x-0' : '-translate-x-full lg:translate-x-0'}
        overflow-y-auto pt-4
      `}>
        <div className="px-3 space-y-2">
          {sidebarLinks.map(createNavigationLink)}
        </div>
      </nav>
    </>
  )
}

export function HomePage() {
    return (
        <div>
          <h1 className="text-4xl font-bold text-gray-900 mb-4">Welcome to Recipio</h1>
          <p className="text-lg text-gray-600">Manage your recipes, meal plans, and grocery lists all in one place.</p>
        </div>
    )
}

export default App
