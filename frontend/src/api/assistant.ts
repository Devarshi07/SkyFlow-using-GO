import { request } from './client';

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
  timestamp: number;
  isLoading?: boolean;
}

export interface ChatResponse {
  response: string;
  conversation_id: string;
  tool_calls?: {
    tool: string;
    params: Record<string, unknown>;
    result: string;
  }[];
}

export interface Conversation {
  id: string;
  user_id: string;
  messages: ChatMessage[];
  created_at: string;
  updated_at: string;
}

export interface ChatRequest {
  message: string;
  conversation_id?: string;
}

export const assistantApi = {
  chat: (message: string, conversationId?: string) =>
    request<ChatResponse>('/assistant/chat', {
      method: 'POST',
      body: JSON.stringify({
        message,
        conversation_id: conversationId || '',
      } as ChatRequest),
    }),

  getConversation: (id: string) =>
    request<Conversation>(`/assistant/conversation/${id}`),
};

// Helper to format a single message for display
export function formatMessageTimestamp(timestamp: number): string {
  const date = new Date(timestamp * 1000); // Convert Unix timestamp to Date
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
}