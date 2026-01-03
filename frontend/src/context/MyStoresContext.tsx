import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import type { ReactNode } from 'react'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { StockCheckerService } from '../gen/stockchecker/v1/service_pb.js'
import type { Store } from '../gen/stockchecker/v1/service_pb.js'
import { useAuth } from './AuthContext'

const STORAGE_KEY = 'pokemon-stock-checker-my-stores'

interface MyStoresContextType {
  stores: Store[]
  addStore: (store: Store) => void
  removeStore: (storeId: string) => void
  isStoreInList: (storeId: string) => boolean
  clearStores: () => void
  isLoading: boolean
}

const MyStoresContext = createContext<MyStoresContextType | undefined>(undefined)

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

// Helper to serialize Store to a plain object for localStorage
function serializeStore(store: Store): Record<string, unknown> {
  return {
    storeId: store.storeId,
    name: store.name,
    address: store.address,
    city: store.city,
    state: store.state,
    postalCode: store.postalCode,
    phone: store.phone,
    distanceMiles: store.distanceMiles,
  }
}

// Helper to deserialize from localStorage back to Store-like object
function deserializeStore(data: Record<string, unknown>): Store {
  return {
    storeId: data.storeId as string,
    name: data.name as string,
    address: data.address as string,
    city: data.city as string,
    state: data.state as string,
    postalCode: data.postalCode as string,
    phone: data.phone as string,
    distanceMiles: data.distanceMiles as number,
    $typeName: 'stockchecker.v1.Store',
  } as Store
}

export function MyStoresProvider({ children }: { children: ReactNode }) {
  const { isAuthenticated, isLoading: authLoading } = useAuth()
  const [stores, setStores] = useState<Store[]>([])
  const [isLoading, setIsLoading] = useState(true)

  // Load stores from API or localStorage
  const loadStores = useCallback(async () => {
    if (authLoading) return

    setIsLoading(true)
    try {
      if (isAuthenticated) {
        // Load from API
        const response = await client.getMyStores({})
        setStores(response.stores)
      } else {
        // Load from localStorage
        const saved = localStorage.getItem(STORAGE_KEY)
        if (saved) {
          const parsed = JSON.parse(saved) as Record<string, unknown>[]
          setStores(parsed.map(deserializeStore))
        } else {
          setStores([])
        }
      }
    } catch (e) {
      console.error('Failed to load stores:', e)
      // Fall back to localStorage if API fails
      try {
        const saved = localStorage.getItem(STORAGE_KEY)
        if (saved) {
          const parsed = JSON.parse(saved) as Record<string, unknown>[]
          setStores(parsed.map(deserializeStore))
        }
      } catch {
        setStores([])
      }
    } finally {
      setIsLoading(false)
    }
  }, [isAuthenticated, authLoading])

  useEffect(() => {
    loadStores()
  }, [loadStores])

  // Save to localStorage when not authenticated
  useEffect(() => {
    if (authLoading || isAuthenticated) return
    try {
      const serialized = stores.map(serializeStore)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(serialized))
    } catch (e) {
      console.error('Failed to save stores to localStorage:', e)
    }
  }, [stores, isAuthenticated, authLoading])

  const addStore = async (store: Store) => {
    // Optimistically update local state
    setStores((prev) => {
      if (prev.some((s) => s.storeId === store.storeId)) {
        return prev
      }
      return [...prev, store]
    })

    if (isAuthenticated) {
      try {
        await client.addMyStore({ store })
      } catch (e) {
        console.error('Failed to add store to API:', e)
        // Revert on failure
        setStores((prev) => prev.filter((s) => s.storeId !== store.storeId))
      }
    }
  }

  const removeStore = async (storeId: string) => {
    const removedStore = stores.find((s) => s.storeId === storeId)

    // Optimistically update local state
    setStores((prev) => prev.filter((s) => s.storeId !== storeId))

    if (isAuthenticated) {
      try {
        await client.removeMyStore({ storeId })
      } catch (e) {
        console.error('Failed to remove store from API:', e)
        // Revert on failure
        if (removedStore) {
          setStores((prev) => [...prev, removedStore])
        }
      }
    }
  }

  const isStoreInList = (storeId: string) => {
    return stores.some((s) => s.storeId === storeId)
  }

  const clearStores = () => {
    setStores([])
    if (!isAuthenticated) {
      localStorage.removeItem(STORAGE_KEY)
    }
  }

  return (
    <MyStoresContext.Provider
      value={{ stores, addStore, removeStore, isStoreInList, clearStores, isLoading }}
    >
      {children}
    </MyStoresContext.Provider>
  )
}

export function useMyStores() {
  const context = useContext(MyStoresContext)
  if (context === undefined) {
    throw new Error('useMyStores must be used within a MyStoresProvider')
  }
  return context
}
