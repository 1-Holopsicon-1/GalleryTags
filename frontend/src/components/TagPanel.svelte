<script lang="ts">
  import { createEventDispatcher } from 'svelte';
  import type { Tag } from '../types';

  export let tags: Tag[] = [];
  export let disabled = false;
  export let activeTagIds: number[] = [];
  export let filterTagIds: Set<number> = new Set();

  const dispatch = createEventDispatcher<{ save: { tagIDs: number[] }; delete: null; filter: { tagId: number } }>();

  let selectedIDs = new Set<number>();
  let activeSet = new Set<number>();

  $: activeSet = new Set(activeTagIds);
  $: selectedIDs = new Set();

  $: folderTags = tags.filter((t) => t.type === 'folder');
  $: labelTags = tags.filter((t) => t.type === 'label');

  function toggleTag(tag: Tag) {
    if (tag.type === 'folder') {
      const otherFolderIDs = folderTags
        .filter((t) => t.id !== tag.id)
        .map((t) => t.id);

      if (selectedIDs.has(tag.id)) {
        selectedIDs.delete(tag.id);
      } else {
        for (const id of otherFolderIDs) {
          selectedIDs.delete(id);
        }
        selectedIDs.add(tag.id);
      }
    } else {
      if (selectedIDs.has(tag.id)) {
        selectedIDs.delete(tag.id);
      } else {
        selectedIDs.add(tag.id);
      }
    }
    selectedIDs = selectedIDs;
    dispatch('save', { tagIDs: [...selectedIDs] });
  }

  function handleContextMenu(e: MouseEvent, tag: Tag) {
    e.preventDefault();
    dispatch('filter', { tagId: tag.id });
  }
</script>

<div class="tag-panel">
  {#if labelTags.length > 0}
    <div class="label-tags">
      {#each labelTags as tag (tag.id)}
        <button
          class="tag-btn"
          class:active={selectedIDs.has(tag.id)}
          class:applied={activeSet.has(tag.id) && !selectedIDs.has(tag.id)}
          class:filter-active={filterTagIds.has(tag.id)}
          style="--tag-color: {tag.color}"
          on:click={() => toggleTag(tag)}
          on:contextmenu|preventDefault={(e) => handleContextMenu(e, tag)}
          disabled={disabled}
        >
          <span class="color-dot" style="background: {tag.color}"></span>
          {tag.name}
        </button>
      {/each}
    </div>
  {/if}

  {#if folderTags.length > 0}
    <div class="folder-tags">
      {#each folderTags as tag (tag.id)}
        <button
          class="tag-btn folder"
          class:active={selectedIDs.has(tag.id)}
          class:applied={activeSet.has(tag.id) && !selectedIDs.has(tag.id)}
          class:filter-active={filterTagIds.has(tag.id)}
          style="--tag-color: {tag.color}"
          on:click={() => toggleTag(tag)}
          on:contextmenu|preventDefault={(e) => handleContextMenu(e, tag)}
          disabled={disabled}
        >
          <span class="folder-icon">dir</span> {tag.name}
        </button>
      {/each}
    </div>
  {/if}
  <button class="delete-btn" on:click={() => dispatch('delete', null)} disabled={disabled}>del</button>
</div>

<style>
  .tag-panel {
    display: flex;
    align-items: stretch;
    gap: 0;
    background: var(--color-surface);
    border-top: 1px solid #2a2a2e;
    flex-shrink: 0;
  }

  .label-tags {
    flex: 1;
    display: flex;
    flex-wrap: wrap;
    align-content: center;
    gap: 6px;
    padding: 8px 12px;
    min-width: 0;
    overflow: hidden;
  }

  .folder-tags {
    display: flex;
    flex-direction: column;
    gap: 0;
    border-left: 1px solid #2a2a2e;
    flex-shrink: 0;
  }

  .tag-btn {
    display: flex;
    align-items: center;
    gap: 6px;
    padding: 4px 12px;
    border-radius: 9999px;
    border: 1px solid #2a2a2e;
    background: transparent;
    color: var(--color-text);
    font-size: 0.8rem;
    cursor: pointer;
    font-family: system-ui, -apple-system, sans-serif;
    transition: all 150ms ease;
    white-space: nowrap;
  }

  .tag-btn:hover:not(:disabled) {
    background: color-mix(in srgb, var(--tag-color, #4a9eff) 15%, transparent);
    border-color: var(--tag-color, #4a9eff);
  }

  .tag-btn.active {
    background: color-mix(in srgb, var(--tag-color, #4a9eff) 25%, transparent);
    border-color: var(--tag-color, #4a9eff);
    color: var(--tag-color, #4a9eff);
  }

  .tag-btn.applied {
    background: color-mix(in srgb, var(--tag-color, #4a9eff) 10%, transparent);
    border-color: color-mix(in srgb, var(--tag-color, #4a9eff) 40%, transparent);
    color: color-mix(in srgb, var(--tag-color, #4a9eff) 60%, var(--color-text));
  }

  .tag-btn.filter-active {
    background: color-mix(in srgb, var(--tag-color, #4a9eff) 35%, transparent);
    border-color: var(--tag-color, #4a9eff);
    color: var(--tag-color, #4a9eff);
    border-style: dashed;
  }

  .tag-btn:disabled {
    opacity: 0.4;
    cursor: default;
  }

  .folder .tag-btn,
  .folder-tags .tag-btn {
    border-radius: 0;
    border: none;
    border-bottom: 1px solid #2a2a2e;
    padding: 6px 12px;
  }

  .folder-tags .tag-btn:last-child {
    border-bottom: none;
  }

  .color-dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .folder-icon {
    font-family: 'JetBrains Mono', monospace;
    font-size: 0.7rem;
    color: #f5a623;
  }

  .delete-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 6px 12px;
    border: none;
    border-top: 1px solid #2a2a2e;
    background: transparent;
    color: #e74c3c;
    font-size: 0.75rem;
    font-family: 'JetBrains Mono', monospace;
    cursor: pointer;
    transition: background 150ms;
  }
  .delete-btn:hover:not(:disabled) { background: #2a1a1a; }
  .delete-btn:disabled { opacity: 0.3; cursor: default; }
</style>
