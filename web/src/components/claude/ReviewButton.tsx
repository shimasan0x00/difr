interface ReviewButtonProps {
  onClick: () => void
  loading: boolean
  disabled?: boolean
}

export function ReviewButton({ onClick, loading, disabled }: ReviewButtonProps) {
  return (
    <button
      type="button"
      onClick={onClick}
      disabled={loading || disabled}
      aria-busy={loading}
      className="px-4 py-2 text-sm bg-purple-700 text-white rounded hover:bg-purple-600 disabled:opacity-50 disabled:cursor-not-allowed"
    >
      {loading ? 'Reviewing...' : 'Auto Review'}
    </button>
  )
}
