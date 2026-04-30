<script lang="ts">
  import { createEventDispatcher } from 'svelte';

  export let title = '';
  export let open = false;

  const dispatch = createEventDispatcher<{ close: void }>();

  function handleKeydown(e: KeyboardEvent) {
    if (!open) return;
    if (e.key === 'Escape') {
      open = false;
      dispatch('close');
    }
  }

  function handleBackdropClick() {
    open = false;
    dispatch('close');
  }
</script>

<svelte:window on:keydown={handleKeydown} />

{#if open}
  <!-- svelte-ignore a11y-no-static-element-interactions a11y-click-events-have-key-events -->
  <div class="overlay" role="dialog" on:click|self={handleBackdropClick}>
    <div class="overlay-content">
      <div class="overlay-header">
        <h3>{title}</h3>
        <button class="icon-btn" on:click={handleBackdropClick}>close</button>
      </div>
      <div class="overlay-body">
        <slot />
      </div>
    </div>
  </div>
{/if}

<style>
  .overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: rgba(0, 0, 0, 0.6);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    animation: fadeIn 150ms ease;
  }

  .overlay-content {
    background: var(--color-surface);
    border-radius: 8px;
    width: 90%;
    max-width: 560px;
    max-height: 80vh;
    overflow-y: auto;
    border: 1px solid #2a2a2e;
  }

  .overlay-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    border-bottom: 1px solid #2a2a2e;
    position: sticky;
    top: 0;
    background: var(--color-surface);
    z-index: 1;
  }

  .overlay-header h3 {
    margin: 0;
    font-size: 0.95rem;
    color: var(--color-text);
    font-family: system-ui, -apple-system, sans-serif;
  }

  .overlay-body {
    padding: 16px;
  }

  .icon-btn {
    background: transparent;
    border: none;
    color: var(--color-muted);
    padding: 4px 10px;
    border-radius: 6px;
    font-size: 0.75rem;
    cursor: pointer;
    font-family: 'JetBrains Mono', monospace;
    transition: all 150ms ease;
  }

  .icon-btn:hover {
    background: #2a2a2e;
    color: var(--color-text);
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
</style>
