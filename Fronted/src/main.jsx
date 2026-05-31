import React, { useMemo, useState } from 'react';
import { createRoot } from 'react-dom/client';
import './styles.css';

const API_BASE_URL = 'http://localhost:6872/api';

function getSessionId() {
  const key = 'paper-agent-session-id';
  const current = window.localStorage.getItem(key);
  if (current) {
    return current;
  }

  const next = crypto.randomUUID ? crypto.randomUUID() : `${Date.now()}-${Math.random()}`;
  window.localStorage.setItem(key, next);
  return next;
}

function ChatPage() {
  const sessionId = useMemo(getSessionId, []);
  const [question, setQuestion] = useState('');
  const [messages, setMessages] = useState([]);
  const [agentMode, setAgentMode] = useState('react');
  const [chatMode, setChatMode] = useState('quick');
  const [loading, setLoading] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');

  function getEndpoint() {
    if (agentMode === 'plan') {
      return chatMode === 'stream' ? '/chat_p_stream' : '/chat_p';
    }
    return chatMode === 'stream' ? '/chat_stream' : '/chat';
  }

  function getModeLabel() {
    return `${agentMode === 'plan' ? 'Plan-Execute' : 'ReAct'} ${chatMode === 'stream' ? '流式对话' : '快速对话'}`;
  }

  async function sendStreamMessage(text) {
    const assistantIndex = messages.length + 1;
    setMessages((items) => [...items, { role: 'assistant', content: '' }]);

    const response = await fetch(`${API_BASE_URL}${getEndpoint()}`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        id: sessionId,
        question: text,
      }),
    });

    if (!response.ok || !response.body) {
      throw new Error(`流式请求失败：${response.status}`);
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder('utf-8');
    let buffer = '';

    for (;;) {
      const { value, done } = await reader.read();
      if (done) {
        break;
      }
      buffer += decoder.decode(value, { stream: true });
      const events = buffer.split('\n\n');
      buffer = events.pop() || '';

      for (const rawEvent of events) {
        const event = parseSSEEvent(rawEvent);
        if (event.event === 'message') {
          setMessages((items) => updateMessage(items, assistantIndex, event.data));
        }
        if (event.event === 'error') {
          throw new Error(event.data || '流式响应失败');
        }
        if (event.event === 'done') {
          return;
        }
      }
    }
  }

  function parseSSEEvent(rawEvent) {
    return rawEvent.split('\n').reduce(
      (event, line) => {
        if (line.startsWith('event:')) {
          event.event = line.slice(6).trim();
        }
        if (line.startsWith('data:')) {
          event.data += line.slice(5).trimStart();
        }
        return event;
      },
      { event: '', data: '' },
    );
  }

  function updateMessage(items, index, chunk) {
    return items.map((item, itemIndex) => {
      if (itemIndex !== index) {
        return item;
      }
      return { ...item, content: `${item.content}${chunk}` };
    });
  }

  async function uploadPaper(event) {
    const file = event.target.files?.[0];
    event.target.value = '';
    if (!file || uploading) {
      return;
    }
    if (!file.name.toLowerCase().endsWith('.pdf')) {
      setError('仅支持上传 PDF 文件');
      return;
    }

    setError('');
    setUploading(true);
    setMessages((items) => [...items, { role: 'user', content: `上传论文：${file.name}` }]);

    try {
      const formData = new FormData();
      formData.append('file', file);
      const response = await fetch(`${API_BASE_URL}/upload`, {
        method: 'POST',
        body: formData,
      });
      const data = await response.json();
      if (!response.ok || data.message !== 'OK') {
        throw new Error(data.message || `上传失败：${response.status}`);
      }

      const info = data.data || {};
      setMessages((items) => [
        ...items,
        {
          role: 'assistant',
          content: `论文已上传并开始构建知识库：${info.fileName || file.name}`,
        },
      ]);
    } catch (err) {
      setError(err.message || '上传失败');
      setMessages((items) => [...items, { role: 'assistant', content: '论文上传失败，请检查后端服务和文件格式。' }]);
    } finally {
      setUploading(false);
    }
  }

  async function sendMessage(event) {
    event.preventDefault();
    const text = question.trim();
    if (!text || loading) {
      return;
    }

    setQuestion('');
    setError('');
    setLoading(true);
    setMessages((items) => [...items, { role: 'user', content: text }]);

    try {
      if (chatMode === 'stream') {
        await sendStreamMessage(text);
        return;
      }

      const response = await fetch(`${API_BASE_URL}${getEndpoint()}`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          id: sessionId,
          question: text,
        }),
      });

      const data = await response.json();
      if (!response.ok || data.message !== 'OK') {
        throw new Error(data.message || `请求失败：${response.status}`);
      }

      const answer = data.data?.answer || data.data?.Answer || data.data?.result || data.data?.Result || '';
      setMessages((items) => [...items, { role: 'assistant', content: answer }]);
    } catch (err) {
      setError(err.message || '请求失败');
      setMessages((items) => [...items, { role: 'assistant', content: '请求失败，请检查后端服务是否已启动。' }]);
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="app-shell">
      <aside className="sidebar">
        <button className="new-chat" type="button" onClick={() => setMessages([])}>
          + 新对话
        </button>
        <div className="side-section">
          <p className="side-label">PaperAgent</p>
          <div className="side-item active">快速对话</div>
        </div>
      </aside>

      <section className="chat-shell">
        <header className="topbar">
          <div className="model-name">PaperAgent</div>
        </header>

        <div className={`messages ${messages.length === 0 ? 'empty' : ''}`}>
          {messages.length === 0 && (
            <div className="welcome">
              <div className="logo-mark">P</div>
              <h1>今天想聊什么？</h1>
              <p>输入问题，开始和 PaperAgent {getModeLabel()}。</p>
            </div>
          )}

          {messages.map((message, index) => (
            <article key={`${message.role}-${index}`} className={`message-row ${message.role}`}>
              <div className="message-inner">
                <div className="avatar">{message.role === 'user' ? '你' : 'P'}</div>
                <div className="message-content">{message.content}</div>
              </div>
            </article>
          ))}
          {loading && (
            <article className="message-row assistant">
              <div className="message-inner">
                <div className="avatar">P</div>
                <div className="message-content typing">
                  <span />
                  <span />
                  <span />
                </div>
              </div>
            </article>
          )}
        </div>

        <div className="composer-wrap">
          {error && <div className="error">{error}</div>}
          <form className="composer" onSubmit={sendMessage}>
            <textarea
              value={question}
              onChange={(event) => setQuestion(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === 'Enter' && !event.shiftKey) {
                  event.preventDefault();
                  sendMessage(event);
                }
              }}
              placeholder={`给 PaperAgent 发送消息，当前为${getModeLabel()}`}
              rows={2}
            />
            <div className="upload-action">
              <input id="paper-upload" type="file" accept="application/pdf,.pdf" onChange={uploadPaper} disabled={uploading} />
              <label htmlFor="paper-upload" className={uploading ? 'upload-button disabled' : 'upload-button'} title="上传论文PDF">
                {uploading ? '…' : '+'}
              </label>
            </div>
            <div className="composer-actions">
              <label className="mode-pill">
                <select value={agentMode} onChange={(event) => setAgentMode(event.target.value)} aria-label="Agent模式">
                  <option value="react">ReAct</option>
                  <option value="plan">Plan-Execute</option>
                </select>
              </label>
              <label className="mode-pill">
                <select value={chatMode} onChange={(event) => setChatMode(event.target.value)} aria-label="输出模式">
                  <option value="quick">快速</option>
                  <option value="stream">流式</option>
                </select>
              </label>
              <button className="send-button" type="submit" disabled={loading || !question.trim()} aria-label="发送消息">
                ↑
              </button>
            </div>
          </form>
          <p className="hint">PaperAgent 可能会出错，请核对重要信息。</p>
        </div>
      </section>
    </main>
  );
}

function App() {
  return <ChatPage />;
}

createRoot(document.getElementById('root')).render(<App />);
