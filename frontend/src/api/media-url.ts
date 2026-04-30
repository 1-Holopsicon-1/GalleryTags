import { GetMediaBaseURL as _GetMediaBaseURL } from '../../wailsjs/go/main/App';

let baseUrl = '';

export async function initMediaUrl(): Promise<void> {
  baseUrl = await _GetMediaBaseURL();
}

export function getMediaUrl(filePath: string): string {
  return baseUrl + encodeURIComponent(filePath);
}
