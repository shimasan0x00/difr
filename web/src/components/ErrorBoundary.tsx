import { Component } from 'react'
import type { ErrorInfo, ReactNode } from 'react'

interface Props {
  children: ReactNode
}

interface State {
  hasError: boolean
  error: Error | null
}

export class ErrorBoundary extends Component<Props, State> {
  constructor(props: Props) {
    super(props)
    this.state = { hasError: false, error: null }
  }

  static getDerivedStateFromError(error: Error): State {
    return { hasError: true, error }
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    console.error('ErrorBoundary caught:', error, errorInfo)
  }

  render() {
    if (this.state.hasError) {
      return (
        <div className="flex items-center justify-center min-h-screen bg-[#0d1117]">
          <div className="text-center space-y-3">
            <p className="text-red-400 text-lg">Something went wrong</p>
            <p className="text-gray-400 text-sm">{this.state.error?.message}</p>
            <button
              type="button"
              onClick={() => this.setState({ hasError: false, error: null })}
              className="px-4 py-2 text-sm bg-gray-700 text-white rounded hover:bg-gray-600"
            >
              Try again
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
