import { useMyStores } from '../context/MyStoresContext'
import { Link } from 'react-router-dom'

export function MyStores() {
  const { stores, removeStore, clearStores } = useMyStores()

  if (stores.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow-md p-8 text-center">
        <div className="text-gray-400 text-5xl mb-4">🏪</div>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">No stores saved yet</h2>
        <p className="text-gray-500 mb-6">
          Search for Best Buy stores and add them to your list to track inventory.
        </p>
        <Link
          to="/stores/search"
          className="inline-block px-6 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
        >
          Search Stores
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">
            My Stores ({stores.length})
          </h2>
          <div className="flex gap-2">
            <Link
              to="/stores/search"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 text-sm transition-colors"
            >
              + Add More
            </Link>
            {stores.length > 0 && (
              <button
                onClick={() => {
                  if (confirm('Are you sure you want to remove all stores?')) {
                    clearStores()
                  }
                }}
                className="px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 text-sm transition-colors"
              >
                Clear All
              </button>
            )}
          </div>
        </div>

        <div className="space-y-3">
          {stores.map((store) => (
            <div
              key={store.storeId}
              className="flex items-center justify-between border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
            >
              <div className="flex-1">
                <div className="font-semibold text-gray-900">{store.name}</div>
                <div className="text-gray-600 text-sm">
                  {store.address}, {store.city}, {store.state} {store.postalCode}
                </div>
                <div className="text-gray-500 text-sm">{store.phone}</div>
              </div>
              <button
                onClick={() => removeStore(store.storeId)}
                className="ml-4 px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 text-sm font-medium transition-colors"
              >
                Remove
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
