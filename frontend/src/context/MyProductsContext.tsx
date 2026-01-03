import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import type { ReactNode } from 'react'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { StockCheckerService } from '../gen/stockchecker/v1/service_pb.js'
import type { Product } from '../gen/stockchecker/v1/service_pb.js'
import { useAuth } from './AuthContext'

const STORAGE_KEY = 'pokemon-stock-checker-my-products'

interface MyProductsContextType {
  products: Product[]
  addProduct: (product: Product) => void
  removeProduct: (sku: string) => void
  isProductInList: (sku: string) => boolean
  clearProducts: () => void
  isLoading: boolean
}

const MyProductsContext = createContext<MyProductsContextType | undefined>(undefined)

// Custom fetch that includes credentials
const fetchWithCredentials: typeof fetch = (input, init) => {
  return fetch(input, { ...init, credentials: 'include' })
}

// Create transport with credentials to send cookies
const transport = createConnectTransport({
  baseUrl: import.meta.env.VITE_API_URL || 'http://localhost:8080',
  fetch: fetchWithCredentials,
})

const client = createClient(StockCheckerService, transport)

// Helper to serialize Product to a plain object for localStorage
function serializeProduct(product: Product): Record<string, unknown> {
  return {
    sku: product.sku,
    name: product.name,
    salePrice: product.salePrice,
    thumbnailUrl: product.thumbnailUrl,
    productUrl: product.productUrl,
  }
}

// Helper to deserialize from localStorage back to Product-like object
function deserializeProduct(data: Record<string, unknown>): Product {
  return {
    sku: data.sku as string,
    name: data.name as string,
    salePrice: data.salePrice as number,
    thumbnailUrl: data.thumbnailUrl as string,
    productUrl: data.productUrl as string,
    $typeName: 'stockchecker.v1.Product',
  } as Product
}

export function MyProductsProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const [products, setProducts] = useState<Product[]>([])
  const [isLoading, setIsLoading] = useState(true)

  // Load products from API or localStorage
  const loadProducts = useCallback(async () => {
    if (authLoading) return

    setIsLoading(true)
    try {
      if (isAuthenticated) {
        // Load from API
        const response = await client.getMyProducts({})
        setProducts(response.products)
      } else {
        // Load from localStorage
        const saved = localStorage.getItem(STORAGE_KEY)
        if (saved) {
          const parsed = JSON.parse(saved) as Record<string, unknown>[]
          setProducts(parsed.map(deserializeProduct))
        } else {
          setProducts([])
        }
      }
    } catch (e) {
      console.error('Failed to load products:', e)
      // Fall back to localStorage if API fails
      try {
        const saved = localStorage.getItem(STORAGE_KEY)
        if (saved) {
          const parsed = JSON.parse(saved) as Record<string, unknown>[]
          setProducts(parsed.map(deserializeProduct))
        }
      } catch {
        setProducts([])
      }
    } finally {
      setIsLoading(false)
    }
  }, [isAuthenticated, authLoading])

  useEffect(() => {
    loadProducts()
  }, [loadProducts])

  // Save to localStorage when not authenticated
  useEffect(() => {
    if (authLoading || isAuthenticated) return
    try {
      const serialized = products.map(serializeProduct)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(serialized))
    } catch (e) {
      console.error('Failed to save products to localStorage:', e)
    }
  }, [products, isAuthenticated, authLoading])

  const addProduct = async (product: Product) => {
    // Optimistically update local state
    setProducts((prev) => {
      if (prev.some((p) => p.sku === product.sku)) {
        return prev
      }
      return [...prev, product]
    })

    if (isAuthenticated) {
      try {
        await client.addMyProduct({ product })
      } catch (e) {
        console.error('Failed to add product to API:', e)
        // Revert on failure
        setProducts((prev) => prev.filter((p) => p.sku !== product.sku))
      }
    }
  }

  const removeProduct = async (sku: string) => {
    const removedProduct = products.find((p) => p.sku === sku)

    // Optimistically update local state
    setProducts((prev) => prev.filter((p) => p.sku !== sku))

    if (isAuthenticated) {
      try {
        await client.removeMyProduct({ sku })
      } catch (e) {
        console.error('Failed to remove product from API:', e)
        // Revert on failure
        if (removedProduct) {
          setProducts((prev) => [...prev, removedProduct])
        }
      }
    }
  }

  const isProductInList = (sku: string) => {
    return products.some((p) => p.sku === sku)
  }

  const clearProducts = () => {
    setProducts([])
    if (!isAuthenticated) {
      localStorage.removeItem(STORAGE_KEY)
    }
  }

  return (
    <MyProductsContext.Provider
      value={{ products, addProduct, removeProduct, isProductInList, clearProducts, isLoading }}
    >
      {children}
    </MyProductsContext.Provider>
  )
}

export function useMyProducts() {
  const context = useContext(MyProductsContext)
  if (context === undefined) {
    throw new Error('useMyProducts must be used within a MyProductsProvider')
  }
  return context
}
