import { useState } from 'react'
import { stockCheckerClient } from '../lib/api'
import { useMyProducts } from '../context/MyProductsContext'
import { useToast } from '../components/Toast'
import { ProductCardSkeleton } from '../components/Skeleton'
import type { Product } from '../gen/stockchecker/v1/service_pb.js'

export function ProductSearch() {
  const [products, setProducts] = useState<Product[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [query, setQuery] = useState('')
  const [hasSearched, setHasSearched] = useState(false)

  const { addProduct, isProductInList } = useMyProducts()
  const { showToast } = useToast()

  const handleSearch = async () => {
    if (!query) return

    setLoading(true)
    setError(null)
    setHasSearched(true)

    try {
      const response = await stockCheckerClient.searchProducts({
        query: query,
      })
      setProducts(response.products)
      if (response.products.length === 0) {
        showToast('No products found matching your search', 'info')
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to search products'
      setError(getFriendlyError(message))
      setProducts([])
    } finally {
      setLoading(false)
    }
  }

  const handleAddProduct = (product: Product) => {
    addProduct(product)
    showToast(`Added ${product.name.substring(0, 40)}...`, 'success')
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleSearch()
    }
  }

  const getFriendlyError = (message: string): string => {
    if (message.includes('fetch') || message.includes('network')) {
      return 'Unable to connect to the server. Please check your internet connection.'
    }
    if (message.includes('timeout')) {
      return 'The request took too long. Please try again.'
    }
    return 'Something went wrong. Please try again.'
  }

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price)
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-4 sm:p-6">
        <h2 className="text-lg sm:text-xl font-semibold mb-4">Search Pokemon Products</h2>
        <div className="flex flex-col sm:flex-row gap-3 sm:gap-4">
          <input
            type="text"
            placeholder="Search by name or SKU"
            value={query}
            onChange={(e) => setQuery(e.target.value)}
            onKeyPress={handleKeyPress}
            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-yellow-500 text-base"
          />
          <button
            onClick={handleSearch}
            disabled={loading || !query}
            className="px-6 py-2 bg-yellow-500 text-gray-900 font-medium rounded-lg hover:bg-yellow-400 disabled:opacity-50 disabled:cursor-not-allowed transition-colors whitespace-nowrap"
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
        <p className="text-gray-500 text-sm mt-2">
          Try: pokemon, prismatic evolutions, surging sparks, 151, paldean fates
        </p>
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
          <div className="grid gap-4 sm:grid-cols-2">
            {[1, 2, 3, 4].map((i) => (
              <ProductCardSkeleton key={i} />
            ))}
          </div>
        </div>
      )}

      {/* Results */}
      {!loading && products.length > 0 && (
        <div className="bg-white rounded-lg shadow-md p-4 sm:p-6">
          <h2 className="text-lg sm:text-xl font-semibold mb-4">
            Found {products.length} product{products.length !== 1 ? 's' : ''}
          </h2>
          <div className="grid gap-4 sm:grid-cols-2">
            {products.map((product) => {
              const inList = isProductInList(product.sku)
              return (
                <div
                  key={product.sku}
                  className="flex border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
                >
                  {product.thumbnailUrl && (
                    <img
                      src={product.thumbnailUrl}
                      alt={product.name}
                      className="w-16 h-16 sm:w-20 sm:h-20 object-contain rounded-lg bg-gray-100 flex-shrink-0"
                      onError={(e) => {
                        (e.target as HTMLImageElement).style.display = 'none'
                      }}
                    />
                  )}
                  <div className="flex-1 ml-3 sm:ml-4 min-w-0">
                    <div className="font-semibold text-gray-900 text-sm sm:text-base line-clamp-2">
                      {product.name}
                    </div>
                    <div className="text-gray-500 text-xs sm:text-sm mt-1">
                      SKU: {product.sku}
                    </div>
                    <div className="text-base sm:text-lg font-bold text-green-600 mt-1">
                      {formatPrice(product.salePrice)}
                    </div>
                    <div className="mt-2 flex flex-wrap gap-2">
                      <button
                        onClick={() => handleAddProduct(product)}
                        disabled={inList}
                        className={`px-3 py-1 rounded text-xs sm:text-sm font-medium transition-colors ${
                          inList
                            ? 'bg-green-100 text-green-700 cursor-default'
                            : 'bg-yellow-500 text-gray-900 hover:bg-yellow-400'
                        }`}
                      >
                        {inList ? '✓ Added' : 'Add'}
                      </button>
                      {product.productUrl && (
                        <a
                          href={product.productUrl}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="px-3 py-1 rounded text-xs sm:text-sm font-medium bg-blue-100 text-blue-700 hover:bg-blue-200 transition-colors"
                        >
                          Best Buy
                        </a>
                      )}
                    </div>
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* Empty state after search */}
      {hasSearched && !loading && products.length === 0 && !error && (
        <div className="bg-white rounded-lg shadow-md p-8 text-center">
          <div className="text-gray-300 text-6xl mb-4">🔍</div>
          <h3 className="text-lg font-semibold text-gray-700 mb-2">No products found</h3>
          <p className="text-gray-500 text-sm">
            We couldn't find any products matching "{query}".<br />
            Try a different search term.
          </p>
        </div>
      )}

      {/* Initial state */}
      {!hasSearched && !loading && (
        <div className="bg-white rounded-lg shadow-md p-8 text-center">
          <div className="text-gray-300 text-6xl mb-4">🎴</div>
          <h3 className="text-lg font-semibold text-gray-700 mb-2">Search for Pokemon cards</h3>
          <p className="text-gray-500 text-sm">
            Enter a product name or SKU above to find Pokemon card products.
          </p>
        </div>
      )}
    </div>
  )
}
