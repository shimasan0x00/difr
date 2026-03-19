import { create } from 'zustand'
import type { Comment, ReviewCategory, Severity } from '../api/types'
import * as api from '../api/client'

interface CommentState {
  comments: Comment[]
  loading: boolean
  saving: boolean
  error: string | null
  addComment: (filePath: string, line: number, body: string, reviewCategory?: ReviewCategory, severity?: Severity) => Promise<void>
  updateComment: (id: string, body: string, reviewCategory?: ReviewCategory, severity?: Severity) => Promise<void>
  removeComment: (id: string) => Promise<void>
  loadComments: (filePath?: string) => Promise<void>
  clearAll: () => Promise<void>
}

export const useCommentStore = create<CommentState>((set, get) => ({
  comments: [],
  loading: false,
  saving: false,
  error: null,
  addComment: async (filePath, line, body, reviewCategory, severity) => {
    if (get().saving) return
    try {
      set({ saving: true, error: null })
      const comment = await api.createComment(filePath, line, body, reviewCategory, severity)
      set((state) => ({ comments: [...state.comments, comment], saving: false }))
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to add comment', saving: false })
    }
  },
  updateComment: async (id, body, reviewCategory, severity) => {
    if (get().saving) return
    try {
      set({ saving: true, error: null })
      const updated = await api.updateComment(id, body, reviewCategory, severity)
      set((state) => ({
        comments: state.comments.map((c) => (c.id === id ? updated : c)),
        saving: false,
      }))
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to update comment', saving: false })
    }
  },
  removeComment: async (id) => {
    if (get().saving) return
    try {
      set({ saving: true, error: null })
      await api.deleteComment(id)
      set((state) => ({ comments: state.comments.filter((c) => c.id !== id), saving: false }))
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to remove comment', saving: false })
    }
  },
  loadComments: async (filePath?) => {
    try {
      set({ loading: true, error: null })
      const comments = await api.fetchComments(filePath)
      set({ comments, loading: false })
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to load comments', loading: false })
    }
  },
  clearAll: async () => {
    const prev = get().comments
    set({ comments: [], error: null })
    try {
      await api.deleteAllComments()
    } catch {
      set({ comments: prev })
    }
  },
}))
