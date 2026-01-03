import { useState } from 'react'
import { stockCheckerClient } from './lib/api'
import type { Store } from './gen/stockchecker/v1/service_pb.js'

function App() {
  const [stores, setStores] = useState<Store[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [postalCode, setPostalCode] = useState('')

  const handleSearch = async () => {
    if (!postalCode) return

    setLoading(true)
    setError(null)

    try {
      const response = await stockCheckerClient.searchStores({
        postalCode: postalCode,
        radiusMiles: 25,
      })
      setStores(response.stores)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to search stores')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-100">
      <header className="bg-blue-600 text-white py-4 px-6 shadow-lg">
        <h1 className="text-2xl font-bold">Pokemon Stock Checker</h1>
      </header>

      <main className="container mx-auto px-4 py-8">
        <div className="bg-white rounded-lg shadow-md p-6 mb-6">
          <h2 className="text-xl font-semibold mb-4">Search Stores</h2>
          <div className="flex gap-4">
            <input
              type="text"
              placeholder="Enter postal code"
              value={postalCode}
              onChange={(e) => setPostalCode(e.target.value)}
              className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
            <button
              onClick={handleSearch}
              disabled={loading || !postalCode}
              className="px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {loading ? 'Searching...' : 'Search'}
            </button>
          </div>
        </div>

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded-lg mb-6">
            {error}
          </div>
        )}

        {stores.length > 0 && (
          <div className="bg-white rounded-lg shadow-md p-6">
            <h2 className="text-xl font-semibold mb-4">Results</h2>
            <div className="space-y-4">
              {stores.map((store) => (
                <div
                  key={store.storeId}
                  className="border border-gray-200 rounded-lg p-4 hover:bg-gray-50"
                >
                  <div className="font-semibold">{store.name}</div>
                  <div className="text-gray-600 text-sm">
                    {store.address}, {store.city}, {store.state} {store.postalCode}
                  </div>
                  <div className="text-gray-500 text-sm">
                    {store.phone} • {store.distanceMiles.toFixed(1)} miles away
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {stores.length === 0 && !loading && !error && (
          <div className="bg-white rounded-lg shadow-md p-6 text-center text-gray-500">
            Enter a postal code to search for nearby Best Buy stores
          </div>
        )}
      </main>
    </div>
  )
}

export default App
