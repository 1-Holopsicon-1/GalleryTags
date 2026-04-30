import { writable, derived, get } from 'svelte/store';
import type { FileInfo } from '../types';
import * as api from '../api/index';

export const files = writable<FileInfo[]>([]);
export const currentIndex = writable<number>(0);
export const processedCount = writable<number>(0);

export const loop = writable<boolean>(
  localStorage.getItem('loop') === 'true'
);
loop.subscribe((v) => localStorage.setItem('loop', String(v)));

export const autoplay = writable<boolean>(
  localStorage.getItem('autoplay') === 'true'
);
autoplay.subscribe((v) => localStorage.setItem('autoplay', String(v)));

export const scanSubdirs = writable<boolean>(
  localStorage.getItem('scanSubdirs') !== 'false'
);
scanSubdirs.subscribe((v) => localStorage.setItem('scanSubdirs', String(v)));

export const sortOrder = writable<string>(
  localStorage.getItem('sortOrder') || 'name'
);
sortOrder.subscribe((v) => localStorage.setItem('sortOrder', v));

export const shuffle = writable<string>(
  localStorage.getItem('shuffle') || 'off'
);
shuffle.subscribe((v) => localStorage.setItem('shuffle', v));

export const slideshowInterval = writable<number>(
  parseInt(localStorage.getItem('slideshowInterval') || '5', 10)
);
slideshowInterval.subscribe((v) => localStorage.setItem('slideshowInterval', String(v)));

export const filterTagIds = writable<Set<number>>(new Set());
const _filteredPaths = writable<Set<string> | null>(null);

export async function applyFilter() {
  const ids = get(filterTagIds);
  if (ids.size === 0) {
    _filteredPaths.set(null);
    return;
  }
  const paths = await api.getFilteredFiles([...ids]);
  _filteredPaths.set(new Set(paths));
}

export function toggleFilterTag(tagId: number) {
  filterTagIds.update(set => {
    const next = new Set(set);
    if (next.has(tagId)) {
      next.delete(tagId);
    } else {
      next.add(tagId);
    }
    return next;
  });
}

export function clearFilter() {
  filterTagIds.set(new Set());
  _filteredPaths.set(null);
}

export const filteredFiles = derived(
  [files, filterTagIds, _filteredPaths],
  ([$files, $filterTagIds, $paths]): FileInfo[] => {
    if ($filterTagIds.size === 0 || $paths === null) {
      return $files;
    }
    return $files.filter(f => $paths.has(f.path));
  }
);

export const currentFile = derived(
  [filteredFiles, currentIndex],
  ([$filtered, $idx]): FileInfo | null => {
    return $idx < $filtered.length ? $filtered[$idx] : null;
  }
);

export const hasPrev = derived(currentIndex, ($idx) => $idx > 0);

export const hasNext = derived(
  [filteredFiles, currentIndex],
  ([$filtered, $idx]) => $idx < $filtered.length - 1
);

export const isDone = derived(
  [filteredFiles, currentIndex],
  ([$filtered, $idx]) => $filtered.length > 0 && $idx >= $filtered.length
);

export const progress = derived(
  [filteredFiles, currentIndex],
  ([$filtered, $idx]) => ({
    current: Math.min($idx + 1, $filtered.length),
    total: $filtered.length,
  })
);

export function resetSession() {
  currentIndex.set(0);
  processedCount.set(0);
  clearFilter();
}
