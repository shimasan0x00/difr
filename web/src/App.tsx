import { useEffect, useMemo } from 'react'
import { fetchDiff } from './api/client'
import { useDiffStore } from './stores/diffStore'
import { useClaudeStore } from './stores/claudeStore'
import { useCommentStore } from './stores/commentStore'
import { DiffViewer } from './components/diff/DiffViewer'
import { FileListPanel } from './components/diff/FileListPanel'
import { ExportButton } from './components/comment/ExportButton'
import { ChatPanel } from './components/claude/ChatPanel'
import { ReviewButton } from './components/claude/ReviewButton'

function App() {
  const files = useDiffStore((s) => s.files)
  const stats = useDiffStore((s) => s.stats)
  const viewMode = useDiffStore((s) => s.viewMode)
  const setViewMode = useDiffStore((s) => s.setViewMode)
  const loading = useDiffStore((s) => s.loading)
  const error = useDiffStore((s) => s.error)
  const selectedFile = useDiffStore((s) => s.selectedFile)
  const selectFile = useDiffStore((s) => s.selectFile)
  const setFiles = useDiffStore((s) => s.setFiles)
  const setError = useDiffStore((s) => s.setError)

  const claudeMessages = useClaudeStore((s) => s.messages)
  const claudeLoading = useClaudeStore((s) => s.loading)
  const claudeConnected = useClaudeStore((s) => s.connected)
  const connect = useClaudeStore((s) => s.connect)
  const sendChat = useClaudeStore((s) => s.sendChat)
  const sendReview = useClaudeStore((s) => s.sendReview)

  const comments = useCommentStore((s) => s.comments)
  const commentError = useCommentStore((s) => s.error)
  const addComment = useCommentStore((s) => s.addComment)
  const updateComment = useCommentStore((s) => s.updateComment)
  const removeComment = useCommentStore((s) => s.removeComment)
  const loadComments = useCommentStore((s) => s.loadComments)

  useEffect(() => {
    const controller = new AbortController()
    fetchDiff(controller.signal)
      .then((res) => setFiles(res.files, res.stats))
      .catch((err) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError(err instanceof Error ? err.message : String(err))
      })
    return () => controller.abort()
  }, [setFiles, setError])

  useEffect(() => {
    loadComments()
  }, [loadComments])

  useEffect(() => {
    connect()
    return () => useClaudeStore.getState().disconnect()
  }, [connect])

  const commentsByFile = useMemo(() => {
    const result = new Map<string, typeof comments>()
    for (const c of comments) {
      const existing = result.get(c.filePath) ?? []
      result.set(c.filePath, [...existing, c])
    }
    return result
  }, [comments])

  const handleSelectFile = (path: string) => {
    selectFile(path)
    document.getElementById(`diff-file-${path}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  const handleReview = () => {
    const diffSummary = files
      .map((f) => `${f.newPath || f.oldPath} (${f.status})`)
      .join('\n')
    sendReview(`Review the following diff files:\n${diffSummary}`)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <p className="text-gray-400">Loading diff...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <p className="text-red-400">Error: {error}</p>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-[#0d1117]">
      <header className="border-b border-gray-700 px-4 py-3 flex items-center justify-between">
        <h1 className="text-xl font-semibold text-white">difr</h1>
        <div className="flex items-center gap-3">
          {comments.length > 0 && <ExportButton />}
          <div className="flex bg-[#161b22] rounded border border-gray-700">
            <button
              type="button"
              onClick={() => setViewMode('split')}
              className={`px-3 py-1 text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500 ${viewMode === 'split' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-gray-200'}`}
            >
              Split
            </button>
            <button
              type="button"
              onClick={() => setViewMode('unified')}
              className={`px-3 py-1 text-xs focus-visible:outline focus-visible:outline-2 focus-visible:outline-blue-500 ${viewMode === 'unified' ? 'bg-gray-700 text-white' : 'text-gray-400 hover:text-gray-200'}`}
            >
              Unified
            </button>
          </div>
          {claudeConnected && (
            <ReviewButton onClick={handleReview} loading={claudeLoading} />
          )}
        </div>
      </header>
      {files.length > 0 && (
        <div className="px-4 py-2 border-b border-gray-700 text-sm text-gray-400">
          <span>{files.length} file{files.length !== 1 ? 's' : ''} changed</span>
          {stats.additions > 0 && <span className="text-green-400 ml-2">+{stats.additions}</span>}
          {stats.deletions > 0 && <span className="text-red-400 ml-2">-{stats.deletions}</span>}
        </div>
      )}
      <div className="flex">
        <main className="flex-1 p-4 space-y-4">
          {commentError && (
            <div className="bg-red-900/30 border border-red-700 rounded-md px-4 py-2 text-red-300 text-sm">
              Comment error: {commentError}
            </div>
          )}
          {files.length === 0 ? (
            <p className="text-gray-400">No changes found.</p>
          ) : (
            <>
              <FileListPanel
                files={files}
                commentsByFile={commentsByFile}
                selectedFile={selectedFile}
                onSelectFile={handleSelectFile}
              />
              {files.map((file) => {
                const filePath = file.newPath || file.oldPath
                return (
                  <div key={filePath} id={`diff-file-${filePath}`}>
                    <DiffViewer
                      file={file}
                      viewMode={viewMode}
                      comments={commentsByFile.get(filePath) ?? []}
                      onAddComment={(line, body) => addComment(filePath, line, body)}
                      onDeleteComment={removeComment}
                      onUpdateComment={updateComment}
                    />
                  </div>
                )
              })}
            </>
          )}
        </main>
        {claudeConnected && (
          <aside className="w-80 border-l border-gray-700 p-4 h-[calc(100vh-57px)] sticky top-[57px]">
            <ChatPanel
              onSend={sendChat}
              messages={claudeMessages}
              loading={claudeLoading}
            />
          </aside>
        )}
      </div>
    </div>
  )
}

export default App
