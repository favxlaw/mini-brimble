import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { api } from '../api/deployments'

export function DeploymentForm() {
  const queryClient = useQueryClient()
  const [name, setName] = useState('')
  const [sourceUrl, setSourceUrl] = useState('')

  const { mutate, isPending, error } = useMutation({
    mutationFn: api.create,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['deployments'] })
      setName('')
      setSourceUrl('')
    },
  })

  const submit = (e: React.FormEvent) => {
    e.preventDefault()
    mutate({ name, source_type: 'git', source_url: sourceUrl })
  }

  return (
    <form onSubmit={submit} style={{ marginBottom: '2rem' }}>
      <h2>New Deployment</h2>

      <div style={{ marginBottom: '0.5rem' }}>
        <input
          placeholder="Name"
          value={name}
          onChange={e => setName(e.target.value)}
          required
        />
      </div>

      <div style={{ marginBottom: '0.5rem' }}>
        <input
          placeholder="https://github.com/you/repo"
          value={sourceUrl}
          onChange={e => setSourceUrl(e.target.value)}
          style={{ width: '400px' }}
          required
        />
      </div>

      <button type="submit" disabled={isPending}>
        {isPending ? 'Deploying...' : 'Deploy'}
      </button>

      {error && <p style={{ color: 'red' }}>{(error as Error).message}</p>}
    </form>
  )
}