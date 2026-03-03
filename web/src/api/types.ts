export type FileStatus = 'added' | 'deleted' | 'modified' | 'renamed'
export type LineType = 'context' | 'add' | 'delete'

export interface DiffLine {
  type: LineType
  content: string
  oldNumber?: number
  newNumber?: number
}

export interface Hunk {
  oldStart: number
  oldLines: number
  newStart: number
  newLines: number
  header: string
  lines: DiffLine[]
}

export interface FileStats {
  additions: number
  deletions: number
}

export interface DiffFile {
  oldPath: string
  newPath: string
  status: FileStatus
  language: string
  isBinary: boolean
  hunks: Hunk[]
  stats: FileStats
}

export interface DiffMeta {
  from: string
  to: string
  mode: string
}

export interface DiffResult {
  files: DiffFile[]
  stats: FileStats
  meta: DiffMeta
}

export interface Comment {
  id: string
  filePath: string
  line: number
  body: string
  createdAt: string
  updatedAt?: string
}

export interface FileContent {
  path: string
  content: string
  isBinary?: boolean
  isTruncated?: boolean
  size: number
}

export interface ChatMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  sessionId?: string | null
}

export interface WSMessage {
  type: 'chat' | 'review' | 'clear'
  content: string
}

export interface WSResponse {
  type: 'session' | 'text' | 'done' | 'error'
  content?: string
  sessionId?: string
  error?: string
}
