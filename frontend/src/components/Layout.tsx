import type { ReactNode } from 'react'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useMyStores } from '../context/MyStoresContext'
import { useMyProducts } from '../context/MyProductsContext'
import { useAuth } from '../context/AuthContext'

interface LayoutProps {
  children: ReactNode
}

export function Layout({ children }: LayoutProps) {
  const location = useLocation()
  const navigate = useNavigate()
  const { stores: myStores } = useMyStores()
  const { products: myProducts } = useMyProducts()
  const { user, isLoading: authLoading, isAuthenticated, login, logout } = useAuth()

  const canCheckStock = myStores.length > 0 && myProducts.length > 0

  const navItems = [
    { path: '/stores', label: 'My Stores', shortLabel: 'Stores', badge: myStores.length || undefined },
    { path: '/stores/search', label: 'Search Stores', shortLabel: 'Search' },
    { path: '/products', label: 'My Products', shortLabel: 'Products', badge: myProducts.length || undefined },
    { path: '/products/search', label: 'Search Products', shortLabel: 'Search' },
  ]

  const handleCheckStock = () => {
    if (canCheckStock) {
      navigate('/check-stock')
    }
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <header className="bg-blue-600 text-white shadow-lg">
        <div className="container mx-auto px-4">
          <div className="flex items-center justify-between py-3 sm:py-4">
            <Link to="/" className="text-lg sm:text-2xl font-bold hover:text-blue-100 transition-colors">
              <span className="hidden sm:inline">Pokemon Stock Checker</span>
              <span className="sm:hidden">Stock Checker</span>
            </Link>
            <div className="flex items-center gap-2 sm:gap-3">
              {!canCheckStock && (
                <span className="text-blue-200 text-xs sm:text-sm hidden md:block">
                  {myStores.length === 0 && myProducts.length === 0
                    ? 'Add stores & products'
                    : myStores.length === 0
                    ? 'Add stores'
                    : 'Add products'}
                </span>
              )}
              <button
                onClick={handleCheckStock}
                disabled={!canCheckStock}
                className="px-3 sm:px-4 py-1.5 sm:py-2 bg-yellow-500 text-gray-900 rounded-lg font-semibold hover:bg-yellow-400 transition-colors disabled:opacity-50 disabled:cursor-not-allowed text-sm sm:text-base"
              >
                <span className="hidden sm:inline">Check Stock</span>
                <span className="sm:hidden">Check</span>
              </button>

              {/* Auth controls */}
              {authLoading ? (
                <div className="w-8 h-8 rounded-full bg-blue-500 animate-pulse" />
              ) : isAuthenticated && user ? (
                <div className="flex items-center gap-2">
                  <img
                    src={user.pictureUrl}
                    alt={user.name}
                    className="w-8 h-8 rounded-full border-2 border-blue-400"
                    title={user.name}
                  />
                  <button
                    onClick={logout}
                    className="hidden sm:block px-3 py-1.5 text-sm text-blue-200 hover:text-white transition-colors"
                  >
                    Logout
                  </button>
                </div>
              ) : (
                <button
                  onClick={login}
                  className="px-3 sm:px-4 py-1.5 sm:py-2 bg-white text-blue-600 rounded-lg font-semibold hover:bg-blue-50 transition-colors text-sm sm:text-base"
                >
                  Login
                </button>
              )}
            </div>
          </div>
          <nav className="flex gap-1 pb-2 overflow-x-auto scrollbar-hide -mx-4 px-4 sm:mx-0 sm:px-0">
            {navItems.map((item, index) => {
              const isActive = location.pathname === item.path
              // Group items: stores together, products together
              const isFirstInGroup = index === 0 || index === 2
              return (
                <Link
                  key={item.path}
                  to={item.path}
                  className={`px-3 sm:px-4 py-1.5 sm:py-2 rounded-t-lg text-xs sm:text-sm font-medium transition-colors whitespace-nowrap flex-shrink-0 ${
                    isActive
                      ? 'bg-gray-100 text-gray-900'
                      : 'text-blue-100 hover:bg-blue-500'
                  } ${!isFirstInGroup ? 'ml-0.5' : index > 0 ? 'ml-2 sm:ml-4' : ''}`}
                >
                  <span className="sm:hidden">{item.shortLabel}</span>
                  <span className="hidden sm:inline">{item.label}</span>
                  {item.badge !== undefined && item.badge > 0 && (
                    <span className={`ml-1.5 sm:ml-2 px-1.5 sm:px-2 py-0.5 text-xs rounded-full ${
                      isActive ? 'bg-blue-600 text-white' : 'bg-blue-500 text-white'
                    }`}>
                      {item.badge}
                    </span>
                  )}
                </Link>
              )
            })}
          </nav>
        </div>
      </header>

      <main className="container mx-auto px-4 py-4 sm:py-8">{children}</main>

      {/* Mobile bottom padding for better scrolling */}
      <div className="h-4 sm:hidden" />
    </div>
  )
}
