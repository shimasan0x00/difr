import type { FileContent } from '../../api/types'
import { HighlightedLine } from './SyntaxHighlight'

interface FileViewerProps {
  file: FileContent
}

const langMap: Record<string, string> = {
  go: 'go',
  ts: 'typescript',
  tsx: 'typescript',
  js: 'javascript',
  jsx: 'javascript',
  py: 'python',
  rs: 'rust',
  java: 'java',
  rb: 'ruby',
  php: 'php',
  c: 'c',
  cpp: 'cpp',
  h: 'c',
  hpp: 'cpp',
  cs: 'csharp',
  swift: 'swift',
  kt: 'kotlin',
  sql: 'sql',
  sh: 'bash',
  bash: 'bash',
  zsh: 'bash',
  yaml: 'yaml',
  yml: 'yaml',
  json: 'json',
  xml: 'html',
  html: 'html',
  css: 'css',
  scss: 'css',
  md: 'markdown',
  toml: 'toml',
}

function detectLanguage(path: string): string {
  const ext = path.split('.').pop() ?? ''
  return langMap[ext] ?? ext
}

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

export function FileViewer({ file }: FileViewerProps) {
  const language = detectLanguage(file.path)

  return (
    <div className="border border-gray-700 rounded-md overflow-hidden">
      <div className="flex items-center justify-between px-4 py-2 bg-[#161b22] border-b border-gray-700">
        <span className="text-sm font-mono text-gray-200">{file.path}</span>
        <span className="text-xs text-gray-400">{formatSize(file.size)}</span>
      </div>
      {file.isBinary ? (
        <div className="px-4 py-3 text-gray-400 text-sm">
          Binary file ({formatSize(file.size)})
        </div>
      ) : file.isTruncated ? (
        <div className="px-4 py-3 text-gray-400 text-sm">
          File too large to display ({formatSize(file.size)})
        </div>
      ) : file.content === '' ? (
        <div className="px-4 py-3 text-gray-400 text-sm">
          Empty file
        </div>
      ) : (
        <div className="text-sm font-mono">
          {file.content.split('\n').map((line, i, arr) => {
            // Don't render trailing empty line from split
            if (i === arr.length - 1 && line === '') return null
            return (
              <div key={i} className="flex hover:bg-[#161b22]">
                <span className="w-12 text-right pr-2 text-gray-500 select-none shrink-0 text-xs leading-6">
                  {i + 1}
                </span>
                <span className="flex-1 px-2 whitespace-pre-wrap break-all leading-6 text-gray-300">
                  <HighlightedLine code={line} language={language} />
                </span>
              </div>
            )
          })}
        </div>
      )}
    </div>
  )
}
