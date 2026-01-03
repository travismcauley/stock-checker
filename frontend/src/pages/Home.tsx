import { Link } from 'react-router-dom'
import { useMyStores } from '../context/MyStoresContext'
import { useMyProducts } from '../context/MyProductsContext'

export function Home() {
  const { stores } = useMyStores()
  const { products } = useMyProducts()

  const canCheckStock = stores.length > 0 && products.length > 0

  return (
    <div className="space-y-6">
      <div className="bg-white rounded-lg shadow-md p-8 text-center">
        <h1 className="text-3xl font-bold text-gray-900 mb-4">
          Pokemon Stock Checker
        </h1>
        <p className="text-gray-600 mb-8 max-w-2xl mx-auto">
          Track Pokemon card inventory at your favorite Best Buy stores.
          Add stores and products to your lists, then check stock availability with one click.
        </p>

        <div className="grid md:grid-cols-2 gap-6 max-w-2xl mx-auto">
          <div className="bg-blue-50 rounded-lg p-6">
            <div className="text-4xl mb-3">🏪</div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">
              {stores.length > 0 ? `${stores.length} Store${stores.length !== 1 ? 's' : ''} Saved` : 'No Stores Yet'}
            </h2>
            <p className="text-gray-600 text-sm mb-4">
              {stores.length > 0
                ? 'Your saved stores are ready for stock checking'
                : 'Start by searching for Best Buy stores near you'
              }
            </p>
            <Link
              to={stores.length > 0 ? '/stores' : '/stores/search'}
              className="inline-block px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors text-sm"
            >
              {stores.length > 0 ? 'View My Stores' : 'Search Stores'}
            </Link>
          </div>

          <div className="bg-yellow-50 rounded-lg p-6">
            <div className="text-4xl mb-3">🎴</div>
            <h2 className="text-xl font-semibold text-gray-900 mb-2">
              {products.length > 0 ? `${products.length} Product${products.length !== 1 ? 's' : ''} Saved` : 'No Products Yet'}
            </h2>
            <p className="text-gray-600 text-sm mb-4">
              {products.length > 0
                ? 'Your saved products are ready for stock checking'
                : 'Search for Pokemon card products to track'
              }
            </p>
            <Link
              to={products.length > 0 ? '/products' : '/products/search'}
              className="inline-block px-4 py-2 bg-yellow-500 text-gray-900 rounded-lg hover:bg-yellow-400 transition-colors text-sm font-medium"
            >
              {products.length > 0 ? 'View My Products' : 'Search Products'}
            </Link>
          </div>
        </div>
      </div>

      {canCheckStock && (
        <div className="bg-green-50 border border-green-200 rounded-lg p-6 text-center">
          <p className="text-green-800 mb-4">
            You have <strong>{stores.length} store{stores.length !== 1 ? 's' : ''}</strong> and{' '}
            <strong>{products.length} product{products.length !== 1 ? 's' : ''}</strong> saved.
            Ready to check stock!
          </p>
          <Link
            to="/check-stock"
            className="inline-block px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700 transition-colors font-semibold"
          >
            Check Stock Now
          </Link>
        </div>
      )}

      {!canCheckStock && (stores.length > 0 || products.length > 0) && (
        <div className="bg-amber-50 border border-amber-200 rounded-lg p-6 text-center">
          <p className="text-amber-800">
            {stores.length === 0
              ? 'Add some stores to check stock availability!'
              : 'Add some products to check stock availability!'}
          </p>
        </div>
      )}
    </div>
  )
}
