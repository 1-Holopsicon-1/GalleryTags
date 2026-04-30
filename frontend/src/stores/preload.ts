import { writable, derived } from 'svelte/store';
import type { FileInfo } from '../types';
import { getMediaUrl } from '../api/media-url';

export type PreloadStatus = 'idle' | 'loading' | 'loaded' | 'error';

export const preloadStatuses = writable<Map<string, PreloadStatus>>(new Map());

export const preloadedPaths = derived(preloadStatuses, ($map) => {
  const set = new Set<string>();
  $map.forEach((status, path) => { if (status === 'loaded') set.add(path); });
  return set;
});

const MAX_CONCURRENT = 3;
let queue: { file: FileInfo; highPriority: boolean }[] = [];
let active = 0;
let currentWindow = new Set<string>();
let handles: Map<string, { img: HTMLImageElement | null; controller: AbortController | null }> = new Map();

function processQueue() {
  while (active < MAX_CONCURRENT && queue.length > 0) {
    // High priority first
    const idx = queue.findIndex((item) => item.highPriority);
    const item = idx >= 0 ? queue.splice(idx, 1)[0] : queue.shift();
    if (!item) break;
    active++;
    loadFile(item.file).finally(() => {
      active--;
      processQueue();
    });
  }
}

function loadFile(file: FileInfo): Promise<void> {
  return new Promise((resolve) => {
    preloadStatuses.update((m) => { const n = new Map(m); n.set(file.path, 'loading'); return n; });

    if (file.type === 'image') {
      const img = new Image();
      const handle: { img: HTMLImageElement | null; controller: AbortController | null } = { img, controller: null };
      handles.set(file.path, handle);

      img.onload = () => {
        preloadStatuses.update((m) => { const n = new Map(m); n.set(file.path, 'loaded'); return n; });
        handle.img = null;
        resolve();
      };
      img.onerror = () => {
        preloadStatuses.update((m) => { const n = new Map(m); n.set(file.path, 'error'); return n; });
        handle.img = null;
        resolve();
      };
      img.src = getMediaUrl(file.path);
    } else {
      const controller = new AbortController();
      const handle: { img: HTMLImageElement | null; controller: AbortController | null } = { img: null, controller };
      handles.set(file.path, handle);

      fetch(getMediaUrl(file.path), { method: 'HEAD', signal: controller.signal })
        .then(() => {
          preloadStatuses.update((m) => { const n = new Map(m); n.set(file.path, 'loaded'); return n; });
          handle.controller = null;
          resolve();
        })
        .catch((e) => {
          if (e.name !== 'AbortError') {
            preloadStatuses.update((m) => { const n = new Map(m); n.set(file.path, 'error'); return n; });
          }
          handle.controller = null;
          resolve();
        });
    }
  });
}

export function startPreload(files: FileInfo[], currentIndex: number, count = 100) {
  const end = Math.min(currentIndex + count, files.length);
  const newWindow = new Set<string>();
  for (let i = currentIndex; i < end; i++) {
    newWindow.add(files[i].path);
  }

  // Cancel loads outside the window.
  const toCancel: string[] = [];
  currentWindow.forEach((path) => {
    if (!newWindow.has(path)) toCancel.push(path);
  });

  for (const path of toCancel) {
    const handle = handles.get(path);
    if (handle) {
      if (handle.img) { handle.img.src = ''; handle.img = null; }
      if (handle.controller) { handle.controller.abort(); handle.controller = null; }
      handles.delete(path);
    }
    preloadStatuses.update((m) => { const n = new Map(m); n.set(path, 'idle'); return n; });
  }
  queue = queue.filter((item) => newWindow.has(item.file.path));

  currentWindow = newWindow;

  const $statuses = getSnapshot();
  const currentPath = files[currentIndex]?.path;
  const currentStatus = currentPath ? $statuses.get(currentPath) : undefined;

  // Current file not loaded/loading — mark high priority to jump the queue.
  if (currentPath && (!currentStatus || currentStatus === 'idle')) {
    queue = queue.filter((item) => item.file.path !== currentPath);
    queue.unshift({ file: files[currentIndex], highPriority: true });
  }

  // Enqueue remaining files in window order.
  for (let i = currentIndex; i < end; i++) {
    const f = files[i];
    const status = $statuses.get(f.path);
    if ((!status || status === 'idle') && !queue.some((item) => item.file.path === f.path)) {
      queue.push({ file: f, highPriority: false });
    }
  }

  processQueue();
}

function getSnapshot(): Map<string, PreloadStatus> {
  let val: Map<string, PreloadStatus> = new Map();
  preloadStatuses.update((m) => { val = new Map(m); return m; });
  return val;
}

export function resetPreload() {
  handles.forEach((handle) => {
    if (handle.img) { handle.img.src = ''; }
    if (handle.controller) { handle.controller.abort(); }
  });
  handles.clear();
  queue = [];
  active = 0;
  currentWindow.clear();
  preloadStatuses.set(new Map());
}