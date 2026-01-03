import { createContext, useContext, useState, useEffect } from 'react'
import type { ReactNode } from 'react'
import type { Store } from '../gen/stockchecker/v1/service_pb.js'

const STORAGE_KEY = 'pokemon-stock-checker-my-stores'

interface MyStoresContextType {
  stores: Store[]
  addStore: (store: Store) => void
  removeStore: (storeId: string) => void
  isStoreInList: (storeId: string) => boolean
  clearStores: () => void
}

const MyStoresContext = createContext<MyStoresContextType | undefined>(undefined)

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
  const [stores, setStores] = useState<Store[]>(() => {
    // Load from localStorage on initial render
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      if (saved) {
        const parsed = JSON.parse(saved) as Record<string, unknown>[]
        return parsed.map(deserializeStore)
      }
    } catch (e) {
      console.error('Failed to load stores from localStorage:', e)
    }
    return []
  })

  // Save to localStorage whenever stores change
  useEffect(() => {
    try {
      const serialized = stores.map(serializeStore)
      localStorage.setItem(STORAGE_KEY, JSON.stringify(serialized))
    } catch (e) {
      console.error('Failed to save stores to localStorage:', e)
    }
  }, [stores])

  const addStore = (store: Store) => {
    setStores((prev) => {
      // Don't add if already in list
      if (prev.some((s) => s.storeId === store.storeId)) {
        return prev
      }
      return [...prev, store]
    })
  }

  const removeStore = (storeId: string) => {
    setStores((prev) => prev.filter((s) => s.storeId !== storeId))
  }

  const isStoreInList = (storeId: string) => {
    return stores.some((s) => s.storeId === storeId)
  }

  const clearStores = () => {
    setStores([])
  }

  return (
    <MyStoresContext.Provider
      value={{ stores, addStore, removeStore, isStoreInList, clearStores }}
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
