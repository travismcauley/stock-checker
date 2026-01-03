import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { AuthProvider, useAuth } from './context/AuthContext'
import { MyStoresProvider } from './context/MyStoresContext'
import { MyProductsProvider } from './context/MyProductsContext'
import { ToastProvider } from './components/Toast'
import { Layout } from './components/Layout'
import { Home } from './pages/Home'
import { StoreSearch } from './pages/StoreSearch'
import { MyStores } from './pages/MyStores'
import { ProductSearch } from './pages/ProductSearch'
import { MyProducts } from './pages/MyProducts'
import { CheckStock } from './pages/CheckStock'
import { Login } from './pages/Login'

function ProtectedApp() {
  const { isAuthenticated, isLoading } = useAuth()

  // Show loading spinner while checking auth
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <div className="text-center">
          <div className="w-12 h-12 border-4 border-blue-600 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    )
  }

  // Show login page if not authenticated
  if (!isAuthenticated) {
    return <Login />
  }

  // Show the app if authenticated
  return (
    <MyStoresProvider>
      <MyProductsProvider>
        <Layout>
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/stores" element={<MyStores />} />
            <Route path="/stores/search" element={<StoreSearch />} />
            <Route path="/products" element={<MyProducts />} />
            <Route path="/products/search" element={<ProductSearch />} />
            <Route path="/check-stock" element={<CheckStock />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Layout>
      </MyProductsProvider>
    </MyStoresProvider>
  )
}

function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <ToastProvider>
          <ProtectedApp />
        </ToastProvider>
      </AuthProvider>
    </BrowserRouter>
  )
}

export default App
