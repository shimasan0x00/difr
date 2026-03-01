import { useMemo, useState } from 'react'
import type { Comment, DiffFile, DiffLine, FileStatus, Hunk } from '../../api/types'
import { HighlightedLine } from './SyntaxHighlight'
import { CommentForm } from '../comment/CommentForm'
import { InlineComment } from '../comment/InlineComment'

interface DiffViewerProps {
  file: DiffFile
  viewMode: 'split' | 'unified'
  comments?: Comment[]
  onAddComment?: (line: number, body: string) => void
  onDeleteComment?: (id: string) => void
  onUpdateComment?: (id: string, body: string) => void
}

export function DiffViewer({ file, viewMode, comments = [], onAddComment, onDeleteComment, onUpdateComment }: DiffViewerProps) {
  const [expanded, setExpanded] = useState(true)
  const displayPath = file.newPath && file.newPath !== '/dev/null' ? file.newPath : file.oldPath

  return (
    <div className="border border-gray-700 rounded-md overflow-hidden">
      <FileHeader
        path={displayPath}
        stats={file.stats}
        status={file.status}
        commentCount={comments.length}
        expanded={expanded}
        onToggle={() => setExpanded(!expanded)}
      />
      {expanded && (
        file.isBinary ? (
          <div className="px-4 py-3 text-gray-400 text-sm">Binary file not shown</div>
        ) : viewMode === 'split' ? (
          <SplitView
            hunks={file.hunks}
            language={file.language}
            comments={comments}
            onAddComment={onAddComment}
            onDeleteComment={onDeleteComment}
            onUpdateComment={onUpdateComment}
          />
        ) : (
          <UnifiedView
            hunks={file.hunks}
            language={file.language}
            comments={comments}
            onAddComment={onAddComment}
            onDeleteComment={onDeleteComment}
            onUpdateComment={onUpdateComment}
          />
        )
      )}
    </div>
  )
}

function FileHeader({
  path,
  stats,
  status,
  commentCount = 0,
  expanded = true,
  onToggle,
}: {
  path: string
  stats: { additions: number; deletions: number }
  status: FileStatus
  commentCount?: number
  expanded?: boolean
  onToggle?: () => void
}) {
  return (
    <div className="flex items-center justify-between px-4 py-2 bg-[#161b22] border-b border-gray-700">
      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onToggle}
          aria-expanded={expanded}
          aria-label="Toggle file"
          className="text-gray-400 hover:text-gray-200 text-xs w-4"
        >
          <span className={`inline-block transition-transform ${expanded ? 'rotate-90' : ''}`}>&#9654;</span>
        </button>
        <StatusBadge status={status} />
        <span className="text-sm font-mono text-gray-200">{path}</span>
      </div>
      <div className="flex items-center gap-2 text-xs">
        {commentCount > 0 && (
          <span className="text-gray-400">{commentCount} comment{commentCount !== 1 ? 's' : ''}</span>
        )}
        {stats.additions > 0 && (
          <span className="text-green-400">+{stats.additions}</span>
        )}
        {stats.deletions > 0 && (
          <span className="text-red-400">-{stats.deletions}</span>
        )}
      </div>
    </div>
  )
}

function StatusBadge({ status }: { status: FileStatus }) {
  const colors: Record<FileStatus, string> = {
    added: 'bg-green-700 text-green-200',
    deleted: 'bg-red-700 text-red-200',
    modified: 'bg-yellow-700 text-yellow-200',
    renamed: 'bg-blue-700 text-blue-200',
  }
  return (
    <span className={`px-1.5 py-0.5 rounded text-xs ${colors[status]}`} aria-label={`File status: ${status}`}>
      {status[0].toUpperCase()}
    </span>
  )
}

interface ViewProps {
  hunks: Hunk[]
  language: string
  comments: Comment[]
  onAddComment?: (line: number, body: string) => void
  onDeleteComment?: (id: string) => void
  onUpdateComment?: (id: string, body: string) => void
}

function InlineComments({
  lineNumber,
  comments,
  onAddComment,
  onDeleteComment,
  onUpdateComment,
  commentingLine,
  setCommentingLine,
}: {
  lineNumber: number | undefined
  comments: Comment[]
  onAddComment?: (line: number, body: string) => void
  onDeleteComment?: (id: string) => void
  onUpdateComment?: (id: string, body: string) => void
  commentingLine: number | null
  setCommentingLine: (line: number | null) => void
}) {
  if (!lineNumber) return null
  const lineComments = comments.filter((c) => c.line === lineNumber)
  const isCommenting = commentingLine === lineNumber

  return (
    <>
      {lineComments.map((c) => (
        <div key={c.id} className="px-4 py-1">
          <InlineComment
            comment={c}
            onDelete={(id) => onDeleteComment?.(id)}
            onUpdate={onUpdateComment}
          />
        </div>
      ))}
      {isCommenting && onAddComment && (
        <div className="px-4 py-1">
          <CommentForm
            onSubmit={(body) => {
              onAddComment(lineNumber, body)
              setCommentingLine(null)
            }}
            onCancel={() => setCommentingLine(null)}
          />
        </div>
      )}
    </>
  )
}

function UnifiedView({ hunks, language, comments, onAddComment, onDeleteComment, onUpdateComment }: ViewProps) {
  const [commentingLine, setCommentingLine] = useState<number | null>(null)

  return (
    <div className="text-sm font-mono">
      {hunks.map((hunk) => (
        <div key={`${hunk.oldStart}-${hunk.newStart}`}>
          {hunk.header && (
            <div className="px-4 py-1 bg-[#1c2128] text-gray-500 text-xs">
              @@ {hunk.header} @@
            </div>
          )}
          {hunk.lines.map((line, j) => {
            const lineNum = line.newNumber ?? line.oldNumber
            const lineKey = `${line.type}-${line.oldNumber ?? 0}-${line.newNumber ?? 0}-${j}`
            return (
              <div key={lineKey}>
                <UnifiedLine
                  line={line}
                  language={language}
                  onClickComment={onAddComment ? () => lineNum && setCommentingLine(lineNum) : undefined}
                />
                <InlineComments
                  lineNumber={lineNum}
                  comments={comments}
                  onAddComment={onAddComment}
                  onDeleteComment={onDeleteComment}
                  onUpdateComment={onUpdateComment}
                  commentingLine={commentingLine}
                  setCommentingLine={setCommentingLine}
                />
              </div>
            )
          })}
        </div>
      ))}
    </div>
  )
}

function UnifiedLine({
  line,
  language,
  onClickComment,
}: {
  line: DiffLine
  language: string
  onClickComment?: () => void
}) {
  const bgColors: Record<string, string> = {
    add: 'bg-[#0d2818]',
    delete: 'bg-[#3d1215]',
    context: '',
  }
  const textColors: Record<string, string> = {
    add: 'text-green-300',
    delete: 'text-red-300',
    context: 'text-gray-300',
  }
  const prefix: Record<string, string> = {
    add: '+',
    delete: '-',
    context: ' ',
  }

  return (
    <div
      className={`flex group ${bgColors[line.type]} hover:brightness-125`}
      data-line-type={line.type}
    >
      <span className="w-12 text-right pr-2 text-gray-500 select-none shrink-0 text-xs leading-6">
        {line.oldNumber ?? ''}
      </span>
      <span className="w-12 text-right pr-2 text-gray-500 select-none shrink-0 text-xs leading-6">
        {line.newNumber ?? ''}
      </span>
      <span className="w-4 text-center text-gray-500 select-none shrink-0 text-xs leading-6">
        {prefix[line.type]}
      </span>
      <span className={`flex-1 px-2 whitespace-pre-wrap break-all leading-6 ${textColors[line.type]}`}>
        <HighlightedLine code={line.content} language={language} />
      </span>
      {onClickComment && (
        <button
          type="button"
          onClick={onClickComment}
          className="w-6 text-center text-transparent group-hover:text-blue-400 focus:text-blue-400 hover:text-blue-300 shrink-0 text-xs leading-6"
          aria-label="Add comment"
        >
          +
        </button>
      )}
    </div>
  )
}

function SplitView({ hunks, language, comments, onAddComment, onDeleteComment, onUpdateComment }: ViewProps) {
  const [commentingLine, setCommentingLine] = useState<number | null>(null)

  const pairedLines = useMemo(() => pairLines(hunks), [hunks])

  return (
    <div className="text-sm font-mono" data-testid="split-view">
      {pairedLines.map((pair, i) => {
        if (pair.type === 'header') {
          return (
            <div key={`header-${i}`} className="px-4 py-1 bg-[#1c2128] text-gray-500 text-xs">
              @@ {pair.header} @@
            </div>
          )
        }

        const lineNum = pair.right?.newNumber ?? pair.left?.newNumber ?? pair.left?.oldNumber
        return (
          <div key={`${pair.left?.oldNumber ?? 0}-${pair.right?.newNumber ?? 0}-${i}`}>
            <div className="flex" data-line-type={pair.left?.type ?? pair.right?.type ?? 'context'}>
              {/* Left side (old) */}
              <div className={`flex-1 flex group/left ${pair.left?.type === 'delete' ? 'bg-[#3d1215]' : ''}`}>
                <span className="w-12 text-right pr-2 text-gray-500 select-none shrink-0 text-xs leading-6">
                  {pair.left?.oldNumber ?? ''}
                </span>
                <span className={`flex-1 px-2 whitespace-pre-wrap break-all leading-6 ${pair.left?.type === 'delete' ? 'text-red-300' : 'text-gray-300'}`}>
                  {pair.left ? <HighlightedLine code={pair.left.content} language={language} /> : ''}
                </span>
                {onAddComment && pair.left?.type === 'delete' && !pair.right && pair.left.oldNumber && (
                  <button
                    type="button"
                    onClick={() => setCommentingLine(pair.left!.oldNumber!)}
                    className="w-6 text-center text-transparent group-hover/left:text-blue-400 focus:text-blue-400 hover:text-blue-300 shrink-0 text-xs leading-6"
                    aria-label="Add comment"
                  >
                    +
                  </button>
                )}
              </div>
              {/* Right side (new) */}
              <div className={`flex-1 flex border-l border-gray-700 group ${pair.right?.type === 'add' ? 'bg-[#0d2818]' : ''}`}>
                <span className="w-12 text-right pr-2 text-gray-500 select-none shrink-0 text-xs leading-6">
                  {pair.right?.newNumber ?? ''}
                </span>
                <span className={`flex-1 px-2 whitespace-pre-wrap break-all leading-6 ${pair.right?.type === 'add' ? 'text-green-300' : 'text-gray-300'}`}>
                  {pair.right ? <HighlightedLine code={pair.right.content} language={language} /> : ''}
                </span>
                {onAddComment && lineNum && (
                  <button
                    type="button"
                    onClick={() => setCommentingLine(lineNum)}
                    className="w-6 text-center text-transparent group-hover:text-blue-400 focus:text-blue-400 hover:text-blue-300 shrink-0 text-xs leading-6"
                    aria-label="Add comment"
                  >
                    +
                  </button>
                )}
              </div>
            </div>
            {lineNum && (
              <InlineComments
                lineNumber={lineNum}
                comments={comments}
                onAddComment={onAddComment}
                onDeleteComment={onDeleteComment}
                onUpdateComment={onUpdateComment}
                commentingLine={commentingLine}
                setCommentingLine={setCommentingLine}
              />
            )}
          </div>
        )
      })}
    </div>
  )
}

interface PairedLine {
  type: 'pair' | 'header'
  left?: DiffLine
  right?: DiffLine
  header?: string
}

/** Pair delete/add lines side-by-side for split view (GitHub-style). */
function pairLines(hunks: Hunk[]): PairedLine[] {
  const result: PairedLine[] = []

  for (const hunk of hunks) {
    if (hunk.header) {
      result.push({ type: 'header', header: hunk.header })
    }

    const lines = hunk.lines
    let i = 0

    while (i < lines.length) {
      const line = lines[i]

      if (line.type === 'context') {
        result.push({ type: 'pair', left: line, right: line })
        i++
      } else if (line.type === 'delete') {
        // Collect consecutive deletes
        const deletes: DiffLine[] = []
        while (i < lines.length && lines[i].type === 'delete') {
          deletes.push(lines[i])
          i++
        }
        // Collect consecutive adds
        const adds: DiffLine[] = []
        while (i < lines.length && lines[i].type === 'add') {
          adds.push(lines[i])
          i++
        }
        // Pair them side by side
        const max = Math.max(deletes.length, adds.length)
        for (let j = 0; j < max; j++) {
          result.push({
            type: 'pair',
            left: j < deletes.length ? deletes[j] : undefined,
            right: j < adds.length ? adds[j] : undefined,
          })
        }
      } else if (line.type === 'add') {
        result.push({ type: 'pair', right: line })
        i++
      } else {
        i++
      }
    }
  }

  return result
}
