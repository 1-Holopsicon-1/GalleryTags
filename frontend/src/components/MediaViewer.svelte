<script lang="ts">
  import type { FileInfo } from '../types';
  import { getMediaUrl } from '../api/media-url';
  import VideoPlayer from './VideoPlayer.svelte';

  export let file: FileInfo;

  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();
  function onVideoEnded() { dispatch('videoEnded'); }
</script>

<div class="media-container">
  {#if file.type === 'video'}
    <VideoPlayer src={getMediaUrl(file.path)} on:ended={onVideoEnded} />
  {:else}
    {#key file.path}
      <img src={getMediaUrl(file.path)} alt={file.name} draggable="false" />
    {/key}
  {/if}
</div>

<style>
  .media-container {
    flex: 1;
    min-height: 0;
    width: 100%;
    position: relative;
    overflow: hidden;
    opacity: 0;
    animation: fadeIn 150ms ease forwards;
  }

  .media-container img,
  .media-container :global(video) {
    position: absolute;
    inset: 0;
    margin: auto;
    max-width: 100%;
    max-height: 100%;
    object-fit: contain;
    border-radius: 4px;
    user-select: none;
  }

  @keyframes fadeIn {
    from { opacity: 0; }
    to { opacity: 1; }
  }
</style>
