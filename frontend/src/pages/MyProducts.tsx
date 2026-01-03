import { Link } from 'react-router-dom'
import { useMyProducts } from '../context/MyProductsContext'

export function MyProducts() {
  const { products, removeProduct, clearProducts } = useMyProducts()

  const formatPrice = (price: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
    }).format(price)
  }

  if (products.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow-md p-8 text-center">
        <div className="text-gray-400 text-5xl mb-4">🎴</div>
        <h2 className="text-xl font-semibold text-gray-700 mb-2">No products saved yet</h2>
        <p className="text-gray-500 mb-6">
          Search for Pokemon card products and add them to your list to track inventory.
        </p>
        <Link
          to="/products/search"
          className="inline-block px-6 py-2 bg-yellow-500 text-gray-900 font-medium rounded-lg hover:bg-yellow-400 transition-colors"
        >
          Search Products
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-xl font-semibold">
            My Products ({products.length})
          </h2>
          <div className="flex gap-2">
            <Link
              to="/products/search"
              className="px-4 py-2 bg-yellow-500 text-gray-900 font-medium rounded-lg hover:bg-yellow-400 text-sm transition-colors"
            >
              + Add More
            </Link>
            {products.length > 0 && (
              <button
                onClick={() => {
                  if (confirm('Are you sure you want to remove all products?')) {
                    clearProducts()
                  }
                }}
                className="px-4 py-2 bg-red-100 text-red-700 rounded-lg hover:bg-red-200 text-sm transition-colors"
              >
                Clear All
              </button>
            )}
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2">
          {products.map((product) => (
            <div
              key={product.sku}
              className="flex border border-gray-200 rounded-lg p-4 hover:bg-gray-50 transition-colors"
            >
              {product.thumbnailUrl && (
                <img
                  src={product.thumbnailUrl}
                  alt={product.name}
                  className="w-20 h-20 object-contain rounded-lg bg-gray-100 flex-shrink-0"
                  onError={(e) => {
                    (e.target as HTMLImageElement).style.display = 'none'
                  }}
                />
              )}
              <div className="flex-1 ml-4 min-w-0">
                <div className="font-semibold text-gray-900 line-clamp-2">
                  {product.name}
                </div>
                <div className="text-gray-500 text-sm mt-1">
                  SKU: {product.sku}
                </div>
                <div className="text-lg font-bold text-green-600 mt-1">
                  {formatPrice(product.salePrice)}
                </div>
                <div className="mt-2 flex gap-2">
                  <button
                    onClick={() => removeProduct(product.sku)}
                    className="px-3 py-1 rounded text-sm font-medium bg-red-100 text-red-700 hover:bg-red-200 transition-colors"
                  >
                    Remove
                  </button>
                  {product.productUrl && (
                    <a
                      href={product.productUrl}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="px-3 py-1 rounded text-sm font-medium bg-blue-100 text-blue-700 hover:bg-blue-200 transition-colors"
                    >
                      View on Best Buy
                    </a>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}
