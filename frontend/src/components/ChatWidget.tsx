import { useState, useRef, useEffect } from 'react';
import { assistantApi } from '../api/assistant';
import type { ChatMessage } from '../api/assistant';
import './ChatWidget.css';

export function ChatWidget() {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [conversationId, setConversationId] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    if (isOpen) {
      scrollToBottom();
      inputRef.current?.focus();
    }
  }, [messages, isOpen]);

  const handleSend = async () => {
    if (!input.trim() || isLoading) return;

    const userMessage: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: input.trim(),
      timestamp: Date.now() / 1000,
    };

    setMessages(prev => [...prev, userMessage]);
    setInput('');
    setIsLoading(true);

    try {
      const response = await assistantApi.chat(userMessage.content, conversationId || undefined);
      setConversationId(response.conversation_id);

      const assistantMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: response.response,
        timestamp: Date.now() / 1000,
      };

      setMessages(prev => [...prev, assistantMessage]);
    } catch (err) {
      const errorMessage: ChatMessage = {
        id: (Date.now() + 1).toString(),
        role: 'assistant',
        content: 'Sorry, I encountered an error. Please try again.',
        timestamp: Date.now() / 1000,
      };
      setMessages(prev => [...prev, errorMessage]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const formatTime = (timestamp: number) => {
    const date = new Date(timestamp * 1000);
    return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
  };

  return (
    <div className={`chat-widget ${isOpen ? 'chat-widget-open' : ''}`}>
      {isOpen && (
        <div className="chat-panel">
          <div className="chat-header">
            <div className="chat-header-content">
              <span className="chat-avatar">&#128640;</span>
              <div>
                <div className="chat-title">SkyFlow Assistant</div>
                <div className="chat-subtitle">AI-powered booking help</div>
              </div>
            </div>
            <button className="chat-close" onClick={() => setIsOpen(false)}>
              &#10005;
            </button>
          </div>

          <div className="chat-messages">
            {messages.length === 0 && (
              <div className="chat-welcome">
                <span className="chat-welcome-icon">&#128640;</span>
                <p>Hi! I'm your SkyFlow assistant. I can help you:</p>
                <ul>
                  <li>Search for flights</li>
                  <li>Book flights and manage reservations</li>
                  <li>View your bookings</li>
                  <li>Send confirmation emails</li>
                </ul>
                <p className="chat-welcome-hint">Ask me anything about flying with SkyFlow!</p>
              </div>
            )}

            {messages.map(msg => (
              <div key={msg.id} className={`chat-message ${msg.role === 'user' ? 'chat-message-user' : 'chat-message-assistant'}`}>
                <div className="chat-message-content">{msg.content}</div>
                <div className="chat-message-time">{formatTime(msg.timestamp)}</div>
              </div>
            ))}

            {isLoading && (
              <div className="chat-message chat-message-assistant chat-message-loading">
                <div className="chat-message-content">
                  <span className="spinner"></span>
                  <span className="chat-loading-text">Thinking...</span>
                </div>
              </div>
            )}

            <div ref={messagesEndRef} />
          </div>

          <div className="chat-input-area">
            <input
              ref={inputRef}
              type="text"
              className="chat-input"
              placeholder="Ask me anything..."
              value={input}
              onChange={e => setInput(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={isLoading}
            />
            <button
              className="chat-send"
              onClick={handleSend}
              disabled={!input.trim() || isLoading}
            >
              &#10148;
            </button>
          </div>
        </div>
      )}

      {!isOpen && (
        <button className="chat-launcher" onClick={() => setIsOpen(true)}>
          <span className="chat-launcher-icon">&#128640;</span>
          <span className="chat-launcher-badge">AI</span>
        </button>
      )}
    </div>
  );
}