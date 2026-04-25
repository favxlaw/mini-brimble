import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { createRouter, createRoute, createRootRoute, RouterProvider, Outlet } from '@tanstack/react-router'
import React from 'react'
import ReactDOM from 'react-dom/client'
import { DeploymentForm } from './components/DeploymentForm'
import { DeploymentList } from './components/DeploymentList'

const rootRoute = createRootRoute({
  component: () => (
    <main style={{ maxWidth: 900, margin: '0 auto', padding: '2rem' }}>
      <h1>mini-brimble</h1>
      <Outlet />
    </main>
  ),
})

const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/',
  component: () => (
    <div>
      <DeploymentForm />
      <DeploymentList />
    </div>
  ),
})

const routeTree = rootRoute.addChildren([indexRoute])
const router = createRouter({ routeTree })

declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}

const queryClient = new QueryClient()

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  </React.StrictMode>
)