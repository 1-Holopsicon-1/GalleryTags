<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import * as api from '../api/index';
  import { loop, autoplay, scanSubdirs, slideshowInterval, sortOrder } from '../stores/files';

  const dispatch = createEventDispatcher<{ change: void }>();

  let inboxPath = '';
  let loading = true;

  onMount(async () => {
    try {
      inboxPath = await api.getInboxPath();
    } catch (e) {
      console.error('Failed to load inbox path:', e);
    } finally {
      loading = false;
    }
  });

  async function browse() {
    try {
      const selected = await api.openDirectoryDialog();
      if (selected) {
        await api.setInboxPath(selected);
        inboxPath = selected;
        dispatch('change');
      }
    } catch (e) {
      console.error('Failed to set inbox path:', e);
    }
  }
</script>

<div class="settings">
  <div class="settings-row">
    <span class="settings-label">Inbox folder</span>
    <div class="path-display" class:empty={!inboxPath}>
      {#if loading}
        <span class="muted">Loading...</span>
      {:else if inboxPath}
        {inboxPath}
      {:else}
        <span class="muted">No folder selected</span>
      {/if}
    </div>
    <button class="btn" on:click={browse}>Browse</button>
  </div>
  <div class="settings-row">
    <span class="settings-label">Loop</span>
    <label class="toggle">
      <input type="checkbox" bind:checked={$loop} />
      <span class="toggle-label">{#if $loop}On{:else}Off{/if}</span>
    </label>
  </div>
  <div class="settings-row">
    <span class="settings-label">Autoplay</span>
    <label class="toggle">
      <input type="checkbox" bind:checked={$autoplay} />
      <span class="toggle-label">{#if $autoplay}On{:else}Off{/if}</span>
    </label>
  </div>
  <div class="settings-row">
    <span class="settings-label">Subdirs</span>
    <label class="toggle">
      <input type="checkbox" bind:checked={$scanSubdirs} on:change={() => dispatch('change')} />
      <span class="toggle-label">{#if $scanSubdirs}On{:else}Off{/if}</span>
    </label>
  </div>
  <div class="settings-row">
    <span class="settings-label">Slide delay</span>
    <input type="number" class="num-input" bind:value={$slideshowInterval} min="1" max="60" />
    <span class="settings-hint">sec</span>
  </div>
  <div class="settings-row">
    <span class="settings-label">Sort by</span>
    <select class="select-input" bind:value={$sortOrder} on:change={() => dispatch('change')}>
      <option value="name">Name A-Z</option>
      <option value="name-desc">Name Z-A</option>
      <option value="mtime">Newest modified</option>
      <option value="mtime-desc">Oldest modified</option>
      <option value="btime">Newest created</option>
      <option value="btime-desc">Oldest created</option>
    </select>
  </div>
</div>

<style>
  .settings { background: var(--color-surface); border-radius: 8px; display: flex; flex-direction: column; gap: 10px; }
  .settings-row { display: flex; align-items: center; gap: 12px; }
  .settings-label { font-size: 0.85rem; color: var(--color-muted); white-space: nowrap; min-width: 90px; font-family: 'JetBrains Mono', monospace; }
  .path-display { flex: 1; font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; color: var(--color-text); background: var(--color-bg); padding: 6px 10px; border-radius: 6px; border: 1px solid #2a2a2e; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; }
  .path-display.empty { color: var(--color-muted); }
  .muted { color: var(--color-muted); }
  .btn { background: var(--color-accent); color: #fff; border: none; padding: 6px 16px; border-radius: 9999px; font-size: 0.85rem; cursor: pointer; white-space: nowrap; font-family: system-ui, sans-serif; transition: background 150ms; }
  .btn:hover { background: #5daaff; }
  .toggle { display: flex; align-items: center; gap: 8px; cursor: pointer; }
  .toggle input { accent-color: var(--color-accent); width: 16px; height: 16px; cursor: pointer; }
  .toggle-label { font-family: 'JetBrains Mono', monospace; font-size: 0.8rem; color: var(--color-muted); }
  .num-input { width: 60px; background: var(--color-bg); border: 1px solid #2a2a2e; border-radius: 6px; padding: 4px 8px; color: var(--color-text); font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; text-align: center; outline: none; }
  .num-input:focus { border-color: var(--color-accent); }
  .settings-hint { font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; color: var(--color-muted); }
  .select-input { flex: 1; background: var(--color-bg); border: 1px solid #2a2a2e; border-radius: 6px; padding: 4px 8px; color: var(--color-text); font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; outline: none; cursor: pointer; -webkit-appearance: none; -moz-appearance: none; appearance: none; background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' fill='%23888'%3E%3Cpath d='M2 4l4 4 4-4'/%3E%3C/svg%3E"); background-repeat: no-repeat; background-position: right 8px center; padding-right: 28px; }
  .select-input:focus { border-color: var(--color-accent); }
  .select-input option { background: #1a1a1d; color: #e8e8e8; }
</style>
