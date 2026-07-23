
(function() {


    if (window.globalPlayer) return;

    window.globalPlayer = {


        audio:          null,
        preloadAudio:   null,
        preloadedIdx:   -1,
        currentTrack:   null,
        currentIdx:     -1,
        playing:        false,
        tracks:         [],
        volume:         80,
        repeatMode:     'off',
        fullscreenUpdateInterval: null,

        queueContextLabel: null,

        localObjectURLs: {},

        coverCache: new Map(),
        coverDB:    null,

        pendingFavorites: new Set(),

        pendingPlaylistAdd: new Map(),

        fetchTimeout: 15000,
        maxRetries:   2,

        async init() {
            await this.initCoverDB();

            if (!this.audio) {
                this.audio = new Audio();
                this.audio.volume = this.volume / 100;
                this.audio.crossOrigin = 'anonymous';

                this.audio.ontimeupdate = () => {
                    this.updateProgress();

                    if (this.audio.duration && !isNaN(this.audio.duration)) {
                        const remaining = this.audio.duration - this.audio.currentTime;
                        if (remaining < 30 && remaining > 0 && !this._preloading) {
                            this._preloadNext();
                        }
                    }
                };

                this.audio.onended = () => {
                    if (this.repeatMode === 'one') {
                        this.audio.currentTime = 0;
                        this.audio.play().catch(() => {});
                    } else {
                        this.nextTrack();
                    }
                };

                this.audio.onplay = () => {
                    this.playing = true;
                    this.updatePlayButton(true);
                    this.updateFullscreenUI();
                    this._pushNowPlaying();
                    this._broadcastState();

                    if (navigator.serviceWorker.controller) {
                        navigator.serviceWorker.controller.postMessage({
                            type: 'PLAYBACK_STARTED',
                            url:  this.audio.src
                        });
                    }
                };

                this.audio.onpause = () => {
                    this.playing = false;
                    this.updatePlayButton(false);
                    this.updateFullscreenUI();
                    this._clearNowPlaying();
                    this._broadcastState();
                };

                this.audio.onerror = (e) => {
                    const err = e.target?.error;

                    if (err && err.code === 1) return;

                    console.warn('Audio error code:', err?.code, err?.message);

                    if (!this._audioRetried && this.currentTrack) {
                        this._audioRetried = true;
                        console.log('Retrying track:', this.currentTrack.title);

                        const retrySrc = this.localObjectURLs[this.currentTrack.id]
                            ? this.localObjectURLs[this.currentTrack.id]
                            : `/api/stream/${this.currentTrack.id}?token=${encodeURIComponent(localStorage.getItem('token')||'')}`;

                        setTimeout(() => {
                            if (this.audio) {
                                this.audio.src = retrySrc;
                                this.audio.load();
                                this.audio.play().catch(() => {
                                    console.error('Retry failed, skipping track');
                                    this.showToast('Ошибка воспроизведения');
                                    setTimeout(() => this.nextTrack(), 1500);
                                });
                            }
                        }, 800);
                        return;
                    }

                    this._audioRetried = false;
                    console.error('Audio error after retry:', e);
                    this.showToast('Ошибка воспроизведения');
                    setTimeout(() => this.nextTrack(), 2000);
                };
            }

            if (!this.preloadAudio) {
                this.preloadAudio = new Audio();
                this.preloadAudio.preload = 'auto';
                this.preloadAudio.volume = 0;
                this.preloadAudio.crossOrigin = 'anonymous';
            }

            this.bindControls();
            await this.loadInitialTracks();
        },

        async initCoverDB() {
            return new Promise((resolve, reject) => {
                const request = indexedDB.open('CoverCache', 1);

                request.onerror = () => reject(request.error);

                request.onsuccess = () => {
                    this.coverDB = request.result;
                    resolve();
                };

                request.onupgradeneeded = (e) => {
                    const db = e.target.result;
                    if (!db.objectStoreNames.contains('covers')) {
                        const store = db.createObjectStore('covers', { keyPath: 'trackId' });
                        store.createIndex('timestamp', 'timestamp');
                    }
                };
            });
        },

        async getCachedCover(trackId, url) {
            if (this.coverCache.has(trackId)) {
                return this.coverCache.get(trackId);
            }

            if (this.coverDB) {
                const tx    = this.coverDB.transaction('covers', 'readonly');
                const store = tx.objectStore('covers');

                const result = await new Promise((resolve) => {
                    const req = store.get(trackId);
                    req.onsuccess = () => resolve(req.result);
                    req.onerror   = () => resolve(null);
                });

                if (result && (Date.now() - result.timestamp) < 30 * 86400000) {
                    const blobUrl = URL.createObjectURL(result.blob);
                    this.coverCache.set(trackId, blobUrl);
                    return blobUrl;
                }
            }

            try {
                const response = await this.fetchWithRetry(url);
                const blob= await response.blob();
                const blobUrl = URL.createObjectURL(blob);
                this.coverCache.set(trackId, blobUrl);

                if (this.coverDB) {
                    const tx    = this.coverDB.transaction('covers', 'readwrite');
                    const store = tx.objectStore('covers');
                    store.put({ trackId, blob, timestamp: Date.now() });
                }

                return blobUrl;
            } catch (error) {
                console.warn('Failed to load cover:', error);
                return null;
            }
        },

        async fetchWithRetry(url, options = {}, retries = this.maxRetries) {
            for (let i = 0; i <= retries; i++) {
                try {
                    const controller = new AbortController();
                    const timeoutId  = setTimeout(() => controller.abort(), this.fetchTimeout);

                    const token = localStorage.getItem('token');
                    if (token) {
                        options.headers = {
                            ...options.headers,
                            'Authorization': `Bearer ${token}`
                        };
                    }

                    const response = await fetch(url, { ...options, signal: controller.signal });
                    clearTimeout(timeoutId);

                    if (response.ok) return response;
                    if (response.status === 404) throw new Error('Not found');
                    if (response.status >= 500 && i < retries) continue;
                    throw new Error(`HTTP ${response.status}`);
                } catch (error) {
                    if (i === retries) throw error;
                    await new Promise(r => setTimeout(r, 1000 * (i + 1)));
                }
            }
            throw new Error('Max retries exceeded');
        },

        showToast(message) {
            const toast = document.getElementById('toast');
            if (toast) {
                toast.textContent = message;
                toast.classList.add('show');
                setTimeout(() => toast.classList.remove('show'), 3000);
            } else {
                console.log(message);
            }
        },

        async loadInitialTracks() {
            try {
                const response = await this.fetchWithRetry('/api/tracks');
                const tracks   = await response.json();
                if (Array.isArray(tracks) && tracks.length) {
                    this.tracks = tracks;
                }
            } catch (error) {
                console.error('Failed to load tracks:', error);
                this.showToast('Не удалось загрузить треки');
            }
        },

        _recordHistory(track) {
            if (!track || !track.id) return;
            if (typeof track.id === 'string' && track.id.startsWith('local_')) return;

            const token = localStorage.getItem('token');
            if (!token) return;

            fetch('/api/history', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ track_id: track.id })
            }).catch(() => {});
        },

        _nowPlayingHeartbeat: null,

        _pushNowPlaying() {
            const track = this.currentTrack;
            if (!track || !track.id) return;
            if (typeof track.id === 'string' && track.id.startsWith('local_')) return;

            const token = localStorage.getItem('token');
            if (!token) return;

            fetch('/api/now-playing', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${token}`
                },
                body: JSON.stringify({ track_id: track.id })
            }).catch(() => {});

            if (this._nowPlayingHeartbeat) clearInterval(this._nowPlayingHeartbeat);

            this._nowPlayingHeartbeat = setInterval(() => {
                if (!this.playing) return;
                fetch('/api/now-playing/ping', {
                    method: 'POST',
                    headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
                }).catch(() => {});
            }, 60000);
        },

        _clearNowPlaying() {
            if (this._nowPlayingHeartbeat) {
                clearInterval(this._nowPlayingHeartbeat);
                this._nowPlayingHeartbeat = null;
            }
            const token = localStorage.getItem('token');
            if (!token) return;
            fetch('/api/now-playing', {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}` }
            }).catch(() => {});
        },

        async _preloadNext() {
            if (this._preloading) return;
            if (!this.tracks.length || this.currentIdx === -1) return;

            const nextIndices = [
                (this.currentIdx + 1) % this.tracks.length,
                (this.currentIdx + 2) % this.tracks.length
            ].filter((v, i, a) => a.indexOf(v) === i && v !== this.preloadedIdx);

            this._preloading = true;

            for (const idx of nextIndices) {
                const nextTrack = this.tracks[idx];
                if (!nextTrack) continue;

                const src = this.localObjectURLs[nextTrack.id]
                    ? this.localObjectURLs[nextTrack.id]
                    : `/api/stream/${nextTrack.id}?token=${encodeURIComponent(localStorage.getItem('token')||'')}`;

                if (navigator.serviceWorker.controller) {
                    navigator.serviceWorker.controller.postMessage({
                        type: 'PRELOAD_TRACK',
                        url:  src
                    });
                }

                if (nextTrack.id) {
                    this.getCachedCover(nextTrack.id, `/api/cover/${nextTrack.id}`).catch(() => {});
                }
            }

            setTimeout(() => { this._preloading = false; }, 5000);
        },

        async loadLocalFile(file) {
            if (!file || !file.type.startsWith('audio/')) {
                this.showToast('❌ Пожалуйста, выберите аудиофайл');
                return;
            }

            if (file.size > 150 * 1024 * 1024) {
                this.showToast('❌ Файл слишком большой (макс. 150 МБ)');
                return;
            }

            const existingId = 'local_' + Date.now() + '_' + file.name;

            if (this.localObjectURLs[existingId]) {
                URL.revokeObjectURL(this.localObjectURLs[existingId]);
            }

            const objectURL = URL.createObjectURL(file);

            const nameWithoutExt = file.name.replace(/\.[^/.]+$/, '');
            const parts          = nameWithoutExt.split(' - ');
            const artist = parts.length >= 2 ? parts[0].trim() : 'Локальный файл';
            const title  = parts.length >= 2 ? parts.slice(1).join(' - ').trim() : nameWithoutExt;

            let duration = 0;
            try {
                const ac          = new (window.AudioContext || window.webkitAudioContext)();
                const arrayBuffer = await file.arrayBuffer();
                const decoded     = await ac.decodeAudioData(arrayBuffer.slice(0));
                duration = decoded.duration;
                ac.close();
            } catch(e) {}

            const localTrack = {
                id:           existingId,
                title,
                artist,
                album:        '',
                duration,
                cover_color:  '#2a1a3a',
                _isLocal:     true,
                _data:        null
            };

            this.localObjectURLs[existingId] = objectURL;

            this.tracks = this.tracks.filter(t => t.id !== existingId);

            this.tracks.unshift(localTrack);
            this.playTrack(0);
            this.showToast('Локальный трек добавлен');
        },

        _broadcastState() {
            document.dispatchEvent(new CustomEvent('player:trackstate', {
                detail: { id: this.currentTrack ? this.currentTrack.id : null, playing: this.playing }
            }));
        },

        setQueueContext(label) {
            this.queueContextLabel = label || null;
            const bar = document.getElementById('playerContextLabel');
            const fs  = document.getElementById('fullscreenContextLabel');
            if (bar) { bar.textContent = this.queueContextLabel || ''; bar.style.display = this.queueContextLabel ? '' : 'none'; }
            if (fs)  { fs.textContent  = this.queueContextLabel || ''; fs.style.display  = this.queueContextLabel ? '' : 'none'; }
        },

        updateUI() {
            const title = document.getElementById('playerTitle');
            const artist = document.getElementById('playerArtist');
            const thumb  = document.getElementById('playerThumb');
            const heart  = document.getElementById('heartBtn');

            if (title) title.textContent = this.currentTrack ? this.currentTrack.title : '—';

            if (artist && this.currentTrack) {
                const artists = this.currentTrack.artists && this.currentTrack.artists.length
                    ? this.currentTrack.artists
                    : (this.currentTrack.artist || '').split(',').map(s => s.trim()).filter(Boolean);

                artist.innerHTML = '';

                artists.forEach((name, i) => {
                    if (i > 0) {
                        const sep = document.createElement('span');
                        sep.textContent = ', ';
                        sep.style.pointerEvents = 'none';
                        artist.appendChild(sep);
                    }

                    const span = document.createElement('span');
                    span.textContent = name;
                    span.style.cssText = 'cursor:pointer;';
                    span.title = 'Открыть страницу артиста';

                    span.addEventListener('click', (e) => {
                        e.stopPropagation();

                        const navigateToArtist = (artistName) => {
                            if (window.globalArtists && window.globalArtists.length) {
                                const found = window.globalArtists.find(
                                    a => a.name.toLowerCase() === artistName.toLowerCase()
                                );
                                if (found && found.id && typeof window.navigateTo === 'function') {
                                    window.navigateTo('/artist/' + found.id);
                                    return;
                                }
                            }
                            fetch('/api/artists')
                                .then(r => r.json())
                                .then(artists => {
                                    window.globalArtists = artists;
                                    const found = artists.find(
                                        a => a.name.toLowerCase() === artistName.toLowerCase()
                                    );
                                    if (found && found.id && typeof window.navigateTo === 'function') {
                                        window.navigateTo('/artist/' + found.id);
                                    }
                                })
                                .catch(() => {});
                        };
                        navigateToArtist(name);
                    });

                    artist.appendChild(span);
                });
                artist.onclick = null;
            } else if (artist) {
                artist.innerHTML = '';
                artist.onclick = null;
            }

            if (thumb && this.currentTrack) {
                thumb.style.background = this.currentTrack.cover_color || '#1a3a1a';
                thumb.innerHTML = '';

                this.getCachedCover(this.currentTrack.id, `/api/cover/${this.currentTrack.id}`)
                    .then(coverUrl => {
                        if (coverUrl && thumb) {
                            const img = document.createElement('img');
                            img.src = coverUrl;
                            img.style.cssText = 'width:100%;height:100%;object-fit:cover;';
                            img.onerror = () => {
                                thumb.innerHTML = `<svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="rgba(179,255,0,.7)" stroke-width="1.5">
                                    <path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
                                </svg>`;
                            };
                            thumb.appendChild(img);
                        } else if (thumb) {
                            thumb.innerHTML = `<svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="rgba(179,255,0,.7)" stroke-width="1.5">
                                <path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
                            </svg>`;
                        }
                    })
                    .catch(() => {
                        if (thumb) {
                            thumb.innerHTML = `<svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="rgba(179,255,0,.7)" stroke-width="1.5">
                                <path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
                            </svg>`;
                        }
                    });
            }

            if (heart && this.currentTrack && !this.pendingFavorites.has(this.currentTrack.id)) {
                this.pendingFavorites.add(this.currentTrack.id);

                fetch('/api/favorites', {
                    headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
                })
                    .then(res => res.json())
                    .then(favs => heart.classList.toggle('liked', favs.some(f => f.id === this.currentTrack.id)))
                    .catch(() => {})
                    .finally(() => this.pendingFavorites.delete(this.currentTrack.id));
            }

            this.updatePlayButton(this.playing);
            this.updateFullscreenUI();
            this._broadcastState();
        },

        updatePlayButton(isPlaying) {
            ['playIcon', 'nowPlayingPlayIcon'].forEach(id => {
                const icon = document.getElementById(id);
                if (icon) {
                    icon.innerHTML = isPlaying
                        ? '<rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/>'
                        : '<polygon points="5 3 19 12 5 21 5 3"/>';
                }
            });
        },

        togglePlay() {
            if (!this.currentTrack && this.tracks.length) {
                this.playTrack(0);
            } else if (this.playing) {
                this.audio.pause();
            } else {
                this.audio.play().catch(error => {
                    console.error('Play failed:', error);
                    this.showToast('⚠️ Не удалось воспроизвести');
                });
            }
        },

        async playTrack(idx, savedTime = 0, shouldPlay = true) {
            if (idx < 0 || idx >= this.tracks.length) return;

            const newTrack = this.tracks[idx];

            if (this.currentTrack && this.currentTrack.id === newTrack.id && this.audio.src) {
                this.currentIdx = idx;
                if (savedTime) this.audio.currentTime = savedTime;
                if (shouldPlay && !this.playing) {
                    await this.audio.play().catch(() => {});
                } else if (!shouldPlay && this.playing) {
                    this.audio.pause();
                }
                this.updateUI();
                return;
            }

            this.currentTrack = newTrack;
            this.currentIdx   = idx;
            this._audioRetried = false;
            this.updateUI();
            this._recordHistory(newTrack);

            if (typeof window.renderRecents === 'function') {
                setTimeout(() => window.renderRecents(), 800);
            }

            this.audio.pause();

            const src = this.localObjectURLs[newTrack.id]
                ? this.localObjectURLs[newTrack.id]
                : `/api/stream/${newTrack.id}?token=${encodeURIComponent(localStorage.getItem('token')||'')}`;

            this._preloadNext();

            const newAudio = new Audio();
            newAudio.volume      = this.audio.volume;
            newAudio.crossOrigin = 'anonymous';

            newAudio.ontimeupdate = this.audio.ontimeupdate;
            newAudio.onended      = this.audio.onended;
            newAudio.onplay       = this.audio.onplay;
            newAudio.onpause      = this.audio.onpause;

            const oldAudio = this.audio;
            this.audio = newAudio;

            this.audio.src = src;
            this.audio.currentTime = savedTime;
            this.audio.load();

            setTimeout(() => {
                oldAudio.src = '';
                oldAudio.load();
            }, 1000);

            if (shouldPlay) {
                const playPromise = this.audio.play();
                if (playPromise !== undefined) {
                    playPromise.catch(error => {
                        console.warn('Auto-play blocked:', error);
                        setTimeout(() => {
                            if (!this.playing && this.currentTrack === newTrack) {
                                this.audio.play().catch(() => {});
                            }
                        }, 1000);
                    });
                }
            }
        },

        updateProgress() {
            if (!this.audio || !this.audio.duration || isNaN(this.audio.duration)) return;

            const percent = (this.audio.currentTime / this.audio.duration) * 100;
            const fill    = document.getElementById('progressFill');
            const cur     = document.getElementById('curTime');
            const dur     = document.getElementById('durTime');

            if (fill) fill.style.width = percent + '%';
            if (cur)  cur.textContent  = this.formatTime(this.audio.currentTime);
            if (dur)  dur.textContent  = this.formatTime(this.audio.duration);

            this.updateFullscreenProgress();
        },

        formatTime(seconds) {
            if (isNaN(seconds) || seconds === Infinity) return '0:00';
            const mins = Math.floor(seconds / 60);
            const secs = Math.floor(seconds % 60);
            return mins + ':' + String(secs).padStart(2, '0');
        },

        nextTrack() {
            if (!this.tracks.length) return;

            const next = this.currentIdx !== -1 ? (this.currentIdx + 1) % this.tracks.length : 0;
            this.playTrack(next);
        },

        prevTrack() {
            if (!this.tracks.length) return;
            const prev = this.currentIdx !== -1
                ? (this.currentIdx - 1 + this.tracks.length) % this.tracks.length
                : 0;
            this.playTrack(prev);
        },

        setVolume(value) {
            this.volume = Math.min(100, Math.max(0, parseInt(value) || 0));
            if (this.audio) this.audio.volume = this.volume / 100;

            const volSlider   = document.getElementById('volSlider');
            const fsVolSlider = document.getElementById('fullscreenVolSlider');
            if (volSlider)   volSlider.value   = this.volume;
            if (fsVolSlider) fsVolSlider.value = this.volume;

            localStorage.setItem('playerVolume', this.volume);
        },

        setProgress(percent) {
            if (this.audio && this.audio.duration && !isNaN(this.audio.duration)) {
                this.audio.currentTime = Math.min(
                    this.audio.duration,
                    Math.max(0, percent * this.audio.duration)
                );
            }
        },
        toggleRepeat() {
            const modes        = ['off', 'all', 'one'];
            const currentIndex = modes.indexOf(this.repeatMode);
            this.repeatMode    = modes[(currentIndex + 1) % modes.length];
            this.updateRepeatButton();
        },

        updateRepeatButton() {
            const labels = {
                off: 'Повтор: выкл',
                all: 'Повтор: все треки',
                one: 'Повтор: один трек'
            };

            ['repeatBtn', 'fullscreenRepeatBtn', 'nowPlayingRepeatBtn'].forEach(id => {
                const btn = document.getElementById(id);
                if (!btn) return;
                btn.classList.toggle('active', this.repeatMode !== 'off');
                btn.classList.toggle('repeat-one', this.repeatMode === 'one');
                btn.title = labels[this.repeatMode];
                const badge = btn.querySelector('.repeat-one-badge');
                if (badge) badge.style.display = this.repeatMode === 'one' ? 'flex' : 'none';
            });
        },

        updateFullscreenUI() {
            const fsTitle  = document.getElementById('fullscreenTitle');
            const fsArtist = document.getElementById('fullscreenArtist');
            const fsCover  = document.getElementById('fullscreenCover');
            const fsBg     = document.getElementById('fullscreenBg');

            if (fsTitle && this.currentTrack) fsTitle.textContent = this.currentTrack.title;

            if (fsArtist && this.currentTrack) {
                const artists = this.currentTrack.artists && this.currentTrack.artists.length
                    ? this.currentTrack.artists
                    : (this.currentTrack.artist || '').split(',').map(s => s.trim()).filter(Boolean);

                fsArtist.innerHTML = '';
                artists.forEach((name, i) => {
                    if (i > 0) {
                        const sep = document.createElement('span');
                        sep.textContent = ', ';
                        sep.style.pointerEvents = 'none';
                        fsArtist.appendChild(sep);
                    }
                    const span = document.createElement('span');
                    span.textContent = name;
                    span.style.cssText = 'cursor:pointer;';
                    span.addEventListener('click', () => {
                        if (typeof window.closeFullscreen === 'function') window.closeFullscreen();
                        const navigateToArtist = (artistName) => {
                            if (window.globalArtists && window.globalArtists.length) {
                                const found = window.globalArtists.find(
                                    a => a.name.toLowerCase() === artistName.toLowerCase()
                                );
                                if (found && found.id && typeof window.navigateTo === 'function') {
                                    window.navigateTo('/artist/' + found.id);
                                    return;
                                }
                            }
                            fetch('/api/artists')
                                .then(r => r.json())
                                .then(artists => {
                                    window.globalArtists = artists;
                                    const found = artists.find(
                                        a => a.name.toLowerCase() === artistName.toLowerCase()
                                    );
                                    if (found && found.id && typeof window.navigateTo === 'function') {
                                        window.navigateTo('/artist/' + found.id);
                                    }
                                })
                                .catch(() => {});
                        };
                        setTimeout(() => navigateToArtist(name), 120);
                    });
                    fsArtist.appendChild(span);
                });
                fsArtist.onclick = null;
            }

            if (fsCover && this.currentTrack) {
                fsCover.innerHTML = '';
                const fallbackColor = this.currentTrack.cover_color || '#1a3a1a';
                fsCover.style.background = fallbackColor;
                this.applyNowColor(fallbackColor);

                this.getCachedCover(this.currentTrack.id, `/api/cover/${this.currentTrack.id}`)
                    .then(coverUrl => {
                        if (coverUrl && fsCover) {
                            const img = document.createElement('img');
                            img.src = coverUrl;
                            img.style.cssText = 'width:100%;height:100%;object-fit:cover;';
                            img.onload = () => {
                                this.getAverageColor(img, (color) => {
                                    if (fsBg && color) fsBg.style.background = color;
                                    if (color) this.applyNowColor(color);
                                });
                            };
                            fsCover.innerHTML = '';
                            fsCover.appendChild(img);
                        } else if (fsCover) {
                            fsCover.innerHTML = `<svg width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="rgba(179,255,0,.7)" stroke-width="1.5">
                                <path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
                            </svg>`;
                        }
                    })
                    .catch(() => {
                        if (fsCover) {
                            fsCover.innerHTML = `<svg width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="rgba(179,255,0,.7)" stroke-width="1.5">
                                <path d="M9 18V5l12-2v13"/><circle cx="6" cy="18" r="3"/><circle cx="18" cy="16" r="3"/>
                            </svg>`;
                        }
                    });

                if (fsBg && !fsBg.style.background) fsBg.style.background = fallbackColor;
            }

            const fsPlayIcon = document.getElementById('fullscreenPlayIcon');
            if (fsPlayIcon) {
                fsPlayIcon.innerHTML = this.playing
                    ? '<rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/>'
                    : '<polygon points="5 3 19 12 5 21 5 3"/>';
            }

            if (typeof window.updateFullscreenNeighbors === 'function') {
                window.updateFullscreenNeighbors();
            }
        },

        applyNowColor(color) {
            try {
                let r, g, b;
                const m = /rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)/i.exec(color || '');
                if (m) {
                    r = +m[1]; g = +m[2]; b = +m[3];
                } else {
                    let hex = (color || '').replace('#', '');
                    if (hex.length === 3) hex = hex.split('').map(c => c + c).join('');
                    if (hex.length !== 6) return;
                    r = parseInt(hex.slice(0, 2), 16);
                    g = parseInt(hex.slice(2, 4), 16);
                    b = parseInt(hex.slice(4, 6), 16);
                }
                if ([r, g, b].some(v => Number.isNaN(v))) return;

                const r1 = r / 255, g1 = g / 255, b1 = b / 255;
                const max = Math.max(r1, g1, b1), min = Math.min(r1, g1, b1);
                let h, s, l = (max + min) / 2;
                if (max === min) { h = s = 0; }
                else {
                    const d = max - min;
                    s = l > .5 ? d / (2 - max - min) : d / (max + min);
                    switch (max) {
                        case r1: h = (g1 - b1) / d + (g1 < b1 ? 6 : 0); break;
                        case g1: h = (b1 - r1) / d + 2; break;
                        default: h = (r1 - g1) / d + 4;
                    }
                    h /= 6;
                }
                const hue2 = (h + 22 / 360) % 1;
                const hslToRgb = (h, s, l) => {
                    if (s === 0) { const v = Math.round(l * 255); return [v, v, v]; }
                    const hue2rgb = (p, q, t) => {
                        if (t < 0) t += 1; if (t > 1) t -= 1;
                        if (t < 1 / 6) return p + (q - p) * 6 * t;
                        if (t < 1 / 2) return q;
                        if (t < 2 / 3) return p + (q - p) * (2 / 3 - t) * 6;
                        return p;
                    };
                    const q = l < .5 ? l * (1 + s) : l + s - l * s;
                    const p = 2 * l - q;
                    return [
                        Math.round(hue2rgb(p, q, h + 1 / 3) * 255),
                        Math.round(hue2rgb(p, q, h) * 255),
                        Math.round(hue2rgb(p, q, h - 1 / 3) * 255)
                    ];
                };
                const l2 = Math.min(.78, Math.max(.10, l + (l < .5 ? .1 : -.08)));
                const [r2, g2, b2] = hslToRgb(hue2, Math.max(s, .32), l2);

                const root = document.documentElement.style;
                root.setProperty('--now-r', r); root.setProperty('--now-g', g); root.setProperty('--now-b', b);
                root.setProperty('--now2-r', r2); root.setProperty('--now2-g', g2); root.setProperty('--now2-b', b2);

                document.dispatchEvent(new CustomEvent('player:nowcolor', { detail: { r, g, b, r2, g2, b2 } }));
            } catch (e) {}
        },

        getAverageColor(img, callback) {
            try {
                const canvas = document.createElement('canvas');
                const w = canvas.width  = 24;
                const h = canvas.height = 24;
                const ctx = canvas.getContext('2d');
                ctx.drawImage(img, 0, 0, w, h);

                const data = ctx.getImageData(0, 0, w, h).data;

                let r = 0, g = 0, b = 0, count = 0;
                for (let i = 0; i < data.length; i += 4) {
                    if (data[i + 3] < 32) continue;
                    r += data[i];
                    g += data[i + 1];
                    b += data[i + 2];
                    count++;
                }

                if (!count) return callback(null);

                r = Math.round(r / count);
                g = Math.round(g / count);
                b = Math.round(b / count);

                callback(`rgb(${r}, ${g}, ${b})`);
            } catch (e) {
                callback(null);
            }
        },

        updateFullscreenProgress() {
            const fsFill = document.getElementById('fullscreenProgressFill');
            const fsCur  = document.getElementById('fullscreenCurTime');
            const fsDur  = document.getElementById('fullscreenDurTime');

            if (this.audio && this.audio.duration && !isNaN(this.audio.duration)) {
                const percent = (this.audio.currentTime / this.audio.duration) * 100;
                if (fsFill) fsFill.style.width  = percent + '%';
                if (fsCur)  fsCur.textContent   = this.formatTime(this.audio.currentTime);
                if (fsDur)  fsDur.textContent   = this.formatTime(this.audio.duration);
            }
        },


        bindControls() {
            const playBtn       = document.getElementById('playBtn');
            const nextBtn       = document.getElementById('nextBtn');
            const prevBtn       = document.getElementById('prevBtn');
            const volSlider     = document.getElementById('volSlider');
            const progressTrack = document.getElementById('progressTrack');
            const heartBtn      = document.getElementById('heartBtn');
            const repeatBtn     = document.getElementById('repeatBtn');


            const savedVolume = localStorage.getItem('playerVolume');
            if (savedVolume !== null) this.setVolume(parseInt(savedVolume));

            if (repeatBtn) repeatBtn.onclick = () => this.toggleRepeat();
            const fsRepeatBtn = document.getElementById('fullscreenRepeatBtn');
            if (fsRepeatBtn) fsRepeatBtn.onclick = () => this.toggleRepeat();
            this.updateRepeatButton();

            if (playBtn)  playBtn.onclick  = () => this.togglePlay();
            if (nextBtn)  nextBtn.onclick  = () => this.nextTrack();
            if (prevBtn)  prevBtn.onclick  = () => this.prevTrack();

            if (volSlider) {
                volSlider.value   = this.volume;
                volSlider.oninput = (e) => this.setVolume(e.target.value);
            }

            const fsVolSlider = document.getElementById('fullscreenVolSlider');
            if (fsVolSlider) {
                fsVolSlider.value   = this.volume;
                fsVolSlider.oninput = (e) => this.setVolume(e.target.value);
            }

            if (progressTrack) {
                progressTrack.onclick = (e) => {
                    const rect = progressTrack.getBoundingClientRect();
                    this.setProgress((e.clientX - rect.left) / rect.width);
                };
            }

            if (heartBtn) {
                heartBtn.onclick = async (e) => {
                    e.stopPropagation();
                    if (this.currentTrack && !this.currentTrack._isLocal) {
                        try {
                            const response = await this.fetchWithRetry('/api/favorites', {
                                method: 'POST',
                                headers: {
                                    'Content-Type': 'application/json',
                                    'Authorization': `Bearer ${localStorage.getItem('token')}`
                                },
                                body: JSON.stringify({ track_id: this.currentTrack.id })
                            });
                            if (response.ok) {
                                heartBtn.classList.toggle('liked');
                                this.showToast('❤️ Добавлено в избранное');
                            }
                        } catch (error) {
                            console.error('Failed to add to favorites:', error);
                        }
                    } else if (this.currentTrack?._isLocal) {
                        this.showToast('Локальные треки нельзя добавить в избранное');
                    }
                };
            }

            const barLeft = document.querySelector('.player-bar-left');
            if (barLeft) {
                barLeft.style.cursor = 'pointer';
                barLeft.onclick = (e) => {
                    if (e.target.closest('#heartBtn'))      return;
                    if (e.target.closest('#playerArtist')) return;
                    if (this.currentTrack && typeof window.openFullscreen === 'function') {
                        window.openFullscreen();
                    }
                };
            }

            this.updateUI();
            this.setQueueContext(this.queueContextLabel);
        },

        async clearCache() {
            if (navigator.serviceWorker.controller) {
                const channel = new MessageChannel();
                channel.port1.onmessage = (e) => {
                    if (e.data.success) {
                        this.showToast('Кеш очищен');
                        setTimeout(() => location.reload(), 1000);
                    }
                };
                navigator.serviceWorker.controller.postMessage(
                    { type: 'CLEAR_CACHE' },
                    [channel.port2]
                );
            } else {
                const keys = await caches.keys();
                await Promise.all(keys.map(key => caches.delete(key)));
                this.showToast('✅ Кеш очищен');
                setTimeout(() => location.reload(), 1000);
            }
        }
    };

    window.addEventListener('pagehide', () => {
        const token = localStorage.getItem('token');
        if (!token) return;
        try {
            fetch('/api/now-playing', {
                method: 'DELETE',
                headers: { 'Authorization': `Bearer ${token}` },
                keepalive: true
            }).catch(() => {});
        } catch (e) {}
    });

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', () => window.globalPlayer.init());
    } else {
        window.globalPlayer.init();
    }

})()