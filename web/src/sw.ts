// @ts-nocheck
const CACHE_NAME = 'keychat-v1'
const STATIC_CACHE = 'keychat-static-v1'
const ASSET_CACHE = 'keychat-assets-v1'

const STATIC_URLS = [
  '/',
  '/static/css/reset.css',
  '/static/css/main.css',
  '/static/icons/keychat.svg',
  '/manifest.json',
  '/sw.js',
]

self.addEventListener('install', (event) => {
  self.skipWaiting()
  event.waitUntil(
    caches.open(STATIC_CACHE).then((cache) => {
      return cache.addAll(STATIC_URLS).catch(() => {})
    }),
  )
})

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((names) => {
      return Promise.all(
        names
          .filter((name) => name !== STATIC_CACHE && name !== ASSET_CACHE)
          .map((name) => caches.delete(name)),
      )
    }),
  )
})

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url)

  if (url.pathname.startsWith('/ws') || url.pathname.startsWith('/api/')) {
    return
  }

  if (url.pathname.startsWith('/dist/assets/')) {
    event.respondWith(
      caches.open(ASSET_CACHE).then(async (cache) => {
        const cached = await cache.match(event.request)
        if (cached) return cached
        try {
          const res = await fetch(event.request)
          if (res.ok) cache.put(event.request, res.clone())
          return res
        } catch {
          return cached || new Response('offline', { status: 503 })
        }
      }),
    )
    return
  }

  event.respondWith(
    caches.match(event.request).then((response) => {
      return response || fetch(event.request).catch(() => {
        return new Response('offline', { status: 503 })
      })
    }),
  )
})

export {}
