<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import type { Tag } from '../types';
  import * as api from '../api/index';

  const dispatch = createEventDispatcher<{ change: void }>();

  const PALETTE = [
    '#4a9eff', '#f5a623', '#e74c3c', '#2ecc71',
    '#9b59b6', '#1abc9c', '#e67e22', '#95a5a6',
  ];

  let tags: Tag[] = [];
  let loading = true;
  let error: string | null = null;
  let showForm = false;
  let editingId: number | null = null;

  let formName = '';
  let formType = 'label';
  let formFolder = '';
  let formColor = PALETTE[0];
  let formHotkey = '';
  let hotkeyRecording = false;

  onMount(loadTags);

  async function loadTags() {
    loading = true;
    error = null;
    try {
      tags = await api.getTags();
    } catch (e: any) {
      error = e?.message || String(e);
    } finally {
      loading = false;
    }
  }

  function openNewForm() {
    editingId = null;
    formName = '';
    formType = 'label';
    formFolder = '';
    formColor = PALETTE[0];
    formHotkey = '';
    showForm = true;
  }

  function openEditForm(tag: Tag) {
    editingId = tag.id;
    formName = tag.name;
    formType = tag.type;
    formFolder = tag.folder;
    formColor = tag.color;
    formHotkey = tag.hotkey;
    showForm = true;
  }

  function cancelForm() {
    showForm = false;
    editingId = null;
  }

  async function browseFolder() {
    try {
      const selected = await api.openDirectoryDialog();
      if (selected) formFolder = selected;
    } catch (e) {
      console.error('Failed to open directory dialog:', e);
    }
  }

  async function saveTag() {
    const name = formName.trim();
    if (!name) return;
    if (formType === 'folder' && !formFolder.trim()) return;

    try {
      if (editingId !== null) {
        await api.updateTag(editingId, name, formType, formFolder, formColor, formHotkey);
      } else {
        await api.createTag(name, formType, formFolder, formColor, formHotkey);
      }
      await loadTags();
      dispatch('change');
      cancelForm();
    } catch (e) {
      console.error('Failed to save tag:', e);
    }
  }

  async function deleteTag(id: number) {
    try {
      await api.deleteTag(id);
      await loadTags();
      dispatch('change');
      if (editingId === id) cancelForm();
    } catch (e) {
      console.error('Failed to delete tag:', e);
    }
  }

  function formatHotkey(e: KeyboardEvent): string {
    if (e.key === 'Escape' || e.key === 'Backspace' || e.key === 'Tab') return '';
    const parts: string[] = [];
    if (e.ctrlKey) parts.push('ctrl');
    if (e.altKey) parts.push('alt');
    if (e.shiftKey) parts.push('shift');
    let key = e.key.toLowerCase();
    if (key === ' ') key = 'space';
    if (parts.length === 0 && key.length > 1) return ''; // ignore standalone special keys
    parts.push(key);
    return parts.join('+');
  }

  function handleHotkeyKeydown(e: KeyboardEvent) {
    e.preventDefault();
    e.stopPropagation();
    const result = formatHotkey(e);
    if (result !== '' || e.key === 'Backspace' || e.key === 'Escape') {
      formHotkey = result;
    }
    hotkeyRecording = false;
  }
</script>

<div class="tag-manager">
  <div class="header">
    <h3 class="title">Tags</h3>
    <button class="btn btn-primary" on:click={openNewForm}>+ New Tag</button>
  </div>

  {#if loading}
    <div class="status">Loading tags...</div>
  {:else if error}
    <div class="status error-msg">
      <p>Error: {error}</p>
      <button class="btn" on:click={loadTags}>Retry</button>
    </div>
  {:else if tags.length === 0 && !showForm}
    <div class="status">No tags yet. Create one to get started.</div>
  {:else}
    <div class="tag-list">
      {#each tags as tag (tag.id)}
        <div class="tag-item">
          <span class="color-dot" style="background-color: {tag.color}"></span>
          <span class="tag-name">{tag.name}</span>
          {#if tag.hotkey}
            <span class="hotkey-badge">{tag.hotkey}</span>
          {/if}
          <span class="badge" class:badge-folder={tag.type === 'folder'}>
            {tag.type === 'folder' ? 'Folder' : 'Label'}
          </span>
          {#if tag.type === 'folder' && tag.folder}
            <span class="tag-folder">{tag.folder}</span>
          {/if}
          <div class="tag-actions">
            <button class="btn-icon" on:click={() => openEditForm(tag)}>edit</button>
            <button class="btn-icon btn-icon-danger" on:click={() => deleteTag(tag.id)}>del</button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  {#if showForm}
    <div class="form">
      <h4 class="form-title">{editingId !== null ? 'Edit Tag' : 'New Tag'}</h4>

      <label class="field">
        <span class="field-label">Name</span>
        <input type="text" class="input" bind:value={formName} placeholder="Tag name" />
      </label>

      <label class="field">
        <span class="field-label">Type</span>
        <div class="type-selector">
          <button class="type-btn" class:active={formType === 'label'} on:click={() => formType = 'label'}>Label</button>
          <button class="type-btn" class:active={formType === 'folder'} on:click={() => formType = 'folder'}>Folder</button>
        </div>
      </label>

      {#if formType === 'folder'}
        <label class="field">
          <span class="field-label">Folder path</span>
          <div class="folder-row">
            <input type="text" class="input" bind:value={formFolder} placeholder="/path/to/folder" />
            <button class="btn btn-small" on:click={browseFolder}>Browse</button>
          </div>
        </label>
      {/if}

      <div class="field">
        <span class="field-label">Color</span>
        <div class="color-picker">
          {#each PALETTE as color}
            <button
              class="color-swatch"
              style="background-color: {color}"
              class:active={formColor === color}
              on:click={() => formColor = color}
            ></button>
          {/each}
        </div>
      </div>

      <label class="field">
        <span class="field-label">Hotkey</span>
        <div class="hotkey-row">
          <input
            type="text"
            class="input hotkey-input"
            readonly
            value={formHotkey || '—'}
            on:focus={() => hotkeyRecording = true}
            on:blur={() => hotkeyRecording = false}
            on:keydown={handleHotkeyKeydown}
          />
          {#if formHotkey}
            <button class="btn btn-small" on:click={() => formHotkey = ''}>Clear</button>
          {/if}
        </div>
        {#if hotkeyRecording}
          <span class="field-hint">Press a key...</span>
        {/if}
      </label>

      <div class="form-actions">
        <button class="btn btn-primary" on:click={saveTag}>Save</button>
        <button class="btn btn-secondary" on:click={cancelForm}>Cancel</button>
      </div>
    </div>
  {/if}
</div>

<style>
  .tag-manager { background: var(--color-surface); border-radius: 8px; }
  .header { display: flex; align-items: center; justify-content: space-between; margin-bottom: 12px; }
  .title { margin: 0; font-size: 0.95rem; color: var(--color-text); font-family: system-ui, sans-serif; }
  .status { color: var(--color-muted); font-size: 0.85rem; text-align: center; padding: 16px 0; }
  .error-msg { color: #e74c3c; }
  .error-msg .btn { margin-top: 8px; }

  .tag-list { display: flex; flex-direction: column; gap: 6px; }
  .tag-item { display: flex; align-items: center; gap: 8px; padding: 8px 10px; background: var(--color-bg); border-radius: 6px; border: 1px solid #2a2a2e; }
  .color-dot { width: 12px; height: 12px; border-radius: 50%; flex-shrink: 0; }
  .tag-name { font-family: 'JetBrains Mono', monospace; font-size: 0.85rem; color: var(--color-text); flex-shrink: 0; }
  .badge { font-size: 0.7rem; padding: 2px 8px; border-radius: 9999px; background: #2a3a50; color: var(--color-accent); flex-shrink: 0; }
  .badge-folder { background: #3a2a1a; color: #f5a623; }
  .hotkey-badge { font-family: 'JetBrains Mono', monospace; font-size: 0.7rem; padding: 1px 6px; border-radius: 4px; background: #2a2a2e; color: #888; border: 1px solid #3a3a3e; }
  .tag-folder { font-family: 'JetBrains Mono', monospace; font-size: 0.75rem; color: var(--color-muted); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; min-width: 0; flex: 1; }
  .tag-actions { display: flex; gap: 4px; flex-shrink: 0; }
  .btn-icon { background: transparent; border: 1px solid #2a2a2e; color: var(--color-muted); padding: 2px 8px; border-radius: 4px; font-size: 0.75rem; cursor: pointer; font-family: 'JetBrains Mono', monospace; transition: all 150ms; }
  .btn-icon:hover { background: #2a2a2e; color: var(--color-text); }
  .btn-icon-danger:hover { background: #3a1a1a; color: #e74c3c; border-color: #e74c3c; }

  .form { margin-top: 12px; padding-top: 12px; border-top: 1px solid #2a2a2e; display: flex; flex-direction: column; gap: 10px; }
  .form-title { margin: 0; font-size: 0.85rem; color: var(--color-text); }
  .field { display: flex; flex-direction: column; gap: 4px; }
  .field-label { font-size: 0.75rem; color: var(--color-muted); font-family: 'JetBrains Mono', monospace; }
  .input { background: var(--color-bg); border: 1px solid #2a2a2e; border-radius: 6px; padding: 6px 10px; color: var(--color-text); font-size: 0.85rem; font-family: 'JetBrains Mono', monospace; outline: none; transition: border-color 150ms; }
  .input:focus { border-color: var(--color-accent); }
  .type-selector { display: flex; gap: 4px; }
  .type-btn { flex: 1; padding: 5px 12px; background: var(--color-bg); border: 1px solid #2a2a2e; border-radius: 6px; color: var(--color-muted); font-size: 0.8rem; cursor: pointer; transition: all 150ms; }
  .type-btn.active { border-color: var(--color-accent); color: var(--color-accent); background: #1a2a3a; }
  .folder-row { display: flex; gap: 8px; }
  .folder-row .input { flex: 1; min-width: 0; }
  .color-picker { display: flex; gap: 6px; }
  .color-swatch { width: 24px; height: 24px; border-radius: 50%; border: 2px solid transparent; cursor: pointer; transition: all 150ms; }
  .color-swatch.active { border-color: #fff; transform: scale(1.15); }
  .color-swatch:hover { transform: scale(1.1); }
  .form-actions { display: flex; gap: 8px; margin-top: 4px; }
  .hotkey-row { display: flex; gap: 8px; }
  .hotkey-input { cursor: pointer; flex: 1; }
  .hotkey-input:focus { border-color: var(--color-accent); cursor: text; }
  .field-hint { font-size: 0.7rem; color: var(--color-muted); font-style: italic; }

  .btn { border: none; padding: 6px 16px; border-radius: 9999px; font-size: 0.85rem; cursor: pointer; font-family: system-ui, sans-serif; transition: background 150ms; background: var(--color-accent); color: #fff; }
  .btn:hover { background: #5daaff; }
  .btn-primary { background: var(--color-accent); color: #fff; }
  .btn-primary:hover { background: #5daaff; }
  .btn-secondary { background: #2a2a2e; color: var(--color-muted); }
  .btn-secondary:hover { background: #3a3a3e; color: var(--color-text); }
  .btn-small { padding: 6px 12px; font-size: 0.8rem; }
</style>
