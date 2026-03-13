import { useEffect, useMemo, useState } from 'react'
import { fetchDiff, fetchFileContent, fetchTrackedFiles, fetchViewMode } from './api/client'
import { useDiffStore } from './stores/diffStore'
import { useClaudeStore } from './stores/claudeStore'
import { useCommentStore } from './stores/commentStore'
import { DiffViewer } from './components/diff/DiffViewer'
import { FileListPanel } from './components/diff/FileListPanel'
import { FileViewer } from './components/diff/FileViewer'
import { ExportButton } from './components/comment/ExportButton'
import { CopyAllButton } from './components/comment/CopyAllButton'
import { ClearButton } from './components/comment/ClearButton'
import { ChatPanel } from './components/claude/ChatPanel'
import { ReviewButton } from './components/claude/ReviewButton'
import { useResizable } from './hooks/useResizable'
import { buildReviewContent } from './utils/buildReviewContent'

function App() {
  const files = useDiffStore((s) => s.files)
  const stats = useDiffStore((s) => s.stats)
  const meta = useDiffStore((s) => s.meta)
  const trackedFiles = useDiffStore((s) => s.trackedFiles)
  const viewMode = useDiffStore((s) => s.viewMode)
  const setViewMode = useDiffStore((s) => s.setViewMode)
  const loading = useDiffStore((s) => s.loading)
  const error = useDiffStore((s) => s.error)
  const selectedFile = useDiffStore((s) => s.selectedFile)
  const selectFile = useDiffStore((s) => s.selectFile)
  const setFiles = useDiffStore((s) => s.setFiles)
  const setError = useDiffStore((s) => s.setError)
  const setTrackedFiles = useDiffStore((s) => s.setTrackedFiles)
  const sidebarTab = useDiffStore((s) => s.sidebarTab)
  const setSidebarTab = useDiffStore((s) => s.setSidebarTab)
  const reviewedFiles = useDiffStore((s) => s.reviewedFiles)
  const toggleReviewed = useDiffStore((s) => s.toggleReviewed)
  const loadReviewedFiles = useDiffStore((s) => s.loadReviewedFiles)
  const clearReviewedFiles = useDiffStore((s) => s.clearReviewedFiles)
  const fileContentCache = useDiffStore((s) => s.fileContentCache)
  const fileContentLoading = useDiffStore((s) => s.fileContentLoading)
  const setFileContent = useDiffStore((s) => s.setFileContent)
  const setFileContentLoading = useDiffStore((s) => s.setFileContentLoading)
  const setFileContentError = useDiffStore((s) => s.setFileContentError)

  const claudeMessages = useClaudeStore((s) => s.messages)
  const claudeLoading = useClaudeStore((s) => s.loading)
  const claudeError = useClaudeStore((s) => s.error)
  const claudeConnected = useClaudeStore((s) => s.connected)
  const claudeSessionId = useClaudeStore((s) => s.sessionId)
  const connect = useClaudeStore((s) => s.connect)
  const sendChat = useClaudeStore((s) => s.sendChat)
  const sendReview = useClaudeStore((s) => s.sendReview)
  const clearMessages = useClaudeStore((s) => s.clearMessages)

  const comments = useCommentStore((s) => s.comments)
  const commentError = useCommentStore((s) => s.error)
  const commentSaving = useCommentStore((s) => s.saving)
  const addComment = useCommentStore((s) => s.addComment)
  const updateComment = useCommentStore((s) => s.updateComment)
  const removeComment = useCommentStore((s) => s.removeComment)
  const loadComments = useCommentStore((s) => s.loadComments)
  const clearAllComments = useCommentStore((s) => s.clearAll)

  useEffect(() => {
    const controller = new AbortController()
    fetchDiff(controller.signal)
      .then((res) => setFiles(res.files, res.stats, res.meta))
      .catch((err) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
        setError(err instanceof Error ? err.message : String(err))
      })
    return () => controller.abort()
  }, [setFiles, setError])

  useEffect(() => {
    const controller = new AbortController()
    fetchTrackedFiles(controller.signal)
      .then(setTrackedFiles)
      .catch((err) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
      })
    return () => controller.abort()
  }, [setTrackedFiles])

  useEffect(() => {
    const controller = new AbortController()
    fetchViewMode(controller.signal)
      .then((mode) => {
        if (mode === 'split' || mode === 'unified') setViewMode(mode)
      })
      .catch((err) => {
        if (err instanceof DOMException && err.name === 'AbortError') return
      })
    return () => controller.abort()
  }, [setViewMode])

  useEffect(() => {
    loadComments()
  }, [loadComments])

  useEffect(() => {
    loadReviewedFiles()
  }, [loadReviewedFiles])

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

  // Determine which files are changed (by path)
  const changedPaths = useMemo(() => {
    const paths = new Set<string>()
    for (const file of files) {
      if (file.newPath && file.newPath !== '/dev/null') paths.add(file.newPath)
      if (file.oldPath && file.oldPath !== '/dev/null') paths.add(file.oldPath)
    }
    return paths
  }, [files])

  // Claude panel should stay visible once initialized (connection attempted)
  // so users can see errors and chat history even when disconnected
  const claudeInitialized = claudeConnected || claudeMessages.length > 0 || !!claudeError

  const [fileListExpanded, setFileListExpanded] = useState(true)
  const [chatExpanded, setChatExpanded] = useState(true)
  const { width: chatWidth, onMouseDown: onChatResizeMouseDown, isResizing: isChatResizing } = useResizable({
    direction: 'left',
    minWidth: 280,
    maxWidth: 640,
    defaultWidth: 320,
    storageKey: 'difr:chatPanelWidth',
  })

  // Whether the selected file is an unchanged file being viewed via FileViewer
  const isViewingUnchangedFile = selectedFile !== null && !changedPaths.has(selectedFile)
  const viewedFileContent = isViewingUnchangedFile ? fileContentCache.get(selectedFile) : null

  const handleSelectFile = (path: string) => {
    selectFile(path)

    // If it's a changed file, scroll to it in the diff list
    if (changedPaths.has(path)) {
      document.getElementById(`diff-file-${path}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
      return
    }

    // For unchanged files, fetch content if not cached
    if (!fileContentCache.has(path)) {
      setFileContentLoading(true)
      fetchFileContent(path)
        .then((fc) => setFileContent(fc))
        .catch((err) => setFileContentError(err instanceof Error ? err.message : String(err)))
    }
  }

  const handleReview = () => {
    sendReview(buildReviewContent(files))
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
      <div className="sticky top-0 z-10 bg-[#0d1117]">
        <header className="border-b border-gray-700 px-4 py-3 flex items-center justify-between">
          <h1 className="text-xl font-semibold text-white flex items-center gap-2">
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 32 32" width="20" height="20">
              <rect width="32" height="32" rx="6" fill="#161b22"/>
              <line x1="16" y1="5" x2="16" y2="27" stroke="#30363d" strokeWidth="1"/>
              <rect x="4" y="9" width="8" height="2" rx="1" fill="#da3633"/>
              <rect x="4" y="14" width="6" height="2" rx="1" fill="#da3633"/>
              <rect x="4" y="19" width="9" height="2" rx="1" fill="#da3633"/>
              <rect x="20" y="9" width="7" height="2" rx="1" fill="#26a641"/>
              <rect x="20" y="14" width="9" height="2" rx="1" fill="#26a641"/>
              <rect x="20" y="19" width="5" height="2" rx="1" fill="#26a641"/>
            </svg>
            difr
          </h1>
          <div className="flex items-center gap-3">
            {comments.length > 0 && <ExportButton />}
            {comments.length > 0 && <CopyAllButton comments={comments} />}
            {(comments.length > 0 || reviewedFiles.size > 0) && (
              <ClearButton
                onClearComments={clearAllComments}
                onClearReviewedFiles={clearReviewedFiles}
                hasComments={comments.length > 0}
                hasReviewedFiles={reviewedFiles.size > 0}
              />
            )}
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
            {claudeInitialized && (
              <ReviewButton onClick={handleReview} loading={claudeLoading} disabled={!claudeConnected} />
            )}
          </div>
        </header>
        {files.length > 0 && (
          <div className="px-4 py-2 border-b border-gray-700 text-sm text-gray-400 flex items-center gap-3">
            {meta && meta.from && meta.to && (
              <span className="text-gray-300 font-mono text-xs">
                <span className="text-gray-500">base:</span>
                <span className="text-gray-400 ml-1">{meta.from}</span>
                <span className="text-gray-500 ml-2">compare:</span>
                <span className="text-gray-400 ml-1">{meta.to}</span>
              </span>
            )}
            <span>{files.length} file{files.length !== 1 ? 's' : ''} changed</span>
            {stats.additions > 0 && <span className="text-green-400 ml-2">+{stats.additions}</span>}
            {stats.deletions > 0 && <span className="text-red-400 ml-2">-{stats.deletions}</span>}
          </div>
        )}
      </div>
      <div className="flex">
        {(files.length > 0 || trackedFiles.length > 0) && (
          <aside className={`${fileListExpanded ? 'w-64' : 'w-10'} shrink-0 border-r border-gray-700 h-[calc(100vh-var(--sticky-header-h,90px))] sticky top-[var(--sticky-header-h,90px)] overflow-y-auto`}>
            <FileListPanel
              files={files}
              trackedFiles={trackedFiles}
              commentsByFile={commentsByFile}
              selectedFile={selectedFile}
              onSelectFile={handleSelectFile}
              expanded={fileListExpanded}
              onToggle={() => setFileListExpanded(!fileListExpanded)}
              activeTab={sidebarTab}
              onTabChange={setSidebarTab}
              reviewedFiles={reviewedFiles}
              onToggleReviewed={toggleReviewed}
            />
          </aside>
        )}
        <main className="flex-1 p-4 space-y-4 min-w-0">
          {commentError && (
            <div className="bg-red-900/30 border border-red-700 rounded-md px-4 py-2 text-red-300 text-sm">
              Comment error: {commentError}
            </div>
          )}
          {isViewingUnchangedFile ? (
            fileContentLoading ? (
              <p className="text-gray-400">Loading file...</p>
            ) : viewedFileContent ? (
              <FileViewer file={viewedFileContent} />
            ) : (
              <p className="text-gray-400">Select a file to view.</p>
            )
          ) : files.length === 0 ? (
            <p className="text-gray-400">No changes found.</p>
          ) : (
            <>
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
                      isReviewed={reviewedFiles.has(filePath)}
                      onToggleReviewed={() => toggleReviewed(filePath)}
                      saving={commentSaving}
                    />
                  </div>
                )
              })}
            </>
          )}
        </main>
        {claudeInitialized && (
          chatExpanded ? (
            <aside
              className={`shrink-0 border-l border-gray-700 h-screen sticky top-0 flex${!isChatResizing ? ' transition-[width] duration-200' : ''}`}
              style={{ width: chatWidth }}
            >
              <div
                className="w-1 cursor-col-resize hover:bg-blue-500/50 active:bg-blue-500/70 shrink-0"
                onMouseDown={onChatResizeMouseDown}
              />
              <div className="flex-1 min-w-0 p-4">
                <ChatPanel
                  onSend={sendChat}
                  messages={claudeMessages}
                  loading={claudeLoading}
                  error={claudeError}
                  connected={claudeConnected}
                  sessionId={claudeSessionId}
                  onClear={clearMessages}
                  onCollapse={() => setChatExpanded(false)}
                />
              </div>
            </aside>
          ) : (
            <aside className="w-10 shrink-0 border-l border-gray-700 h-screen sticky top-0 transition-[width] duration-200">
              <button
                type="button"
                onClick={() => setChatExpanded(true)}
                className="w-full h-full flex flex-col items-center pt-3 gap-2 text-gray-400 hover:text-gray-200"
                aria-label="Expand chat panel"
              >
                <svg className="w-5 h-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                </svg>
                <span
                  className={`w-2 h-2 rounded-full ${!claudeConnected ? 'bg-red-400' : claudeSessionId ? 'bg-green-400' : 'bg-yellow-400'}`}
                  title={!claudeConnected ? 'Disconnected' : claudeSessionId ? 'Session active' : 'No active session'}
                />
              </button>
            </aside>
          )
        )}
      </div>
    </div>
  )
}

export default App
