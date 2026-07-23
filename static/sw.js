const CACHE_NAME  = 'music-cache-v2';
const AUDIO_CACHE = 'audio-cache-v2';
const COVER_CACHE = 'cover-cache-v2';
const API_CACHE   = 'api-cache-v2';

const MAX_AUDIO_FILES = 20;
const MAX_COVERS      = 100;
const MAX_API_ITEMS   = 50;

const STRATEGIES = {
    COVER: 'stale-while-revalidate',
    API:   'cache-first',
    AUDIO: 'network-first'
};

self.addEventListener('install', e => {
    e.waitUntil(self.skipWaiting());
});

self.addEventListener('activate', e => {
    e.waitUntil(
        Promise.all([
            caches.keys().then(keys => Promise.all(
                keys
                    .filter(k =>
                        k !== CACHE_NAME &&
                        k !== AUDIO_CACHE &&
                        k !== COVER_CACHE &&
                        k !== API_CACHE
                    )
                    .map(k => caches.delete(k))
            )),

            self.clients.claim()
        ])
    );
});

async function cleanOldCache(cache, maxItems) {
    const keys = await cache.keys();
    if (keys.length > maxItems) {
        const toDelete = keys.slice(0, keys.length - maxItems);
        await Promise.all(toDelete.map(key => cache.delete(key)));
    }
}

async function getCachedOrFetch(request, cacheName, options = {}) {
    const { strategy = 'stale-while-revalidate', maxAge = 86400000 } = options;

    const cache = await caches.open(cacheName);

    if (strategy === 'cache-first') {
        const cached = await cache.match(request);

        if (cached) {
            const cachedDate = cached.headers.get('sw-cached-date');
            if (cachedDate && Date.now() - parseInt(cachedDate) < maxAge) {
                return cached;
            }
        }

        const response = await fetch(request.clone());

        if (response.ok && response.status === 200) {
            const responseToCache = response.clone();

            const headers = new Headers(responseToCache.headers);
            headers.set('sw-cached-date', Date.now().toString());

            const newResponse = new Response(responseToCache.body, {
                status:     responseToCache.status,
                statusText: responseToCache.statusText,
                headers:    headers
            });

            cache.put(request, newResponse);
            await cleanOldCache(cache, MAX_API_ITEMS);
        }

        return response;
    }

    if (strategy === 'stale-while-revalidate') {
        const cached = await cache.match(request);

        const fetchPromise = fetch(request.clone())
            .then(async response => {
                if (response.ok && response.status === 200) {
                    const responseToCache = response.clone();
                    const headers = new Headers(responseToCache.headers);
                    headers.set('sw-cached-date', Date.now().toString());
                    const newResponse = new Response(responseToCache.body, {
                        status:     responseToCache.status,
                        statusText: responseToCache.statusText,
                        headers:    headers
                    });
                    cache.put(request, newResponse);
                    await cleanOldCache(cache, MAX_COVERS);
                }
                return response;
            })
            .catch(err => {
                console.warn('Fetch failed for:', request.url, err);
                if (cached) return cached;
                throw err;
            });

        return cached || fetchPromise;
    }

    if (strategy === 'network-first') {
        try {
            const response = await fetch(request.clone());

            if (response.ok && response.status === 200) {
                const cache = await caches.open(cacheName);
                const responseToCache = response.clone();
                const headers = new Headers(responseToCache.headers);
                headers.set('sw-cached-date', Date.now().toString());
                const newResponse = new Response(responseToCache.body, {
                    status:     responseToCache.status,
                    statusText: responseToCache.statusText,
                    headers:    headers
                });
                cache.put(request, newResponse);
                await cleanOldCache(cache, MAX_AUDIO_FILES);
            }

            return response;
        } catch (err) {
            const cached = await caches.match(request);
            if (cached) return cached;
            throw err;
        }
    }

    return fetch(request);
}

let nextTrackPreloadController = null;

self.addEventListener('message', event => {

    if (event.data.type === 'PRELOAD_TRACK') {
        const { url } = event.data;

        if (nextTrackPreloadController) {
            nextTrackPreloadController.abort();
        }

        nextTrackPreloadController = new AbortController();

        fetch(url, {
            signal: nextTrackPreloadController.signal,
            headers: { 'Range': 'bytes=0-524288' }
        }).catch(() => {});
    }

    if (event.data.type === 'CLEAR_CACHE') {
        Promise.all([
            caches.delete(AUDIO_CACHE),
            caches.delete(COVER_CACHE),
            caches.delete(API_CACHE)
        ]).then(() => {
            event.ports[0].postMessage({ success: true });
        });
    }
});

self.addEventListener('fetch', e => {
    const url     = e.request.url;
    const request = e.request;

    if (request.method !== 'GET') return;

    if (url.includes('/api/stream/') || url.includes('/api/jamendo/stream')) {

        if (request.headers.get('range')) {
            e.respondWith(fetch(request));
            return;
        }

        e.respondWith(
            getCachedOrFetch(request, AUDIO_CACHE, {
                strategy: 'network-first',
                maxAge: 7 * 86400000
            })
        );
        return;
    }

    if (url.includes('/api/cover/')) {
        e.respondWith(
            getCachedOrFetch(request, COVER_CACHE, {
                strategy: 'stale-while-revalidate',
                maxAge: 30 * 86400000
            })
        );
        return;
    }

    if (url.includes('/api/')) {

        if (
            url.includes('/search')    ||
            url.includes('/favorites') ||
            url.includes('/profile')   ||
            url.includes('/playlists') ||
            url.includes('/history')
        ) {
            return;
        }

        e.respondWith(
            getCachedOrFetch(request, API_CACHE, {
                strategy: 'cache-first',
                maxAge: 5 * 60000
            })
        );
        return;
    }

    if (url.match(/\.(js|css|html|png|jpg|jpeg|gif|svg|webp)$/i)) {
        e.respondWith(
            caches.match(request).then(cached => {
                const fetchPromise = fetch(request)
                    .then(response => {
                        if (response.ok && response.status === 200) {
                            const responseToCache = response.clone();
                            caches.open(CACHE_NAME).then(cache => {
                                cache.put(request, responseToCache);
                            });
                        }
                        return response;
                    })
                    .catch(() => cached);

                return cached || fetchPromise;
            })
        );
        return;
    }
});