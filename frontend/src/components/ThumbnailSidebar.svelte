<script lang="ts">
  import { createEventDispatcher, tick } from 'svelte';
  import type { FileInfo } from '../types';
  import { getMediaUrl } from '../api/media-url';
  import { preloadedPaths } from '../stores/preload';

  export let files: FileInfo[];
  export let currentIndex: number;

  const dispatch = createEventDispatcher<{ navigate: number }>();

  let sidebarVisible = false;
  let hideTimer: ReturnType<typeof setTimeout> | null = null;
  let scrollContainer: HTMLElement;

  $: if (currentIndex >= 0) {
    tick().then(() => {
      const el = scrollContainer?.querySelector('.thumb.current');
      el?.scrollIntoView({ block: 'nearest', behavior: 'smooth' });
    });
  }

  function handleHoverZoneEnter() {
    showSidebar();
  }

  function showSidebar() {
    sidebarVisible = true;
    if (hideTimer) clearTimeout(hideTimer);
  }

  function handleSidebarMouseLeave() {
    hideTimer = setTimeout(() => { sidebarVisible = false; }, 500);
  }

  function navigateTo(i: number) {
    dispatch('navigate', i);
  }
</script>

<div class="sidebar-hover-zone" on:mouseenter={handleHoverZoneEnter}>
  <!-- svelte-ignore a11y-no-static-element-interactions -->
  <div class="sidebar" class:visible={sidebarVisible}
    on:mouseleave={handleSidebarMouseLeave}
  >
    <div class="sidebar-scroll" bind:this={scrollContainer}>
      {#each files as file, i}
        <button
          class="thumb"
          class:current={i === currentIndex}
          on:click={() => navigateTo(i)}
          title={file.name}
        >
          {#if file.type === 'video'}
            <div class="thumb-video-badge">vid</div>
          {/if}
          <img src={getMediaUrl(file.path)} alt={file.name} loading="lazy" />
        </button>
      {/each}
    </div>
  </div>
</div>

<style>
  .sidebar-hover-zone {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: 40px;
    z-index: 50;
  }

  .sidebar {
    position: fixed;
    left: 0;
    top: 0;
    bottom: 0;
    width: 88px;
    background: var(--color-surface);
    border-right: 1px solid #2a2a2e;
    z-index: 50;
    transform: translateX(-100%);
    transition: transform 300ms ease, opacity 300ms ease;
    opacity: 0;
  }

  .sidebar.visible {
    transform: translateX(0);
    opacity: 1;
  }

  .sidebar-scroll {
    display: flex;
    flex-direction: column;
    gap: 2px;
    padding: 4px;
    height: 100%;
    overflow-y: auto;
    overflow-x: hidden;
  }

  .sidebar-scroll::-webkit-scrollbar { width: 4px; }
  .sidebar-scroll::-webkit-scrollbar-track { background: transparent; }
  .sidebar-scroll::-webkit-scrollbar-thumb { background: #2a2a2e; border-radius: 2px; }

  .thumb {
    width: 80px;
    height: 80px;
    border-radius: 4px;
    overflow: hidden;
    position: relative;
    border: 2px solid transparent;
    background: #111;
    cursor: pointer;
    padding: 0;
    flex-shrink: 0;
    transition: border-color 150ms ease;
  }

  .thumb:hover { border-color: #3a3a3e; }
  .thumb.current { border-color: var(--color-accent); }

  .thumb img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

  .thumb-video-badge {
    position: absolute;
    bottom: 2px;
    right: 2px;
    font-size: 0.55rem;
    background: rgba(0, 0, 0, 0.7);
    color: var(--color-text);
    padding: 1px 4px;
    border-radius: 2px;
    font-family: 'JetBrains Mono', monospace;
    z-index: 1;
  }
</style>