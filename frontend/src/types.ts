export type Status = 'pending' | 'building' | 'deploying' | 'running' | 'failed'
export type SourceType = 'git' | 'upload'

export interface Deployment {
  id: string
  name: string
  source_type: SourceType
  source_url?: string
  status: Status
  image_tag?: string
  container_id?: string
  host_port?: number
  live_url?: string
  error?: string
  created_at: string
  updated_at: string
}

export interface CreateDeploymentRequest {
  name: string
  source_type: SourceType
  source_url?: string
}