import { useState, useMemo } from 'react'
import { Link } from 'react-router-dom'
import { stockCheckerClient } from '../lib/api'
import { useMyStores } from '../context/MyStoresContext'
import { useMyProducts } from '../context/MyProductsContext'
import { useToast } from '../components/Toast'
import { TableRowSkeleton } from '../components/Skeleton'
import type { StockStatus } from '../gen/stockchecker/v1/service_pb.js'

type SortField = 'status' | 'product' | 'store' | 'price'
type SortDirection = 'asc' | 'desc'

export function CheckStock() {
  const { stores } = useMyStores()
  const { products } = useMyProducts()
  const { showToast } = useToast()

  const [results, setResults] = useState<StockStatus[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [hasChecked, setHasChecked] = useState(false)
  const [sortField, setSortField] = useState<SortField>('status')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  const canCheck = stores.length > 0 && products.length > 0

  const handleCheckStock = async () => {
    if (!canCheck) return

    setLoading(true)
    setError(null)
    setResults([])
    setHasChecked(true)

    try {
      const storeIds = stores.map((s) => s.storeId)
      const skus = products.map((p) => p.sku)

      const response = await stockCheckerClient.checkStock({
        storeIds,
        skus,
      })

      setResults(response.results)

      // Show summary toast
      const inStock = response.results.filter((r) => r.inStock).length
      if (inStock > 0) {
        showToast(`Found ${inStock} in-stock item${inStock !== 1 ? 's' : ''}!`, 'success')
      } else {
        showToast('No items currently in stock', 'info')
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Failed to check stock'
      setError(getFriendlyError(message))
      showToast('Failed to check stock', 'error')
    } finally {
      setLoading(false)
    }
  }

  const getFriendlyError = (message: string): string => {
    if (message.includes('fetch') || message.includes('network')) {
      return 'Unable to connect to the server. Please check your internet connection.'
    }
    if (message.includes('timeout')) {
      return 'The request took too long. Please try again.'
    }
    if (message.includes('rate limit')) {
      return 'Too many requests. Please wait a moment and try again.'
    }
    return 'Something went wrong. Please try again.'
  }

  // Sort results
  const sortedResults = useMemo(() => {
    const sorted = [...results]

    sorted.sort((a, b) => {
      let comparison = 0

      switch (sortField) {
        case 'status':
          const statusA = a.inStock ? (a.lowStock ? 1 : 2) : 0
          const statusB = b.inStock ? (b.lowStock ? 1 : 2) : 0
          comparison = statusB - statusA
          break
        case 'product':
          comparison = (a.product?.name || '').localeCompare(b.product?.name || '')
          break
        case 'store':
          comparison = (a.store?.name || '').localeCompare(b.store?.name || '')
          break
        case 'price':
          comparison = (a.product?.salePrice || 0) - (b.product?.salePrice || 0)
          break
      }

      return sortDirection === 'asc' ? comparison : -comparison
    })

    return sorted
  }, [results, sortField, sortDirection])

  const handleSort = (field: SortField) => {
    if (field === sortField) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
    } else {
      setSortField(field)
      setSortDirection(field === 'status' ? 'desc' : 'asc')
    }
  }

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price)
  }

  const getStatusBadge = (status: StockStatus) => {
    if (!status.inStock) {
      return (
        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-red-100 text-red-800">
          Out of Stock
        </span>
      )
    }
    if (status.lowStock) {
      return (
        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800">
          Low Stock
        </span>
      )
    }
    return (
      <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
        In Stock
      </span>
    )
  }

  const getSortIcon = (field: SortField) => {
    if (field !== sortField) return '↕'
    return sortDirection === 'asc' ? '↑' : '↓'
  }

  // Summary stats
  const stats = useMemo(() => {
    const inStock = results.filter((r) => r.inStock && !r.lowStock).length
    const lowStock = results.filter((r) => r.lowStock).length
    const outOfStock = results.filter((r) => !r.inStock).length
    return { inStock, lowStock, outOfStock, total: results.length }
  }, [results])

  if (!canCheck) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6 sm:p-8 text-center">
        <div className="text-gray-300 text-5xl sm:text-6xl mb-4">📦</div>
        <h2 className="text-lg sm:text-xl font-semibold text-gray-700 mb-2">Not ready to check stock</h2>
        <p className="text-gray-500 text-sm sm:text-base mb-6">
          You need to add both stores and products before checking stock.
        </p>
        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          {stores.length === 0 && (
            <Link
              to="/stores/search"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Add Stores
            </Link>
          )}
          {products.length === 0 && (
            <Link
              to="/products/search"
              className="px-4 py-2 bg-yellow-500 text-gray-900 rounded-lg hover:bg-yellow-400 transition-colors font-medium"
            >
              Add Products
            </Link>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4 sm:space-y-6">
      {/* Header */}
      <div className="bg-white rounded-lg shadow-md p-4 sm:p-6">
        <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
          <div>
            <h2 className="text-lg sm:text-xl font-semibold">Check Stock</h2>
            <p className="text-gray-500 text-sm mt-1">
              {products.length} product{products.length !== 1 ? 's' : ''} × {stores.length} store{stores.length !== 1 ? 's' : ''}
            </p>
          </div>
          <button
            onClick={handleCheckStock}
            disabled={loading}
            className="px-6 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-semibold flex items-center justify-center gap-2"
          >
            {loading ? (
              <>
                <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                  <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                  <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                </svg>
                Checking...
              </>
            ) : hasChecked ? (
              'Re-check Stock'
            ) : (
              'Check Stock Now'
            )}
          </button>
        </div>
      </div>

      {/* Error */}
      {error && (
        <div className="bg-red-50 border border-red-200 rounded-lg p-4 flex items-start gap-3">
          <span className="text-red-500 text-xl">⚠</span>
          <div>
            <p className="text-red-800 font-medium">Check failed</p>
            <p className="text-red-600 text-sm mt-1">{error}</p>
          </div>
        </div>
      )}

      {/* Loading skeleton */}
      {loading && (
        <>
          <div className="grid grid-cols-3 gap-2 sm:gap-4">
            {[1, 2, 3].map((i) => (
              <div key={i} className="bg-gray-100 rounded-lg p-3 sm:p-4 text-center animate-pulse">
                <div className="h-6 sm:h-8 w-8 sm:w-12 bg-gray-200 rounded mx-auto mb-1" />
                <div className="h-3 sm:h-4 w-16 sm:w-20 bg-gray-200 rounded mx-auto" />
              </div>
            ))}
          </div>
          <div className="bg-white rounded-lg shadow-md overflow-hidden">
            <table className="min-w-full divide-y divide-gray-200">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-4 sm:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Status</th>
                  <th className="px-4 sm:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Product</th>
                  <th className="px-4 sm:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase hidden sm:table-cell">Price</th>
                  <th className="px-4 sm:px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">Store</th>
                </tr>
              </thead>
              <tbody className="bg-white divide-y divide-gray-200">
                {[1, 2, 3, 4, 5].map((i) => (
                  <TableRowSkeleton key={i} />
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}

      {/* Results */}
      {hasChecked && !loading && results.length > 0 && (
        <>
          {/* Summary */}
          <div className="grid grid-cols-3 gap-2 sm:gap-4">
            <div className="bg-green-50 rounded-lg p-3 sm:p-4 text-center">
              <div className="text-xl sm:text-2xl font-bold text-green-600">{stats.inStock}</div>
              <div className="text-xs sm:text-sm text-green-800">In Stock</div>
            </div>
            <div className="bg-yellow-50 rounded-lg p-3 sm:p-4 text-center">
              <div className="text-xl sm:text-2xl font-bold text-yellow-600">{stats.lowStock}</div>
              <div className="text-xs sm:text-sm text-yellow-800">Low Stock</div>
            </div>
            <div className="bg-red-50 rounded-lg p-3 sm:p-4 text-center">
              <div className="text-xl sm:text-2xl font-bold text-red-600">{stats.outOfStock}</div>
              <div className="text-xs sm:text-sm text-red-800">Out of Stock</div>
            </div>
          </div>

          {/* Mobile Card View */}
          <div className="sm:hidden space-y-3">
            {sortedResults.map((result, index) => (
              <div
                key={`${result.product?.sku}-${result.store?.storeId}-${index}`}
                className={`bg-white rounded-lg shadow-md p-4 ${!result.inStock ? 'opacity-60' : ''}`}
              >
                <div className="flex justify-between items-start mb-2">
                  {getStatusBadge(result)}
                  <span className="font-semibold text-gray-900">
                    {formatPrice(result.product?.salePrice || 0)}
                  </span>
                </div>
                <div className="text-sm font-medium text-gray-900 line-clamp-2 mb-1">
                  {result.product?.name}
                </div>
                <div className="text-xs text-gray-500 mb-2">SKU: {result.product?.sku}</div>
                <div className="text-sm text-gray-600">
                  {result.store?.name}
                  <span className="text-gray-400"> • {result.store?.city}, {result.store?.state}</span>
                </div>
              </div>
            ))}
          </div>

          {/* Desktop Table View */}
          <div className="hidden sm:block bg-white rounded-lg shadow-md overflow-hidden">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200">
                <thead className="bg-gray-50">
                  <tr>
                    <th
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('status')}
                    >
                      Status {getSortIcon('status')}
                    </th>
                    <th
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('product')}
                    >
                      Product {getSortIcon('product')}
                    </th>
                    <th
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('price')}
                    >
                      Price {getSortIcon('price')}
                    </th>
                    <th
                      className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider cursor-pointer hover:bg-gray-100"
                      onClick={() => handleSort('store')}
                    >
                      Store {getSortIcon('store')}
                    </th>
                    <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                      Pickup
                    </th>
                  </tr>
                </thead>
                <tbody className="bg-white divide-y divide-gray-200">
                  {sortedResults.map((result, index) => (
                    <tr
                      key={`${result.product?.sku}-${result.store?.storeId}-${index}`}
                      className={`hover:bg-gray-50 ${!result.inStock ? 'opacity-60' : ''}`}
                    >
                      <td className="px-6 py-4 whitespace-nowrap">
                        {getStatusBadge(result)}
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm font-medium text-gray-900 line-clamp-2">
                          {result.product?.name}
                        </div>
                        <div className="text-xs text-gray-500">
                          SKU: {result.product?.sku}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="text-sm font-semibold text-gray-900">
                          {formatPrice(result.product?.salePrice || 0)}
                        </div>
                      </td>
                      <td className="px-6 py-4">
                        <div className="text-sm font-medium text-gray-900">
                          {result.store?.name}
                        </div>
                        <div className="text-xs text-gray-500">
                          {result.store?.city}, {result.store?.state}
                        </div>
                      </td>
                      <td className="px-6 py-4 whitespace-nowrap">
                        {result.pickupEligible ? (
                          <span className="text-green-600 text-sm">✓ Available</span>
                        ) : (
                          <span className="text-gray-400 text-sm">—</span>
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </>
      )}

      {/* Empty state after check */}
      {hasChecked && !loading && results.length === 0 && !error && (
        <div className="bg-white rounded-lg shadow-md p-6 sm:p-8 text-center">
          <div className="text-gray-300 text-5xl sm:text-6xl mb-4">😔</div>
          <h2 className="text-lg sm:text-xl font-semibold text-gray-700 mb-2">No results found</h2>
          <p className="text-gray-500 text-sm">
            We couldn't find any stock information for your products at your stores.
          </p>
        </div>
      )}
    </div>
  )
}
