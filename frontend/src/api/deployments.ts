import type { CreateDeploymentRequest, Deployment } from '../types'

const BASE = '/api'

export const api = {
  list: (): Promise<Deployment[]> =>
    fetch(`${BASE}/deployments`).then(r => r.json()),

  get: (id: string): Promise<Deployment> =>
    fetch(`${BASE}/deployments/${id}`).then(r => r.json()),

  create: (req: CreateDeploymentRequest): Promise<Deployment> =>
    fetch(`${BASE}/deployments`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(req),
    }).then(r => r.json()),

  logsUrl: (id: string) => `${BASE}/deployments/${id}/logs`,
}