import { useState } from 'react'
import { stockCheckerClient } from '../lib/api'
import { useMyStores } from '../context/MyStoresContext'
import { useToast } from '../components/Toast'
import { StoreCardSkeleton } from '../components/Skeleton'
import type { Store } from '../gen/stockchecker/v1/service_pb.js'

export function StoreSearch() {
  const [stores, setStores] = useState<Store[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [postalCode, setPostalCode] = useState('')
  const [hasSearched, setHasSearched] = useState(false)

  const { addStore, isStoreInList } = useMyStores()
  const { showToast } = useToast()

  const handleSearch = async () => {
    if (!postalCode) return

    setLoading(true)
    setError(null)
    setHasSearched(true)

    try {
      const response = await stockCheckerClient.searchStores({
        postalCode: postalCode,
        radiusMiles: 25,
      })
      setStores(response.stores)
      if (response.stores.length === 0) {
        showToast('No stores found in that area', 'info')
      }
    } catch (err) {
      console.error('Store search error:', err)
      const message = err instanceof Error ? err.message : 'Failed to search stores'
      setError(message)
      setStores([])
    } finally {
      setLoading(false)
    }
  }

  const handleAddStore = (store: Store) => {
    addStore(store)
    showToast(`Added ${store.name}`, 'success')
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-4 sm:p-6">
        <h2 className="text-lg sm:text-xl font-semibold mb-4">Search Best Buy Stores</h2>
        <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
          <input
            type="text"
            placeholder="Enter postal code (e.g., 94102)"
            value={postalCode}
            onChange={(e) => setPostalCode(e.target.value)}
            onKeyPress={handleKeyPress}
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 text-base"
          />
          <button
            onClick={handleSearch}
            disabled={loading || !postalCode}
            className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium whitespace-nowrap"
          >
            {loading ? (
              <span className="flex items-center justify-center gap-2">
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Searching...
              </span>
            ) : (
              'Search'
            )}
          </button>
        </div>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
          <span className="text-red-500 text-xl">⚠</span>
          <div>
            <p className="text-red-800 font-medium">Search failed</p>
            <p className="text-red-600 text-sm mt-1">{error}</p>
          </div>
        </div>
      )}

      {/* Loading skeletons */}
      {loading && (
        <div className="bg-white rounded-lg shadow-md p-4 sm:p-6">
          <div className="h-6 w-32 bg-gray-200 rounded animate-pulse mb-4" />
          <div className="space-y-3">
            {[1, 2, 3].map((i) => (
              <StoreCardSkeleton key={i} />
            ))}
          </div>
        </div>
      )}

      {/* Results */}
      {!loading && stores.length > 0 && (
        <div className="bg-white rounded-lg shadow-md p-4 sm:p-6">
          <h2 className="text-lg sm:text-xl font-semibold mb-4">
            Found {stores.length} store{stores.length !== 1 ? 's' : ''}
          </h2>
          <div className="space-y-3">
            {stores.map((store) => {
              const inList = isStoreInList(store.storeId)
              return (
                <div
                  key={store.storeId}
                  className="flex flex-col sm:flex-row sm:items-center justify-between border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors gap-3"
                >
                  <div className="flex-1 min-w-0">
                    <div className="font-semibold text-gray-900">{store.name}</div>
                    <div className="text-gray-600 text-sm truncate">
                      {store.address}, {store.city}, {store.state} {store.postalCode}
                    </div>
                    <div className="text-gray-500 text-sm">
                      {store.phone} • {store.distanceMiles.toFixed(1)} miles away
                    </div>
                  </div>
                  <button
                    onClick={() => handleAddStore(store)}
                    disabled={inList}
                    className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors whitespace-nowrap ${
                      inList
                        ? 'bg-green-100 text-green-700 cursor-default'
                        : 'bg-blue-600 text-white hover:bg-blue-700'
                    }`}
                  >
                    {inList ? '✓ Added' : 'Add to My Stores'}
                  </button>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Empty state */}
      {hasSearched && !loading && stores.length === 0 && !error && (
        <div className="bg-white rounded-lg shadow-md p-8 text-center">
          <div className="text-gray-300 text-6xl mb-4">🏪</div>
          <h3 className="text-lg font-semibold text-gray-700 mb-2">No stores found</h3>
          <p className="text-gray-500 text-sm">
            We couldn't find any Best Buy stores near that location.<br />
            Try a different postal code or increase your search area.
          </p>
        </div>
      )}

      {/* Initial state */}
      {!hasSearched && !loading && (
        <div className="bg-white rounded-lg shadow-md p-8 text-center">
          <div className="text-gray-300 text-6xl mb-4">📍</div>
          <h3 className="text-lg font-semibold text-gray-700 mb-2">Find stores near you</h3>
          <p className="text-gray-500 text-sm">
            Enter a postal code above to search for Best Buy stores in your area.
          </p>
        </div>
      )}
    </div>
  )
}
