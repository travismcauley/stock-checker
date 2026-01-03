import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
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

function App() {
  return (
    <BrowserRouter>
      <ToastProvider>
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
      </ToastProvider>
    </BrowserRouter>
  )
}

export default App
