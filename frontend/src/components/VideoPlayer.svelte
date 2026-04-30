<script lang="ts">
    import { autoplay } from "../stores/files";

    export let src: string = "";
    let video: HTMLVideoElement;
    let errorMsg = "";

    $: if (video && src) {
        errorMsg = "";
        video.src = src;
        video.load();
    }

    function handleError(e: Event) {
        const v = e.target as HTMLVideoElement;
        const err = v.error;
        const codes: Record<number, string> = {
            1: "ABORTED",
            2: "NETWORK",
            3: "DECODE",
            4: "SRC_NOT_SUPPORTED",
        };
        errorMsg = `Video error: ${err?.code} (${codes[err?.code ?? 0] ?? "UNKNOWN"}): ${err?.message ?? "no message"} | src: ${src}`;
        console.error(errorMsg);
    }
</script>

{#if errorMsg}
    <div
        style="color: #e74c3c; font-size: 0.8rem; padding: 8px; word-break: break-all;"
    >
        {errorMsg}
    </div>
{/if}

<!-- svelte-ignore a11y-media-has-caption -->
<video
    bind:this={video}
    controls
    preload="metadata"
    autoplay={$autoplay}
    on:error={handleError}
    on:ended
>
    Your browser does not support the video element.
</video>

<style>
    video {
        max-width: 100%;
        max-height: 100%;
        object-fit: contain;
        border-radius: 4px;
    }
</style>
