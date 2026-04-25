import { useEffect, useRef, useState } from 'react'
import { api } from '../api/deployments'

interface Props {
  deploymentId: string
}

export function LogStream({ deploymentId }: Props) {
  const [lines, setLines] = useState<string[]>([])
  const bottomRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    setLines([])

    // EventSource is the native browser SSE API.
    // It opens a persistent connection to the server and
    // calls onmessage each time the server sends a line.
    const es = new EventSource(api.logsUrl(deploymentId))

    es.onmessage = (e) => {
      setLines(prev => [...prev, e.data])
    }

    es.onerror = () => es.close()

    return () => es.close()
  }, [deploymentId])

  // Auto-scroll to bottom as new lines arrive
  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [lines])

  return (
    <div style={{
      background: '#111',
      color: '#eee',
      fontFamily: 'monospace',
      fontSize: '0.85rem',
      padding: '1rem',
      height: '300px',
      overflowY: 'auto',
      borderRadius: '4px',
    }}>
      {lines.length === 0 && <span style={{ color: '#666' }}>Waiting for logs...</span>}
      {lines.map((line, i) => (
        <div key={i}>{line}</div>
      ))}
      <div ref={bottomRef} />
    </div>
  )
}