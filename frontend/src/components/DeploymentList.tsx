import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { api } from '../api/deployments'
import type { Deployment, Status } from '../types'
import { LogStream } from './LogStream'

const statusColor: Record<Status, string> = {
  pending:   '#888',
  building:  '#f5a623',
  deploying: '#4a90e2',
  running:   '#7ed321',
  failed:    '#d0021b',
}

export function DeploymentList() {
  const [selected, setSelected] = useState<string | null>(null)

  const { data: deployments = [], isLoading } = useQuery({
    queryKey: ['deployments'],
    queryFn: api.list,
    refetchInterval: 3000,
  })

  if (isLoading) return <p>Loading...</p>
  if (deployments.length === 0) return <p>No deployments yet.</p>

  return (
    <div>
      <h2>Deployments</h2>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr style={{ textAlign: 'left', borderBottom: '1px solid #333' }}>
            <th>Name</th>
            <th>Status</th>
            <th>Image</th>
            <th>URL</th>
            <th>Logs</th>
          </tr>
        </thead>
        <tbody>
          {deployments.map((d: Deployment) => (
            <>
              <tr key={d.id} style={{ borderBottom: '1px solid #222', padding: '0.5rem 0' }}>
                <td>{d.name}</td>
                <td>
                  <span style={{ color: statusColor[d.status] }}>
                    {d.status}
                  </span>
                </td>
                <td>{d.image_tag ?? '—'}</td>
                <td>
                  {d.live_url
                    ? <a href={d.live_url} target="_blank" rel="noreferrer">{d.live_url}</a>
                    : '—'
                  }
                </td>
                <td>
                  <button onClick={() => setSelected(selected === d.id ? null : d.id)}>
                    {selected === d.id ? 'hide' : 'logs'}
                  </button>
                </td>
              </tr>
              {selected === d.id && (
                <tr key={`${d.id}-logs`}>
                  <td colSpan={5}>
                    <LogStream deploymentId={d.id} />
                  </td>
                </tr>
              )}
            </>
          ))}
        </tbody>
      </table>
    </div>
  )
}