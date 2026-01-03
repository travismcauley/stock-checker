interface SkeletonProps {
  className?: string
}

export function Skeleton({ className = '' }: SkeletonProps) {
  return (
    <div
      className={`animate-pulse bg-gray-200 rounded ${className}`}
    />
  )
}

export function StoreCardSkeleton() {
  return (
    <div className="flex items-center justify-between border border-gray-200 rounded-lg p-4">
      <div className="flex-1 space-y-2">
        <Skeleton className="h-5 w-48" />
        <Skeleton className="h-4 w-64" />
        <Skeleton className="h-4 w-40" />
      </div>
      <Skeleton className="h-9 w-32 ml-4" />
    </div>
  )
}

export function ProductCardSkeleton() {
  return (
    <div className="flex border border-gray-200 rounded-lg p-4">
      <Skeleton className="w-20 h-20 rounded-lg flex-shrink-0" />
      <div className="flex-1 ml-4 space-y-2">
        <Skeleton className="h-5 w-full" />
        <Skeleton className="h-4 w-24" />
        <Skeleton className="h-5 w-16" />
        <div className="flex gap-2 mt-2">
          <Skeleton className="h-7 w-32" />
          <Skeleton className="h-7 w-28" />
        </div>
      </div>
    </div>
  )
}

export function TableRowSkeleton() {
  return (
    <tr>
      <td className="px-6 py-4"><Skeleton className="h-6 w-20" /></td>
      <td className="px-6 py-4">
        <Skeleton className="h-5 w-full mb-1" />
        <Skeleton className="h-4 w-24" />
      </td>
      <td className="px-6 py-4"><Skeleton className="h-5 w-16" /></td>
      <td className="px-6 py-4">
        <Skeleton className="h-5 w-32 mb-1" />
        <Skeleton className="h-4 w-24" />
      </td>
      <td className="px-6 py-4"><Skeleton className="h-5 w-16" /></td>
    </tr>
  )
}
