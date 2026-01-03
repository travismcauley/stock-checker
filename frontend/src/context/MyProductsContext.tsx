import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import type { Product } from '../gen/stockchecker/v1/service_pb.js'

const STORAGE_KEY = 'pokemon-stock-checker-my-products'

interface MyProductsContextType {
  products: Product[]
  addProduct: (product: Product) => void
  removeProduct: (sku: string) => void
  isProductInList: (sku: string) => boolean
  clearProducts: () => void
}

const MyProductsContext = createContext<MyProductsContextType | undefined>(undefined)

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
  const [products, setProducts] = useState<Product[]>(() => {
    // Load from localStorage on initial render
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      if (saved) {
        const parsed = JSON.parse(saved) as Record<string, unknown>[]
        return parsed.map(deserializeProduct)
      }
    } catch (e) {
      console.error('Failed to load products from localStorage:', e)
    }
    return []
  })

  // Save to localStorage whenever products change
  useEffect(() => {
    try {
      const serialized = products.map(serializeProduct)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(serialized))
    } catch (e) {
      console.error('Failed to save products to localStorage:', e)
    }
  }, [products])

  const addProduct = (product: Product) => {
    setProducts((prev) => {
      // Don't add if already in list
      if (prev.some((p) => p.sku === product.sku)) {
        return prev
      }
      return [...prev, product]
    })
  }

  const removeProduct = (sku: string) => {
    setProducts((prev) => prev.filter((p) => p.sku !== sku))
  }

  const isProductInList = (sku: string) => {
    return products.some((p) => p.sku === sku)
  }

  const clearProducts = () => {
    setProducts([])
  }

  return (
    <MyProductsContext.Provider
      value={{ products, addProduct, removeProduct, isProductInList, clearProducts }}
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
