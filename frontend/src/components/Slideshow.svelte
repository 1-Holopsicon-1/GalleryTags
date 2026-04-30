<script lang="ts">
  import { onMount } from 'svelte';
  import { initMediaUrl } from '../api/media-url';
  import * as api from '../api/index';
  import { files, currentIndex, currentFile, processedCount, hasPrev, isDone, progress, resetSession, loop, scanSubdirs, slideshowInterval, sortOrder, shuffle, filterTagIds, filteredFiles, applyFilter, toggleFilterTag, clearFilter } from '../stores/files';
  import { tags, loadTags } from '../stores/tags';
  import { startPreload, resetPreload, preloadedPaths } from '../stores/preload';
  import MediaViewer from './MediaViewer.svelte';
  import TagPanel from './TagPanel.svelte';
  import Overlay from './Overlay.svelte';
  import Settings from './Settings.svelte';
  import TagManager from './TagManager.svelte';
  import ThumbnailSidebar from './ThumbnailSidebar.svelte';

  let loading = true;
  let needsSetup = false;
  let showSettings = false;
  let showTagManager = false;
  let toast: { message: string; type: string } | null = null;
  let toastTimeout: ReturnType<typeof setTimeout> | null = null;
  let tagPanelDisabled = false;
  let fileTagIds: number[] = [];
  let slideshowMode = false;
  let slideshowPaused = false;
  let showTopBar = false;
  let showTagPanel = false;
  let hideTimerTop: ReturnType<typeof setTimeout> | null = null;
  let hideTimerBottom: ReturnType<typeof setTimeout> | null = null;
  let slideTimer: ReturnType<typeof setInterval> | null = null;
  let jumpIndex = $progress.current;
  let jumpFocused = false;
  let scanVersion = 0;

  $: if ($currentFile) {
    api.getFileTags($currentFile.path).then((tags) => {
      fileTagIds = tags.map((t) => t.id);
    }).catch(() => {
      fileTagIds = [];
    });
    if (slideshowMode) startSlideTimer();
  } else {
    fileTagIds = [];
  }

  $: if ($files.length > 0 && $currentIndex >= 0) {
    startPreload($files, $currentIndex, $shuffle === 'all' ? 1 : 100);
  }

  $: if (!jumpFocused) jumpIndex = $progress.current;

  onMount(() => {
    initMediaUrl();
    loadTags();
    loadFiles();
  });

  async function loadFiles() {
    resetPreload();
    const version = ++scanVersion;
    loading = true;
    try {
      const result = await api.scanInbox($scanSubdirs, $sortOrder);
      if (version !== scanVersion) return;
      const list = Array.isArray(result) ? result : [];
      needsSetup = list.length === 0;
      const currentPath = $currentFile?.path;
      files.set(list);
      if (currentPath) {
        const idx = list.findIndex(f => f.path === currentPath);
        currentIndex.set(idx >= 0 ? idx : Math.min($currentIndex, list.length - 1));
      } else {
        resetSession();
      }
    } catch {
      if (version !== scanVersion) return;
      needsSetup = true;
      showToast('Failed to scan inbox. Check settings.', 'error');
    } finally {
      if (version === scanVersion) loading = false;
    }
  }

  async function handleSave(e: CustomEvent) {
    if (!$currentFile) return;
    try {
      const result = await api.applyTags($currentFile.path, e.detail.tagIDs);
      processedCount.update((n) => n + 1);
      if (result.moved) {
        showToast('File moved.', 'info');
        const idx = $files.findIndex(f => f.path === result.newPath);
        if (idx >= 0) {
          files.update(list => { list[idx] = { ...list[idx], path: result.newPath }; return list; });
        }
      }
      fileTagIds = e.detail.tagIDs;
    } catch (e: any) {
      showToast(e?.message || 'Failed to save tags.', 'error');
    }
  }

  function handleNext() {
    currentIndex.update((i) => {
      if (i + 1 >= $filteredFiles.length && $loop) return 0;
      return i + 1;
    });
  }

  function goPrev() { currentIndex.update((i) => Math.max(0, i - 1)); }

  function goNext() {
    if ($shuffle === 'off') {
      currentIndex.update((i) => {
        if (i + 1 >= $filteredFiles.length && $loop) return 0;
        return i + 1;
      });
    } else {
      goRandom();
    }
  }

  function goRandom() {
    if ($shuffle === 'loaded') {
      const loaded = $preloadedPaths;
      const candidates = $filteredFiles.filter((f, i) => i !== $currentIndex && loaded.has(f.path));
      if (candidates.length === 0) {
        goRandomAll();
        return;
      }
      const pick = candidates[Math.floor(Math.random() * candidates.length)];
      const idx = $filteredFiles.indexOf(pick);
      if (idx >= 0) currentIndex.set(idx);
    } else {
      goRandomAll();
    }
  }

  function goRandomAll() {
    if ($filteredFiles.length <= 1) return;
    let next: number;
    do {
      next = Math.floor(Math.random() * $filteredFiles.length);
    } while (next === $currentIndex);
    currentIndex.set(next);
  }

  function buildHotkey(e: KeyboardEvent): string {
    const parts: string[] = [];
    if (e.ctrlKey) parts.push('ctrl');
    if (e.altKey) parts.push('alt');
    if (e.shiftKey) parts.push('shift');
    let key = e.key.toLowerCase();
    if (key === ' ') key = 'space';
    parts.push(key);
    return parts.join('+');
  }

  async function handleHotkeyTag(tagId: number) {
    if (!$currentFile) return;
    const current = new Set(fileTagIds);
    if (current.has(tagId)) {
      current.delete(tagId);
    } else {
      // Enforce single folder tag
      const tag = $tags.find(t => t.id === tagId);
      if (tag?.type === 'folder') {
        for (const t of $tags) {
          if (t.type === 'folder' && t.id !== tagId) current.delete(t.id);
        }
      }
      current.add(tagId);
    }
    const ids = [...current];
    try {
      const result = await api.applyTags($currentFile.path, ids);
      processedCount.update((n) => n + 1);
      if (result.moved) {
        const idx = $files.findIndex(f => f.path === result.newPath);
        if (idx >= 0) {
          files.update(list => { list[idx] = { ...list[idx], path: result.newPath }; return list; });
        }
      }
      fileTagIds = ids;
    } catch (e: any) {
      showToast(e?.message || 'Failed to apply tag.', 'error');
    }
  }

  function handleKeydown(e: KeyboardEvent) {
    if ((e.target as HTMLElement).tagName === 'INPUT' || (e.target as HTMLElement).tagName === 'SELECT') return;
    if (showSettings || showTagManager) return;

    // Check tag hotkeys first
    if (!e.ctrlKey || e.altKey || e.shiftKey || e.key.length === 1) {
      const hk = buildHotkey(e);
      const tag = $tags.find(t => t.hotkey === hk);
      if (tag) {
        e.preventDefault();
        handleHotkeyTag(tag.id);
        return;
      }
    }

    if (e.key === 'Escape' && slideshowMode) { e.preventDefault(); exitSlideshow(); return; }
    if (e.key === 'f' || e.key === 'F') { e.preventDefault(); toggleSlideshow(); return; }
    if (e.key === ' ' && slideshowMode) { e.preventDefault(); togglePause(); return; }
    if (e.key === 'ArrowRight') { e.preventDefault(); goNext(); }
    else if (e.key === 'ArrowLeft') { e.preventDefault(); goPrev(); }
    else if (e.key === 'Delete') { e.preventDefault(); handleTrash(); }
  }

  function toggleSlideshow() {
    if (slideshowMode) exitSlideshow();
    else startSlideshow();
  }

  function togglePause() {
    if (slideshowPaused) {
      slideshowPaused = false;
      startSlideTimer();
    } else {
      slideshowPaused = true;
      stopSlideTimer();
    }
  }

  function cycleShuffle() {
    if ($shuffle === 'off') shuffle.set('all');
    else if ($shuffle === 'all') shuffle.set('loaded');
    else shuffle.set('off');
  }

  function startSlideshow() {
    if (!$currentFile) return;
    slideshowMode = true;
    showTopBar = false;
    showTagPanel = false;
    startSlideTimer();
    document.documentElement.requestFullscreen?.().catch(() => {});
  }

  function exitSlideshow() {
    slideshowMode = false;
    slideshowPaused = false;
    showTopBar = false;
    showTagPanel = false;
    stopSlideTimer();
    if (document.fullscreenElement) document.exitFullscreen?.().catch(() => {});
  }

  function startSlideTimer() {
    stopSlideTimer();
    if (slideshowPaused) return;
    if ($currentFile?.type === 'video') return; // wait for videoEnded
    slideTimer = setInterval(() => {
      if ($isDone && !$loop) { exitSlideshow(); return; }
      goNext();
    }, $slideshowInterval * 1000);
  }

  function handleVideoEnded() {
    if (slideshowMode) {
      if ($isDone && !$loop) { exitSlideshow(); return; }
      goNext();
    }
  }

  function stopSlideTimer() {
    if (slideTimer) { clearInterval(slideTimer); slideTimer = null; }
  }

  function handleSlideshowMouseMove(e: MouseEvent) {
    if (!slideshowMode) return;
    if (e.clientY < 80) showTopBarControls();
    if (e.clientY > window.innerHeight - 140) showTagPanelControls();
  }

  function showTopBarControls() {
    if (!slideshowMode) return;
    showTopBar = true;
    if (hideTimerTop) clearTimeout(hideTimerTop);
    hideTimerTop = setTimeout(() => { showTopBar = false; }, 3000);
  }

  function showTagPanelControls() {
    if (!slideshowMode) return;
    showTagPanel = true;
    if (hideTimerBottom) clearTimeout(hideTimerBottom);
    hideTimerBottom = setTimeout(() => { showTagPanel = false; }, 3000);
  }

  function showToast(message: string, type: string) {
    if (toastTimeout) clearTimeout(toastTimeout);
    toast = { message, type };
    toastTimeout = setTimeout(() => { toast = null; }, 3000);
  }

  async function handleSettingsChange() {
    showSettings = false;
    await loadFiles();
  }

  async function handleTagManagerChange() {
    showTagManager = false;
    await loadTags();
  }

  function handleFullscreenChange() {
    if (!document.fullscreenElement && slideshowMode) {
      exitSlideshow();
    }
  }

  function handleJumpKeydown(e: KeyboardEvent) {
    if (e.key === 'Enter') {
      e.preventDefault();
      const idx = parseInt(String(jumpIndex), 10) - 1;
      if (idx >= 0 && idx < $filteredFiles.length) {
        currentIndex.set(idx);
      } else {
        jumpIndex = $progress.current;
      }
      (e.target as HTMLElement).blur();
    } else if (e.key === 'Escape') {
      jumpIndex = $progress.current;
      (e.target as HTMLElement).blur();
    } else if (e.key.length === 1 && !/\d/.test(e.key)) {
      e.preventDefault();
    }
  }

  function handleJumpBlur() {
    jumpFocused = false;
    jumpIndex = $progress.current;
  }

  async function handleTrash() {
    if (!$currentFile) return;
    try {
      await api.trashFile($currentFile.path);
      const idx = $currentIndex;
      files.update((list) => { list.splice(idx, 1); return [...list]; });
      if ($currentIndex >= $filteredFiles.length) {
        currentIndex.set(Math.max(0, $filteredFiles.length - 1));
      }
      showToast('Moved to trash.', 'info');
    } catch (e: any) {
      showToast(e?.message || 'Failed to trash file.', 'error');
    }
  }

  async function handleFilter(e: CustomEvent) {
    toggleFilterTag(e.detail.tagId);
    await applyFilter();
    currentIndex.set(0);
  }

  function handleClearFilter() {
    clearFilter();
  }
</script>

<svelte:window on:keydown={handleKeydown} on:fullscreenchange={handleFullscreenChange} on:contextmenu|preventDefault={() => {}} />

<div class="slideshow" class:slideshow-mode={slideshowMode} on:mousemove={handleSlideshowMouseMove}>
  <!-- Thumbnail sidebar -->
  {#if $filteredFiles.length > 0}
    <ThumbnailSidebar files={$filteredFiles} currentIndex={$currentIndex} on:navigate={(e) => currentIndex.set(e.detail)} />
  {/if}

  <!-- Top bar -->
  <div class="top-bar" class:hidden={slideshowMode && !showTopBar}>
    <div class="top-left">
      {#if $currentFile}
        <span class="filename" title={$currentFile.path}>{$currentFile.name}</span>
        {#if $filterTagIds.size > 0}
          <span class="filter-badge">
            filter: {$tags.filter(t => $filterTagIds.has(t.id)).map(t => t.name).join(' + ')}
            <button class="filter-clear" on:click={handleClearFilter}>x</button>
          </span>
        {/if}
        <span class="progress">
          <input type="text" class="progress-input" inputmode="numeric" bind:value={jumpIndex} on:keydown={handleJumpKeydown} on:blur={handleJumpBlur} on:focus={() => jumpFocused = true} size={String(jumpIndex).length} />
          / {$filteredFiles.length}
        </span>
      {:else if $isDone}
        <span class="filename">Done</span>
      {:else}
        <span class="filename">GalleryTags</span>
      {/if}
    </div>
    <div class="top-right">
      <button class="icon-btn" class:shuffle-active={$shuffle !== 'off'} on:click={cycleShuffle} title={$shuffle === 'off' ? 'Shuffle: off' : $shuffle === 'all' ? 'Shuffle: all' : 'Shuffle: loaded'}>{$shuffle === 'off' ? 'rnd' : $shuffle === 'all' ? 'rnd*' : 'rnd↯'}</button>
      {#if slideshowMode}
        <button class="icon-btn" on:click={togglePause}>{slideshowPaused ? 'play' : 'pause'}</button>
        <button class="icon-btn" on:click={exitSlideshow}>exit</button>
      {:else}
        <button class="icon-btn" on:click={startSlideshow}>play</button>
      {/if}
      <button class="icon-btn" on:click={() => { showTagManager = !showTagManager; showSettings = false; }}>tags</button>
      <button class="icon-btn" on:click={() => { showSettings = !showSettings; showTagManager = false; }}>settings</button>
    </div>
  </div>

  <!-- Content -->
  <div class="content">
    {#if loading}
      <div class="center"><div class="spinner"></div><span>Loading...</span></div>
    {:else if needsSetup && $filteredFiles.length === 0 && $filterTagIds.size === 0}
      <div class="center">
        <h2>No inbox configured</h2>
        <p>Open settings to select a folder.</p>
        <button class="btn-primary" on:click={() => showSettings = true}>Open Settings</button>
      </div>
    {:else if $filteredFiles.length === 0 && $filterTagIds.size > 0}
      <div class="center">
        <h2>No files match filter</h2>
        <button class="btn-secondary" on:click={handleClearFilter}>Clear filter</button>
      </div>
    {:else if $filteredFiles.length === 0}
      <div class="center">
        <h2>No files found</h2>
        <button class="btn-secondary" on:click={loadFiles}>Rescan</button>
      </div>
    {:else if $isDone && !$loop}
      <div class="center">
        <h2>All done</h2>
        <p>{$processedCount} file{$processedCount !== 1 ? 's' : ''} processed</p>
        <button class="btn-secondary" on:click={loadFiles}>Rescan</button>
      </div>
    {:else if $currentFile}
      <MediaViewer file={$currentFile} on:videoEnded={handleVideoEnded} />
      {#if !slideshowMode}
        <div class="nav">
          <button class="nav-btn" on:click={goPrev} disabled={!$hasPrev}>← Prev</button>
          <button class="nav-btn" on:click={goNext}>Next →</button>
        </div>
      {/if}
    {/if}
  </div>

  <!-- Tag panel -->
  {#if ($currentFile && !$isDone) || $filterTagIds.size > 0}
    <div class="tag-panel-wrap" class:panel-hidden={slideshowMode && !showTagPanel}>
      <TagPanel tags={$tags} activeTagIds={fileTagIds} filterTagIds={$filterTagIds} disabled={tagPanelDisabled} on:save={handleSave} on:delete={handleTrash} on:filter={handleFilter} />
    </div>
  {/if}
</div>

<!-- Overlays outside .slideshow -->
<Overlay title="Settings" bind:open={showSettings} on:close={() => showSettings = false}>
  <Settings on:change={handleSettingsChange} />
</Overlay>

<Overlay title="Tag Manager" bind:open={showTagManager} on:close={() => showTagManager = false}>
  <TagManager on:change={handleTagManagerChange} />
</Overlay>

<!-- Toast -->
{#if toast}
  <div class="toast" class:toast-error={toast.type === 'error'}>
    {toast.message}
  </div>
{/if}

<style>
  .slideshow { display: flex; flex-direction: column; height: 100vh; background: var(--color-bg); overflow: hidden; }

  .top-bar { display: flex; align-items: center; justify-content: space-between; padding: 8px 16px; background: var(--color-surface); border-bottom: 1px solid #2a2a2e; flex-shrink: 0; min-height: 40px; transition: opacity 300ms, transform 300ms; }
  .top-bar.hidden { opacity: 0; transform: translateY(-100%); pointer-events: none; position: absolute; z-index: 10; width: 100%; }
  .top-left { display: flex; align-items: center; gap: 12px; min-width: 0; flex: 1; }
  .filename { font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; color: var(--color-text); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .progress { font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; color: var(--color-muted); flex-shrink: 0; display: flex; align-items: center; gap: 2px; }
  .progress-input { min-width: 24px; max-width: 56px; background: transparent; border: 1px solid transparent; border-radius: 4px; padding: 0 4px; color: var(--color-muted); font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; text-align: center; outline: none; transition: all 150ms; }
  .progress-input:hover { border-color: #2a2a2e; }
  .progress-input:focus { background: var(--color-bg); border-color: var(--color-accent); color: var(--color-text); }
  .top-right { display: flex; gap: 4px; flex-shrink: 0; }
  .icon-btn { background: transparent; border: 1px solid #2a2a2e; color: var(--color-muted); padding: 4px 10px; border-radius: 6px; font-size: 0.75rem; cursor: pointer; font-family: 'JetBrains Mono', monospace; transition: all 150ms; }
  .icon-btn:hover { background: #2a2a2e; color: var(--color-text); }
  .shuffle-active { background: #2a2a2e; color: var(--color-accent); border-color: var(--color-accent); }

  .filter-badge { display: inline-flex; align-items: center; gap: 6px; background: rgba(74, 158, 255, 0.15); border: 1px dashed var(--color-accent); border-radius: 9999px; padding: 2px 8px; font-size: 0.7rem; color: var(--color-accent); font-family: 'JetBrains Mono', monospace; }
  .filter-clear { background: none; border: none; color: var(--color-accent); cursor: pointer; padding: 0 2px; font-size: 0.7rem; }
  .filter-clear:hover { color: #e74c3c; }

  .content { flex: 1; display: flex; flex-direction: column; align-items: center; justify-content: center; overflow: hidden; min-height: 0; padding: 16px; }
  .slideshow-mode .content { padding: 0; }
  .center { text-align: center; color: var(--color-text); display: flex; flex-direction: column; align-items: center; gap: 8px; }
  .center h2 { margin: 0; font-size: 1.4rem; }
  .center p { color: var(--color-muted); margin: 0; }
  .spinner { width: 24px; height: 24px; border: 2px solid #2a2a2e; border-top-color: var(--color-accent); border-radius: 50%; animation: spin 0.8s linear infinite; }
  .center span { color: var(--color-muted); font-size: 0.9rem; }
  @keyframes spin { to { transform: rotate(360deg); } }

  .nav { display: flex; gap: 8px; padding: 8px 0; flex-shrink: 0; }
  .nav-btn { background: var(--color-surface); border: 1px solid #2a2a2e; color: var(--color-text); padding: 6px 20px; border-radius: 9999px; font-size: 0.85rem; cursor: pointer; transition: all 150ms; }
  .nav-btn:hover:not(:disabled) { background: #2a2a2e; }
  .nav-btn:disabled { opacity: 0.3; cursor: default; }

  .btn-primary { background: var(--color-accent); color: #fff; border: none; padding: 8px 20px; border-radius: 9999px; font-size: 0.85rem; cursor: pointer; transition: all 150ms; }
  .btn-primary:hover { background: #5daaff; }
  .btn-secondary { background: #2a2a2e; color: var(--color-muted); border: none; padding: 8px 20px; border-radius: 9999px; font-size: 0.85rem; cursor: pointer; transition: all 150ms; }
  .btn-secondary:hover { background: #3a3a3e; color: var(--color-text); }

  .tag-panel-wrap { transition: opacity 300ms, transform 300ms; }
  .panel-hidden { opacity: 0; transform: translateY(100%); pointer-events: none; position: absolute; bottom: 0; width: 100%; z-index: 10; }

  .toast { position: fixed; bottom: 120px; left: 50%; transform: translateX(-50%); background: var(--color-surface); color: var(--color-text); padding: 8px 20px; border-radius: 9999px; font-size: 0.85rem; border: 1px solid #2a2a2e; z-index: 200; animation: toastIn 150ms ease; white-space: nowrap; }
  .toast-error { border-color: #e74c3c; color: #e74c3c; }
  @keyframes toastIn { from { opacity: 0; transform: translateX(-50%) translateY(8px); } to { opacity: 1; transform: translateX(-50%) translateY(0); } }
</style>
